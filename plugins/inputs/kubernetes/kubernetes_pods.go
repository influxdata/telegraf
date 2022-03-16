package kubernetes

type Pods struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Items      []Item `json:"items"`
}

type Item struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type Spec struct {
	Containers []Container `json:"containers"`
}

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}
