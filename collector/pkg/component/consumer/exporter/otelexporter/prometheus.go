// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

var (
	mu   sync.Mutex
	srv  *http.Server
)

func StartServer(exporter *prometheus.Exporter, telemetry *component.TelemetryTools, port string) error {
	mu.Lock()
	defer mu.Unlock()

	if srv != nil {
		if err := srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("failed to stop server: %w", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", exporter.ServeHTTP)

	srv = &http.Server{
		Addr:    port,
		Handler: mux,
	}

	telemetry.Logger.Infof("Prometheus Server listening at port: [%s]", port)

	go func() {
		srv.ListenAndServe()
		telemetry.Logger.Infof("Prometheus gracefully shutdown the http server...\n")
	}()

	return nil
}
