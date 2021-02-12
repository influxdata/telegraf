Kubernetes auditing provides a security-relevant chronological set of records
documenting the sequence of activities that have affected system by individual
users, administrators or other components of the system. It allows cluster
administrator to answer the following questions:

* what happened?
* when did it happen?
* who initiated it?
* on what did it happen?
* where was it observed?
* from where was it initiated?
* to where was it going?

You can keep reading the [official
docs](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/) to know
more about Kubernetes Audit log.

There are two different ways to get audit logs, one is tailing a file, the other
one is configuring a webhook. This plugin catch the webhooks sends from the
kube-apiserver and it stores them in InfluxDB.

## How to try it

### 0. Requirements:

* minikube installed
* an influxdb running on port 8086
* you need a way to make telegraf available from minikube network. I used ngrok
  to proxy telegraf to the outside

### 1. Start telegraf

When you have the proxied URL for your telegraf you should replace the url [the kubeconfig
here](./example/webhook.yaml) with the right one.

Now you can start  telegraf using [this
configuration](./example/telegraf.config).

### 2. Kubernetes via minikube

First of all you need to start the VM passing few other flags to expose the
audit log from the apiserver:

```
minikube start \
    --extra-config=apiserver.audit-webhook-config-file=/var/lib/localkube/certs/hack/example/webhook.yaml \
    --extra-config=apiserver.audit-policy-file=/var/lib/localkube/certs/hack/example/policy.yaml \
    --extra-config=apiserver.audit-webhook-mode=batch
```

We also need to mount the example directory inside the VM or the
apiserver won't be able to load the audit policy and the kubeconfig.

```
minikube mount $TELEGRAF_PATH/pluigins/inputs/webhooks/kubernetes_audit/example:/var/lib/localkube/certs/hack/example
```

The mount location inside the VM is currently an hack explain
[here](https://github.com/kubernetes/minikube/issues/2741#issuecomment-398683171).

### 3. The end
If everything is working right you should be able to query the audit logs in
your influxdb.
