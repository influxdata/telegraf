# Kubernetes Input Plugin

**This plugin is experimental and may cause high cardinality issues with moderate to large Kubernetes deployments**

This input plugin talks to the kubelet api using the `/stats/summary` endpoint to gather metrics about the running pods and containers for a single host. It is assumed that this plugin is running as part of a `daemonset` within a kubernetes installation. This means that telegraf is running on every node within the cluster. Therefore, you should configure this plugin to talk to its locally running kubelet.

To find the ip address of the host you are running on you can issue a command like the following:
```
$ curl -s $API_URL/api/v1/namespaces/$POD_NAMESPACE/pods/$HOSTNAME --header "Authorization: Bearer $TOKEN" --insecure | jq -r '.status.hostIP'
```
In this case we used the downward API to pass in the `$POD_NAMESPACE` and `$HOSTNAME` is the hostname of the pod which is set by the kubernetes API.

## Summary Data

```json
{
  "node": {
   "nodeName": "node1",
   "systemContainers": [
    {
     "name": "kubelet",
     "startTime": "2016-08-25T18:46:52Z",
     "cpu": {
      "time": "2016-09-27T16:57:31Z",
      "usageNanoCores": 56652446,
      "usageCoreNanoSeconds": 101437561712262
     },
     "memory": {
      "time": "2016-09-27T16:57:31Z",
      "usageBytes": 62529536,
      "workingSetBytes": 62349312,
      "rssBytes": 47509504,
      "pageFaults": 4769397409,
      "majorPageFaults": 13
     },
     "rootfs": {
      "availableBytes": 84379979776,
      "capacityBytes": 105553100800
     },
     "logs": {
      "availableBytes": 84379979776,
      "capacityBytes": 105553100800
     },
     "userDefinedMetrics": null
   },
   {
    "name": "bar",
    "startTime": "2016-08-25T18:46:52Z",
    "cpu": {
     "time": "2016-09-27T16:57:31Z",
     "usageNanoCores": 56652446,
     "usageCoreNanoSeconds": 101437561712262
    },
    "memory": {
     "time": "2016-09-27T16:57:31Z",
     "usageBytes": 62529536,
     "workingSetBytes": 62349312,
     "rssBytes": 47509504,
     "pageFaults": 4769397409,
     "majorPageFaults": 13
    },
    "rootfs": {
     "availableBytes": 84379979776,
     "capacityBytes": 105553100800
    },
    "logs": {
     "availableBytes": 84379979776,
     "capacityBytes": 105553100800
    },
    "userDefinedMetrics": null
   }
   ],
   "startTime": "2016-08-25T18:46:52Z",
   "cpu": {
    "time": "2016-09-27T16:57:41Z",
    "usageNanoCores": 576996212,
    "usageCoreNanoSeconds": 774129887054161
   },
   "memory": {
    "time": "2016-09-27T16:57:41Z",
    "availableBytes": 10726387712,
    "usageBytes": 12313182208,
    "workingSetBytes": 5081538560,
    "rssBytes": 35586048,
    "pageFaults": 351742,
    "majorPageFaults": 1236
   },
   "network": {
    "time": "2016-09-27T16:57:41Z",
    "rxBytes": 213281337459,
    "rxErrors": 0,
    "txBytes": 292869995684,
    "txErrors": 0
   },
   "fs": {
    "availableBytes": 84379979776,
    "capacityBytes": 105553100800,
    "usedBytes": 16754286592
   },
   "runtime": {
    "imageFs": {
     "availableBytes": 84379979776,
     "capacityBytes": 105553100800,
     "usedBytes": 5809371475
    }
   }
  },
  "pods": [
   {
    "podRef": {
     "name": "foopod",
     "namespace": "foons",
     "uid": "6d305b06-8419-11e6-825c-42010af000ae"
    },
    "startTime": "2016-09-26T18:45:42Z",
    "containers": [
     {
      "name": "foocontainer",
      "startTime": "2016-09-26T18:46:43Z",
      "cpu": {
       "time": "2016-09-27T16:57:32Z",
       "usageNanoCores": 846503,
       "usageCoreNanoSeconds": 56507553554
      },
      "memory": {
       "time": "2016-09-27T16:57:32Z",
       "usageBytes": 30789632,
       "workingSetBytes": 30789632,
       "rssBytes": 30695424,
       "pageFaults": 10761,
       "majorPageFaults": 0
      },
      "rootfs": {
       "availableBytes": 84379979776,
       "capacityBytes": 105553100800,
       "usedBytes": 57344
      },
      "logs": {
       "availableBytes": 84379979776,
       "capacityBytes": 105553100800,
       "usedBytes": 24576
      },
      "userDefinedMetrics": null
     }
    ],
    "network": {
     "time": "2016-09-27T16:57:34Z",
     "rxBytes": 70749124,
     "rxErrors": 0,
     "txBytes": 47813506,
     "txErrors": 0
    },
    "volume": [
     {
      "availableBytes": 7903948800,
      "capacityBytes": 7903961088,
      "usedBytes": 12288,
      "name": "volume1"
     },
     {
      "availableBytes": 7903956992,
      "capacityBytes": 7903961088,
      "usedBytes": 4096,
      "name": "volume2"
     },
     {
      "availableBytes": 7903948800,
      "capacityBytes": 7903961088,
      "usedBytes": 12288,
      "name": "volume3"
     },
     {
      "availableBytes": 7903952896,
      "capacityBytes": 7903961088,
      "usedBytes": 8192,
      "name": "volume4"
     }
    ]
   }
  ]
 }
 ```

 ### Daemonset YAML

