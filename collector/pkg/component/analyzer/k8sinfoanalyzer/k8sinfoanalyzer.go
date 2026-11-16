// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package k8sinfoanalyzer

import (
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"go.uber.org/zap/zapcore"
)

const Type analyzer.Type = "k8sinfoanalyzer"

type K8sInfoAnalyzer struct {
	cfg             *Config
	nextConsumers   []consumer.Consumer
	telemetry       *component.TelemetryTools
	stopProfileChan chan struct{}
}

func New(cfg interface{}, telemetry *component.TelemetryTools, consumer []consumer.Consumer) analyzer.Analyzer {
	config, ok := cfg.(*Config)
	if !ok {
		telemetry.Logger.Panic("Cannot convert k8sinfoanalyzer config")
	}
	return &K8sInfoAnalyzer{
		cfg:           config,
		nextConsumers: consumer,
		telemetry:     telemetry,
	}
}

func (a *K8sInfoAnalyzer) sendToNextConsumer() {
	timer := time.NewTicker(time.Duration(a.cfg.SendDataGroupInterval) * time.Second)
	for {
		select {
		case <-a.stopProfileChan:
			return
		case <-timer.C:
			func() {
				dataGroups := kubernetes.GetWorkloadDataGroup()
				for _, nextConsumer := range a.nextConsumers {
					for _, dataGroup := range dataGroups {
						nextConsumer.Consume(dataGroup)
						if ce := a.telemetry.Logger.Check(zapcore.DebugLevel, ""); ce != nil {
							a.telemetry.Logger.Debug("K8sInfoAnalyzer send to consumer workload name=:\n" +
								dataGroup.Labels.GetStringValue(constlabels.WorkloadName))
						}
					}
				}
			}()
		}
	}
}

func (a *K8sInfoAnalyzer) Start() error {
	a.stopProfileChan = make(chan struct{})
	go a.sendToNextConsumer()
	return nil
}

func (a *K8sInfoAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	return nil
}

func (a *K8sInfoAnalyzer) Shutdown() error {
	close(a.stopProfileChan)
	return nil
}

func (a *K8sInfoAnalyzer) Type() analyzer.Type {
	return Type
}

func (a *K8sInfoAnalyzer) ConsumableEvents() []string {
	return nil
}
