package kubernetes

type pods struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Items      []item `json:"items"`
}

type item struct {
	Metadata metadata `json:"metadata"`
	Spec     spec     `json:"spec"`
}

type metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type spec struct {
	Containers []container `json:"containers"`
}

type container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}
