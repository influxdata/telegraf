package ecs

import (
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Ecs config object
type Ecs struct {
	EndpointURL string `toml:"endpoint_url"`
	Timeout     internal.Duration

	ContainerNameInclude []string `toml:"container_name_include"`
	ContainerNameExclude []string `toml:"container_name_exclude"`

	ContainerStatusInclude []string `toml:"container_status_include"`
	ContainerStatusExclude []string `toml:"container_status_exclude"`

	LabelInclude []string `toml:"ecs_label_include"`
	LabelExclude []string `toml:"ecs_label_exclude"`

	newClient func(timeout time.Duration) (*EcsClient, error)

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

var sampleConfig = `
  ## ECS metadata url
  # endpoint_url = "http://169.254.170.2"

  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  # container_name_include = []
  # container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "RUNNING" state will be captured.
  ## Possible values are "NONE", "PULLED", "CREATED", "RUNNING",
  ## "RESOURCES_PROVISIONED", "STOPPED".
  # container_status_include = []
  # container_status_exclude = []

  ## ecs labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  ecs_label_include = [ "com.amazonaws.ecs.*" ]
  ecs_label_exclude = []

  ## Timeout for queries.
  # timeout = "5s"
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

	mergeTaskStats(task, stats)

	taskTags := map[string]string{
		"cluster":  task.Cluster,
		"task_arn": task.TaskARN,
		"family":   task.Family,
		"revision": task.Revision,
	}

	// accumulate metrics
	ecs.accTask(task, taskTags, acc)
	ecs.accContainers(task, taskTags, acc)

	return nil
}

func initSetup(ecs *Ecs) error {
	if ecs.client == nil {
		var err error
		var c *EcsClient
		c, err = ecs.newClient(ecs.Timeout.Duration)
		if err != nil {
			return err
		}

		c.BaseURL, err = url.Parse(ecs.EndpointURL)
		if err != nil {
			return err
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

func (ecs *Ecs) accTask(task *Task, tags map[string]string, acc telegraf.Accumulator) {
	taskFields := map[string]interface{}{
		"revision":       task.Revision,
		"desired_status": task.DesiredStatus,
		"known_status":   task.KnownStatus,
		"limit_cpu":      task.Limits["CPU"],
		"limit_mem":      task.Limits["Memory"],
	}

	acc.AddFields("ecs_task", taskFields, tags, task.PullStoppedAt)
}

func (ecs *Ecs) accContainers(task *Task, taskTags map[string]string, acc telegraf.Accumulator) {
	for _, c := range task.Containers {
		if !ecs.containerNameFilter.Match(c.Name) {
			continue
		}

		if !ecs.statusFilter.Match(strings.ToUpper(c.KnownStatus)) {
			continue
		}

		// add matching ECS container Labels
		containerTags := map[string]string{
			"id":   c.ID,
			"name": c.Name,
		}
		for k, v := range c.Labels {
			if ecs.labelFilter.Match(k) {
				containerTags[k] = v
			}
		}
		tags := mergeTags(taskTags, containerTags)

		parseContainerStats(c, acc, tags)
	}
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
		ecs.ContainerStatusInclude = []string{"RUNNING"}
	}

	// ECS uses uppercase status names, normalizing for comparison.
	for i, include := range ecs.ContainerStatusInclude {
		ecs.ContainerStatusInclude[i] = strings.ToUpper(include)
	}
	for i, exclude := range ecs.ContainerStatusExclude {
		ecs.ContainerStatusExclude[i] = strings.ToUpper(exclude)
	}

	filter, err := filter.NewIncludeExcludeFilter(ecs.ContainerStatusInclude, ecs.ContainerStatusExclude)
	if err != nil {
		return err
	}
	ecs.statusFilter = filter
	return nil
}

func init() {
	inputs.Add("ecs", func() telegraf.Input {
		return &Ecs{
			EndpointURL:    "http://169.254.170.2",
			Timeout:        internal.Duration{Duration: 5 * time.Second},
			newClient:      NewClient,
			filtersCreated: false,
		}
	})
}
