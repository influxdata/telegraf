# Kubernete_State Plugin
[kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) is an open source project designed to generate metrics derived from the state of Kubernetes objects – the abstractions Kubernetes uses to represent your cluster. With this information you can monitor details such as:

State of nodes, pods, and jobs
Compliance with replicaSets specs
Resource requests and min/max limits

The Kubernete State Plugin gathers information based on [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics)

### Configuration:

```toml
[[inputs.kube_state]]
  ## URL for the kubelet
  url = "https://1.1.1.1"

  ## Namespace to use
  namespace = "default"

  ## Use bearer token for authorization
  #  bearer_token = "abc123"

  ## Set response_timeout (default 5 seconds)
  #  response_timeout = "5s"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be - "configmaps", "daemonsets", deployments", "nodes",
  ## "persistentvolumes", "persistentvolumeclaims", "pods", "statefulsets"
  #  resource_exclude = [ "deployments", "nodes", "statefulsets" ]

  ## Optional Resources to include when gathering
  ## Overrides resource_exclude if both set.
  #  resource_include = [ "deployments", "nodes", "statefulsets" ]

  ## Optional max age for config map
  #  max_config_map_age = "1h"

  ## Optional TLS Config
  ## Use TLS but skip chain & host verification
  #  insecure_skip_verify = false
```

### Metrics:

#### kube_configmap
```
TS: created

Fields:
    gauge
Tags:
    name
    namespace
    resource_version
```

#### kube_daemonset
```
TS: now

Fields:
    metadata_generation
    created
    status_current_number_scheduled
    status_desired_number_scheduled
    status_number_available
    status_number_misscheduled
    status_number_ready
    status_number_unavailable
    status_updated_number_scheduled

Tags:
    name
    namespace
```

#### kube_deployment
```
TS: now

Fields:
    status_replicas_available
    status_replicas_unavailable
    created
Tags:
    namespace
    name
```

#### kube_node
```
TS: now

Fields:
    status_allocatable_cpu_cores
    status_allocatable_memory_bytes
    status_allocatable_pods
    status_capacity_pods
    status_capacity_cpu_cores
    status_capacity_memory_bytes
Tags:
    name
```

#### kube_persistentvolume
```
TS: now

Fields:
    status_available
    status_bound
    status_failed
    status_pending
    status_released

Tags:
    name
    storageclass
    status
```

#### kube_persistentvolumeclaim
```
TS: now

Fields:
    status_lost
    status_pending
    status_bound

Tags:
    name
    namespace
    status
    storageclass
```

#### kube_pod_container
```
TS: now

Fields:
    status_restarts_total
    status_running
    status_terminated
    status_terminated_reason
    resource_requests_cpu_cores
    resource_requests_memory_bytes
    resource_requests_limits_cpu_cores
    resource_requests_limits_memory_bytes

Tags:
    name
    namespace
    node
    pod
```

#### kube_pod_status
```
TS: now

Fields:
    ready
Tags:
    name
    namespace
    node
    reason  [Completed,ContainerCannotRun,Error,OOMKilled]
```

#### kube_statefulset
```
TS: statefulset creation time

Fields:
    metadata_generation
    replicas
    status_replicas
    status_replicas_current
    status_replicas_ready
    status_replicas_updated
    status_observed_generation

Tags:
    name
    namespace
```


### Example Output:

