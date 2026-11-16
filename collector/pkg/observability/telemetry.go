// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package observability

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	exportmetric "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
)

const (
	clusterIdEnv              = "CLUSTER_ID"
	userIdEnv                 = "USER_ID"
	regionIdEnv               = "REGION_ID"
	KindlingServiceNamePrefix = "kindling"
	StdoutKindExporter        = "stdout"
	OtlpGrpcKindExporter      = "otlp"
	PrometheusKindExporter    = "prometheus"
)

var (
	selfTelemetryOnce   sync.Once
	agentCPUTimeSeconds metric.Float64CounterObserver
	agentMemUsedBytes   metric.Float64GaugeObserver
)

const (
	resourcePerformance = "resource"
)

const (
	agentCPUTimeSecondsMetric  = "kindling_telemetry_agent_cpu_time_seconds"
	agentMemoryUsedBytesMetric = "kindling_telemetry_agent_memory_used_bytes"
)

type otelLoggerHandler struct {
	logger *zap.Logger
}

func (h *otelLoggerHandler) Handle(err error) {
	h.logger.Warn("Opentelemetry-go encountered an error: ", zap.Error(err))
}

func RegisterExtraMetrics(selectors []string, mp metric.MeterProvider) {
	for _, selector := range selectors {
		switch selector {
		case resourcePerformance:
			registerAgentResourcePerformanceMetrics(mp)
		}
	}
}

func registerAgentResourcePerformanceMetrics(mp metric.MeterProvider) (err error) {
	proc, _ := process.NewProcess(int32(os.Getpid()))
	meter := mp.Meter("kindling")
	selfTelemetryOnce.Do(func() {
		agentCPUTimeSeconds, err = meter.NewFloat64CounterObserver(agentCPUTimeSecondsMetric, func(ctx context.Context, result metric.Float64ObserverResult) {
			cpuTime, _ := proc.Times()
			result.Observe(cpuTime.User, attribute.String("type", "user"))
			result.Observe(cpuTime.System, attribute.String("type", "system"))
		})
		if err != nil {
			return
		}
		agentMemUsedBytes, err = meter.NewFloat64GaugeObserver(agentMemoryUsedBytesMetric, func(ctx context.Context, result metric.Float64ObserverResult) {
			mem, _ := proc.MemoryInfo()
			result.Observe(float64(mem.RSS), attribute.String("type", "rss"))
			result.Observe(float64(mem.VMS), attribute.String("type", "vms"))
		})
		if err != nil {
			return
		}
	})
	return nil
}

func InitTelemetry(logger *zap.Logger, config *Config) (metric.MeterProvider, error) {
	otel.SetErrorHandler(&otelLoggerHandler{logger: logger})
	hostName, err := os.Hostname()
	if err != nil {
		logger.Info("Cannot get hostname; set hostname unknown: ", zap.Error(err))
		hostName = "unknown"
	}

	clusterId, ok := os.LookupEnv(clusterIdEnv)
	if !ok {
		logger.Info("[CLUSTER_ID] is not found in env variable which will be set [noclusteridset]")
		clusterId = "noclusteridset"
	}

	serviceName := KindlingServiceNamePrefix + "-" + clusterId
	rs, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceInstanceIDKey.String(hostName),
		),
	)

	if config.ExportKind == PrometheusKindExporter {
		// Create controller
		c := controller.New(
			otelprocessor.NewFactory(
				selector.NewWithInexpensiveDistribution(),
				aggregation.CumulativeTemporalitySelector(),
			),
			controller.WithCollectPeriod(0), // In pull mode, make prometheu deceide the collect Period
			controller.WithResource(rs),
		)

		promConfig := prometheus.Config{
			DefaultHistogramBoundaries: exponentialInt64NanosecondsBoundaries,
		}
		exp, err := prometheus.New(promConfig, c)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize self-telemetry prometheus %w", err)
		}

		go func() {
			err := StartServer(exp, logger, config.PromCfg.Port)
			if err != nil {
				logger.Warn("error starting self-telemetry server: ", zap.Error(err))
			}
		}()

		mp := exp.MeterProvider()
		RegisterExtraMetrics(config.PromCfg.ExtraMetrics, mp)
		return mp, nil
	} else {
		var collectPeriod time.Duration

		if config.ExportKind == StdoutKindExporter {
			collectPeriod = config.StdoutCfg.CollectPeriod
		} else if config.ExportKind == OtlpGrpcKindExporter {
			collectPeriod = config.OtlpGrpcCfg.CollectPeriod
		} else {
			return nil, fmt.Errorf("no self-telemetry exporter kind matched, current: [%v]", config.ExportKind)
		}

		exporters, err := newExporters(context.Background(), config, logger)
		if err != nil {
			return nil, fmt.Errorf("error happened when creating self-telemetry exporter: %w", err)
		}

		cont := controller.New(
			otelprocessor.NewFactory(selector.NewWithInexpensiveDistribution(), exporters),
			controller.WithExporter(exporters),
			controller.WithCollectPeriod(collectPeriod),
			controller.WithResource(rs),
		)

		if err = cont.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to start self-telemetry controller: %w", err)
		}

		return cont, nil
	}
}

// Crete new opentelemetry-go exporter.
func newExporters(context context.Context, cfg *Config, logger *zap.Logger) (exportmetric.Exporter, error) {
	logger.Sugar().Infof("Initializing self-observability exporter whose type is %s", cfg.ExportKind)
	switch cfg.ExportKind {
	case StdoutKindExporter:
		metricExp, err := stdoutmetric.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter, %w", err)
		}
		return metricExp, nil
	case OtlpGrpcKindExporter:
		metricExporter, err := otlpmetricgrpc.New(context,
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithEndpoint(cfg.OtlpGrpcCfg.Endpoint),
			otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetrySettings{
				Enabled:         true,
				InitialInterval: 300 * time.Millisecond,
				MaxInterval:     5 * time.Second,
				MaxElapsedTime:  15 * time.Second,
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter, %w", err)
		}
		return metricExporter, nil
	default:
		return nil, errors.New("failed to create exporter, no exporter kind is provided")
	}
}

func StartServer(exporter *prometheus.Exporter, logger *zap.Logger, port string) error {
	serveMux := http.ServeMux{}
	serveMux.HandleFunc("/metrics", exporter.ServeHTTP)

	srv := http.Server{
		Addr:    port,
		Handler: &serveMux,
	}

	logger.Sugar().Infof("Prometheus Server listening at port: [%s]", port)
	err := srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	logger.Sugar().Infof("Prometheus gracefully shutdown the http server...\n")

	return nil
}
