  # eBPF CPU Profiling → Pyroscope

Kindling can sample CPU call stacks with eBPF (~97Hz), symbolize them with a
minimal BCC symbolizer baked into `libkindling`, and push gzip pprof profiles to
[Grafana Pyroscope](https://grafana.com/oss/pyroscope/) every **15 seconds**.

This path is **independent** of the existing Camera On/OffCPU flame graph
(`cpuanalyzer` → `cameraexporter` → `camera-front`).

## Enable / disable

Edit the collector ConfigMap / config file:

- `deploy/agent/kindling-collector-config.yml`
- `collector/docker/kindling-collector-config.yml`

```yaml
receivers:
  cgoreceiver:
    ebpf_profiling:
      endpoint:
        # Empty = disabled (default). Non-empty = start sampling + push.
        url: "http://pyroscope:4040/"
      # Empty / omit = sample all non-kernel processes (CAPTURE_ALL).
      # Non-empty = only these host PIDs (whitelist + PROFILING_ENABLED).
      pids: []
      # Example: pids: [1234, 5678]
```

Apply via the usual agent ConfigMap (`kindlingcfg`) and restart the agent pod.

**Sampling scope** (static, read at startup only):

| `pids` | Behavior |
|--------|----------|
| omitted / `[]` | `CAPTURE_ALL` — all non-kernel threads |
| `[pid, ...]` | whitelist — only those host PIDs |

There is no runtime hot-reload of the PID list; change YAML and restart the agent.

## What gets pushed

| Item | Value |
|------|--------|
| Profile name | `process_cpu` (`__name__`) |
| Labels | `service_name=kindling`, `host_ip` (`MY_NODE_IP`), `pid` |
| Interval | 15s (hardcoded in `cgoreceiver.Start`) |
| Sample rate | 97Hz |

## Runtime requirements

The stock `deploy/agent/kindling-deploy.yml` already satisfies these:

- `privileged: true` (eBPF + `perf_event_open`)
- `hostPID: true` / host `/proc` mounts (user-space symbolization)
- host `/sys` mount; `kindling-probe-loader` mounts debugfs under `/sys/kernel/debug` if missing

Without those, CPU stack sampling will not work even if `url` is set.

## Probe build notes (BCC symbolizer)

CPU symbolization uses a **vendored BCC v0.30 subset** (`bcc_syms_minimal`), not
a system `libbcc` package.

| Piece | Path |
|-------|------|
| Fetch/build | `probe/cmake/modules/bcc_minimal.cmake` |
| Sources | `probe/src/core/symbolizer/` |
| Link into | `libkindling.so` (`probe/src/CMakeLists.txt`) |

Build-host packages (in addition to existing probe deps such as Qt5 / cmake):

- `elfutils-devel` / `libelf-dev`
- `zlib-devel` / `zlib1g-dev`
- network access at **cmake configure** time (FetchContent downloads BCC v0.30.0)

The agent runtime image only needs the prebuilt `libkindling.so`; BCC is linked in statically.

## Acceptance checklist

| Check | How |
|-------|-----|
| Default off | Leave `endpoint.url` empty; agent logs skip / does not start CPU sampling |
| Push works | Set `url` to Pyroscope base (e.g. `http://<host>:4040/`); after ~15s UI shows `process_cpu` |
| Symbols | Flame graph has function names or at least `module+offset` frames |
| Stop cleanly | Stop agent; sampling stops (no stuck perf attachments / growing maps) |

Full eBPF→Pyroscope verification needs a privileged agent (see Runtime requirements) plus a reachable Pyroscope.

## Related code

- Collector loop: `collector/pkg/component/receiver/cgoreceiver/profiling.go`
- Sender: `collector/pkg/component/consumer/exporter/pyroscopeexporter/simple_sender.go`
- pprof builder: `collector/pkg/ebpf/pprof/`
