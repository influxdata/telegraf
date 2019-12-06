package kubernetes

import "encoding/json"

type Podlist struct {
	Kind       string                       `json:"kind"`
	ApiVersion string                       `json:"apiVersion"`
	Items      []map[string]json.RawMessage `json:"items"`
}

type PodInfo struct {
	Name      string            `json:"name"`
	NameSpace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}
