# eBPF CPU profiling

Kindling samples Native/C/C++, Go and CPython CPU stacks at approximately 97 Hz
and sends gzip pprof profiles to Pyroscope. Native unwinding uses ELF unwind
metadata and frame pointers. Go symbols are read from pclntab; Python frames are
resolved while their code objects are still available. Kernel symbols and native
fallbacks use the BCC symbolizer linked into `libkindling.so`.

This path is independent of Camera On/OffCPU profiling. It does not include Java
or allocation/memory profiling.

## Configuration

Set `receivers.cgoreceiver.ebpf_profiling` in either deployment configuration:

- `deploy/agent/kindling-collector-config.yml`
- `collector/docker/kindling-collector-config.yml`

```yaml
receivers:
  cgoreceiver:
    ebpf_profiling:
      endpoint:
        url: "http://pyroscope:4040/" # Empty disables CPU profiling.
        remote_timeout: 10s
        min_backoff_period: 500ms
        max_backoff_period: 30s
        max_backoff_retries: 10
      sender_worker_count: 2
      sender_queue_capacity: 200
      pids: [1234, 5678] # Host PIDs.
      metrics:
        enable: true
        interval: 60s
```

An empty `endpoint.url` disables profiling. An omitted or empty `pids` list also
disables CPU sampling. Set `pids: [0]` to explicitly sample all processes
(`CAPTURE_ALL`), or list positive host PIDs such as `[1234, 5678]` for whitelist
sampling. The all-process marker `0` must be used alone; mixing it with other
PIDs is a configuration error. Configuration is read at startup, so restart the
agent after editing the file or ConfigMap. No page configuration handler or
runtime configuration endpoint is included.

The profile name is `process_cpu`. Labels include `service_name=kindling`,
`host_ip` from `MY_NODE_IP`, `pid`, and `container_id` when known. Export defaults
to every 15 seconds. `metrics.enable` enables periodic self-monitoring logs for
BPF samples/errors, PID reports, mapping synchronization, stack aggregation,
symbolization and push/retry outcomes.

## Build and deployment

Collector, probe and agent-libs must be upgraded together. The old aggregated
`cpu_counts`/`cpu_stacks` protocol is replaced by typed trace events and native
unwind maps. Initialize the pinned submodule before building:

```sh
git submodule update --init --recursive
```

Build the collector with Go 1.23 or later. The preprocessing dependency is pinned
to commit `7ca45ed3c933` from the `kindling-preprocess` branch of
[`harmonycloud/otel-ebpf-profiler`](https://github.com/harmonycloud/otel-ebpf-profiler/tree/kindling-preprocess)
in `collector/go.mod`.
Build `libkindling.so` first and make it available to the CGO linker and runtime
(e.g. `CGO_LDFLAGS=-L/path/to/probe/build/src` and
`LD_LIBRARY_PATH=/path/to/probe/build/src`). Existing probe dependencies include
Qt5, libelf and zlib. CMake fetches the BCC v0.30 subset and spdlog v1.14 headers;
`USE_BUNDLED_SPDLOG=OFF` uses installed spdlog headers instead.

Runtime requires permission for BPF/perf events, visibility into host PIDs and
`/proc`, readable executable mappings and kernel symbols, and access to the
Pyroscope endpoint. The agent deployment uses privileged mode and host PID
visibility. Python preprocessing reads the target process's runtime metadata;
permissions or an unsupported runtime can prevent Python frame collection.

## Validation

Run `go test ./pkg/profiling/...` from `collector/` with the newly built probe
library available. The tests cover frame decoding, aggregation, symbolization,
process mapping lifetimes, Python preprocessing, PID rosters and self-monitoring.
Live acceptance also requires a privileged host: select representative Native,
Go and Python processes, confirm named `process_cpu` profiles in Pyroscope,
inspect metrics, restart with updated PIDs, and verify clean shutdown. Unit tests and a
successful build alone do not verify BPF loading or live export.

## Code entry points

- Probe trace reader and map writer: `probe/src/core/profiling/`
- Pure C boundary: `probe/src/cgo/profiling_cgo.h` (identical collector copy)
- Receiver lifecycle: `collector/pkg/component/receiver/cgoreceiver/profiling.go`
- CPU service, preprocessing and symbolization: `collector/pkg/profiling/`
- Pyroscope sender: `collector/pkg/component/consumer/exporter/pyroscopeexporter/`
