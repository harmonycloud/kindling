// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package aggregateprocessor

import (
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator/defaultaggregator"
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/processor"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

const Type = "aggregateprocessor"

var exponentialInt64Boundaries = []int64{10e6, 20e6, 50e6, 80e6, 130e6, 200e6, 300e6,
	400e6, 500e6, 700e6, 1e9, 2e9, 5e9, 30e9}

var tcpConnectLabelSelectors = newTcpConnectLabelSelectors()

type AggregateProcessor struct {
	cfg          *Config
	telemetry    *component.TelemetryTools
	nextConsumer consumer.Consumer

	aggregator               aggregator.Aggregator
	netRequestLabelSelectors *aggregator.LabelSelectors
	tcpLabelSelectors        *aggregator.LabelSelectors
	stopCh                   chan struct{}
	ticker                   *time.Ticker
}

func New(config interface{}, telemetry *component.TelemetryTools, nextConsumer consumer.Consumer) processor.Processor {
	cfg := config.(*Config)
	p := &AggregateProcessor{
		cfg:          cfg,
		telemetry:    telemetry,
		nextConsumer: nextConsumer,

		aggregator:               defaultaggregator.NewDefaultAggregator(toAggregatedConfig(cfg.AggregateKindMap)),
		netRequestLabelSelectors: newNetRequestLabelSelectors(),
		tcpLabelSelectors:        newTcpLabelSelectors(),
		stopCh:                   make(chan struct{}),
		ticker:                   time.NewTicker(time.Duration(cfg.TickerInterval) * time.Second),
	}
	go p.runTicker()
	return p
}

func toAggregatedConfig(m map[string][]AggregatedKindConfig) *defaultaggregator.AggregatedConfig {
	ret := &defaultaggregator.AggregatedConfig{KindMap: make(map[string][]defaultaggregator.KindConfig)}
	for k, v := range m {
		kindConfig := make([]defaultaggregator.KindConfig, len(v))
		for i, kind := range v {
			if kind.OutputName == "" {
				kind.OutputName = k
			}
			kindConfig[i] = newKindConfig(&kind)
		}
		ret.KindMap[k] = kindConfig
	}
	return ret
}

func newKindConfig(rawConfig *AggregatedKindConfig) (kindConfig defaultaggregator.KindConfig) {
	kind := defaultaggregator.GetAggregatorKind(rawConfig.Kind)
	switch kind {
	case defaultaggregator.HistogramKind:
		var boundaries []int64
		if rawConfig.ExplicitBoundaries != nil {
			boundaries = rawConfig.ExplicitBoundaries
		} else {
			boundaries = exponentialInt64Boundaries
		}
		return defaultaggregator.KindConfig{
			OutputName:         rawConfig.OutputName,
			Kind:               kind,
			ExplicitBoundaries: boundaries,
		}
	default:
		return defaultaggregator.KindConfig{
			OutputName: rawConfig.OutputName,
			Kind:       kind,
		}
	}
}

// TODO: Graceful shutdown
func (p *AggregateProcessor) runTicker() {
	for {
		select {
		case <-p.stopCh:
			return
		case <-p.ticker.C:
			aggResults := p.aggregator.Dump()
			for _, agg := range aggResults {
				err := p.nextConsumer.Consume(agg)
				if err != nil {
					p.telemetry.Logger.Warn("Error happened when consuming aggregated recordersMap",
						zap.Error(err))
				}
			}
		}
	}
}

func (p *AggregateProcessor) Consume(dataGroup *model.DataGroup) error {
	switch dataGroup.Name {
	case constnames.NetRequestMetricGroupName:
		var abnormalDataErr error
		// The abnormal recordersMap will be treated as trace in later processing.
		// Must trace be merged into metrics in this place? Yes, because we have to generate histogram metrics,
		// trace recordersMap should not be recorded again, otherwise the percentiles will be much higher.
		if p.isSampled(dataGroup) {
			dataGroup.Name = constnames.SingleNetRequestMetricGroup
			abnormalDataErr = p.nextConsumer.Consume(dataGroup)
		}
		dataGroup.Name = constnames.AggregatedNetRequestMetricGroup
		p.aggregator.Aggregate(dataGroup, p.netRequestLabelSelectors)
		return abnormalDataErr
	case constnames.TcpRttMetricGroupName:
		fallthrough
	case constnames.TcpRetransmitMetricGroupName:
		fallthrough
	case constnames.TcpDropMetricGroupName:
		p.aggregator.Aggregate(dataGroup, p.tcpLabelSelectors)
		return nil
	case constnames.TcpConnectMetricGroupName:
		p.aggregator.Aggregate(dataGroup, tcpConnectLabelSelectors)
		return nil
	default:
		p.aggregator.Aggregate(dataGroup, p.netRequestLabelSelectors)
		return nil
	}
}

// TODO: make it configurable instead of hard-coded
func newNetRequestLabelSelectors() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.Pid, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.Comm, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.Protocol, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.IsServer, VType: aggregator.BooleanType},
		aggregator.LabelSelector{Name: constlabels.ContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DnatIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DnatPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DstContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstContainer, VType: aggregator.StringType},

		aggregator.LabelSelector{Name: constlabels.IsError, VType: aggregator.BooleanType},
		aggregator.LabelSelector{Name: constlabels.IsSlow, VType: aggregator.BooleanType},
		aggregator.LabelSelector{Name: constlabels.HttpStatusCode, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DnsRcode, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.SqlErrCode, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.ContentKey, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DnsDomain, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.KafkaTopic, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.RocketMQErrCode, VType: aggregator.IntType},
	)
}

func newTcpLabelSelectors() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.SrcNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.SrcContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DnatIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DnatPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DstContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstContainer, VType: aggregator.StringType},
	)
}

func newTcpConnectLabelSelectors() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.Pid, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.Comm, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DnatIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DnatPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DstContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.Errno, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.Success, VType: aggregator.BooleanType},
	)
}

func (p *AggregateProcessor) isSampled(dataGroup *model.DataGroup) bool {
	randSeed := rand.Intn(100)
	if isAbnormal(dataGroup) {
		if (randSeed < p.cfg.SamplingRate.SlowData) && dataGroup.Labels.GetBoolValue(constlabels.IsSlow) {
			return true
		}
		if (randSeed < p.cfg.SamplingRate.ErrorData) && dataGroup.Labels.GetBoolValue(constlabels.IsError) {
			return true
		}
	} else {
		if randSeed < p.cfg.SamplingRate.NormalData {
			return true
		}
	}
	return false
}

// shouldAggregate returns true if the dataGroup is slow or has errors.
func isAbnormal(g *model.DataGroup) bool {
	return g.Labels.GetBoolValue(constlabels.IsSlow) || g.Labels.GetBoolValue(constlabels.IsError) ||
		g.Labels.GetIntValue(constlabels.ErrorType) > constlabels.NoError
}
