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
  url = "http://1.1.1.1:10255"

  ## Use bearer token for authorization
  #  bearer_token = /path/to/bearer/token

  ## Set response_timeout (default 5 seconds)
  #  response_timeout = "5s"

  ## Optional TLS Config
  #  tls_ca = /path/to/cafile
  #  tls_cert = /path/to/certfile
  #  tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  #  insecure_skip_verify = false

  ## Worker pool for kube_state_metric plugin only
  #  empty this field will use default value 30
  #  max_connections = 5

  ## Optional Max Config Map Age, this will ignore
  ## old config map for the second time of gathering.
  ## Blank will use the default value 24 hours
  #  max_config_map_age = "24h"

  ## Optional Max Job Age, this will ignore
  ## old job created for kube_job metrics for the second time of gathering.
  ## Blank will collect all jobs
  #  max_job_age = "24h"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be "cronjobs", "daemonsets", "deployments",
  ## "jobs", "limitranges", "nodes", "pods", "replicasets", "replicationcontrollers",
  ## "resourcequotas", "services", "statefulsets", "persistentvolumes",
  ## "persistentvolumeclaims", "namespaces", "horizontalpodautoscalers",
  ## "endpoints", "secrets", "configmaps"
  #  resource_exclude = [ "cronjobs", "daemonsets", "deployments" ]

  ## Optional Resouce List Check Interval, leave blank will use the default
  #  value of 1 hour. This is the interval to check available resource lists.
  #  resouce_list_check_interval = "1h"
