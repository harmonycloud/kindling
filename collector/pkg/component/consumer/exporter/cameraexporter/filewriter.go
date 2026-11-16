// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cameraexporter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/filepathhelper"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

type fileWriter struct {
	config *fileConfig
	logger *component.TelemetryLogger
}

func newFileWriter(cfg *fileConfig, logger *component.TelemetryLogger) (*fileWriter, error) {
	// StoragePath must be an absolute path
	if !path.IsAbs(cfg.StoragePath) {
		return nil, fmt.Errorf("storage_path must be an absolute path")
	}
	// Check if there is the directory.
	// If the directory is not found, we create one
	if err := os.MkdirAll(cfg.StoragePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create fileWriter: %w", err)
	}
	return &fileWriter{
		config: cfg,
		logger: logger,
	}, nil
}

func (fw *fileWriter) write(group *model.DataGroup) {
	groupName := group.Name
	switch groupName {
	case constnames.SingleNetRequestMetricGroup:
		fw.writeTrace(group)
	case constnames.SpanEvent:
		fw.writeTrace(group)
	case constnames.CameraEventGroupName:
		fw.writeCpuEvents(group)
	}
}

func (fw *fileWriter) pidFilePath(workloadName string, podName string, containerName string, pid int64) string {
	dirName := workloadName + "_" + podName + "_" + containerName + "_" + strconv.FormatInt(pid, 10)
	return path.Join(fw.config.StoragePath, dirName)
}

func getFileName(protocol string, contentKey string, timestamp uint64, isServer bool) string {
	var isServerString string
	if isServer {
		isServerString = "true"
	} else {
		isServerString = "false"
	}
	encodedContent := base64.URLEncoding.EncodeToString([]byte(contentKey))
	return getDateString(int64(timestamp)) + "_" + protocol + "_" + encodedContent + "_" + isServerString
}

func (fw *fileWriter) writeTrace(group *model.DataGroup) {
	pathElements := filepathhelper.GetFilePathElements(group, group.Timestamp)
	// Create the directory first for saving its profile files.
	// If there has been such a directory, it will do nothing and return nil.
	baseDir := fw.pidFilePath(pathElements.WorkloadName, pathElements.PodName, pathElements.ContainerName, pathElements.Pid)
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		fw.logger.Errorf("Failed to create pid directory: %v", err)
		return
	}
	// /$path/podName_containerName_pid/protocol_contentKey_timestamp_isServer
	fileName := getFileName(pathElements.Protocol, pathElements.ContentKey, pathElements.Timestamp, pathElements.IsServer) // Check whether we need to roll over the files
	err := fw.writeFile(baseDir, fileName, group)
	if err != nil {
		fw.logger.Errorf("Failed to write trace to file: %v", err)
		return
	}
}

func (fw *fileWriter) writeFile(baseDir string, fileName string, group *model.DataGroup) error {
	// Check whether the count of files is greater than MaxCount
	err := fw.rotateFiles(baseDir)
	if err != nil {
		fw.logger.Warnf("can't rotate files in %s: %v", baseDir, err)
	}
	filePath := filepath.Join(baseDir, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("can't create new file: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fw.logger.Warnf("Failed to close the file %s", filePath)
		}
	}(f)
	fw.logger.Debugf("Create a trace file at [%s]", filePath)
	bytes, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("can't marshal DataGroup: %w", err)
	}
	_, err = f.Write(bytes)
	return err
}

func (fw *fileWriter) rotateFiles(baseDir string) error {
	// No constrains set
	if fw.config.MaxFileCountEachProcess <= 0 {
		return nil
	}
	// Get all files path
	toBeRotated, err := getDirEntryInTimeOrder(baseDir)
	if err != nil {
		return fmt.Errorf("can't get files list: %w", err)
	}
	// No need to rotate
	if len(toBeRotated) < fw.config.MaxFileCountEachProcess {
		return nil
	}
	// Remove the older files and remove half of them one time to decrease the frequency
	// of deleting files. Note this is different from rotating log files. We could delete
	// one file at a time for log files because the action "rotate" is in a low frequency
	// in that case.
	toBeRotated = toBeRotated[:len(toBeRotated)-fw.config.MaxFileCountEachProcess/2+1]
	// Remove the stale files asynchronously
	go func() {
		for _, dirEntry := range toBeRotated {
			_ = os.Remove(filepath.Join(baseDir, dirEntry.Name()))
			fw.logger.Infof("Rotate trace files [%s]", dirEntry.Name())
		}
	}()
	return nil
}

// getDirEntryInTimeOrder returns the directory entries slice in chronological order.
// The result files are sorted based on their modification time.
func getDirEntryInTimeOrder(path string) ([]os.DirEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)
	dirs, err := f.ReadDir(-1)
	// Sort the files based on their modification time. We don't sort them based on the
	// timestamp in the file name because they are similar but the latter one costs more CPU
	// considering that we have to split the file name first.
	sort.Slice(dirs, func(i, j int) bool {
		fileInfoA, err := dirs[i].Info()
		if err != nil {
			return false
		}
		fileInfoB, err := dirs[j].Info()
		if err != nil {
			return false
		}
		return fileInfoA.ModTime().Before(fileInfoB.ModTime())
	})
	return dirs, err
}

const dividingLine = "\n------\n"

func (fw *fileWriter) writeCpuEvents(group *model.DataGroup) {
	traceTimestamp := group.Labels.GetIntValue(constlabels.Timestamp)
	pathElements := filepathhelper.GetFilePathElements(group, uint64(traceTimestamp))
	baseDir := fw.pidFilePath(pathElements.WorkloadName, pathElements.PodName, pathElements.ContainerName, pathElements.Pid)
	fileName := getFileName(pathElements.Protocol, pathElements.ContentKey, pathElements.Timestamp, pathElements.IsServer)
	filePath := filepath.Join(baseDir, fileName)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		fw.logger.Infof("Couldn't open the trace file %s when append CpuEvents: %v. "+
			"Maybe the file has been rotated.", filePath, err)
		return
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fw.logger.Warnf("Failed to close the file %s, %v", filePath, err)
		}
	}(f)
	_, err = f.Write([]byte(dividingLine))
	if err != nil {
		fw.logger.Errorf("Failed to append CpuEvents to the file %s: %v", filePath, err)
		return
	}
	eventsBytes, _ := json.Marshal(group)
	_, err = f.Write(eventsBytes)
	if err != nil {
		fw.logger.Errorf("Failed to append CpuEvents to the file %s: %v", filePath, err)
		return
	}
	fw.logger.Debugf("Write CpuEvents to trace files [%s]", filePath)
}

func (fw *fileWriter) name() string {
	return storageFile
}
