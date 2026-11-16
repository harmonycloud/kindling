# K8sInfo Analyzer

K8sInfoAnalyzer is a component that analyzes k8s workload infomation received, receive `model.DataGroup` from metadata
and send to exporter
## Configuration
See [config.go](./config.go) for the config specification.

## Received Data (Input)

The `DataGroup` contains the following fields:
- `Name` is always `k8s_workload_metric_group`.
- `Lables` contains the following fields:
  - `namespace`: The namespace of the workload.
  - `workload_name `: The name of the workload.
  - `workload_kind`: The kind of the workload.
- `Metrics` contains the following fields:
  - `kindling_k8s_workload_info`

An example is as follows.
```json
{
  "Name": "k8s_workload_metric_group",
  "Metrics":{
    "kindling_k8s_workload_info": 1
  }
  "Labels": {
    "namespace": "kindling",
    "workload_kind": "Deployment",
    "workload_name": "testdemo1",
    }
}
```