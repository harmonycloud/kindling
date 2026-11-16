// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package prometheusexporter

import (
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	adapter3 "github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/tools/adapter"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

const Type = "prometheus"

type prometheusExporter struct {
	cfg                  *Config
	metricAggregationMap map[string]MetricAggregationKind
	customLabels         []attribute.KeyValue
	telemetry            *component.TelemetryTools
	collector            *collector
	registry             *prometheus.Registry
	shutdownFunc         func() error
	handler              http.Handler
	adapters             map[string][]adapter3.Adapter
	adapter              adapter3.Adapter
}

func NewExporter(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Panic("Cannot convert Component config", zap.String("componentType", Type))
	}

	collector := newCollector(cfg, telemetry.Logger)
	registry := prometheus.NewRegistry()
	_ = registry.Register(collector)

	netAdapter := adapter3.NewNetAdapter(nil, &adapter3.NetAdapterConfig{
		StoreTraceAsMetric: cfg.AdapterConfig.NeedTraceAsMetric,
		StoreTraceAsSpan:   false,
		StorePodDetail:     cfg.AdapterConfig.NeedPodDetail,
		StoreExternalSrcIP: cfg.AdapterConfig.StoreExternalSrcIP,
	})
	simpleAdapter := adapter3.NewSimpleAdapter([]string{constnames.TcpRttMetricGroupName, constnames.TcpRetransmitMetricGroupName,
		constnames.TcpDropMetricGroupName}, nil)

	prometheusExporter := &prometheusExporter{
		cfg:       cfg,
		collector: collector,
		handler: promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				ErrorHandling: promhttp.ContinueOnError,
				ErrorLog:      newPromLogger(telemetry.Logger),
			},
		),
		// metricAggregationMap: cfg.MetricAggregationMap,
		telemetry: telemetry,
		adapters: map[string][]adapter3.Adapter{
			constnames.NetRequestMetricGroupName:       {netAdapter},
			constnames.AggregatedNetRequestMetricGroup: {netAdapter},
			constnames.SingleNetRequestMetricGroup:     {netAdapter},
			constnames.TcpRttMetricGroupName:           {simpleAdapter},
			constnames.TcpRetransmitMetricGroupName:    {simpleAdapter},
			constnames.TcpDropMetricGroupName:          {simpleAdapter},
		},
		adapter: simpleAdapter,
	}

	go func() {
		_ = prometheusExporter.Start(context.Background())
	}()
	return prometheusExporter
}

func (p *prometheusExporter) findInstrumentKind(metricName string) (MetricAggregationKind, bool) {
	kind, find := p.metricAggregationMap[metricName]
	return kind, find
}

func (p *prometheusExporter) Start(_ context.Context) error {
	ln, err := net.Listen("tcp", p.cfg.PromCfg.Endpoint)
	if err != nil {
		return err
	}

	p.shutdownFunc = ln.Close

	mux := http.NewServeMux()
	mux.Handle("/metrics", p.handler)
	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()

	return nil
}

type promLogger struct {
	realLog *component.TelemetryLogger
}

func newPromLogger(zapLog *component.TelemetryLogger) *promLogger {
	return &promLogger{
		realLog: zapLog,
	}
}

func (l *promLogger) Println(v ...interface{}) {
	l.realLog.Error(fmt.Sprintln(v...))
}
