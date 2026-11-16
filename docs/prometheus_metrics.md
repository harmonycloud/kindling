# Prometheus Metrics Description
## Service Metrics
Service metrics are generated from the server-side events, which are used to show the quality of service. 
### Metrics List
| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| `kindling_entity_request_total` | Counter | Total number of requests |
| `kindling_entity_request_duration_nanoseconds_total` | Counter | Total duration of requests |
| `kindling_entity_request_send_bytes_total` | Counter | Total size of payload sent |
| `kindling_entity_request_receive_bytes_total` | Counter | Total size of payload received |
| `kindling_entity_request_average_duration_nanoseconds_count` | Histogram | Count of average duration of requests <br> **Disabled by default. See Note 3 for how to enable it.**|
| `kindling_entity_request_average_duration_nanoseconds_sum` | Histogram | Sum of average duration of requests <br> **Disabled by default. See Note 3 for how to enable it.**|
| `kindling_entity_request_average_duration_nanoseconds_bucket` | Histogram | Histogram buckets of average duration of requests <br> **Disabled by default. See Note 3 for how to enable it.**|
### Labels List
| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| `node` | worker-1 | Node name represented in Kubernetes cluster |
| `namespace` | default | Namespace of the pod |
| `workload_kind` | daemonset | K8sResourceType |
| `workload_name` | api-ds | K8sResourceName |
| `service` | api | One of the services that target this pod |
| `pod` | api-ds-xxxx | The name of the pod |
| `container` | api-container | The name of the container |
| `container_id` | 1a2b3c4d5e6f | The shorten container id which contains 12 characters |
| `ip` | 10.1.11.23 | The IP address of the entity |
| `port` | 80 | The listening port of the entity |
| `protocol` | http | The application layer protocol the requests use |
| `request_content` | /test/api | The request content of the requests |
| `response_content` | 200 | The response content of the requests |
| `is_slow` | false | (Only applicable to `kindling_entity_request_total`)<br>Whether the requests are considered as slow |
### Notes
**Note 1**: The label `namespace` holds a value `NOT_FOUND_INTERNAL` when the `container_id` and the IP can't be found in the current Kubernetes cluster, in which case the entity isn't maintained by the current Kubernetes.

**Note 2**: The labels `request_content` and `response_content` hold different values when `protocol` is different.

- When protocol is `http`:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| `request_content` | /test/api | Endpoint of HTTP request. URL has been truncated to avoid high-cardinality. |
| `response_content` | 200 | 'Status Code' of HTTP response. |

- When protocol is `dns`:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| `request_content` | www.google.com | Domain to be queried |
| `response_content` | 0 | "rcode" of DNS response. Including 0, 1, 2, 3, 4 |

