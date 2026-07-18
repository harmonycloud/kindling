package cgoreceiver

/*
#cgo LDFLAGS: -L ./ -lkindling  -lstdc++ -ldl
#cgo CFLAGS: -I .
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include "cgo_func.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"
	"unsafe"

	exporter "github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/pyroscopeexporter"
	"github.com/Kindling-project/kindling/collector/pkg/ebpf/pprof"
	pushv1 "github.com/grafana/pyroscope/api/gen/proto/go/push/v1"
	"github.com/prometheus/prometheus/model/labels"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type SymOptions struct {
	DebugSyms bool
	PerfMap   bool
	MapFiles  bool
}

func (r *CgoReceiver) makeProfilingSender() (*exporter.AsyncSender, error) {
	cfgEndpoint := r.cfg.EbpfProfiling.Endpoint
	if cfgEndpoint.URL == "" {
		r.telemetry.Logger.Error("eBPF profiling endpoint URL is empty; skip periodic sampling")
		return nil, fmt.Errorf("profiling endpoint url is empty")
	}
	if cfgEndpoint.RemoteTimeout == 0 {
		cfgEndpoint.RemoteTimeout = 10 * time.Second
	}
	minBackoff := cfgEndpoint.MinBackoff
	if minBackoff == 0 {
		minBackoff = 500 * time.Millisecond
	}
	maxBackoff := cfgEndpoint.MaxBackoff
	if maxBackoff == 0 {
		maxBackoff = 30 * time.Second
	}
	maxRetries := cfgEndpoint.MaxBackoffRetries
	if maxRetries == 0 {
		maxRetries = 10
	}
	workers := r.cfg.EbpfProfiling.SenderWorkerCount
	if workers <= 0 {
		workers = 2
	}
	queueCap := r.cfg.EbpfProfiling.SenderQueueCapacity
	if queueCap <= 0 {
		queueCap = 200
	}
	return exporter.NewAsyncSender(cfgEndpoint.URL, cfgEndpoint.RemoteTimeout, workers, queueCap, minBackoff, maxBackoff, maxRetries), nil
}

// EnablePeriodicFlamegraphSampling runs CPU stack sampling and pushes to Pyroscope.
// Empty ebpf_profiling.pids uses CAPTURE_ALL; non-empty uses PROFILING_ENABLED whitelist.
func (r *CgoReceiver) EnablePeriodicFlamegraphSampling(interval time.Duration) {
	r.initSymbolizer()
	sender, err := r.makeProfilingSender()
	if err != nil {
		r.telemetry.Logger.Error("Failed to init profiling sender", zap.Error(err))
		r.closeSymbolizer()
		return
	}

	go func() {
		defer func() {
			sender.Stop()
			r.closeSymbolizer()
		}()
		r.startCpuSampling()
		roundNumber := 0
		for {
			timer := time.NewTimer(interval)
			select {
			case <-r.stopCh:
				timer.Stop()
				r.stopCpuSampling()
				r.telemetry.Logger.Info("Periodic CPU sampling stopped")
				return
			case <-timer.C:
			}
			currentTime := time.Now()
			if err := r.collectCpuProfiles(sender); err != nil {
				r.telemetry.Logger.Error("Failed to collect periodic CPU profile", zap.Error(err))
			}
			roundNumber++
			if roundNumber%10 == 0 {
				C.clearStacksMap()
				C.maintainBccCaches(C.int64_t(60))
			}
			r.telemetry.Logger.Info("CPU FlamegraphSampling once cost", zap.Duration("cost", time.Since(currentTime)))
		}
	}()
}

func (r *CgoReceiver) collectCpuProfiles(sender *exporter.AsyncSender) error {
	var keySet C.struct_sample_key_set
	defer C.freeProfileKeys(&keySet)
	if ret := C.getProfileKeys(&keySet); ret != 0 {
		return fmt.Errorf("getProfileKeys failed with ret: %d", ret)
	}
	if keySet.nr_keys == 0 {
		r.telemetry.Logger.Info("no profile data available")
		return nil
	}

	builders := pprof.NewProfileBuilders(97)
	keysSlice := (*[1 << 30]C.struct_sample_key)(unsafe.Pointer(keySet.keys))[:keySet.nr_keys:keySet.nr_keys]
	pidSet := make(map[int64]struct{}, keySet.nr_keys)
	for i := uint32(0); i < uint32(keySet.nr_keys); i++ {
		sample := C.struct_bpf_profile_data{}
		if ret := C.getProfileData(keysSlice[i], &sample); ret != 0 {
			continue
		}
		stack, err := r.MergeStack(&sample, C.GoString(&keysSlice[i].comm[0]))
		if err != nil || stack == nil {
			if err != nil {
				r.telemetry.Logger.Error("MergeStack err: ", zap.Error(err))
			}
			continue
		}
		pidSet[int64(sample.pid)] = struct{}{}
		builder := builders.BuilderForTarget(uint64(sample.pid), nil)
		builder.AddSample(stack, uint64(sample.count))
	}

	labeler := NewLabeler(r.telemetry.GetZapLogger(), "process_cpu")
	perPidLabels := labeler.BuildLabelsForPIDs(pidSet)
	r.telemetry.Logger.Info("Cpu profiling", zap.Int("builders.Builders", len(builders.Builders)))
	return r.sendPprofSeries(sender, builders, perPidLabels)
}

func (r *CgoReceiver) startCpuSampling() {
	pids := r.cfg.EbpfProfiling.Pids
	if len(pids) == 0 {
		r.telemetry.Logger.Info("eBPF CPU profiling: CAPTURE_ALL (no pids configured)")
		C.startCpuSampelingAll()
		return
	}
	r.telemetry.Logger.Info("eBPF CPU profiling: whitelist mode", zap.Int("pid_count", len(pids)))
	if ret := C.setProfilePids((*C.uint32_t)(unsafe.Pointer(&pids[0])), C.uint32_t(len(pids))); ret != 0 {
		r.telemetry.Logger.Error("setProfilePids failed", zap.Int("ret", int(ret)))
		return
	}
	C.startCpuSampeling()
}

func (r *CgoReceiver) stopCpuSampling() {
	C.stopCpuSampeling()
}

func (r *CgoReceiver) initSymbolizer() {
	r.symOnce.Do(func() {
		if C.symbolizerInit() != 0 {
			r.telemetry.Logger.Panic("InitSymbolizer failed", zap.Error(fmt.Errorf("kindling_symbolizer_init failed")))
		}
		r.symInited = true
	})
}

func (r *CgoReceiver) closeSymbolizer() {
	if r.symInited {
		C.symbolizerShutdown()
		r.symInited = false
	}
}

func (r *CgoReceiver) SymbolizeProcessNames(pid uint32, addrs []uint64) ([]string, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("kindling_symbolize_process_names addrs == 0")
	}
	var cNames **C.char
	var outN C.uint32_t
	rc := C.symbolizeProcessNamesWithFallback(
		C.uint32_t(pid),
		(*C.uint64_t)(unsafe.Pointer(&addrs[0])), C.uint32_t(len(addrs)),
		boolToCInt(r.symOpts.DebugSyms), boolToCInt(r.symOpts.PerfMap), boolToCInt(r.symOpts.MapFiles),
		&cNames, &outN,
	)
	if rc != 0 {
		return nil, fmt.Errorf("kindling_symbolize_process_names failed")
	}
	defer C.freeNames(cNames, outN)
	return cStringsToSlice(cNames, outN), nil
}

func (r *CgoReceiver) SymbolizeKernelNames(addrs []uint64) ([]string, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var cNames **C.char
	var outN C.uint32_t
	rc := C.symbolizeKernelNames(
		(*C.uint64_t)(unsafe.Pointer(&addrs[0])), C.uint32_t(len(addrs)),
		&cNames, &outN,
	)
	if rc != 0 {
		return nil, fmt.Errorf("kindling_symbolize_kernel_names failed")
	}
	defer C.freeNames(cNames, outN)
	return cStringsToSlice(cNames, outN), nil
}

func (r *CgoReceiver) MergeStack(sample *C.struct_bpf_profile_data, comm string) ([]string, error) {
	userNames, err := r.SymbolizeProcessNames(uint32(sample.pid), (*[127]uint64)(unsafe.Pointer(&sample.user_stack))[:])
	if err != nil {
		return nil, err
	}
	kernelNames, err := r.SymbolizeKernelNames((*[127]uint64)(unsafe.Pointer(&sample.kernel_stack))[:])
	if err != nil {
		return nil, err
	}
	stack := make([]string, 0, len(userNames)+len(kernelNames)+1)
	stack = append(stack, kernelNames...)
	stack = append(stack, userNames...)
	stack = append(stack, comm)
	if len(stack) == 1 {
		return nil, nil
	}
	return stack, nil
}

type Labeler struct {
	Base   labels.Labels
	logger *zap.Logger
}

func NewLabeler(logger *zap.Logger, profileName string) *Labeler {
	hostIp := os.Getenv("MY_NODE_IP")
	return &Labeler{
		Base: labels.Labels{
			{Name: "__name__", Value: profileName},
			{Name: "__delta__", Value: "false"},
			{Name: "service_name", Value: "kindling"},
			{Name: "host_ip", Value: hostIp},
		},
		logger: logger,
	}
}

// BuildLabelsForPIDs builds minimal per-pid labels (no PID control-plane container lookup).
func (l *Labeler) BuildLabelsForPIDs(pidSet map[int64]struct{}) map[int64]labels.Labels {
	res := make(map[int64]labels.Labels, len(pidSet))
	for pid := range pidSet {
		lbs := append(append(labels.Labels(nil), l.Base...), labels.Label{
			Name: "pid", Value: strconv.FormatInt(pid, 10),
		})
		res[pid] = lbs
	}
	return res
}

func (r *CgoReceiver) sendPprofSeries(sender *exporter.AsyncSender, builders *pprof.ProfileBuilders, perPidLabels map[int64]labels.Labels) error {
	const (
		maxSeriesPerReq = 80
		maxBytesPerReq  = 1 << 17 // 128 kB
	)
	curBytes := 0
	protoSeries := make([]*pushv1.RawProfileSeries, 0, maxSeriesPerReq)
	flush := func() {
		if len(protoSeries) == 0 {
			return
		}
		_copy := make([]*pushv1.RawProfileSeries, len(protoSeries))
		copy(_copy, protoSeries)
		_ = exporter.BulkSend(_copy, sender)
		protoSeries = protoSeries[:0]
		curBytes = 0
	}

	for pid, builder := range builders.Builders {
		buf := bytes.NewBuffer(nil)
		_, err := builder.Write(buf)
		if err != nil {
			return fmt.Errorf("pprof encode %w", err)
		}
		rawProfile := buf.Bytes()
		realSamples := []*exporter.RawSample{{RawProfile: rawProfile}}
		lbs := append(labels.Labels(nil), perPidLabels[int64(pid)]...)
		protoLabels, rawSample, _ := exporter.PrepareDataForTransmission(lbs, realSamples)
		if len(protoSeries) >= maxSeriesPerReq || curBytes >= maxBytesPerReq {
			flush()
		}
		series := &pushv1.RawProfileSeries{Labels: protoLabels, Samples: rawSample}
		protoSeries = append(protoSeries, series)
		curBytes += proto.Size(series)
	}
	flush()
	return nil
}

func cStringsToSlice(cNames **C.char, n C.uint32_t) []string {
	count := int(n)
	if count == 0 || cNames == nil {
		return nil
	}
	arr := (*[1 << 30]*C.char)(unsafe.Pointer(cNames))[:count:count]
	out := make([]string, count)
	for i := 0; i < count; i++ {
		out[i] = C.GoString(arr[i])
	}
	return out
}

func boolToCInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