```
./telegraf --config sample.conf --input-filter kube_state --test

kube_persistentvolume,host=ip-192-168-1-7.ec2.internal,label_failure_domain_beta_kubernetes_io_region=us-east-1,label_failure_domain_beta_kubernetes_io_zone=us-east-1c,persistentvolume=pvc-05f362fe-7a31-11e8-87cb-1232e142048e,status=bound,storageclass=ebs-1 status_available=0i,status_bound=1i,status_failed=0i,status_pending=0i,status_released=0i 1532310328000000000
kube_service,cluster_ip=10.96.0.1,host=ip-192-168-1-7.ec2.internal,label_component=apiserver,label_provider=kubernetes,namespace=default,service=kubernetes,type=ClusterIP gauge=1i 1521811667000000000
kube_node_status_conditions,condition=outofdisk,host=ip-192-168-1-7.ec2.internal,node=ip-180-12-0-10.ec2.internal,status=false gauge=1i 1529116717000000000
kube_node,container_runtime_version=docker://18.3.1,host=ip-192-168-1-7.ec2.internal,kernel_version=4.14.48-coreos-r2,kubelet_version=v1.10.3,kubeproxy_version=v1.10.3,label_beta_kubernetes_io_arch=amd64,label_beta_kubernetes_io_instance_type=r4.4xlarge,label_beta_kubernetes_io_os=linux,label_failure_domain_beta_kubernetes_io_region=us-east-1,label_failure_domain_beta_kubernetes_io_zone=us-east-1c,label_kubernetes_io_hostname=ip-180-12-0-10,node=ip-180-12-0-10.ec2.internal,os_image=Container\ Linux\ by\ CoreOS\ 1745.7.0\ (Rhyolite),provider_id=aws:///us-east-1c/i-0c00dca5166f05cec,spec_unschedulable=false created=1527602419i,status_allocatable_cpu_cores=16i,status_allocatable_ephemeral_storage_bytes=44582761194i,status_allocatable_hugepages_1Gi_bytes=0i,status_allocatable_hugepages_2Mi_bytes=0i,status_allocatable_memory_bytes=128732758016i,status_allocatable_pods=110i,status_capacity_cpu_cores=16i,status_capacity_ephemeral_storage_bytes=49536401408i,status_capacity_hugepages_1Gi_bytes=0i,status_capacity_hugepages_2Mi_bytes=0i,status_capacity_memory_bytes=128837615616i,status_capacity_pods=110i 1532310329000000000
kube_node_status_conditions,condition=ready,host=ip-192-168-1-7.ec2.internal,node=ip-180-12-0-16.ec2.internal,status=true gauge=1i 1529149835000000000
kube_secret,host=ip-192-168-1-7.ec2.internal,namespace=default,resource_version=309,secret=default-token-xz8jw gauge=1i 1521811687000000000
kube_configmap,configmap=ingress-controller-leader-nginx,host=ip-192-168-1-7.ec2.internal,namespace=ingress-nginx,resource_version=25230493 gauge=1i 1521812871000000000
kube_namespace,annotation_kubectl_kubernetes_io_last_applied_configuration={"apiVersion":"v1"\,"kind":"Namespace"\,"metadata":{"annotations":{}\,"name":"monitoring"\,"namespace":""}\,"spec":{"finalizers":["kubernetes"]}}\n,host=ip-192-168-1-7.ec2.internal,namespace=monitoring,status_phase=active created=1521812741i,status_phase_code=1i 1532310329000000000
kube_persistentvolumeclaim,host=ip-192-168-1-7.ec2.internal,namespace=jenkins,persistentvolumeclaim=jenkins-home-claim,status=bound,storageclass=ebs-1,volumename=pvc-ef66a844-0599-11e8-842c-02aa4bc06eb8 resource_requests_storage_bytes=53687091200i,status_bound=1i,status_lost=0i,status_pending=0i 1532310329000000000
kube_endpoint,endpoint=kubernetes,host=ip-192-168-1-7.ec2.internal,namespace=default address_available=1i,address_not_ready=0i,created=1521811667i 1532310329000000000
kube_pod,created_by_kind=ReplicaSet,created_by_name=fluxbot-5f886fcc49,host=ip-192-168-1-7.ec2.internal,label_app=fluxbot,label_pod_template_hash=1944297705,namespace=fluxbot,node=ip-180-12-20-96.ec2.internal,owner_is_controller=true,owner_kind=ReplicaSet,owner_name=fluxbot-5f886fcc49,pod=fluxbot-5f886fcc49-dwzhl gauge=1i 1530260807000000000
kube_pod_status,host=ip-192-168-1-7.ec2.internal,host_ip=180.12.20.96,namespace=fluxbot,node=ip-180-12-20-96.ec2.internal,pod=fluxbot-5f886fcc49-dwzhl,pod_ip=10.244.11.22,ready=true,scheduled=false,status_phase=running start_time=1530260807i,status_phase_failed=0i,status_phase_pending=0i,status_phase_running=1i,status_phase_succeeded=0i,status_phase_unknown=0i 1530260829000000000
kube_pod,created_by_kind=DaemonSet,created_by_name=forwarder,host=ip-192-168-1-7.ec2.internal,label_app=forwarder,label_controller_revision_hash=2401403235,label_pod_template_generation=3,namespace=frontline,node=ip-180-12-10-18.ec2.internal,owner_is_controller=true,owner_kind=DaemonSet,owner_name=forwarder,pod=forwarder-rxwhn gauge=1i 1530777209000000000
kube_pod_container,container=forwarder,container_id=docker://54abe32d0094479d3dcb9ba43a9d19cc343d9774e238d6f1932aebd862ed1d99,host=ip-192-168-1-7.ec2.internal,image=quay.io/influxdb/forward-proxy:latest,image_id=docker-pullable://quay.io/influxdb/forward-proxy@sha256:2919e3fe5ca7c8ed4f2a592576eca5a662c35c7d6a6a6c081d5b00c49a7fe662,namespace=frontline,node_name=ip-180-12-10-18.ec2.internal,pod_name=forwarder-rxwhn status_ready=1i,status_restarts_total=0i,status_running=1i,status_terminated=0i,status_waiting=0i 1532310329000000000
kube_pod_container,container=nginx-ingress-controller,container_id=docker://652b8ea63ff36df31649c672353bde03af90628c76182d05962c71d1c76b3d8a,host=ip-192-168-1-7.ec2.internal,image=quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.16.2,image_id=docker-pullable://quay.io/kubernetes-ingress-controller/nginx-ingress-controller@sha256:84ed5290a91c53b4c224a724a251347dfc8cf2bca4be06e32f642c396eb02429,namespace=ingress-nginx,node_name=ip-180-12-0-10.ec2.internal,pod_name=nginx-ingress-controller-6b6f867648-5cvcn status_ready=1i,status_restarts_total=0i,status_running=1i,status_terminated=0i,status_waiting=0i 1532310329000000000
```