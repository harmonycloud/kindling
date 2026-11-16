// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/processor"
	"github.com/Kindling-project/kindling/collector/pkg/component/receiver"
)

const (
	ReceiversKey  = "receivers"
	AnalyzersKey  = "analyzers"
	ProcessorsKey = "processors"
	ExportersKey  = "exporters"
)

var ComponentsKeyMap = []string{ReceiversKey, AnalyzersKey, ProcessorsKey, ExportersKey}

type ComponentsFactory struct {
	Receivers  map[string]ReceiverFactory
	Analyzers  map[string]AnalyzerFactory
	Processors map[string]ProcessorFactory
	Exporters  map[string]ExporterFactory
}

type NewReceiverFunc func(cfg interface{}, telemetry *component.TelemetryTools, analyzerManager *analyzer.Manager) receiver.Receiver
type NewAnalyzerFunc func(cfg interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer
type NewProcessorFunc func(cfg interface{}, telemetry *component.TelemetryTools, consumer consumer.Consumer) processor.Processor
type NewExporterFunc func(cfg interface{}, telemetry *component.TelemetryTools) exporter.Exporter

type ReceiverFactory struct {
	NewFunc NewReceiverFunc
	Config  interface{}
}

type AnalyzerFactory struct {
	NewFunc NewAnalyzerFunc
	Config  interface{}
}

type ProcessorFactory struct {
	NewFunc NewProcessorFunc
	Config  interface{}
}

type ExporterFactory struct {
	NewFunc NewExporterFunc
	Config  interface{}
}

func NewComponentsFactory() *ComponentsFactory {
	return &ComponentsFactory{
		Receivers:  make(map[string]ReceiverFactory),
		Analyzers:  make(map[string]AnalyzerFactory),
		Processors: make(map[string]ProcessorFactory),
		Exporters:  make(map[string]ExporterFactory),
	}
}
func (c *ComponentsFactory) RegisterReceiver(
	name string,
	f NewReceiverFunc,
	config interface{},
) {
	c.Receivers[name] = ReceiverFactory{
		NewFunc: f,
		Config:  config,
	}
}

func (c *ComponentsFactory) RegisterAnalyzer(
	name string,
	f NewAnalyzerFunc,
	config interface{},
) {
	c.Analyzers[name] = AnalyzerFactory{
		NewFunc: f,
		Config:  config,
	}
}

func (c *ComponentsFactory) RegisterProcessor(
	name string,
	f NewProcessorFunc,
	config interface{},
) {
	c.Processors[name] = ProcessorFactory{
		NewFunc: f,
		Config:  config,
	}
}

func (c *ComponentsFactory) RegisterExporter(
	name string,
	f NewExporterFunc,
	config interface{},
) {
	c.Exporters[name] = ExporterFactory{
		NewFunc: f,
		Config:  config,
	}
}

// mapStructureDecoderConfigFunc is a function that is used to configure the mapstructure decoder.
// ZeroFields option is set to true to allow the map/slice in the configuration file to override
// the default values in the config struct.
var mapStructureDecoderConfigFunc = func(dc *mapstructure.DecoderConfig) {
	dc.ZeroFields = true
}

func (c *ComponentsFactory) ConstructConfig(viper *viper.Viper) error {
	for _, componentKind := range ComponentsKeyMap {
		switch componentKind {
		case ReceiversKey:
			for k, factory := range c.Receivers {
				key := ReceiversKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config, mapStructureDecoderConfigFunc)
				if err != nil {
					return err
				}
			}
		case AnalyzersKey:
			for k, factory := range c.Analyzers {
				key := AnalyzersKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config, mapStructureDecoderConfigFunc)
				if err != nil {
					return err
				}
			}
		case ProcessorsKey:
			for k, factory := range c.Processors {
				key := ProcessorsKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config, mapStructureDecoderConfigFunc)
				if err != nil {
					return err
				}
			}
		case ExportersKey:
			for k, factory := range c.Exporters {
				key := ExportersKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config, mapStructureDecoderConfigFunc)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
