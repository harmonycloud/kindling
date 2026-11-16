# Changelog
### Notes
1. All notable changes to this project will be documented in this file.
2. Records in this file are not identical to the title of their Pull Requests. A detailed description is necessary for understanding what changes are and why they are made.

## v0.9.1 - 2024-02-26
### Enhancements
- Improved the garbage collection efficiency of the otelexporter component, resulting in a noticeable reduction in the CPU usage of the agent. [#623](https://github.com/KindlingProject/kindling/pull/623)

## v0.9.0 - 2024-02-02
### Enhancements
- Add periodic memory cleanup for OtelExporter. This can slightly slow down the rate of memory growth. Users can configure the restart period in hours. Disabled by deafult. ([#577](https://github.com/KindlingProject/kindling/pull/577))
- Added an extra application for collecting k8s metadata named metadata-provider,with an API that list/watch k8s metadata. More Detail at [readme of metaprovider](collector/pkg/metadata/metaprovider/readme.md) ([#580](https://github.com/KindlingProject/kindling/pull/580),[#595](https://github.com/KindlingProject/kindling/pull/595),[#596](https://github.com/KindlingProject/kindling/pull/596))

### Bug fixes
- Fix the bug that the agent writes a lot of kernel logs when using kernel module as the driver. ([#590](https://github.com/KindlingProject/kindling/pull/590))
- Fix the bug that the `View Mode` and `Show Services` in the topology panel are not displayed. ([#579](https://github.com/KindlingProject/kindling/pull/579))

## v0.8.1 - 2023-09-01
### Enhancements
- Improve the Grafana plugin's performance by reducing the amount of data requiring queries. Now the plugin queries through Grafana's api proxy. ([#555](https://github.com/KindlingProject/kindling/pull/555))
- Expand the histogram bucket of otelexpoerter (Add 1500ms). ([#563](https://github.com/KindlingProject/kindling/pull/563))
- Set default values of `store_external_src_ip` and `StoreExternalSrcIP` to false to reduce occurrences of unexpected src IP data. ([#562](https://github.com/KindlingProject/kindling/pull/562))
- Increase maximum throughput capacity for event handling. Optimized the `networkanalyzer` component of the probe analyzer by utilizing Go's goroutines, enabling concurrent execution. ([#558](https://github.com/KindlingProject/kindling/pull/558))
- Add a new configuration option ignore_dns_rcode3_error to allow users to specify whether DNS responses with RCODE 3 should be treated as errors. ([#566](https://github.com/KindlingProject/kindling/pull/566))
- Improve event processing efficiency with batch event retrieval in cgo. ([#560](https://github.com/KindlingProject/kindling/pull/560))
- Reduce the data volume of the `kindling_k8s_workload_info` metric by having each agent only send workloads present on its own node.([#554](https://github.com/KindlingProject/kindling/pull/554))
- Provide a new self metric for probe events, including the count of skipped and dropped events. ([#553](https://github.com/KindlingProject/kindling/pull/553))

### Bug fixes
- Fix the bug where DNS resolution would fail when UDP packets were received out of order. ([#565](https://github.com/KindlingProject/kindling/pull/565))
- Add periodic cleanup of javatraces data to prevent continuous memory growth when trace-profiling is enabled. ([#514](https://github.com/KindlingProject/kindling/pull/514))

In this release, we have a new contributor @YDMsama. Thanks and welcome! 🥳

## v0.8.0 - 2023-06-30
### New features
- Provide a new metric called kindling_k8s_workload_info, which supports workload filtering for k8s, thus preventing frequent crashes of Grafana topology. Please refer to the [doc](http://kindling.harmonycloud.cn/docs/usage/grafana-topology-plugin/) for any limitations.([#530](https://github.com/KindlingProject/kindling/pull/530))
- Added support for displaying trace-profiling data by querying from Elasticsearch. ([#528](https://github.com/KindlingProject/kindling/pull/528))
- Display scheduler run queue latency on Trace-Profiling chart. To learn more about the concept of 'Run Queue Latency', refer to [this blog post](https://www.brendangregg.com/blog/2016-10-08/linux-bcc-runqlat.html). You can also find a use case for this feature in [this blog post](http://kindling.harmonycloud.cn/blogs/use-cases/optimize-cpu/). ([#494](https://github.com/KindlingProject/kindling/pull/494))
### Enhancements
- Upgrade the Grafana version to 8.5.26 ([#533](https://github.com/KindlingProject/kindling/pull/533))
- MySQL CommandLine Case: Ignore quit command and get sql with CLIENT_QUERY_ATTRIBUTES([#523](https://github.com/KindlingProject/kindling/pull/523))
- ⚠️Breaking change: Refactor the data format of on/off CPU events from "string" to "array". Note that the old data format cannot be parsed using the new version of the front-end.([#512](https://github.com/KindlingProject/kindling/pull/512) [#520](https://github.com/KindlingProject/kindling/pull/520))

### Bug fixes
- Fix the bug where the DNS domain is not obtained when DNS transport over TCP. ([#524](https://github.com/KindlingProject/kindling/pull/524))
- Fix panic: send on closed channel. ([#519](https://github.com/KindlingProject/kindling/pull/519))
- Fix the bug that the event detail panel doesn't hide when switching profiles.([#513](https://github.com/KindlingProject/kindling/pull/513))
- Fix span data deduplication issue.([#511](https://github.com/KindlingProject/kindling/pull/511))

In this release, we have a new contributor @hwz779866221. Thanks and welcome! 🥳

## v0.7.2 - 2023-04-24
### Enhancements
- Add an option `WithMemory` to OpenTelemetry's Prometheus exporter. It allows users to control whether metrics that haven't been updated in the most recent interval are reported. ([#501](https://github.com/KindlingProject/kindling/pull/501))
- Add a config to cgoreceiver for suppressing events according to processes' comm ([#495](https://github.com/KindlingProject/kindling/pull/495))
- Add `bind` syscall support to get the listening ip and port of a server. ([#493](https://github.com/KindlingProject/kindling/pull/493))
- Add an option `enable_fetch_replicaset` to control whether to fetch ReplicaSet metadata. The default value is false which aims to release pressure on Kubernetes API server. ([#492](https://github.com/KindlingProject/kindling/pull/492))

### Bug fixes
- Fix the memory leak issue by deleting vtid-tid map to avoid OOM. ([#499](https://github.com/KindlingProject/kindling/pull/499))
- Fix unrunnable bug due to the error Insufficient parameters of TCP retransmit. ([#499](https://github.com/KindlingProject/kindling/pull/499))
- Fix the bug that in `cpuanalyzer`, no segments are sent if they contain no cpuevents. Now segments are sent as long as they contain events, regardless of what the events are. ([#502](https://github.com/KindlingProject/kindling/pull/502))
- Fix the bug that the default configs of slice/map are not overridden. ([#497](https://github.com/KindlingProject/kindling/pull/497))

Thanks for the significant help of @yanhongchang to provide OOM-killed information on #499.

## v0.7.1 - 2023-03-01
### New features
- Support trace-profiling sampling to reduce data output. One trace is sampled every five seconds for each endpoint by default. ([#446](https://github.com/KindlingProject/kindling/pull/446)[#462](https://github.com/KindlingProject/kindling/pull/462))

### Enhancements
- **Upgrade the golang version to v1.19 in the requirement**. ([#463](https://github.com/KindlingProject/kindling/pull/463))
- Improve Kindling Event log format. ([#455](https://github.com/KindlingProject/kindling/pull/455))

### Bug fixes
- Fix security alerts(CVE-2022-41721, CVE-2022-27664) by upgrading package `golang.org/x/net`.([#463](https://github.com/KindlingProject/kindling/pull/463))
- Fix the potential endless loop in the rocketmq parser. ([#465](https://github.com/KindlingProject/kindling/pull/465))
- Fix retransmission count is not consistent with the real value on Linux 4.7 or higher. ([#450](https://github.com/KindlingProject/kindling/pull/450))
- Reduce the cases pods are not found when they are daemonset. ([#439](https://github.com/KindlingProject/kindling/pull/439) @llhhbc)
- Collector subscribes `sendmmsg` events to fix the bug that some DNS requests are missed. ([#430](https://github.com/KindlingProject/kindling/pull/430))
- Fix the bug that the agent panics when it receives DeletedFinalStateUnknown by watching K8s metadata. ([#456](https://github.com/KindlingProject/kindling/pull/456))

In this release, we have a new contributor @llhhbc. Thanks and welcome! 🥳

## v0.7.0 - 2023-02-16
### New features
- Add a new simplified chart to display the trace-profiling data. It mixes `span` with profiling and is more user-friendly. Try the demo now on the [website](http://kindling.harmonycloud.cn/).([#443](https://github.com/KindlingProject/kindling/pull/443))
- Add trace to cpuevents to display the payload of network flows. ([#442](https://github.com/KindlingProject/kindling/pull/442))
- Support Attach Agent for NoAPM Java Application. ([#431](https://github.com/KindlingProject/kindling/pull/431))

### Enhancements
- Add an option edge_events_window_size to allow users to reduce the size of the files by narrowing the time window where seats the edge events. ([#437](https://github.com/KindlingProject/kindling/pull/437))
- Rename the camera profiling file to make the timestamp of the profiling files readable. ([#434](https://github.com/KindlingProject/kindling/pull/434))
- When using the file writer in `cameraexporter`, we rotate files in chronological order now and rotate half of files one time. ([#420](https://github.com/KindlingProject/kindling/pull/420))
- Support to identify the MySQL protocol with statements `commit` and `set`. ([#417](https://github.com/KindlingProject/kindling/pull/417))

### Bug fixes
- Fix the bug that TCP metrics are not aggregated correctly. ([#444](https://github.com/KindlingProject/kindling/pull/444))
- Fix the bug that cpuanalyzer missed some trigger events due to the incorrect variable reference. This may cause some traces can't correlate with on/off CPU data. ([#424](https://github.com/KindlingProject/kindling/pull/424))

## v0.6.0 - 2022-12-21
### New features
- Support to configure `snaplen` through startup args.([#387](https://github.com/KindlingProject/kindling/pull/387))
- Add tracing span data in cpu events. ([#384](https://github.com/KindlingProject/kindling/pull/384))
- Add a new tool: A debug tool for Trace Profiling is provided for developers to troubleshoot problems.([#363](https://github.com/KindlingProject/kindling/pull/363))
- Support the protocol RocketMQ.([#328](https://github.com/KindlingProject/kindling/pull/328))

### Enhancements
- Add self-monitor tool: include kernel event log and gdb information of exit.([#398](https://github.com/KindlingProject/kindling/pull/398))
- Adjust max depth of stack trace to 20. ([#399](https://github.com/KindlingProject/kindling/pull/399))
- Add the field `end_timestamp` to the trace data to make it easier for querying. ([#380](https://github.com/KindlingProject/kindling/pull/380))
- Add `request_tid` and `response_tid` for trace labels.([#379](https://github.com/KindlingProject/kindling/pull/379))
- Add no_response_threshold(120s) for No response requests. ([#376](https://github.com/KindlingProject/kindling/pull/376))
- Add payload for all protocols.([#375](https://github.com/KindlingProject/kindling/pull/375))
- Add a new clustering method "blank" that is used to reduce the cardinality of metrics as much as possible. ([#372](https://github.com/KindlingProject/kindling/pull/372))
- Modify the configuration file structure and add parameter fields for subscription events. ([#368](https://github.com/KindlingProject/kindling/pull/368))


### Bug fixes
- Add the missing timestamp of TCP connect data and filter the incorrect one without srcPort.([#405](https://github.com/KindlingProject/kindling/pull/405))
- Fix the bug that multiple events cannot be correlated when they are in one ON-CPU data. ([#395](https://github.com/KindlingProject/kindling/pull/395))
- Add the missed latency field for `cgoEvent` to fix the bug where the `request_sent_time` in `single_net_request_metric_group` is always 0. ([#394](https://github.com/KindlingProject/kindling/pull/394))
- Fix http-100 request is detected as NOSUPPORT([#393](https://github.com/KindlingProject/kindling/pull/393))
- Fix the wrong thread name in the trace profiling function. ([#385](https://github.com/KindlingProject/kindling/pull/385))
- Remove "reset" method of ScheduledTaskRoutine to fix a potential dead-lock issue. ([#369](https://github.com/KindlingProject/kindling/pull/369))
- Fix the bug where the pod metadata with persistent IP in the map is deleted incorrectly due to the deleting mechanism with a delay. ([#374](https://github.com/KindlingProject/kindling/pull/374))
- Fix the bug that when the response is nil, the NAT IP and port are not added to the labels of the "DataGroup". ([#378](https://github.com/KindlingProject/kindling/pull/378))
- Fix potential deadlock of exited thread delay queue. ([#373](https://github.com/KindlingProject/kindling/pull/373))
- Fix the bug that cpuEvent cache size continuously increases even if trace profiling is not enabled.([#362](https://github.com/KindlingProject/kindling/pull/362))
- Fix the bug that duplicate CPU events are indexed into Elasticsearch. ([#359](https://github.com/KindlingProject/kindling/pull/359))
- Implement the delay queue for exited thread, so as to avoid losing the data in the period before the thread exits. ([#365](https://github.com/KindlingProject/kindling/pull/365))
- Fix the bug of incomplete records when threads arrive at the cpu analyzer for the first time. ([#364](https://github.com/KindlingProject/kindling/pull/364))

## v0.5.0 - 2022-11-02
### New features
- Add a new feature: Trace Profiling. See more details about it on our [website](http://kindling.harmonycloud.cn). ([#335](https://github.com/KindlingProject/kindling/pull/335))

### Enhancements
- Add request and response payload of `Redis` protocol message to `Span` data. ([#325](https://github.com/KindlingProject/kindling/pull/325))

### Bug fixes
- Fix the topology node naming error in the default namespace.([#346](https://github.com/KindlingProject/kindling/pull/346))
- Fix the bug that if `ReadBytes` receives negative numbers as arguments, the program panics with the error of slice outofbound. ([#327](https://github.com/KindlingProject/kindling/pull/327))

## v0.4.1 - 2022-09-21
### Enhancements
- When processing Redis' Requests, add additional labels to describe the key information of the message. Check [Metrics Document](https://github.com/KindlingProject/kindling/blob/main/docs/prometheus_metrics.md) for more details. ([#321](https://github.com/KindlingProject/kindling/pull/321))

### Bug fixes
- Fix the bug when the kernel does not support some kprobe, the probe crashes. ([#320](https://github.com/KindlingProject/kindling/pull/320))

## v0.4.0 - 2022-09-19
### Enhancements
- Optimize the log output. ([#299](https://github.com/KindlingProject/kindling/pull/299))
- Print logs when subscribing to events. Print a warning message if there is no event the agent subscribes to. ([#290](https://github.com/KindlingProject/kindling/pull/290))
- Allow the collector run in the non-Kubernetes environment by setting the option `enable` `false` under the `k8smetadataprocessor` section. ([#285](https://github.com/KindlingProject/kindling/pull/285))
- Add a new environment variable: IS_PRINT_EVENT. When the value is true, sinsp events can be printed to the stdout. ([#283](https://github.com/KindlingProject/kindling/pull/283))
- Declare the 9500 port in the agent's deployment file ([#282](https://github.com/KindlingProject/kindling/pull/282))

### Bug fixes
- Avoid printing logs to console when both `observability.logger.file_level` and `observability.logger.console_level` are set to none([#316](https://github.com/KindlingProject/kindling/pull/316))
- Fix the userAttributes array out of range error caused by userAttNumber exceeding 8
- Fix the bug where no HTTP headers were got. ([#301](https://github.com/KindlingProject/kindling/pull/301))
- Fix the bug that need_trace_as_span options cannot take effect ([#292](https://github.com/KindlingProject/kindling/pull/292))
- Fix connection failure rate data lost when change topology layout in the Grafana plugin. ([#289](https://github.com/KindlingProject/kindling/pull/289))
- Fix the bug that the external topologys' metric name is named with `kindling_entity_request` prefix. Change the prefix of these metrics to `kindling_topology_request` ([#287](https://github.com/KindlingProject/kindling/pull/287))
- Fix the bug where the table name of SQL is missed if there is no trailing character at the end of the table name. ([#284](https://github.com/KindlingProject/kindling/pull/284))

## v0.3.0 - 2022-06-29
### New features
- Add an option name `debug_selector` to filter debug_log from different components ([#300](https://github.com/KindlingProject/kindling/pull/300))
- Add a URL clustering method to reduce the cardinality of the entity metrics. Configuration options are provided to choose which method to use. ([#268](https://github.com/KindlingProject/kindling/pull/268)) 
- Display connection failure metrics in the Grafana-plugin ([#255](https://github.com/KindlingProject/kindling/pull/255)) 
- Add the metrics that describe how many times the TCP connections have been made ([#234](https://github.com/KindlingProject/kindling/pull/234) [#235](https://github.com/KindlingProject/kindling/pull/235) [#236](https://github.com/KindlingProject/kindling/pull/236) [#237](https://github.com/KindlingProject/kindling/pull/237))
- Add a histogram aggregator in defaultAggregator ([#226](https://github.com/KindlingProject/kindling/pull/226))
- (Experimental) Support Protocol Dubbo2 ([#184](https://github.com/KindlingProject/kindling/pull/184)) 

### Enhancements
- Improve the go project layout ([#273](https://github.com/KindlingProject/kindling/pull/273))
- Correct the configurations and disable the `dubbo` protocol parser by default since it is still experimental now. ([#270](https://github.com/KindlingProject/kindling/pull/270))
- Implement self-metrics using opentelemetry for cgoreceiver ([#269](https://github.com/KindlingProject/kindling/pull/269))
- Use cgo to replace UDS for transferring data from the probe to the collector to improve the performance ([#264](https://github.com/KindlingProject/kindling/pull/264))
- Add command labels in tcp connect metrics and span attributes ([#260](https://github.com/KindlingProject/kindling/pull/260))
- Use the tcp_close events to generate the srtt metric ([#256](https://github.com/KindlingProject/kindling/pull/256))
- Remove the histogram metrics by default to reduce the number of metrics ([#253](https://github.com/KindlingProject/kindling/pull/253)) 
- k8sprocessor: use src IP for further searching if the dst IP is a loopback address ([#251](https://github.com/KindlingProject/kindling/pull/251))
- docs:update developer links ([#247](https://github.com/KindlingProject/kindling/pull/247)) 
- Add some self metrics for agent cpu and memory usage ([#243](https://github.com/KindlingProject/kindling/pull/243))
- Export the trace of MySQL request when it contains an error ([#241](https://github.com/KindlingProject/kindling/pull/241))
- Block in the application instead of the udsreceiver after running ([#240](https://github.com/KindlingProject/kindling/pull/240)) 
- Decouple the logic of dispatching events from receivers ([#232](https://github.com/KindlingProject/kindling/pull/232)) 
- Search for k8s metadata using `src_ip` when no containerid found ([#233](https://github.com/KindlingProject/kindling/pull/233))
- Record the containers with `hostport` mode and fill the pod information of them in k8sprocessor ([#219](https://github.com/KindlingProject/kindling/pull/219))
- Support building Grafana-plugin by using Actions ([#218](https://github.com/KindlingProject/kindling/pull/218))
- Improve metrics description doc ([#216](https://github.com/KindlingProject/kindling/pull/216)) 
- Update deployment files needed for releasing ([#215](https://github.com/KindlingProject/kindling/pull/215)) 

### Bug fixes 
- docs: fix language issues in documents ([#258](https://github.com/KindlingProject/kindling/pull/258))
- Fix the bug where the pod information is missed after it is restarted ([#245](https://github.com/KindlingProject/kindling/pull/245))
- Grafana-plugin: delete yarn.lock to remove unnecessary dependencies ([#244](https://github.com/KindlingProject/kindling/pull/244)) 
- Fix the bug that the container name is incorrect when multiple containers in the pod don't specify ports by setting it empty. ([#238](https://github.com/KindlingProject/kindling/pull/238))
- Fix the bug that sometimes the workload kind is `ReplicaSet` ([#230](https://github.com/KindlingProject/kindling/pull/230)) 
- Fix "no such file or directory" when using the kubeconfig file. [#225](https://github.com/KindlingProject/kindling/pull/225)
- Fix several bugs in the Grafana plugin. ([#220](https://github.com/KindlingProject/kindling/pull/220))

## v0.2.0 - 2022-05-07
### Features
- Provide a kindling Prometheus exporter that can support integration with Prometheus easily. See kindling's metrics from the kindling [website](http://kindling.harmonycloud.cn/docs/usage/prometheus-metric/).
- Support network performance, DNS performance, service network maps, and workload performance analysis.
- Support HTTP, MySQL, and REDIS request analysis.
- Provide a Grafana-plugin with four built-in dashboards to support basic analysis features.



