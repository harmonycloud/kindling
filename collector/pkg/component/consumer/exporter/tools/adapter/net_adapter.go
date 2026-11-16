// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
)

type NetMetricGroupAdapter struct {
	*NetAdapterManager
	*NetAdapterConfig
}

type NetAdapterConfig struct {
	StoreTraceAsMetric bool
	StoreTraceAsSpan   bool
	StorePodDetail     bool
	StoreExternalSrcIP bool
}

func (n *NetMetricGroupAdapter) Adapt(dataGroup *model.DataGroup, attrType AttrType) ([]*AdaptedResult, error) {
	switch dataGroup.Name {
	case constnames.AggregatedNetRequestMetricGroup:
		return n.dealWithPreAggMetricGroups(dataGroup, attrType), nil
	case constnames.SingleNetRequestMetricGroup:
		return n.dealWithSingleMetricGroup(dataGroup, attrType), nil
	default:
		return nil, nil
	}
}

func (n *NetMetricGroupAdapter) dealWithSingleMetricGroup(dataGroup *model.DataGroup, attrType AttrType) []*AdaptedResult {
	requestTotalTime, ok := dataGroup.GetMetric(constvalues.RequestTotalTime)
	if !ok {
		return nil
	}
	results := make([]*AdaptedResult, 0, 2)
	if n.StoreTraceAsMetric {
		// Since TraceAsMetric has to be aggregated again, the attrType of TraceAsMetric must be `AttributeMap`
		results = append(results, createMetric(
			[]*model.Metric{model.NewMetric(constnames.TraceAsMetric, requestTotalTime.GetData())},
			dataGroup, n.traceToMetricAdapter,
			AttributeMap))
	}
	if n.StoreTraceAsSpan {
		results = append(results, createTrace([]*model.Metric{requestTotalTime}, dataGroup, n.traceToSpanAdapter, attrType))
	}
	return results
}

func (n *NetMetricGroupAdapter) dealWithPreAggMetricGroups(dataGroup *model.DataGroup, attrType AttrType) []*AdaptedResult {
	results := make([]*AdaptedResult, 0, 4)
	isServer := dataGroup.Labels.GetBoolValue(constlabels.IsServer)
	srcNamespace := dataGroup.Labels.GetStringValue(constlabels.SrcNamespace)
	if n.StoreExternalSrcIP && srcNamespace == constlabels.ExternalClusterNamespace && isServer {
		externalAdapterCache := n.detailTopologyAdapter
		externalTopology := n.createNetMetricResults(dataGroup, externalAdapterCache, attrType, true)
		results = append(results, externalTopology...)
	}

	var metricAdapterCache [2]*LabelConverter
	if n.StorePodDetail {
		if isServer {
			metricAdapterCache = n.detailEntityAdapter
		} else {
			metricAdapterCache = n.detailTopologyAdapter
		}
	} else {
		if isServer {
			metricAdapterCache = n.aggEntityAdapter
		} else {
			metricAdapterCache = n.aggTopologyAdapter
		}
	}
	metrics := n.createNetMetricResults(dataGroup, metricAdapterCache, attrType, false)
	return append(results, metrics...)
}

func (n *NetMetricGroupAdapter) createNetMetricResults(dataGroup *model.DataGroup, adapter [2]*LabelConverter, attrType AttrType, toTopology bool) (tmpResults []*AdaptedResult) {
	values := dataGroup.Metrics
	isServer := dataGroup.Labels.GetBoolValue(constlabels.IsServer)
	metricsExceptRequestCount := make([]*model.Metric, 0, len(values))
	requestCount := make([]*model.Metric, 0, 1)
	for _, metric := range dataGroup.Metrics {
		if metric.Name != constvalues.RequestCount {
			metricsExceptRequestCount = append(metricsExceptRequestCount, model.NewMetric(constnames.ToKindlingNetMetricName(metric.Name, isServer && !toTopology), metric.GetData()))
		} else {
			requestCount = append(requestCount, model.NewMetric(constnames.ToKindlingNetMetricName(metric.Name, isServer && !toTopology), metric.GetData()))
		}
	}
	tmpResults = make([]*AdaptedResult, 0, 2)
	if len(metricsExceptRequestCount) > 0 {
		tmpResults = append(tmpResults, createMetric(metricsExceptRequestCount, dataGroup, adapter[0], attrType))
	}
	if len(requestCount) > 0 {
		tmpResults = append(tmpResults, createMetric(requestCount, dataGroup, adapter[1], attrType))
	}
	return
}

