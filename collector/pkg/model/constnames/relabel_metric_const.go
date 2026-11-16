// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package constnames

import "github.com/Kindling-project/kindling/collector/pkg/model/constvalues"

// key1: originName key2: isServer
var metricNameDictionary = map[string]map[bool]string{
	constvalues.RequestIo:                 {true: EntityRequestIoMetric, false: TopologyRequestIoMetric},
	constvalues.ResponseIo:                {true: EntityResponseIoMetric, false: TopologyResponseIoMetric},
	constvalues.RequestTotalTime:          {true: EntityRequestLatencyTotalMetric, false: TopologyRequestLatencyTotalMetric},
	constvalues.RequestCount:              {true: EntityRequestCountMetric, false: TopologyRequestCountMetric},
	constvalues.RequestTotalTime + "_avg": {true: EntityRequestLatencyAverageMetric, false: TopologyRequestLatencyAverageMetric},
	constvalues.RequestTimeHistogram:      {true: EntityRequestTimeHistogramMetric, false: TopologyRequestTimeHistogramMetric},
}

const (
	TopologyRequestIoMetric  = "request_bytes_total"
	TopologyResponseIoMetric = "response_bytes_total"
	// TopologyRequestLatencyAverageMetric is a histogram
	TopologyRequestLatencyAverageMetric = "average_duration_nanoseconds"
	TopologyRequestLatencyTotalMetric   = "duration_nanoseconds_total"
	TopologyRequestCountMetric          = "total"
	// TopologyRequestTimeHistogramMetric is a histogram
	TopologyRequestTimeHistogramMetric = "request_time_histogram"

	EntityRequestIoMetric  = "receive_bytes_total"
	EntityResponseIoMetric = "send_bytes_total"
	// EntityRequestLatencyAverageMetric is a histogram
	EntityRequestLatencyAverageMetric = "average_duration_nanoseconds"
	EntityRequestLatencyTotalMetric   = "duration_nanoseconds_total"
	EntityRequestCountMetric          = "total"
	EntityRequestTimeHistogramMetric  = "request_time_histogram"

	TraceAsMetric           = NPMPrefixKindling + "_trace_request_duration_nanoseconds"
	TcpRttMetricName        = "kindling_tcp_srtt_microseconds"
	TcpRetransmitMetricName = "kindling_tcp_retransmit_total"
	TcpDropMetricName       = "kindling_tcp_packet_loss_total"
	K8sWorkLoadMetricName   = "kindling_k8s_workload_info"

	TcpConnectTotalMetric    = "kindling_tcp_connect_total"
	TcpConnectDurationMetric = "kindling_tcp_connect_duration_nanoseconds_total"
)

const (
	NPMPrefixKindling = "kindling"

	EntityPrefix   = "entity"
	TopologyPrefix = "topology"
)

func ToKindlingNetMetricName(origName string, isServer bool) string {
	if names, ok := metricNameDictionary[origName]; !ok {
		return ""
	} else {
		return getKindlingPrefix(isServer) + "request_" + names[isServer]
	}
}

//ToKindlingDetailMetricName For ServerDetail Metric
func ToKindlingDetailMetricName(origName string, protocol string) string {
	if names, ok := metricNameDictionary[origName]; !ok {
		return ""
	} else {
		return getKindlingPrefix(true) + protocol + "_" + names[true]
	}
}

func getKindlingPrefix(isServer bool) string {
	var kindMark string
	if isServer {
		kindMark = EntityPrefix
	} else {
		kindMark = TopologyPrefix
	}
	return NPMPrefixKindling + "_" + kindMark + "_"
}
