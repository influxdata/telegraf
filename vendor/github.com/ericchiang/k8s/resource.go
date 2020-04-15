package k8s

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

// Option represents optional call parameters, such as label selectors.
type Option interface {
	updateURL(base string, v url.Values) string
	updateDelete(r Resource, d *deleteOptions)
}

type optionFunc func(base string, v url.Values) string

func (f optionFunc) updateDelete(r Resource, d *deleteOptions)  {}
func (f optionFunc) updateURL(base string, v url.Values) string { return f(base, v) }

type deleteOptions struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`

	GracePeriod   *int64 `json:"gracePeriodSeconds,omitempty"`
	Preconditions struct {
		UID string `json:"uid,omitempty"`
	} `json:"preconditions"`
	PropagationPolicy string `json:"propagationPolicy"`
}

// QueryParam can be used to manually set a URL query parameter by name.
func QueryParam(name, value string) Option {
	return optionFunc(func(base string, v url.Values) string {
		v.Set(name, value)
		return base
	})
}

type deleteOptionFunc func(r Resource, d *deleteOptions)

func (f deleteOptionFunc) updateDelete(r Resource, d *deleteOptions)  { f(r, d) }
func (f deleteOptionFunc) updateURL(base string, v url.Values) string { return base }

func DeleteAtomic() Option {
	return deleteOptionFunc(func(r Resource, d *deleteOptions) {
		d.Preconditions.UID = *r.GetMetadata().Uid
	})
}

// DeletePropagationOrphan orphans the dependent resources during a delete.
func DeletePropagationOrphan() Option {
	return deleteOptionFunc(func(r Resource, d *deleteOptions) {
		d.PropagationPolicy = "Orphan"
	})
}

// DeletePropagationBackground deletes the resources and causes the garbage
// collector to delete dependent resources in the background.
func DeletePropagationBackground() Option {
	return deleteOptionFunc(func(r Resource, d *deleteOptions) {
		d.PropagationPolicy = "Background"
	})
}

// DeletePropagationForeground deletes the resources and causes the garbage
// collector to delete dependent resources and wait for all dependents whose
// ownerReference.blockOwnerDeletion=true.  API sever will put the "foregroundDeletion"
// finalizer on the object, and sets its deletionTimestamp.  This policy is
// cascading, i.e., the dependents will be deleted with Foreground.
func DeletePropagationForeground() Option {
	return deleteOptionFunc(func(r Resource, d *deleteOptions) {
		d.PropagationPolicy = "Foreground"
	})
}

func DeleteGracePeriod(d time.Duration) Option {
	seconds := int64(d / time.Second)
	return deleteOptionFunc(func(r Resource, d *deleteOptions) {
		d.GracePeriod = &seconds
	})
}

// ResourceVersion causes watch operations to only show changes since
// a particular version of a resource.
func ResourceVersion(resourceVersion string) Option {
	return QueryParam("resourceVersion", resourceVersion)
}

// Timeout declares the timeout for list and watch operations. Timeout
// is only accurate to the second.
func Timeout(d time.Duration) Option {
	return QueryParam(
		"timeoutSeconds",
		strconv.FormatInt(int64(d/time.Second), 10),
	)
}

// Subresource is a way to interact with a part of an API object without needing
// permissions on the entire resource. For example, a node isn't able to modify
// a pod object, but can update the "pods/status" subresource.
//
// Common subresources are "status" and "scale".
//
// See https://kubernetes.io/docs/reference/api-concepts/
func Subresource(name string) Option {
	return optionFunc(func(base string, v url.Values) string {
		return base + "/" + name
	})
}

type resourceType struct {
	apiGroup   string
	apiVersion string
	name       string
	namespaced bool
}

var (
	resources     = map[reflect.Type]resourceType{}
	resourceLists = map[reflect.Type]resourceType{}
)

// Resource is a Kubernetes resource, such as a Node or Pod.
type Resource interface {
	GetMetadata() *metav1.ObjectMeta
}

// Resource is list of common Kubernetes resources, such as a NodeList or
// PodList.
type ResourceList interface {
	GetMetadata() *metav1.ListMeta
}

func Register(apiGroup, apiVersion, name string, namespaced bool, r Resource) {
	rt := reflect.TypeOf(r)
	if _, ok := resources[rt]; ok {
		panic(fmt.Sprintf("resource registered twice %T", r))
	}
	resources[rt] = resourceType{apiGroup, apiVersion, name, namespaced}
}

func RegisterList(apiGroup, apiVersion, name string, namespaced bool, l ResourceList) {
	rt := reflect.TypeOf(l)
	if _, ok := resources[rt]; ok {
		panic(fmt.Sprintf("resource registered twice %T", l))
	}
	resourceLists[rt] = resourceType{apiGroup, apiVersion, name, namespaced}
}

func urlFor(endpoint, apiGroup, apiVersion, namespace, resource, name string, options ...Option) string {
	basePath := "apis/"
	if apiGroup == "" {
		basePath = "api/"
	}

	var p string
	if namespace != "" {
		p = path.Join(basePath, apiGroup, apiVersion, "namespaces", namespace, resource, name)
	} else {
		p = path.Join(basePath, apiGroup, apiVersion, resource, name)
	}
	e := ""
	if strings.HasSuffix(endpoint, "/") {
		e = endpoint + p
	} else {
		e = endpoint + "/" + p
	}
	if len(options) == 0 {
		return e
	}

	v := url.Values{}
	for _, option := range options {
		e = option.updateURL(e, v)
	}
	if len(v) == 0 {
		return e
	}
	return e + "?" + v.Encode()
}

func urlForPath(endpoint, path string) string {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if strings.HasSuffix(endpoint, "/") {
		return endpoint + path
	}
	return endpoint + "/" + path
}

func resourceURL(endpoint string, r Resource, withName bool, options ...Option) (string, error) {
	t, ok := resources[reflect.TypeOf(r)]
	if !ok {
		return "", fmt.Errorf("unregistered type %T", r)
	}
	meta := r.GetMetadata()
	if meta == nil {
		return "", errors.New("resource has no object meta")
	}
	switch {
	case t.namespaced && (meta.Namespace == nil || *meta.Namespace == ""):
		return "", errors.New("no resource namespace provided")
	case !t.namespaced && (meta.Namespace != nil && *meta.Namespace != ""):
		return "", errors.New("resource not namespaced")
	case withName && (meta.Name == nil || *meta.Name == ""):
		return "", errors.New("no resource name provided")
	}
	name := ""
	if withName {
		name = *meta.Name
	}
	namespace := ""
	if t.namespaced {
		namespace = *meta.Namespace
	}

	return urlFor(endpoint, t.apiGroup, t.apiVersion, namespace, t.name, name, options...), nil
}

func resourceGetURL(endpoint, namespace, name string, r Resource, options ...Option) (string, error) {
	t, ok := resources[reflect.TypeOf(r)]
	if !ok {
		return "", fmt.Errorf("unregistered type %T", r)
	}

	if !t.namespaced && namespace != "" {
		return "", fmt.Errorf("type not namespaced")
	}
	if t.namespaced && namespace == "" {
		return "", fmt.Errorf("no namespace provided")
	}

	return urlFor(endpoint, t.apiGroup, t.apiVersion, namespace, t.name, name, options...), nil
}

func resourceListURL(endpoint, namespace string, r ResourceList, options ...Option) (string, error) {
	t, ok := resourceLists[reflect.TypeOf(r)]
	if !ok {
		return "", fmt.Errorf("unregistered type %T", r)
	}

	if !t.namespaced && namespace != "" {
		return "", fmt.Errorf("type not namespaced")
	}

	return urlFor(endpoint, t.apiGroup, t.apiVersion, namespace, t.name, "", options...), nil
}

func resourceWatchURL(endpoint, namespace string, r Resource, options ...Option) (string, error) {
	t, ok := resources[reflect.TypeOf(r)]
	if !ok {
		return "", fmt.Errorf("unregistered type %T", r)
	}

	if !t.namespaced && namespace != "" {
		return "", fmt.Errorf("type not namespaced")
	}

	url := urlFor(endpoint, t.apiGroup, t.apiVersion, namespace, t.name, "", options...)
	if strings.Contains(url, "?") {
		url = url + "&watch=true"
	} else {
		url = url + "?watch=true"
	}
	return url, nil
}