func createMetric(metrics []*model.Metric, dataGroup *model.DataGroup, adapter *LabelConverter, addrType AttrType) *AdaptedResult {
	switch addrType {
	case AttributeMap:
		attrs, free := adapter.transform(dataGroup)
		return &AdaptedResult{
			ResultType:   Metric,
			AttrsMap:     attrs,
			Metrics:      metrics,
			Timestamp:    dataGroup.Timestamp,
			FreeAttrsMap: free,
		}
	case AttributeList:
		attrs, free := adapter.convert(dataGroup)
		return &AdaptedResult{
			ResultType:    Metric,
			AttrsList:     attrs,
			Metrics:       metrics,
			Timestamp:     dataGroup.Timestamp,
			FreeAttrsList: free,
		}
	}
	return nil
}

func createTrace(metrics []*model.Metric, dataGroup *model.DataGroup, adapter *LabelConverter, addrType AttrType) *AdaptedResult {
	switch addrType {
	case AttributeMap:
		attrs, free := adapter.transform(dataGroup)
		return &AdaptedResult{
			ResultType:   Trace,
			AttrsMap:     attrs,
			Metrics:      metrics,
			Timestamp:    dataGroup.Timestamp,
			FreeAttrsMap: free,
		}
	case AttributeList:
		attrs, free := adapter.convert(dataGroup)
		return &AdaptedResult{
			ResultType:    Trace,
			AttrsList:     attrs,
			Metrics:       metrics,
			Timestamp:     dataGroup.Timestamp,
			FreeAttrsList: free,
		}
	}
	return nil
}

type NetAdapterManager struct {
	detailEntityAdapter   [2]*LabelConverter
	aggEntityAdapter      [2]*LabelConverter
	detailTopologyAdapter [2]*LabelConverter
	aggTopologyAdapter    [2]*LabelConverter
	traceToSpanAdapter    *LabelConverter
	traceToMetricAdapter  *LabelConverter
}

func createNetAdapterManager(constLabels []attribute.KeyValue) *NetAdapterManager {
	// TODO deal Error
	aggEntityAdapterWithIsSlow, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{isSlowDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	detailEntityAdapterWithIsSlow, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{entityInstanceMetricDicList, entityDetailMetricDicList, isSlowDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	aggTopologyAdapterWithIsSlow, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{isSlowDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(removeDstPodInfoForNonExternal()).
		withConstLabels(constLabels).
		build()

	detailTopologyAdapterWithIsSlow, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList, isSlowDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(replaceDstIpOrDstPortByDNat()).
		withConstLabels(constLabels).
		build()

	aggEntityAdapter, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	detailEntityAdapter, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{entityInstanceMetricDicList, entityDetailMetricDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	aggTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(removeDstPodInfoForNonExternal()).
		withConstLabels(constLabels).
		build()

	detailTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(replaceDstIpOrDstPortByDNat()).
		withConstLabels(constLabels).
		build()

	traceToSpanAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, SpanDicList, dNatDicList}).
		withExtraLabels(spanProtocol, updateProtocolKey).
		withValueToLabels(traceSpanStatus, getTraceSpanStatusLabels).
		withConstLabels(constLabels).
		build()

	traceToMetricAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList, dNatDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withValueToLabels(traceStatus, getTraceStatusLabels).
		withConstLabels(constLabels).
		build()

	return &NetAdapterManager{
		aggEntityAdapter:      [2]*LabelConverter{aggEntityAdapter, aggEntityAdapterWithIsSlow},
		detailEntityAdapter:   [2]*LabelConverter{detailEntityAdapter, detailEntityAdapterWithIsSlow},
		aggTopologyAdapter:    [2]*LabelConverter{aggTopologyAdapter, aggTopologyAdapterWithIsSlow},
		detailTopologyAdapter: [2]*LabelConverter{detailTopologyAdapter, detailTopologyAdapterWithIsSlow},
		traceToSpanAdapter:    traceToSpanAdapter,
		traceToMetricAdapter:  traceToMetricAdapter,
	}
}

func NewNetAdapter(
	customLabels []attribute.KeyValue,
	config *NetAdapterConfig,
) *NetMetricGroupAdapter {
	return &NetMetricGroupAdapter{
		NetAdapterManager: createNetAdapterManager(customLabels),
		NetAdapterConfig:  config,
	}
}
