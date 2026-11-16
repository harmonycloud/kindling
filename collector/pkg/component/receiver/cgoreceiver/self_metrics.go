// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cgoreceiver

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

var once sync.Once

const (
	eventReceivedMetric    = "kindling_telemetry_cgoreceiver_events_total"
	channelSizeMetric      = "kindling_telemetry_cgoreceiver_channel_size"
	probeEventMetric       = "kindling_telemetry_cgoreceiver_probe_event_total"
	dropProbeEventMetric   = "kindling_telemetry_cgoreceiver_dropped_probe_event_total"
	preemptionsMetric      = "kindling_telemetry_cgoreceiver_preemptions_total"
	skippedEventMetric     = "kindling_telemetry_cgoreceiver_skipped_events_total"
	suppressedThreadMetric = "kindling_telemetry_cgoreceiver_suppressed_thread_total"
)

func newSelfMetrics(meterProvider metric.MeterProvider, receiver *CgoReceiver) {
	once.Do(func() {
		meter := metric.Must(meterProvider.Meter("kindling"))
		meter.NewInt64CounterObserver(eventReceivedMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				for name, value := range receiver.stats.getStats() {
					result.Observe(value, attribute.String("name", name))
				}
			}, metric.WithDescription("The total number of the events received by cgoreceiver"))
		meter.NewInt64GaugeObserver(channelSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(len(receiver.eventChannel)))
			}, metric.WithDescription("The current number of events contained in the channel. The maximum size is 300,000."))
		meter.NewInt64CounterObserver(probeEventMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				receiver.probeCounterMutex.RLock()
				result.Observe(receiver.probeCounter.evts)
				receiver.probeCounterMutex.RUnlock()
			}, metric.WithDescription("The events seen by driver"))
		meter.NewInt64CounterObserver(dropProbeEventMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				receiver.probeCounterMutex.RLock()
				result.Observe(receiver.probeCounter.dropsBuffer, attribute.String("reason", "full buffer"))
				result.Observe(receiver.probeCounter.dropsPf, attribute.String("reason", "invalid memory access"))
				result.Observe(receiver.probeCounter.dropsBug, attribute.String("reason", "invalid condition"))
				result.Observe(receiver.probeCounter.drops-receiver.probeCounter.dropsBuffer-receiver.probeCounter.dropsPf-receiver.probeCounter.dropsBug, attribute.String("reason", "others"))
				receiver.probeCounterMutex.RUnlock()
			}, metric.WithDescription("The dropped events"))
		meter.NewInt64CounterObserver(preemptionsMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				receiver.probeCounterMutex.RLock()
				result.Observe(receiver.probeCounter.preemptions)
				receiver.probeCounterMutex.RUnlock()
			}, metric.WithDescription("The preemptions"))
		meter.NewInt64CounterObserver(skippedEventMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				receiver.probeCounterMutex.RLock()
				result.Observe(receiver.probeCounter.suppressed)
				receiver.probeCounterMutex.RUnlock()
			}, metric.WithDescription("Number of events skipped due to the tid being in a set of suppressed tids"))
		meter.NewInt64CounterObserver(suppressedThreadMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				receiver.probeCounterMutex.RLock()
				result.Observe(receiver.probeCounter.tidsSuppressed)
				receiver.probeCounterMutex.RUnlock()
			}, metric.WithDescription("Number of threads currently being suppressed"))

	})
}

type eventCounter interface {
	add(name string, value int64)
	getStats() map[string]int64
}

type atomicInt64Counter struct {
	v int64
}

type probeCounter struct {
	evts           int64
	drops          int64
	dropsBuffer    int64
	dropsPf        int64
	dropsBug       int64
	preemptions    int64
	suppressed     int64
	tidsSuppressed int64
}

func (c *atomicInt64Counter) add(value int64) {
	atomic.AddInt64(&c.v, value)
}

func (c *atomicInt64Counter) get() int64 {
	return atomic.LoadInt64(&c.v)
}

type dynamicStats struct {
	stats map[string]*atomicInt64Counter
}

func newDynamicStats(subEvents []SubEvent) *dynamicStats {
	ret := &dynamicStats{
		stats: make(map[string]*atomicInt64Counter),
	}
	for _, event := range subEvents {
		var rawName string
		nameSegments := strings.Split(event.Name, "-")
		if len(nameSegments) > 1 {
			rawName = nameSegments[1]
		} else {
			rawName = nameSegments[0]
		}
		ret.stats[rawName] = &atomicInt64Counter{0}
	}
	ret.stats[constnames.OtherEvent] = &atomicInt64Counter{0}
	return ret
}

func (s *dynamicStats) add(name string, value int64) {
	c, ok := s.stats[name]
	if ok {
		c.add(value)
	} else {
		c = s.stats[constnames.OtherEvent]
		c.add(value)
	}
}

func (s *dynamicStats) getStats() map[string]int64 {
	ret := make(map[string]int64, len(s.stats))
	for k, v := range s.stats {
		ret[k] = v.get()
	}
	return ret
}
