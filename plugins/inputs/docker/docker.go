package system

import (
	"github.com/influxdata/telegraf"
	docker "github.com/influxdata/telegraf/internal/docker"
	"github.com/influxdata/telegraf/plugins/inputs"

	godocker "github.com/fsouza/go-dockerclient"
)

type Docker struct {
	Endpoint       string
	ContainerNames []string
	client         *godocker.Client
}

var sampleConfig = `
  # Docker Endpoint
  #   To use TCP, set endpoint = "tcp://[ip]:[port]"
  #   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"
  # Only collect metrics for these containers, collect all if empty
  container_names = []
`

func (d *Docker) Description() string {
	return "Read metrics about docker containers"
}

func (d *Docker) SampleConfig() string { return sampleConfig }

func (d *Docker) Gather(acc telegraf.Accumulator) error {
	var err error

	d.client, err = docker.CreateClient(d.Endpoint)
	if err != nil {
		return err
	}

	if d.client != nil {
		err = docker.GatherContainerMetrics(d.client, nil, d.ContainerNames, acc)
	}
	return err
}

func init() {
	inputs.Add("docker", func() telegraf.Input {
		return &Docker{}
	})
}
