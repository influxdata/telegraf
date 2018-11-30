package ecs

import (
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Ecs config object
type Ecs struct {
	EcsV2   string `toml:"ecsv2_url"`
	EnvCfg  bool   `toml:"envcfg"`
	Timeout internal.Duration

	ContainerNameInclude []string `toml:"container_name_include"`
	ContainerNameExclude []string `toml:"container_name_exclude"`

	ContainerStatusInclude []string `toml:"container_status_include"`
	ContainerStatusExclude []string `toml:"container_status_exclude"`

	LabelInclude []string `toml:"ecs_label_include"`
	LabelExclude []string `toml:"ecs_label_exclude"`

	newEnvClient func() (*EcsClient, error)
	newClient    func(timeout time.Duration) (*EcsClient, error)

	client              Client
	filtersCreated      bool
	labelFilter         filter.Filter
	containerNameFilter filter.Filter
	statusFilter        filter.Filter
}

const (
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB
)

var (
	sizeRegex       = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgGtTpP])?[bB]?$`)
	containerStates = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
)

var sampleConfig = `
  ## ECS metadata url
  # ecsv2_url = "169.254.170.2"

  ## Set to true to configure from env vars
  envcfg = false

  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  container_name_include = []
  container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  # container_status_include = []
  # container_status_exclude = []

  ## ecs labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  ecs_label_include = [ "com.amazonaws.ecs.*" ]
  ecs_label_exclude = []

  ## Timeout for docker list, info, and stats commands
  timeout = "5s"
`

// Description describes ECS plugin
func (ecs *Ecs) Description() string {
	return "Read metrics about docker containers from Fargate/ECS v2 meta endpoints."
}

// SampleConfig returns the ECS example config
func (ecs *Ecs) SampleConfig() string {
	return sampleConfig
}

// Gather is the entrypoint for telegraf metrics collection
func (ecs *Ecs) Gather(acc telegraf.Accumulator) error {
	err := initSetup(ecs)
	if err != nil {
		return err
	}

	task, stats, err := PollSync(ecs.client)
	if err != nil {
		return err
	}

	mergeTaskStats(&task, stats)

	// accumulate metrics
	ecs.accTask(task, acc)
	ecs.accContainers(task, acc)

	return nil
}

func initSetup(ecs *Ecs) error {
	if ecs.client == nil {
		var c *EcsClient
		var err error
		if ecs.EnvCfg {
			c, err = ecs.newEnvClient()
		} else {
			c, err = ecs.newClient(ecs.Timeout.Duration)
		}
		if err != nil {
			return err
		}

		c.BaseURL = &url.URL{
			Scheme: ecsMetaScheme,
			Host:   ecs.EcsV2,
		}

		ecs.client = c
	}

	// Create filters
	if !ecs.filtersCreated {
		err := ecs.createContainerNameFilters()
		if err != nil {
			return err
		}
		err = ecs.createContainerStatusFilters()
		if err != nil {
			return err
		}
		err = ecs.createLabelFilters()
		if err != nil {
			return err
		}
		ecs.filtersCreated = true
	}

	return nil
}

//region: metrics

func (ecs *Ecs) accTask(task Task, acc telegraf.Accumulator) {
	parseTaskStats(task, acc)
}

func (ecs *Ecs) accContainers(task Task, acc telegraf.Accumulator) {
	taskTags := taskTags(task)

	for _, c := range task.Containers {
		match := false
		if ecs.containerNameFilter.Match(c.Name) {
			match = true
		}
		if ecs.statusFilter.Match(c.KnownStatus) {
			match = true
		}
		for k := range c.Labels {
			if ecs.labelFilter.Match(k) {
				match = true
			}
		}
		if !match {
			log.Printf("container %v did not match any filters", c.ID)
			continue
		}
		parseContainerStats(c, acc, taskTags)
	}
}

//endregion: metrics

//region: tags

// taskTags accepts ECS task metadata and parses out tags for the task
func taskTags(task Task) map[string]string {
	return map[string]string{
		"cluster":  task.Cluster,
		"task_arn": task.TaskARN,
		"family":   task.Family,
		"revision": task.Revision,
	}
}

// containerTags accepts ECS container metadata and parses out tags for the container
func containerTags(container Container) map[string]string {
	tags := map[string]string{
		"id":   container.ID,
		"name": container.Name,
	}
	// add all ECS container Labels
	for k, v := range container.Labels {
		tags[k] = v
	}
	return tags
}

// returns a new map with the same content values as the input map
func copyTags(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}
	return out
}

// returns a new map with the merged content values of the two input maps
func mergeTags(a map[string]string, b map[string]string) map[string]string {
	c := copyTags(a)
	for k, v := range b {
		c[k] = v
	}
	return c
}

//endregion: tags

//region: filters

func (ecs *Ecs) createContainerNameFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(ecs.ContainerNameInclude, ecs.ContainerNameExclude)
	if err != nil {
		return err
	}
	ecs.containerNameFilter = filter
	return nil
}

func (ecs *Ecs) createLabelFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(ecs.LabelInclude, ecs.LabelExclude)
	if err != nil {
		return err
	}
	ecs.labelFilter = filter
	return nil
}

func (ecs *Ecs) createContainerStatusFilters() error {
	if len(ecs.ContainerStatusInclude) == 0 && len(ecs.ContainerStatusExclude) == 0 {
		ecs.ContainerStatusInclude = []string{"running"}
	}
	filter, err := filter.NewIncludeExcludeFilter(ecs.ContainerStatusInclude, ecs.ContainerStatusExclude)
	if err != nil {
		return err
	}
	ecs.statusFilter = filter
	return nil
}

//endregion: filters

func init() {
	inputs.Add("ecs", func() telegraf.Input {
		return &Ecs{
			EcsV2:          "169.254.170.2",
			Timeout:        internal.Duration{Duration: 5 * time.Second},
			EnvCfg:         true,
			newEnvClient:   NewEnvClient,
			newClient:      NewClient,
			filtersCreated: false,
		}
	})
}
