package processors

import (
	"os"

	"github.com/influxdata/telegraf"
)

const (
	containerNameEnvName        = "SEMATEXT_CONTAINER_NAME"
	containerIDEnvName          = "SEMATEXT_CONTAINER_ID"
	containerImageNameEnvName   = "SEMATEXT_CONTAINER_IMAGE_NAME"
	containerImageTagEnvName    = "SEMATEXT_CONTAINER_IMAGE_TAG"
	containerImageDigestEnvName = "SEMATEXT_CONTAINER_IMAGE_DIGEST"
	k8sPodEnvName               = "SEMATEXT_K8S_POD_NAME"
	k8sNamespaceEnvName         = "SEMATEXT_K8S_NAMESPACE"
	k8sClusterEnvName           = "SEMATEXT_K8S_CLUSTER"

	containerImageTag       = "container.name"
	containerHostnameTag    = "container.hostname"
	containerIDTag          = "container.id"
	containerImageNameTag   = "container.image.name"
	containerImageTagTag    = "container.image.tag"
	containerImageDigestTag = "container.image.digest"
	k8sPodNameTag           = "kubernetes.pod.name"
	k8sNamespaceIDTag       = "kubernetes.namespace"
	k8sClusterTag           = "kubernetes.cluster.name"
)

// ContainerTags is a metric processor that injects container tags read from env variables
type ContainerTags struct {
	tags map[string]string
}

// NewContainerTags creates new instance of container MetricProcessor
func NewContainerTags() MetricProcessor {
	tags := make(map[string]string)
	tags[containerImageTag] = os.Getenv(containerNameEnvName)
	tags[containerIDTag] = os.Getenv(containerIDEnvName)
	tags[containerImageNameTag] = os.Getenv(containerImageNameEnvName)
	tags[containerImageTagTag] = os.Getenv(containerImageTagEnvName)
	tags[containerImageDigestTag] = os.Getenv(containerImageDigestEnvName)

	tags[k8sPodNameTag] = os.Getenv(k8sPodEnvName)
	tags[k8sNamespaceIDTag] = os.Getenv(k8sNamespaceEnvName)
	tags[k8sClusterTag] = os.Getenv(k8sClusterEnvName)

	// fill container.hostname only when running in a container env
	if tags[containerIDTag] != "" {
		hostname, err := os.Hostname()
		if err != nil {
			// use container id value for container.hostname when the hostname can't be read
			hostname = tags[containerIDTag]
		}
		tags[containerHostnameTag] = hostname
	}

	return &ContainerTags{
		tags: tags,
	}
}

// Process is a method where ContainerTags processor injects container tags from env variables to metric
func (c *ContainerTags) Process(metric telegraf.Metric) error {
	for tag, value := range c.tags {
		if value != "" {
			metric.AddTag(tag, value)
		}
	}
	return nil
}

// Close clears the resources processor used, no-op in this case
func (c *ContainerTags) Close() {}
