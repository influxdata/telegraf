package kubernetes

import "encoding/json"

type Pods struct {
	Kind       string                       `json:"kind"`
	ApiVersion string                       `json:"apiVersion"`
	Items      []map[string]json.RawMessage `json:"items"`
}

type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}
