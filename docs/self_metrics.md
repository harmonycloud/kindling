# Self Observability Metrics Description
## cgoreceiver
### kindling_telemetry_cgoreceiver_events_total
- Description: The total number of the events received by cgoreceiver. 
- Metric Type: counter
- Unit: count
- Labels: Additional labels except [the common ones](#common-labels).


| **Label Name** | **Description**        | **Example** |
|----------------|------------------------|-------------|
| name           | The name of the event. | write       |


### kindling_telemetry_cgoreceiver_channel_size
- Description: The current number of events contained in the channel. Cgoreceiver uses a channel to receive events from `cgo`. This channel is able to accommodate a maximum size of 300,000 events. No events can be received if the channel is full.
- Metric Type: Gauge
- Unit: count
- Labels: No other labels except [the common ones](#common-labels).

## networkanalyzer
### kindling_telemetry_netanalyer_messagepair_size
- Description: The size of the message pairs stored in the map. Message pairs are the middle data structure of "traces". This metric is used to identify how many "traces" have not finished yet.
- Metric Type: Gauge
- Unit: count
- Labels: Additional labels except [the common ones](#common-labels).


| **Label Name** | **Description**                               | **Example** |
|----------------|-----------------------------------------------|-------------|
| type           | The type of the message pair. `tcp` or `udp`. | tcp         |

### kindling_telemetry_netanalyer_parsedrequest_total
- Description: The count of traces that the agent has processed.
- Metric Type: counter
- Unit: count
- Labels: Additional labels except [the common ones](#common-labels).


| **Label Name** | **Description**               | **Example** |
|----------------|-------------------------------|-------------|
| protocol       | The protocol of the requests. | http        |


## tcpconnectanalyzer
### kindling_telemetry_tcpconnectanalyzer_map_size
- Description: The current number of the connections stored in the map. This map accomodates the events related to the metric "TCP connect".
- Metric Type: gauge
- Unit: count
- Labels: No other labels except [the common ones](#common-labels).


## conntracker
### kindling_telemetry_conntracker_cache_size
- Description: The current number of the conntrack records stored in the map.
- Metric Type: gauge
- Unit: count
- Labels: Additional labels except [the common ones](#common-labels).


| **Label Name** | **Description**                                | **Example** |
|----------------|------------------------------------------------|-------------|
| type           | The type of the records. `general` or `orphan` | general     |


### kindling_telemetry_conntracker_cache_max_size
- Description: The maximum size of the cache map. The default value is 130,000. It can be configured in the configuration file.
- Metric Type: gauge
- Unit: count
- Labels: No other labels except [the common ones](#common-labels).

### kindling_telemetry_conntracker_operation_times_total
- Description: The total operation times the conntracker does to the cache map. This metric can reflect the load of the conntracker module.
- Metric Type: counter
- Unit: count
- Labels: Additional labels except [the common ones](#common-labels).


| **Label Name** | **Description**                                                           | **Example** |
|----------------|---------------------------------------------------------------------------|-------------|
| op             | The opreation names. Could be `add`, `drop`, `remove`, `get`, or `evict`. | add         |


### kindling_telemetry_conntracker_errors_total
- Description: The total count of errors the conntracker encounters. This metric can reflect the load of the conntracker module. In most cases, the error type is `enobuf` that means there are too many records the conntracker generates and there is no buffer to receive them.
- Metric Type: counter
- Unit: count
- Labels: Additional labels except [the common ones](#common-labels).

| **Label Name** | **Description**                                                     | **Example** |
|----------------|---------------------------------------------------------------------|-------------|
| type           | The error types. Could be `enobuf`, `read_errors`, or `msg_errors`. | enobuf      |


### kindling_telemetry_conntracker_sampling_rate
- Description: The sampling rate of the conntracker module. This rate may be automatically decreased if the load is too high.
- Metric Type: counter
- Unit: percent
- Labels: No other labels except [the common ones](#common-labels).


### kindling_telemetry_conntracker_throttles_total
- Description: The total count of the records being throttled due to the high load.
- Metric Type: counter
- Unit: count
- Labels: No other labels except [the common ones](#common-labels).


## otelexporter
### kindling_telemetry_otelexporter_metricgroups_received_total
- Description: The total count of the data received by `otelexporter`.
- Metric Type: counter
- Unit: count
- Labels: Additional labels except [the common ones](#common-labels).

| **Label Name** | **Description**              | **Example**                     |
|----------------|------------------------------|---------------------------------|
| name           | The name of the `DataGroup`. | single_net_request_metric_group |


### kindling_telemetry_otelexporter_cardinality_size
- Deprecated.

## Common labels
| **Label Name**       | **Description**                                                    | **Example**      |
|----------------------|--------------------------------------------------------------------|------------------|
| service.instance.id  | The host name where the agent locates in.                          | worker-149       |
| service.name         | The cluster name which is composed of "kindling" and "cluster ID". | kindling-abcd123 |
| instrumentation.name | A constant "kindling".                                             | kindling         |