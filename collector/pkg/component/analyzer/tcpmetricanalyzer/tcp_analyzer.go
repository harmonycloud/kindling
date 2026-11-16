// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tcpmetricanalyzer

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	conntrackerpackge "github.com/Kindling-project/kindling/collector/pkg/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

const (
	TcpMetric analyzer.Type = "tcpmetricanalyzer"
)

type TcpMetricAnalyzer struct {
	consumers   []consumer.Consumer
	conntracker conntrackerpackge.Conntracker
	telemetry   *component.TelemetryTools
}

func NewTcpMetricAnalyzer(_ interface{}, telemetry *component.TelemetryTools, nextConsumers []consumer.Consumer) analyzer.Analyzer {
	retAnalyzer := &TcpMetricAnalyzer{
		consumers: nextConsumers,
		telemetry: telemetry,
	}
	conntracker, err := conntrackerpackge.NewConntracker(nil)
	if err != nil {
		telemetry.Logger.Warn("Conntracker cannot work as expected:", zap.Error(err))
	}
	retAnalyzer.conntracker = conntracker
	return retAnalyzer
}

func (a *TcpMetricAnalyzer) Start() error {
	return nil
}

func (a *TcpMetricAnalyzer) ConsumableEvents() []string {
	return []string{
		constnames.TcpCloseEvent,
		constnames.TcpRcvEstablishedEvent,
		constnames.TcpDropEvent,
		constnames.TcpRetransmitSkbEvent,
	}
}

// ConsumeEvent gets the event from the previous component
func (a *TcpMetricAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	var dataGroup *model.DataGroup
	var err error
	switch event.Name {
	case constnames.TcpCloseEvent:
		fallthrough
	case constnames.TcpRcvEstablishedEvent:
		dataGroup, err = a.generateRtt(event)
	case constnames.TcpDropEvent:
		dataGroup, err = a.generateDrop(event)
	case constnames.TcpRetransmitSkbEvent:
		dataGroup, err = a.generateRetransmit(event)
	default:
		return nil
	}
	if err != nil {
		if ce := a.telemetry.Logger.Check(zapcore.DebugLevel, "Event Skip, "); ce != nil {
			ce.Write(
				zap.Error(err),
			)
		}
		return nil
	}
	if dataGroup == nil {
		return nil
	}
	if ce := a.telemetry.Logger.Check(zapcore.DebugLevel, ""); ce != nil {
		a.telemetry.Logger.Debug("TcpMetricAnalyzer To NextProcess:\n" + dataGroup.String())
	}
	var retError error
	for _, nextConsumer := range a.consumers {
		err := nextConsumer.Consume(dataGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}
	return retError
}

func (a *TcpMetricAnalyzer) generateRtt(event *model.KindlingEvent) (*model.DataGroup, error) {
	// Only client-side has rtt metric
	labels, err := a.getTupleLabels(event)
	if err != nil {
		return nil, err
	}
	// Unit is microsecond
	rtt := event.GetUintUserAttribute("rtt")
	// rtt is zero when the kprobe is invoked in the first time, which should be filtered
	if rtt == 0 {
		return nil, nil
	}
	metric := model.NewIntMetric(constnames.TcpRttMetricName, int64(rtt))
	return model.NewDataGroup(constnames.TcpRttMetricGroupName, labels, event.Timestamp, metric), nil
}

func (a *TcpMetricAnalyzer) generateRetransmit(event *model.KindlingEvent) (*model.DataGroup, error) {
	labels, err := a.getTupleLabels(event)
	if err != nil {
		return nil, err
	}

	var segs = int64(1)
	p_segs := event.GetUserAttribute("segs")
	if p_segs != nil && p_segs.GetValueType() == model.ValueType_INT32 {
		segs = p_segs.GetIntValue()
	}

	metric := model.NewIntMetric(constnames.TcpRetransmitMetricName, segs)
	return model.NewDataGroup(constnames.TcpRetransmitMetricGroupName, labels, event.Timestamp, metric), nil
}

func (a *TcpMetricAnalyzer) generateDrop(event *model.KindlingEvent) (*model.DataGroup, error) {
	labels, err := a.getTupleLabels(event)
	if err != nil {
		return nil, err
	}
	metric := model.NewIntMetric(constnames.TcpDropMetricName, 1)
	return model.NewDataGroup(constnames.TcpDropMetricGroupName, labels, event.Timestamp, metric), nil
}

func (a *TcpMetricAnalyzer) getTupleLabels(event *model.KindlingEvent) (*model.AttributeMap, error) {
	// Note: Here sIp/dIp doesn't mean IP from client/server side for sure.
	// sIp stands for the IP which sends tcp flow.
	sIp := event.GetUserAttribute("sip")
	sPort := event.GetUserAttribute("sport")
	dIp := event.GetUserAttribute("dip")
	dPort := event.GetUserAttribute("dport")

	if sIp == nil || sPort == nil || dIp == nil || dPort == nil {
		return nil, fmt.Errorf("one of sip or dip or dport is nil for event %s", event.Name)
	}
	sIpString := model.IPLong2String(uint32(sIp.GetUintValue()))
	sPortUint := sPort.GetUintValue()
	dIpString := model.IPLong2String(uint32(dIp.GetUintValue()))
	dPortUint := dPort.GetUintValue()

	labels := model.NewAttributeMap()
	labels.AddStringValue(constlabels.SrcIp, sIpString)
	labels.AddIntValue(constlabels.SrcPort, int64(sPortUint))
	labels.AddStringValue(constlabels.DstIp, dIpString)
	labels.AddIntValue(constlabels.DstPort, int64(dPortUint))

	dNatIp, dNatPort := a.findDNatTuple(sIpString, sPortUint, dIpString, dPortUint)
	labels.AddStringValue(constlabels.DnatIp, dNatIp)
	labels.AddIntValue(constlabels.DnatPort, dNatPort)
	return labels, nil
}

func (a *TcpMetricAnalyzer) findDNatTuple(sIp string, sPort uint64, dIp string, dPort uint64) (string, int64) {
	dNat := a.conntracker.GetDNATTupleWithString(sIp, dIp, uint16(sPort), uint16(dPort), 0)
	if dNat == nil {
		return "", -1
	}
	dNatIp := dNat.ReplSrcIP.String()
	dNatPort := dNat.ReplSrcPort
	return dNatIp, int64(dNatPort)
}

// Shutdown cleans all the resources used by the analyzer
func (a *TcpMetricAnalyzer) Shutdown() error {
	return nil
}

// Type returns the type of the analyzer
func (a *TcpMetricAnalyzer) Type() analyzer.Type {
	return TcpMetric
}
