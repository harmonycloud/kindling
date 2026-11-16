// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package prometheusexporter

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/tools/adapter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"go.uber.org/zap"
)

func (p *prometheusExporter) Consume(dataGroup *model.DataGroup) error {
	if dataGroup == nil {
		// no need consume
		return nil
	}
	if ce := p.telemetry.Logger.Check(zap.DebugLevel, ""); ce != nil {
		p.telemetry.Logger.Debug("exporter receives a dataGroup:\n" + dataGroup.String())
	}

	if adapters, ok := p.adapters[dataGroup.Name]; ok {
		for i := 0; i < len(adapters); i++ {
			results, err := adapters[i].Adapt(dataGroup, adapter.AttributeMap)
			if err != nil {
				p.telemetry.Logger.Error("Failed to adapt dataGroup", zap.Error(err))
			}
			if len(results) > 0 {
				p.Export(results)
			}
		}
	} else {
		results, err := p.adapter.Adapt(dataGroup, adapter.AttributeMap)
		if err != nil {
			p.telemetry.Logger.Error("Failed to adapt dataGroup", zap.Error(err))
		}
		if len(results) > 0 {
			p.Export(results)
		}
	}
	return nil
}

func (p *prometheusExporter) Export(results []*adapter.AdaptedResult) {
	for i := 0; i < len(results); i++ {
		result := results[i]
		switch result.ResultType {
		case adapter.Metric:
			p.exportMetric(result)
		default:
			p.telemetry.Logger.Error("Unexpected ResultType", zap.String("type", string(result.ResultType)))
		}
		result.Free()
	}
}

func (p *prometheusExporter) exportMetric(result *adapter.AdaptedResult) {
	p.collector.recordMetricGroups(model.NewDataGroup("", result.AttrsMap, result.Timestamp, result.Metrics...))
}