```

### Metrics:

#### kube_configmap
- ts: configmap created
- fields:
  - gauge (always 1)
- tags:
  - namespace
  - configmap
  - resource_version

#### kube_cronjob
- ts: now
- fields:
  - status_active ( # of active jobs)
  - spec_starting_deadline_seconds
  - next_schedule_time
  - schedule
  - created
  - status_last_schedule_time
- tags:
  - namespace
  - cronjob
  - concurrency_policy
  - label_* 

#### kube_daemonset
- ts: now
- fields:
  - metadata_generation
  - created
  - status_current_number_scheduled
  - status_desired_number_scheduled
  - status_number_available
  - status_number_misscheduled
  - status_number_ready
  - status_number_unavailable
  - status_updated_number_scheduled
- tags:
  - namespace
  - daemonset
  - label_*

#### kube_deployment
- ts: now
- fields:
  - created
  - spec_replicas
  - metadata_generation
  - spec_strategy_rollingupdate_max_unavailable
  - spec_strategy_rollingupdate_max_surge
  - status_replicas
  - status_replicas_available
  - status_replicas_unavailable
  - status_replicas_updated
  - status_observed_generation
- tags:
  - namespace
  - deployment
  - spec_paused ("true", "false")

#### kube_endpoint
- ts: now
- fields:
  - created
  - address_available
  - address_not_ready
- tags:
  - namespace
  - endpoint

#### kube_hpa
- ts: hpa created
- fields:
  - metadata_generation
  - spec_max_replicas
  - spec_min_replicas
- tags:
  - namespace
  - hpa
  - label_*

#### kube_hpa_status
- ts: now
- fields:
  - current_replicas
  - desired_replicas
  - condition_true (1 = "true", 0 = "false")
  - condition_false (1 = "true", 0 = "false")
  - condition_unknown
- tags:
  - namespace
  - hpa
  - condition ("true", "false", "unkown")
  
#### kube_job
- ts: job completion time
- measurement & fields:
  - status_succeeded
  - status_failed
  - status_active
  - spec_parallelism
  - spec_completions
  - created
  - spec_active_deadline_seconds
  - status_start_time
- tags:
  - namespace
  - job_name
  - label_*

#### kube_job_condition
- ts: job condition last transition time
- fields:
  - completed (1 = "true", 0 = "false")
  - failed (1 = "true", 0 = "false")
- tags:
  - namespace
  - job_name
  - condition

#### kube_limitrange
- ts: now
- fields:
  - created
  - min_pod_cpu
  - min_pod_memory
  - min_pod_storage
  - min_pod_ephemeral_storage
  - min_container_cpu
  - min_container_memory
  - min_container_storage
  - min_container_ephemeral_storage
  - min_persistentvolumeclaim_cpu
  - min_persistentvolumeclaim_memory
  - min_persistentvolumeclaim_storage
  - min_persistentvolumeclaim_ephemeral_storage
  - max_pod_cpu
  - max_pod_memory
  - max_pod_storage
  - max_pod_ephemeral_storage
  - max_container_cpu
  - max_container_memory
  - max_container_storage
  - max_container_ephemeral_storage
  - max_persistentvolumeclaim_cpu
  - max_persistentvolumeclaim_memory
  - max_persistentvolumeclaim_storage
  - max_persistentvolumeclaim_ephemeral_storage
  - default_pod_cpu
  - default_pod_memory
  - default_pod_storage
  - default_pod_ephemeral_storage
  - default_container_cpu
  - default_container_memory
  - default_container_storage
  - default_container_ephemeral_storage
  - default_persistentvolumeclaim_cpu
  - default_persistentvolumeclaim_memory
  - default_persistentvolumeclaim_storage
  - default_persistentvolumeclaim_ephemeral_storage
  - default_request_pod_cpu
  - default_request_pod_memory
  - default_request_pod_storage
  - default_request_pod_ephemeral_storage
  - default_request_container_cpu
  - default_request_container_memory
  - default_request_container_storage
  - default_request_container_ephemeral_storage
  - default_request_persistentvolumeclaim_cpu
  - default_request_persistentvolumeclaim_memory
  - default_request_persistentvolumeclaim_storage
  - default_request_persistentvolumeclaim_ephemeral_storage
  - max_limit_request_ratio_pod_cpu
  - max_limit_request_ratio_pod_memory
  - max_limit_request_ratio_pod_storage
  - max_limit_request_ratio_pod_ephemeral_storage
  - max_limit_request_ratio_container_cpu
  - max_limit_request_ratio_container_memory
  - max_limit_request_ratio_container_storage
  - max_limit_request_ratio_container_ephemeral_storage
  - max_limit_request_ratio_persistentvolumeclaim_cpu
  - max_limit_request_ratio_persistentvolumeclaim_memory
  - max_limit_request_ratio_persistentvolumeclaim_storage
  - max_limit_request_ratio_persistentvolumeclaim_ephemeral_storage
- tags:
  - namespace
  - limitrange

#### kube_namespace
- ts: now
- fields:
  - created
  - status_phase_code (1="Active", 0="Terminating")
- tags
  - namespace
  - status_phase
  - label_*
  - annotation_*

#### kube_node
- ts: now
- fields:
  - created
  - status_capacity_cpu_cores
  - status_capacity_ephemera_storage_bytes
  - status_capacity_memory_bytes
  - status_capacity_pods
  - status_capacity_*
  - status_allocatable_cpu_cores
  - status_allocatable_ephemera_storage_bytes
  - status_allocatable_memory_bytes
  - status_allocatable_pods
  - status_allocatable_*
- tags:
  - node
  - kernel_version
  - os_image
  - container_runtime_version
  - kubelet_version
  - kubeproxy_version
  - status_phase
  - provider_id
  - spec_unschedulable
  - label_*

#### kube_node_spec_taint
- ts: now
- fields:
  - gauge (always 1)
- tags:
  - node
  - key
  - value
  - effect

#### kube_node_status_conditions  
- ts: condition last transition time
- fields:
  - gauge (always 1)
- tags:
  - node
  - condition
  - status

#### kube_persistentvolume
- ts: now
- fields:
  - status_pending (1 = "true", 0 = "false")
  - status_available (1 = "true", 0 = "false")
  - status_bound (1 = "true", 0 = "false")
  - status_released (1 = "true", 0 = "false")
  - status_failed (1 = "true", 0 = "false")
- tags:
  - persistentvolume
  - storageclass
  - status
  - label_*

#### kube_persistentvolumeclaim
- ts: now
- fields:
  - status_lost (1 = "true", 0 = "false")
  - status_bound (1 = "true", 0 = "false")
  - status_failed (1 = "true", 0 = "false")
  - resource_requests_storage_bytes
- tags:
  - namespace
  - persistentvolumeclaim
  - storageclass
  - volumename
  - status
  - label_*

#### kube_pod

- ts: pod created
- fields:
  - gauge (always 1)
- tags:
  - namespace
  - pod
  - node
  - created_by_kind
  - created_by_name
  - owner_kind
  - owner_name
  - owner_is_controller ("true", "false")
  - label_*

#### kube_pod_status

- ts: now
- fields:
  - status_phase_pending (1 = "true", 0 = "false")
  - status_phase_succeeded (1 = "true", 0 = "false")
  - status_phase_failed (1 = "true", 0 = "false")
  - status_phase_running (1 = "true", 0 = "false")
  - status_phase_unknown (1 = "true", 0 = "false")
  - completion_time
  - scheduled_time
- tags:
  - namespace
  - pod
  - node
  - host_ip
  - pod_ip
  - status_phase ("pending", "succeeded", "failed", "running", "unknown")
  - ready ("true", "false")
  - scheduled ("true", "false")

#### kube_pod_container

- ts: now
- fields:
  - status_restarts_total
  - status_waiting (1 = "true", 0 = "false")
  - status_running (1 = "true", 0 = "false")
  - status_terminated (1 = "true", 0 = "false")
  - status_ready (1 = "true", 0 = "false")
  - resource_requests_cpu_cores
  - resource_requests_memory_bytes
  - resource_requests_storage_bytes
  - resource_requests_ephemeral_storage_bytes
  - resource_limits_cpu_cores
  - resource_limits_memory_bytes
  - resource_limits_storage_bytes
  - resource_limits_ephemeral_storage_bytes
- tags:
  - namespace
  - pod_name
  - node_name
  - container
  - image
  - image_id
  - container_id
  - status_waiting_reason
  - status_terminated_reason


#### kube_pod_volume
- ts: now
- fields:
  - read_only (1 = "true", 0 = "false")
- tags:
  - namespace
  - pod
  - volume
  - persistentvolumeclaim

#### kube_replicasets  
- ts: now
- fields:
  - created
  - metadata_generation
  - spec_replicas
  - status_replicas
  - status_fully_labeled_replicas
  - status_ready_replicas
  - status_observed_generation
- tags
  - namespace
  - replicaset 
  
#### kube_replicationcontroller
- ts: now
- fields:
  - created
  - metadata_generation
  - spec_replicas
  - status_replicas
  - status_fully_labeled_replicas
  - status_ready_replicas
  - status_available_replicas
  - status_observed_generation

- tags
  - namespace
  - replicationcontroller

#### kube_resourcequota
- ts: resourcequota creation time
- fields:
  - gauge
- tags:
  - namespace
  - resourcequota
  - resource
  - type ("hard", "used")

#### kube_secret
- ts: secret creation time
- fields:
  - gauge (always 1)
- tags:
  - namespace
  - secret
  - resource_version
  - label_*

#### kube_service
- ts: service creation time
- fields:
  - gauge (always 1)
- tags:
  - namespace
  - service
  - type
  - cluster_ip
  - label_*

#### kube_statefulset
- ts: statefulset creation time
- fields:
  - metadata_generation
  - replicas
  - status_replicas
  - status_replicas_current
  - status_replicas_ready
  - status_replicas_updated
  - status_observed_generation
- tags:
  - namespace
  - statefulset
  - label_*

### Sample Queries:

  ```
    select gauge from kube_configmap where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select status_active from kube_cronjob where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select status_current_number_scheduled,status_desired_number_scheduled,
    status_number_available,
    status_number_misscheduled,
    status_number_ready, status_number_unavailable, status_updated_number_scheduled from kube_daemonset where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select spec_replicas, metadata_generation, status_replicas, status_replicas_available, status_replicas_unavailable, status_replicas_updated, status_observed_generation from kube_deployment where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select address_available, address_not_ready from kube_endpoint where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select metadata_generation, spec_max_replicas, spec_min_replicas from kube_hpa where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select current_replicas, desired_replicas, condition_true,condition_false from kube_hpa_status where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select status_succeeded, status_failed, status_active, spec_parallelism, spec_completions from kube_job where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select completed, failed from kube_job_condition where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select max_container_memory, default_request_persistentvolumeclaim_cpu from kube_limitrange where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select status_phase_code from kube_namespace where label_xf = fs1 AND time > now() - 1h GROUP BY label_xf
  ```

  ```
    select status_capacity_cpu_cores, status_capacity_ephemeral_storage_bytes from kube_node where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select gauge from kube_node_spec_taint where namespace = ns1 AND key = s1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select gauge from kube_node_status_conditions where namespace = ns1 and condition = outofdisk AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select status_lost, status_pending, status_bound from kube_persistentvolume where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select status_lost, status_pending, status_bound from kube_persistentvolumeclaim where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select gauge from kube_pod where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select scheduled_time, status_phase_pending, status_phase_succeeded, status_phase_failed, status_phase_running, status_phase_unknown from kube_pod_status where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select status_restarts_total, status_waiting, status_running, status_terminated, status_ready from kube_pod_container where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select read_only from kube_pod_spec_volumes where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select metadata_generation, status_replicas, status_fully_labeled_replicas, status_ready_replicas, status_observed_generation, spec_replicas from kube_replicasets where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select metadata_generation, status_replicas, status_fully_labeled_replicas, status_ready_replicas,   status_available_replicas, status_observed_generation, spec_replicas from kube_replicationcontroller where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select gauge from kube_resourcequota where namespace = ns1 AND type = used AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select gauge from kube_secret where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select gauge from kube_service where namespace = ns1 AND time > now() - 1h GROUP BY namespace
  ```

  ```
    select metadata_generation, status_replicas, status_replicas_current, status_replicas_ready, status_replicas_updated from kube_statefulset where namespace = ns1 AND time > now() - 1h GROUP BY namespace
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