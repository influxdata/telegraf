package system

import (
	kube "k8s.io/kubernetes/pkg/client/unversioned"

	godocker "github.com/fsouza/go-dockerclient"

	"github.com/influxdata/telegraf"
	docker "github.com/influxdata/telegraf/internal/docker"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Kubernetes struct {
	DockerEndpoint   string
	ContainerNames   []string
	DockerClient     *godocker.Client
	KubernetesClient *kube.Client
}

var sampleConfig = `
  # Docker Endpoint
  #   To use TCP, set docker_endpoint = "tcp://[ip]:[port]"
  #   To use environment variables (ie, docker-machine), set docker_endpoint = "ENV"
  docker_endpoint = "unix:///var/run/docker.sock"
  # Only collect metrics for these containers, collect all if empty
  container_names = []
`

func (k *Kubernetes) Description() string {
	return "Read metrics about docker containers running on a kubernetes cluster"
}

func (k *Kubernetes) SampleConfig() string { return sampleConfig }

func (k *Kubernetes) Gather(acc telegraf.Accumulator) error {
	var err error

	k.KubernetesClient, err = kube.NewInCluster()
	if err != nil {
		return err
	}

	k.DockerClient, err = docker.CreateClient(k.DockerEndpoint)
	if err != nil {
		return err
	}

	if k.DockerClient != nil {
		err = docker.GatherContainerMetrics(k.DockerClient, k.KubernetesClient, k.ContainerNames, acc)
	}

	return err
}

func init() {
	inputs.Add("kubernetes", func() telegraf.Input {
		return &Kubernetes{}
	})
}
