// Package processmanager manages loading and unloading of per-process unwind metadata.
package processmanager

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	lru "github.com/elastic/go-freelru"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter/interpreterconfig"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/process"
	pmebpf "github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
	eim "github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/execinfomanager"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/lpm"
	"go.opentelemetry.io/ebpf-profiler/nativeunwind"
	"go.opentelemetry.io/ebpf-profiler/periodiccaller"
	"go.opentelemetry.io/ebpf-profiler/times"
	"go.opentelemetry.io/ebpf-profiler/util"
)

const (
	elfInfoCacheSize = 16384
	elfInfoCacheTTL  = 6 * time.Hour
)

var dummyPrefix = lpm.Prefix{Key: 0, Length: 64}

// Config contains dependencies used to create a ProcessManager.
type Config struct {
	InterpretersConfig    interpreterconfig.Config
	ExecutableUnloadDelay time.Duration
	EbpfHandler           pmebpf.EbpfHandler
	StackDeltaProvider    nativeunwind.StackDeltaProvider
	IncludeEnvVars        libpf.Set[string]
	ProcessMetaEnrichers  []process.MetaEnricher
}

// New creates a ProcessManager for preprocess (SynchronizeProcess / mapping sync).
func New(ctx context.Context, cfg Config) (*ProcessManager, error) {
	if cfg.ExecutableUnloadDelay <= 0 {
		cfg.ExecutableUnloadDelay = time.Minute
	}
	elfInfoCache, err := lru.New[util.OnDiskFileIdentifier, elfInfo](elfInfoCacheSize,
		util.OnDiskFileIdentifier.Hash32)
	if err != nil {
		return nil, fmt.Errorf("unable to create elfInfoCache: %v", err)
	}
	elfInfoCache.SetLifetime(elfInfoCacheTTL)

	em, err := eim.NewExecutableInfoManager(cfg.StackDeltaProvider, cfg.EbpfHandler, cfg.InterpretersConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create ExecutableInfoManager: %v", err)
	}

	periodiccaller.Start(ctx, cfg.ExecutableUnloadDelay, func() {
		if err := em.CleanupUnused(cfg.ExecutableUnloadDelay); err != nil {
			slog.Error(fmt.Sprintf("Failed to cleanup unused executables: %v", err))
		}
	})

	metaEnrichers := make([]process.MetaEnricher, 0, len(cfg.ProcessMetaEnrichers)+2)
	if len(cfg.IncludeEnvVars) > 0 {
		metaEnrichers = append(metaEnrichers, process.NewEnvVarsEnricher(cfg.IncludeEnvVars))
	}

	selfContainerEnricher, err := process.NewSelfContainerIDEnricher()
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to detect self container ID via inode: %v", err))
	} else {
		metaEnrichers = append(metaEnrichers, selfContainerEnricher)
	}
	metaEnrichers = append(metaEnrichers, cfg.ProcessMetaEnrichers...)

	pm := &ProcessManager{
		interpreterTracerEnabled: em.NumInterpreterLoaders() > 0,
		eim:                      em,
		interpreters:             make(map[libpf.PID]map[util.OnDiskFileIdentifier]interpreter.Instance),
		exitEvents:               make(map[libpf.PID]times.KTime),
		pidToProcessInfo:         make(map[libpf.PID]*processInfo),
		ebpf:                     cfg.EbpfHandler,
		elfInfoCache:             elfInfoCache,
		metaEnrichers:            metaEnrichers,
	}

	return pm, nil
}

// Close releases resources held by the process manager.
func (pm *ProcessManager) Close() {}
