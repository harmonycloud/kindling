// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package filepathhelper

import (
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

type FilePathElements struct {
	WorkloadName  string
	PodName       string
	ContainerName string
	Pid           int64
	IsServer      bool
	Protocol      string
	Timestamp     uint64
	ContentKey    string
}

func (f *FilePathElements) ToAttributes() *model.AttributeMap {
	attributes := model.NewAttributeMap()
	attributes.AddBoolValue(constlabels.IsServer, f.IsServer)
	if f.IsServer {
		attributes.AddStringValue(constlabels.DstWorkloadName, f.WorkloadName)
		attributes.AddStringValue(constlabels.DstPod, f.PodName)
		attributes.AddStringValue(constlabels.DstContainer, f.ContainerName)
	} else {
		attributes.AddStringValue(constlabels.SrcWorkloadName, f.WorkloadName)
		attributes.AddStringValue(constlabels.SrcPod, f.PodName)
		attributes.AddStringValue(constlabels.SrcContainer, f.ContainerName)
	}
	attributes.AddStringValue(constlabels.Protocol, f.Protocol)
	attributes.AddStringValue(constlabels.ContentKey, f.ContentKey)
	attributes.AddIntValue(constlabels.Timestamp, int64(f.Timestamp))
	return attributes
}

func GetFilePathElements(group *model.DataGroup, timestamp uint64) FilePathElements {
	pid := group.Labels.GetIntValue(constlabels.Pid)
	isServer := group.Labels.GetBoolValue(constlabels.IsServer)
	var workloadName string
	var podName string
	var containerName string
	if isServer {
		workloadName = group.Labels.GetStringValue(constlabels.DstWorkloadName)
		containerName = group.Labels.GetStringValue(constlabels.DstContainer)
		podName = group.Labels.GetStringValue(constlabels.DstPod)
	} else {
		workloadName = group.Labels.GetStringValue(constlabels.SrcWorkloadName)
		containerName = group.Labels.GetStringValue(constlabels.SrcContainer)
		podName = group.Labels.GetStringValue(constlabels.SrcPod)
	}
	if len(workloadName) == 0 {
		workloadName = "NoWorkloadName"
	}
	if len(containerName) == 0 {
		containerName = "NoContainerName"
	}
	if len(podName) == 0 {
		podName = "NoPodName"
	}
	protocol := group.Labels.GetStringValue(constlabels.Protocol)
	contentKey := group.Labels.GetStringValue(constlabels.ContentKey)
	return FilePathElements{
		WorkloadName:  workloadName,
		PodName:       podName,
		ContainerName: containerName,
		Pid:           pid,
		IsServer:      isServer,
		Protocol:      protocol,
		Timestamp:     timestamp,
		ContentKey:    contentKey,
	}
}
