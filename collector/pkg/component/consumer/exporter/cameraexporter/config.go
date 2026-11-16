// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cameraexporter

const (
	storageFile          = "file"
	storageElasticsearch = "elasticsearch"
)

type Config struct {
	Storage    string      `mapstructure:"storage"`
	EsConfig   *esConfig   `mapstructure:"es_config"`
	FileConfig *fileConfig `mapstructure:"file_config"`
}

type esConfig struct {
	EsHost      string `mapstructure:"es_host"`
	IndexSuffix string `mapstructure:"index_suffix"`
}

type fileConfig struct {
	// StoragePath is the ABSOLUTE path of the directory where the profile file should be saved
	StoragePath string `mapstructure:"storage_path"`
	// Storage constrains for each process
	MaxFileCountEachProcess int `mapstructure:"max_file_count_each_process"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Storage: storageFile,
		FileConfig: &fileConfig{
			StoragePath:             "/tmp/kindling/",
			MaxFileCountEachProcess: 50,
		},
	}
}