- When protocol is `mysql`:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| `request_content` | select employee | SQL of MySQL. SQL has been truncated to avoid high-cardinality. The format is ['operation' 'space' 'table' '*']. |
| `response_content` | 1064 | Error code of MySQL. Only applicable when the response is in error type. See [codes introduction](https://dev.mysql.com/doc/mysql-errors/5.7/en/error-reference-introduction.html).|

- When protocol is `kafka`:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| `request_content` | user-msg-topic | Topic of Kafka request. |
| `response_content` |  | Empty temporarily. |

- When protocol is `dubbo`:

| **Label** | **Example**                   | **Notes**                           |
| --- |-------------------------|--------------------------|
| `request_content` | io.kindling.dubbo.api.service.OrderService#order | Service Info. The format of service is `package.class#method`                                                                                            |
| `response_content` | 20 | "error_code" of Dubbo. 20 means OK, more details at the [docs](https://dubbo.apache.org/en/blog/2018/10/05/introduction-to-the-dubbo-protocol/#dubbo-protocol-details). |

- When protocol is `redis`:

| **Label** | **Example**                   | **Notes**                           |
| --- |-------------------------|--------------------------|
| `request_content` | GET | The command of the Redis request. |
| `response_content` | noerror | The value is either `error` or `noerror`. |

- When protocol is `rocketmq`:

| **Label** | **Example** | **Notes**                                                         |
| --- |-------------|-------------------------------------------------------------------|
| `request_content` | TopicTest   | Topic of RocketMQ request.                                        |
| `response_content` | 0           | response code of RocketMQ. 0 means OK, others mean Error [docs](https://github.com/apache/rocketmq/blob/fcfe26e4443dd24b1055899266d1bd81060ee118/common/src/main/java/org/apache/rocketmq/common/protocol/ResponseCode.java) |

- For other cases, the `request_content` and `response_content` are both empty.

**Note 3**: The histogram metric `kindling_entity_request_average_duration_nanoseconds_*` is disabled by default as it could be high-cardinality. If this metric is needed, please add a new line to the `exporters.otelexporter.metric_aggregation_map` section of the configuration file.
```yaml
exporters:
  otelexporter:
    metric_aggregation_map:
      # add the following line
      kindling_entity_request_average_duration_nanoseconds: histogram 
```

## Topology Metrics

Topology metrics are typically generated from the client-side events, which are used to show the service dependencies map, so the metrics are called "topology". Some timeseries may be generated from the server-side events, which contain a non-empty label `dst_container_id`. These timeseries are generated only when the source IP is not the pod's IP inside the Kubernetes cluster, which are useful when there is no agent installed on the client-side. 

### Metrics List
| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| `kindling_topology_request_total` | Counter | Total number of requests |
| `kindling_topology_request_duration_nanoseconds_total` | Counter |  Total duration of requests |
| `kindling_topology_request_request_bytes_total` | Counter | Total size of payload sent |
| `kindling_topology_request_response_bytes_total` | Counter | Total size of payload received |
| `kindling_topology_request_average_duration_nanoseconds_count` | Histogram | Count of average duration of requests<br> **Disabled by default. See Note 3 for how to enable it.** |​
| `kindling_topology_request_average_duration_nanoseconds_sum` | Histogram | Sum of average duration of requests<br> **Disabled by default. See Note 3 for how to enable it.** |
| `kindling_topology_request_average_duration_nanoseconds_bucket` | Histogram | Histogram buckets of average duration of requests<br> **Disabled by default. See Note 3 for how to enable it.** |

### Labels List
| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| `src_node` | slave-node1 | Which node the source pod is on |
| `src_namespace` | default | Namespace of the source pod |
| `src_workload_kind` | deployment | Workload kind of the source pod |
| `src_workload_name` | business1 | Workload name of the source pod |
| `src_service` | business1-svc | One of the services that target the source pod |
| `src_pod` | business1-0 | The name of the source pod |
| `src_container` | business-container | The name of the source container |
| `src_container_id` | 1a2b3c4d5e6f | The shorten container id which contains 12 characters |
| `src_ip` | 10.1.11.23 | The IP address of the source |
| `dst_node` | slave-node2 | Which node the destination pod is on |
| `dst_namespace` | default | Namespace of the destination pod |
| `dst_workload_kind` | deployment | Workload kind of the destination pod |
| `dst_workload_name` | business2 | Workload name of the destination pod |
| `dst_service` | business2-svc | One of the services that target the destination pod |
| `dst_pod` | business2-0 | The name of the destination pod |
| `dst_container` | business-container | The name of the source container |
| `dst_container_id` | 2b3c4d5e6f7e | (Only applicable to the timeseries generated from the server-side)<br>The shorten container id which contains 12 characters |
| `dst_ip` | 10.1.11.24 | The IP address of the destination |
| `dst_port` | 80 | The listening port of the destination container  |
| `protocol` | http | The application layer protocol the requests use |
| `status_code` | 200 | Different values for different protocols  |

### Notes
**Note 1**: We define two custom terms for the label `src_namespace` and `dst_namespace`, which are `NOT_FOUND_INTERNAL` and `NOT_FOUND_EXTERNAL`. The meanings are described as follows. These terms also apply to other metrics in this doc.

These two terms are composed of two parts.
1. **NOT_FOUND**: `NOT_FOUND` means the IP is neither a pod's one nor a service's one in the current Kubernetes cluster. The IP could belong to a host or an external service. 
2. **INTERNAL or EXTERNAL**: There are two cases in which `INTERNAL` will be set. The first case is when the IP belongs to a node that resides in the current Kubernetes cluster. The second case is when the `source` or `destination` is running on the same host with the kindling agent, which is generally applicable for non-Kubernetes clusters. `EXTERNAL` is set for other cases if the IP is `NOT_FOUND`. Note another Kubernetes cluster is also considered "external".

**Note 2**: The field "status_code" holds different values when "protocol" is different.

- **http**: `Status Code` of HTTP response.
- **dns**: `rcode` of DNS response.
- **mysql**: `Error Code` of the error response.
- **dubbo**: `Error Code` of Dubbo request.
- **redis**: `0` if there is no error; `1` otherwise.
- **rocketmq**: `Response Code` of RocketMQ response.
- **others**: empty temporarily.

**Note 3**: The histogram metric `kindling_topology_request_average_duration_nanoseconds_*` is disabled by default as it could be high-cardinality. If this metric is needed, please add a new line to the `exporters.otelexporter.metric_aggregation_map` section of the configuration file.
```yaml
exporters:
  otelexporter:
    metric_aggregation_map:
      # add the following line
      kindling_topology_request_average_duration_nanoseconds: histogram 
```
## Trace As Metric
We made some rules for considering whether a request is abnormal. For the abnormal request, the detail request information is considered as useful for debugging or profiling. We name this kind of data "trace". It is not a good practice to store such data in Prometheus as some labels are high-cardinality, so we picked up some labels from the original ones to generate a new kind of metric, which is called "Trace As Metric". The following table shows what labels this metric contains.  

### Metrics List
| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| `kindling_trace_request_duration_nanoseconds` | Gauge | The specific request duration |

### Labels List
| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| `src_node` | slave-node1 | Which node the source pod is on |
| `src_namespace` | default | Namespace of the source pod |
| `src_workload_kind` | deployment | Workload kind of the source pod |
| `src_workload_name` | business1 | Workload name of the source pod |
| `src_service` | business1-svc | One of the services that target the source pod |
| `src_pod` | business1-0 | The name of the source pod |
| `src_container` | business-container | The name of the source container |
| `src_container_id` | 1a2b3c4d5e6f | (Only applicable when is_server is false)<br>The shorten container id which contains 12 characters |
| `src_ip` | 10.1.11.23 | The IP address of the source |
| `dst_node` | slave-node2 | Which node the destination pod is on |
| `dst_namespace` | default | Namespace of the destination pod |
| `dst_workload_kind` | deployment | Workload kind of the destination pod |
| `dst_workload_name` | business2 | Workload name of the destination pod |
| `dst_service` | business2-svc | One of the services that target the destination pod |
| `dst_pod` | business2-0 | The name of the destination pod |
| `dst_container` | business-container | The name of the destination container |
| `dst_container_id` | 2b3c4d5e6f7e | (Only applicable when is_server is true)<br>The shorten container id which contains 12 characters |
| `dst_ip` | 10.1.11.24 | The IP address of the destination. This is the original IP before DNAT |
| `dst_port` | 80 | The listening port of the destination container |
| `dnat_ip` | 192.168.12.3 | The IP address of the destination after DNAT if applicable |
| `dnat_port` | 80 | The listening port of the destination container after DNAT if applicable |
| `protocol` | http | The application layer protocol the requests use |
| `is_server` | true | True if the data is from the server-side, false otherwise |
| `request_content` | /test/api | Different values when protocol is different. Refer to service metric |
| `response_content` | 200 | Different values when protocol is different. Refer to service metric |
| `request_duration_status` | 1 | The total duration spent for sending request and receiving response.<br>1(green): latency <= 800ms<br>2(yellow): 800<latency<1500<br>3(red): latency >= 1500 |
| `request_reqxfer_status` | 2 |  ReqXfe indicates the duration for transferring request payload. <br>1(green): latency <= 200ms<br>2(yellow): 200<latency<1000<br>3(red): latency >= 1000 |
| `request_processing_status` | 3 | Processing indicates the duration until receiving the first byte. <br>1(green): latency <= 200ms<br>2(yellow): 200<latency<1000<br>3(red): latency >= 1000 |
| `response_rspxfer_status` | 1 | RspXfer indicates the duration for transferring response bopayloaddy.<br>1(green): latency <= 200ms<br>2(yellow): 200<latency<1000<br>3(red): latency >= 1000 |

## TCP Status Metrics

### Metrics List
| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| `kindling_tcp_srtt_microseconds` | Gauge | Smoothed round trip time of the TCP socket |
| `kindling_tcp_packet_loss_total` | Counter | Total number of dropped packets |
| `kindling_tcp_retransmit_total` | Counter | Total number of resending segments  |

### Labels List
| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| `src_node` | slave-node1 | Which node the source pod is on |
| `src_namespace` | default | Namespace of the source pod |
| `src_workload_kind` | deployment | Workload kind of the source pod |
| `src_workload_name` | business1 | Workload name of the source pod |
| `src_service` | business1-svc | One of the services that target the source pod |
| `src_pod` | business1-0 | The name of the source pod |
| `src_container` | business-container | The name of the source container |
| `src_ip` | 10.1.11.23 | Pod's IP by default. If the source is not a pod in Kubernetes, this is the IP address of an external entity |
| `src_port` | 80 | The listening port of the source container, if applicable  |
| `dst_node` | slave-node2 | Which node the destination pod is on |
| `dst_namespace` | default | Namespace of the destination pod |
| `dst_workload_kind` | deployment | Workload kind of the destination pod |
| `dst_workload_name` | business2 | Workload name of the destination pod |
| `dst_service` | business2-svc | One of the services that target the destination pod |
| `dst_pod` | business2-0 | The name of the destination pod  |
| `dst_container` | business-container | The name of the destination container |
| `dst_ip` | 10.1.11.24 | Pod's IP by default. If the destination is not a pod in Kubernetes, this is the IP address of an external entity |
| `dst_port` | 80 | The listening port of the destination container, if applicable |

### Notes
**Note 1**: Before Kindling v0.7.0, the `kindling_tcp_retransmit_total` was used to count how many retransmit events happened, which is less than total number of resending segments since Linux may resend mutiple segments during one retransmit event.

## TCP Socket Connects Metrics

### Metrics List
| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| `kindling_tcp_connect_total` | Counter | Total number of successfully and unsuccessfully established TCP connections |
| `kindling_tcp_connect_duration_nanoseconds_total` | Counter | Total duration of the successfully established TCP connections |

### Labels List
| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| `pid` | 1024 | The client's process ID|
| `comm` | java | The client's process command|
| `src_node` | slave-node1 | Which node the source pod is on |
| `src_namespace` | default | Namespace of the source pod |
| `src_workload_kind` | deployment | Workload kind of the source pod |
| `src_workload_name` | business1 | Workload name of the source pod |
| `src_service` | business1-svc | One of the services that target the source pod |
| `src_pod` | business1-0 | The name of the source pod |
| `src_container` | business-container | The name of the source container |
| `src_container_id` | 1a2b3c4d5e6f | The shorten container id which contains 12 characters |
| `src_ip` | 10.1.11.23 | Pod's IP by default. If the source is not a pod in Kubernetes, this is the IP address of an external entity |
| `dst_node` | slave-node2 | Which node the destination pod is on |
| `dst_namespace` | default | Namespace of the destination pod |
| `dst_workload_kind` | deployment | Workload kind of the destination pod |
| `dst_workload_name` | business2 | Workload name of the destination pod |
| `dst_service` | business2-svc | One of the services that target the destination pod |
| `dst_pod` | business2-0 | The name of the destination pod  |
| `dst_container` | business-container | The name of the destination container |
| `dst_ip` | 10.1.11.24 | Pod's IP by default. If the destination is not a pod in Kubernetes, this is the IP address of an external entity |
| `dst_port` | 80 | The listening port of the destination container, if applicable |
| `dnat_ip` | 192.168.12.3 | The IP address of the destination after DNAT if applicable |
| `dnat_port` | 80 | The listening port of the destination container after DNAT if applicable |
| `success` | true | Whether the TCP connection is successfully established |
| `errno` | 0 | The error number of the TCP connection. 0 if no error. Note it could also be 0 even if there is an error. |

### Notes
**Note 1**: The field `success` for `kindling_tcp_connect_duration_nanoseconds_total` is always `true`.

**Note 2**: The field `errno` is not `0` only if the TCP socket is blocking and there is an error happened. There are multiple possible values it could contain. See the `ERRORS` section of the [connect(2) manual](https://man7.org/linux/man-pages/man2/connect.2.html) for more details.

**Note 3**: The field `pid` and `comm` will not exist if you set `need_process_info` to `false` (default is false), that will reduce the pressure of Prometheus.

## PromQL Example
Here are some examples of how to use these metrics in Prometheus, which can help you understand them faster.

| **Describe** | **PromQL** |
| --- | --- |
| Request counts | `sum(increase(kindling_entity_request_total{namespace="$namespace",workload_name="$workload"}[5m])) by(namespace, workload_name)` |
| DNS request counts | `sum(increase(kindling_topology_request_total{src_namespace="$namespace",src_workload_name="$workload", protocol="dns"}[5m])) by (src_workload_name)` |
| Latency | `sum by(namespace, workload_name) (increase(kindling_entity_request_duration_nanoseconds_total{namespace="$namespace", workload_name="$workload"}[5m])) / sum by(namespace, workload_name) (increase(kindling_entity_request_total{namespace="$namespace", workload_name="$workload"}[5m]))` |
| Error ratio of HTTP requests | `sum (increase(kindling_entity_request_total{namespace="$namespace",workload_name="$workload",protocol="http",response_content=~"4..\|5.."}[5m])) / sum (increase(kindling_entity_request_total{namespace="$namespace",workload_name="$workload",protocol="http"}[5m])) * 100` |
| Request latency quantile | `histogram_quantile(0.99, rate(kindling_topology_request_average_duration_nanoseconds_bucket{dst_namespace="$namespace", dst_workload_name="$workload",protocol="http"}[5m]))` |
| Retransmit times | `sum(increase(kindling_tcp_retransmit_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[5m]))` |
| Packets lost count | `sum(increase(kindling_tcp_packet_loss_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[5m]))` |
| Network sent bytes | `sum(increase(kindling_topology_request_request_bytes_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[5m]))` |
| Network received bytes | `sum(increase(kindling_topology_request_response_bytes_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[5m]))` |
