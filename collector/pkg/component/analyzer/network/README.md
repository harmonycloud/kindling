# Network Metrics Analyzer

NetworkAnalyzer is a component that analyzes syscall events received, parses the application layer protocol 
and generates a `model.DataGroup` which contains the details on every request.

## Configuration
See [config.go](./config.go) for the config specification.

## Consumable Events (Input)
- syscall_exit-writev
- syscall_exit-readv
- syscall_exit-write
- syscall_exit-read
- syscall_exit-sendto
- syscall_exit-recvfrom
- syscall_exit-sendmsg
- syscall_exit-recvmsg
- syscall_exit-sendmmsg

## Generated Data (Output)
`NetworkAnalyzer` generates a `model.DataGroup` for every request and then sends it to the next consumer. There are 
a few things to note here:
- The `DataGroup` used in `NetworkAnalyzer` is from a [sync.Pool](./datagroup_pool.go), which contains a pool of
reusable `DataGroup` objects, to reduce garbage collection (GC) pressure. 
- The `DataGroup` MAY be sent directly to the next consumer, so when it is returned to the pool, it may contain additional labels.
- The labels will not be removed when the `DataGroup` is put back to the pool, as the combination of labels is usually 
same in most cases.

The `DataGroup` contains the following fields:
- `Name` is always `net_request_metric_group`.
- `Value` contains the following six fields:
  - `request_sent_time`: The time when the request is sent.
  - `waiting_ttfb_time`: The time to first byte.
  - `content_download_time`: The time to download the content.
  - `request_total_time`: The total time of the request.
  - `request_io`: The total number of bytes sent.
  - `response_io`: The total number of bytes received.
- `Lables` contains the following fields:
  - `pid`: The ID of the process.
  - `comm`: The name of the process.
  - `container_id`: The ID of the container.
  - `content_key`: The key of the content. For example, the path of the HTTP request.
  - `dst_ip`: The IP address of the destination.
  - `dst_port`: The port of the destination.
  - `dnat_ip`: The IP address of the destination NAT.
  - `dnat_port`: The port of the destination NAT.
  - `end_timestamp`: The time when the request is finished.
  - `error_type`: The type of the error.
  - `is_error`: Whether the request contains an error.
  - `is_server`: Whether the request is captured from the server side.
  - `is_slow`: Whether the request is slow.
  - `protocol`: The protocol of the request.
  - `request_tid`: The ID of the thread that sends/receives the request.
  - `request_payload`: The payload of the request.
  - `response_tid`: The ID of the thread that receives/sends the response.
  - `response_payload`: The payload of the response.
  - `http_method`: The HTTP method of the request.
  - `http_status_code`: The HTTP status code of the response.
  - `http_url`: The URL of the request.
  - `dns_domain`: The domain of the DNS request.
  - `dns_id`: The ID of the DNS request.
  - `dns_rcode`: The RCODE of the DNS response.
  - You may also find the following *empty* fields. They are there because `DataGroup` is sent to the next consumer, 
and these labels are used by them. After that, the `DataGroup` is reused but no labels are removed.
    - `src_container_id`: The ID of the source container.
    - `src_ip`: The IP address of the source.
    - `src_port`: The port of the source.
    - `src_container_id`: The ID of the source container.
    - `src_container`: The name of the source container.
    - `src_pod`: The name of the source pod.
    - `src_workload_kind`: The kind of the source workload.
    - `src_workload_name`: The name of the source workload.
    - `src_service`: The name of the source service.
    - `src_namespace`: The namespace of the source.
    - `src_node`: The name of the source node.
    - `src_node_ip`: The IP address of the source node.
    - `dst_container_id`: The ID of the destination container.
    - `dst_container`: The name of the destination container.
    - `dst_pod`: The name of the destination pod.
    - `dst_workload_kind`: The kind of the destination workload.
    - `dst_workload_name`: The name of the destination workload.
- `Timestamp` is the time when the request is sent/received.

An example is as follows.
```json
{
  "Name": "net_request_metric_group",
  "Values": {
    "request_sent_time": 30361,
    "waiting_ttfb_time": 644909,
    "content_download_time": 6875,
    "request_total_time": 682145,
    "request_io": 48,
    "response_io": 132
  },
  "Labels": {
    "comm": "wrk",
    "container_id": "",
    "content_key": "/hello",
    "dnat_ip": "",
    "dnat_port": -1,
    "dns_domain": "",
    "dns_id": 0,
    "dns_rcode": 0,
    "dst_container": "",
    "dst_container_id": "",
    "dst_ip": "10.244.8.209",
    "dst_port": 8080,
    "dst_namespace": "",
    "dst_node": "",
    "dst_node_ip": "",
    "dst_pod": "",
    "dst_service": "",
    "dst_workload_kind": "",
    "dst_workload_name": "",
    "end_timestamp": 1683205041396831651,
    "error_type": 0,
    "http_method": "GET",
    "http_status_code": 200,
    "http_url": "/hello",
    "is_error": false,
    "is_server": false,
    "is_slow": false,
    "pid": 23412,
    "protocol": "http",
    "request_payload": "GET /hello HTTP/1.1\r\nHost: 10.244.8.209:8080\r\n\r\n",
    "request_tid": 23414,
    "response_payload": "HTTP/1.1 200 \r\nContent-Type: text/plain;charset=UTF-8\r\nContent-Length: 18\r\nDate: Thu, 04 May 2023 12:57:21 GMT\r\n\r\nhello,spring boot!",
    "response_tid": 23414,
    "src_container": "",
    "src_container_id": "",
    "src_ip": "10.244.2.0",
    "src_port": 54652,
    "src_namespace": "",
    "src_node": "",
    "src_node_ip": "",
    "src_pod": "",
    "src_service": "",
    "src_workload_kind": "",
    "src_workload_name": ""
  },
  "Timestamp": 1683204084753831028
}
```