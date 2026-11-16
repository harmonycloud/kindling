# Deployment files for Kubernetes
The files under this directory are needed when releasing the Kindling.

## agent
This directory contains files used for deploying `kindling-agent` in Kubernetes.

### eBPF CPU profiling (Pyroscope)

Optional continuous CPU stack sampling is configured under
`receivers.cgoreceiver.ebpf_profiling` in `agent/kindling-collector-config.yml`.

- Leave `endpoint.url` empty to keep the feature **off** (default).
- Set it to a Pyroscope base URL (e.g. `http://pyroscope:4040/`) to enable push of `process_cpu` profiles.

See [docs/ebpf_cpu_profiling.md](../docs/ebpf_cpu_profiling.md) for privileges,
BCC build deps, and label details. Stock `kindling-deploy.yml` (`privileged`,
`hostPID`, host `/sys`) is required for this path — same as existing eBPF tracing.

## grafana-with-plugins
This directory contains files used for deploying `Grafana` in Kubernetes. 

## recompile-probe
 The files under this directory provide a convenient way to build a new `kindling-agent` image with `drivers` **built locally**. The script compiles the driver codes and produces `drivers` based on the specific kernel version, and rebuilds a container image base on the `drivers` and the latest `kindling-agent` image. 