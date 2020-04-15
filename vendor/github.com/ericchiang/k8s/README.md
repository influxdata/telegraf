# A simple Go client for Kubernetes

[![GoDoc](https://godoc.org/github.com/ericchiang/k8s?status.svg)](https://godoc.org/github.com/ericchiang/k8s)
[![Build Status](https://travis-ci.org/ericchiang/k8s.svg?branch=master)](https://travis-ci.org/ericchiang/k8s)

A slimmed down Go client generated using Kubernetes' [protocol buffer][protobuf] support. This package behaves similarly to [official Kubernetes' Go client][client-go], but only imports two external dependencies.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ericchiang/k8s"
    corev1 "github.com/ericchiang/k8s/apis/core/v1"
)

func main() {
    client, err := k8s.NewInClusterClient()
    if err != nil {
        log.Fatal(err)
    }

    var nodes corev1.NodeList
    if err := client.List(context.Background(), "", &nodes); err != nil {
        log.Fatal(err)
    }
    for _, node := range nodes.Items {
        fmt.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
    }
}
```

## Requirements

* Go 1.7+ (this package uses "context" features added in 1.7)
* Kubernetes 1.3+ (protobuf support was added in 1.3)
* [github.com/golang/protobuf/proto][go-proto] (protobuf serialization)
* [golang.org/x/net/http2][go-http2] (HTTP/2 support)

## Usage

### Create, update, delete

The type of the object passed to `Create`, `Update`, and `Delete` determine the resource being acted on.

```go
configMap := &corev1.ConfigMap{
    Metadata: &metav1.ObjectMeta{
        Name:      k8s.String("my-configmap"),
        Namespace: k8s.String("my-namespace"),
    },
    Data: map[string]string{"hello": "world"},
}

if err := client.Create(ctx, configMap); err != nil {
    // handle error
}

configMap.Data["hello"] = "kubernetes"

if err := client.Update(ctx, configMap); err != nil {
    // handle error
}

if err := client.Delete(ctx, configMap); err != nil {
    // handle error
}
```

### Get, list, watch

Getting a resource requires providing a namespace (for namespaced objects) and a name.

```go
// Get the "cluster-info" configmap from the "kube-public" namespace
var configMap corev1.ConfigMap
err := client.Get(ctx, "kube-public", "cluster-info", &configMap)
```

When performing a list operation, the namespace to list or watch is also required.

```go
// Pods from the "custom-namespace"
var pods corev1.PodList
err := client.List(ctx, "custom-namespace", &pods)
```

A special value `AllNamespaces` indicates that the list or watch should be performed on all cluster resources.

```go
// Pods in all namespaces
var pods corev1.PodList
err := client.List(ctx, k8s.AllNamespaces, &pods)
```

Watches require a example type to determine what resource they're watching. `Watch` returns an type which can be used to receive a stream of events. These events include resources of the same kind and the kind of the event (added, modified, deleted).

```go
// Watch configmaps in the "kube-system" namespace
var configMap corev1.ConfigMap
watcher, err := client.Watch(ctx, "kube-system", &configMap)
if err != nil {
    // handle error
}
defer watcher.Close()

for {
    cm := new(corev1.ConfigMap)
    eventType, err := watcher.Next(cm)
    if err != nil {
        // watcher encountered and error, exit or create a new watcher
    }
    fmt.Println(eventType, *cm.Metadata.Name)
}
```

Both in-cluster and out-of-cluster clients are initialized with a primary namespace. This is the recommended value to use when listing or watching.

```go
client, err := k8s.NewInClusterClient()
if err != nil {
    // handle error
}

// List pods in the namespace the client is running in.
var pods corev1.PodList
err := client.List(ctx, client.Namespace, &pods)
```

### Custom resources

Client operations support user defined resources, such as resources provided by [CustomResourceDefinitions][crds] and [aggregated API servers][custom-api-servers].  To use a custom resource, define an equivalent Go struct then register it with the `k8s` package. By default the client will use JSON serialization when encoding and decoding custom resources.

```go
import (
    "github.com/ericchiang/k8s"
    metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

type MyResource struct {
    Metadata *metav1.ObjectMeta `json:"metadata"`
    Foo      string             `json:"foo"`
    Bar      int                `json:"bar"`
}

// Required for MyResource to implement k8s.Resource
func (m *MyResource) GetMetadata() *metav1.ObjectMeta {
    return m.Metadata
}

type MyResourceList struct {
    Metadata *metav1.ListMeta `json:"metadata"`
    Items    []MyResource     `json:"items"`
}

// Require for MyResourceList to implement k8s.ResourceList
func (m *MyResourceList) GetMetadata() *metav1.ListMeta {
    return m.Metadata
}

func init() {
    // Register resources with the k8s package.
    k8s.Register("resource.example.com", "v1", "myresources", true, &MyResource{})
    k8s.RegisterList("resource.example.com", "v1", "myresources", true, &MyResourceList{})
}
```

Once registered, the library can use the custom resources like any other.

```go
func do(ctx context.Context, client *k8s.Client, namespace string) error {
    r := &MyResource{
        Metadata: &metav1.ObjectMeta{
            Name:      k8s.String("my-custom-resource"),
            Namespace: &namespace,
        },
        Foo: "hello, world!",
        Bar: 42,
    }
    if err := client.Create(ctx, r); err != nil {
        return fmt.Errorf("create: %v", err)
    }
    r.Bar = -8
    if err := client.Update(ctx, r); err != nil {
        return fmt.Errorf("update: %v", err)
    }
    if err := client.Delete(ctx, r); err != nil {
        return fmt.Errorf("delete: %v", err)
    }
    return nil
}
```

If the custom type implements [`proto.Message`][proto-msg], the client will prefer protobuf when encoding and decoding the type.

### Label selectors

Label selectors can be provided to any list operation.

```go
l := new(k8s.LabelSelector)
l.Eq("tier", "production")
l.In("app", "database", "frontend")

var pods corev1.PodList
err := client.List(ctx, "custom-namespace", &pods, l.Selector())
```

### Subresources

Access subresources using the `Subresource` option.

```go
err := client.Update(ctx, &pod, k8s.Subresource("status"))
```

### Creating out-of-cluster clients

Out-of-cluster clients can be constructed by either creating an `http.Client` manually or parsing a [`Config`][config] object. The following is an example of creating a client from a kubeconfig:

```go
import (
    "io/ioutil"

    "github.com/ericchiang/k8s"

    "github.com/ghodss/yaml"
)

// loadClient parses a kubeconfig from a file and returns a Kubernetes
// client. It does not support extensions or client auth providers.
func loadClient(kubeconfigPath string) (*k8s.Client, error) {
    data, err := ioutil.ReadFile(kubeconfigPath)
    if err != nil {
        return nil, fmt.Errorf("read kubeconfig: %v", err)
    }

    // Unmarshal YAML into a Kubernetes config object.
    var config k8s.Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("unmarshal kubeconfig: %v", err)
    }
    return k8s.NewClient(&config)
}
```

### Errors

Errors returned by the Kubernetes API are formatted as [`unversioned.Status`][unversioned-status] objects and surfaced by clients as [`*k8s.APIError`][k8s-error]s. Programs that need to inspect error codes or failure details can use a type cast to access this information.

```go
// createConfigMap creates a configmap in the client's default namespace
// but does not return an error if a configmap of the same name already
// exists.
func createConfigMap(client *k8s.Client, name string, values map[string]string) error {
    cm := &v1.ConfigMap{
        Metadata: &metav1.ObjectMeta{
            Name:      &name,
            Namespace: &client.Namespace,
        },
        Data: values,
    }

    err := client.Create(context.TODO(), cm)

    // If an HTTP error was returned by the API server, it will be of type
    // *k8s.APIError. This can be used to inspect the status code.
    if apiErr, ok := err.(*k8s.APIError); ok {
        // Resource already exists. Carry on.
        if apiErr.Code == http.StatusConflict {
            return nil
        }
    }
    return fmt.Errorf("create configmap: %v", err)
}
```

[client-go]: https://github.com/kubernetes/client-go
[go-proto]: https://godoc.org/github.com/golang/protobuf/proto
[go-http2]: https://godoc.org/golang.org/x/net/http2
[protobuf]: https://developers.google.com/protocol-buffers/
[unversioned-status]: https://godoc.org/github.com/ericchiang/k8s/api/unversioned#Status
[k8s-error]: https://godoc.org/github.com/ericchiang/k8s#APIError
[config]: https://godoc.org/github.com/ericchiang/k8s#Config
[string]: https://godoc.org/github.com/ericchiang/k8s#String
[crds]: https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/
[custom-api-servers]: https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/
[proto-msg]: https://godoc.org/github.com/golang/protobuf/proto#Message 
