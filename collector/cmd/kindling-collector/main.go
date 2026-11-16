// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	_ "net/http/pprof"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kindling-project/kindling/collector/internal/application"
	"github.com/Kindling-project/kindling/collector/pkg/version"
)

func main() {
	// Print version information
	log.Printf("GitCommitInfo:%s\n", version.Version())
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	app, err := application.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}

	// Register signal handler
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	// Block until a signal is received.
	select {
	case sig := <-sigCh:
		log.Printf("Received signal [%v], and will exit", sig)
		if err = app.Shutdown(); err != nil {
			log.Printf("Error happened when shutting down: %v", err)
			os.Exit(1)
		}
	}
}
