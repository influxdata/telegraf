package podman

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/docker"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Podman struct {
	Endpoint string
	Timeout  config.Duration

	ContainerInclude []string `toml:"container_name_include"`
	ContainerExclude []string `toml:"container_name_exclude"`

	Log             telegraf.Logger
	client          Client
	engineHost      string
	serverVersion   string
	filtersCreated  bool
	containerFilter filter.Filter
}

var sampleConfig = `
  ## Podman Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  endpoint = "unix:///var/run/podman.sock"

  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  container_name_include = []
  container_name_exclude = []

  ## Timeout for podman list, info, and stats commands
  timeout = "5s"

  `

func (p *Podman) SampleConfig() string { return sampleConfig }

func (p *Podman) Description() string {
	return "Read metrics about podman containers"
}

func (p *Podman) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()
	if p.client == nil {
		c, err := NewClient(p.Endpoint, ctx)
		if err != nil {
			return err
		}
		p.client = c
	}

	now := time.Now()

	// Get system info
	info, err := p.client.Info()
	if err != nil {
		return err
	}

	//Describes the host distribution for libpod
	p.engineHost = info.Host.Distribution.Distribution
	p.serverVersion = info.Version.Version

	tags := map[string]string{
		"engine_host":    p.engineHost,
		"server_version": p.serverVersion,
	}

	fields := map[string]interface{}{
		"n_cpus":               info.Host.CPUs,
		"n_containers":         info.Store.ContainerStore.Number,
		"n_containers_running": info.Store.ContainerStore.Running,
		"n_containers_stopped": info.Store.ContainerStore.Stopped,
		"n_containers_paused":  info.Store.ContainerStore.Paused,
		"n_images":             info.Store.ImageStore.Number,
		//"n_listener_events": info.Host.NEventsListener,
	}

	acc.AddFields("podman", fields, tags, now)
	acc.AddFields("podman", map[string]interface{}{"memory_total": info.Host.MemTotal}, tags, now)

	// Create label filters if not already created
	if !p.filtersCreated {
		err := p.createContainerFilters()
		if err != nil {
			return err
		}
		p.filtersCreated = true
	}

	ctxList, cancel := context.WithTimeout(p.client.Background(), time.Duration(p.Timeout))
	defer cancel()
	containers, err := p.client.ContainerList(ctxList, nil)
	if err != nil {
		return err
	}

	// Get container data
	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, container := range containers {
		go func(c entities.ListContainer) {
			defer wg.Done()
			if err := p.gatherContainer(c, acc); err != nil {
				acc.AddError(err)
			}
		}(container)
	}
	wg.Wait()
	return nil
}

func (p *Podman) gatherContainer(container entities.ListContainer, acc telegraf.Accumulator) error {
	// Parse container name
	var cname string
	for _, name := range container.Names {
		trimmedName := strings.TrimPrefix(name, "/")
		if !strings.Contains(trimmedName, "/") {
			cname = trimmedName
			break
		}
	}

	if cname == "" {
		return nil
	}

	if !p.containerFilter.Match(cname) {
		return nil
	}

	if !p.containerFilter.Match(cname) {
		return nil
	}

	imageName, imageVersion := docker.ParseImage(container.Image)

	tags := map[string]string{
		"engine_host":       p.engineHost,
		"server_version":    p.serverVersion,
		"container_name":    cname,
		"container_image":   imageName,
		"container_version": imageVersion,
		"pod_name":          container.PodName,
	}

	ctxStats, cancel := context.WithTimeout(p.client.Background(), time.Duration(p.Timeout))
	defer cancel()
	containerStats, err := p.client.ContainerStats(ctxStats, cname)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("%#v\n", containerStats)
	fields := map[string]interface{}{
		"container_id": containerStats.ContainerID,
		"cpu":          containerStats.CPU,
		"mem_usage":    containerStats.MemUsage,
		"mem_limit":    containerStats.MemLimit,
	}
	acc.AddFields("podman", fields, tags, time.Now())
	return nil
}

func (p *Podman) createContainerFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(p.ContainerInclude, p.ContainerExclude)
	if err != nil {
		return err
	}
	p.containerFilter = filter
	return nil
}

func init() {
	inputs.Add("podman", func() telegraf.Input {
		return &Podman{
			filtersCreated: false,
		}
	})
}