```yaml
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: telegraf
  namespace: telegraf
spec:
  template:
    metadata:
      labels:
        app: telegraf
    spec:
      serviceAccount: telegraf
      containers:
        - name: telegraf
          image: quay.io/org/image:latest
          imagePullPolicy: IfNotPresent
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: "HOST_PROC"
            value: "/rootfs/proc"
          - name: "HOST_SYS"
            value: "/rootfs/sys"
          volumeMounts:
          - name: sysro
            mountPath: /rootfs/sys
            readOnly: true
          - name: procro
            mountPath: /rootfs/proc
            readOnly: true
          - name: varrunutmpro
            mountPath: /var/run/utmp
            readOnly: true
          - name: logger-redis-creds
            mountPath: /var/run/secrets/deis/redis/creds
      volumes:
      - name: sysro
        hostPath:
          path: /sys
      - name: procro
        hostPath:
          path: /proc
      - name: varrunutmpro
        hostPath:
          path: /var/run/utmp
```

### Line Protocol

#### kubernetes_pod_container
```
kubernetes_pod_container,host=ip-10-0-0-0.ec2.internal,
container_name=deis-controller,namespace=deis,
node_name=ip-10-0-0-0.ec2.internal, pod_name=deis-controller-3058870187-xazsr, cpu_usage_core_nanoseconds=2432835i,cpu_usage_nanocores=0i,
logsfs_avaialble_bytes=121128271872i,logsfs_capacity_bytes=153567944704i,
logsfs_used_bytes=20787200i,memory_major_page_faults=0i,
memory_page_faults=175i,memory_rss_bytes=0i,
memory_usage_bytes=0i,memory_working_set_bytes=0i,
rootfs_available_bytes=121128271872i,rootfs_capacity_bytes=153567944704i,
rootfs_used_bytes=1110016i 1476477530000000000
 ```

#### kubernetes_pod_volume
```
kubernetes_pod_volume,host=ip-10-0-0-0.ec2.internal,name=default-token-f7wts,
namespace=kube-system,node_name=ip-10-0-0-0.ec2.internal,
pod_name=kubernetes-dashboard-v1.1.1-t4x4t, available_bytes=8415240192i,
capacity_bytes=8415252480i,used_bytes=12288i 1476477530000000000
```

#### kubernetes_pod_network
```
kubernetes_pod_network,host=ip-10-0-0-0.ec2.internal,namespace=deis,
node_name=ip-10-0-0-0.ec2.internal,pod_name=deis-controller-3058870187-xazsr,
rx_bytes=120671099i,rx_errors=0i,
tx_bytes=102451983i,tx_errors=0i 1476477530000000000
```
