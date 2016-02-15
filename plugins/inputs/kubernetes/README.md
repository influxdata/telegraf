# Kubernetes Input Plugin

The Kubernetes plugin can gather metrics from Kubernetes services
- APIServer
- Scheduler
- Controller Manager
- Kubelet

### Configuration:

```toml
# Get metrics from Kubernetes services    
  [[inputs.kubernetes.apiserver]]
    url = "http://mtl-nvcbladea-15.nuance.com:8080"
    endpoint = "/metrics"
    timeout = 5.0
    # includes only metrics which match one of the
    # following regexp
    includes = ["apiserver_.*"]

  [[inputs.kubernetes.scheduler]]
    url = "http://mtl-nvcbladea-15.nuance.com:10251"
    endpoint = "/metrics"
    timeout = 1.0
    # DO NOT include metrics which match one of the
    # following regexp
    excludes = ["scheduler_.*"]

  [[inputs.kubernetes.controllermanager]]
    url = "http://mtl-nvcbladea-15.nuance.com:10252"

  [[inputs.kubernetes.kubelet]]
    url = "http://mtl-nvcbladea-15.nuance.com:4194"
    # You should increase metric_buffer_limit
    # Because of number of kubelet metrics
    # otherwise you can limit metrics with
    # the following 'excludes' argument
    excludes = ["container_.*"]
```

### Measurements & Fields:

This input plugin get measurements and fields from Kubernetes services.
If new metrics appear on K8S services, this plugin will grab it without
modification.

- http_request_duration_microseconds"
    - 0.5
    - 0.9
    - 0.99
    - count
    - sum
- .......
    - .......
    - .......
    - .......
    - .......
    - .......

### Tags:

This input plugin get metrics tags from Kubernetes services.
If new tags appear on K8S services, this plugin will grab it without
modification.

- kubeservice
- serverURL
- handler
- .......
- .......
- .......

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter example -test
process_cpu_seconds_total,kubeservice=kubelet,serverURL=http://mtl-blade19-02.nuance.com:4194/metrics counter=452366.52 1455436043283475033
ssh_tunnel_open_fail_count,kubeservice=kubelet,serverURL=http://mtl-blade19-02.nuance.com:4194/metrics counter=0 1455436043284057879
get_token_count,kubeservice=kubelet,serverURL=http://mtl-blade19-02.nuance.com:4194/metrics counter=0 1455436043285030409
http_requests_total,code=200,handler=prometheus,kubeservice=kubelet,method=get,serverURL=http://mtl-blade19-02.nuance.com:4194/metrics counter=116 1455436043285364118
kubelet_generate_pod_status_latency_microseconds,kubeservice=kubelet,serverURL=http://mtl-blade19-02.nuance.com:4194/metrics 0.5=94441,0.9=186928,0.99=236698,count=0,sum=128409528225 1455436043285715318
process_max_fds,kubeservice=kubelet,serverURL=http://mtl-blade19-02.nuance.com:4194/metrics gauge=1000000 1455436043286061204
```
