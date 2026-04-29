//go:generate ../../../tools/readme_config_includer/generator
package docker_log

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/docker"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type DockerLogs struct {
	Endpoint              string          `toml:"endpoint"`
	FromBeginning         bool            `toml:"from_beginning"`
	Timeout               config.Duration `toml:"timeout"`
	LabelInclude          []string        `toml:"docker_label_include"`
	LabelExclude          []string        `toml:"docker_label_exclude"`
	ContainerInclude      []string        `toml:"container_name_include"`
	ContainerExclude      []string        `toml:"container_name_exclude"`
	ContainerStateInclude []string        `toml:"container_state_include"`
	ContainerStateExclude []string        `toml:"container_state_exclude"`
	IncludeSourceTag      bool            `toml:"source_tag"`

	common_tls.ClientConfig

	client          *client.Client
	labelFilter     filter.Filter
	containerFilter filter.Filter
	stateFilter     filter.Filter
	wg              sync.WaitGroup
	mu              sync.Mutex
	containerList   map[string]context.CancelFunc

	// State of the plugin mapping container-ID to the timestamp of the
	// last record processed
	lastRecord    map[string]time.Time
	lastRecordMtx sync.Mutex
}

func (*DockerLogs) SampleConfig() string {
	return sampleConfig
}

func (d *DockerLogs) Init() error {
	if d.Endpoint == "" {
		d.Endpoint = "unix:///var/run/docker.sock"
	}

	switch d.Endpoint {
	case "ENV":
		c, err := client.New(client.FromEnv)
		if err != nil {
			return fmt.Errorf("creating client from environment failed: %w", err)
		}
		d.client = c
	default:
		options := []client.Opt{
			client.WithUserAgent("engine-api-cli-1.0"),
			client.WithHost(d.Endpoint),
		}
		tlsConfig, err := d.ClientConfig.TLSConfig()
		if err != nil {
			return fmt.Errorf("creating TLS configuration failed: %w", err)
		}
		if tlsConfig != nil {
			transport := &http.Transport{TLSClientConfig: tlsConfig}
			options = append(options, client.WithHTTPClient(&http.Client{Transport: transport}))
		}
		c, err := client.New(options...)
		if err != nil {
			return fmt.Errorf("creating client failed: %w", err)
		}
		d.client = c
	}

	// Create label filter
	labelFilter, err := filter.NewIncludeExcludeFilter(d.LabelInclude, d.LabelExclude)
	if err != nil {
		return fmt.Errorf("creating label filter failed: %w", err)
	}
	d.labelFilter = labelFilter

	// Create container filter
	containerFilter, err := filter.NewIncludeExcludeFilter(d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return fmt.Errorf("creating container filter failed: %w", err)
	}
	d.containerFilter = containerFilter

	// Create container state filter
	if len(d.ContainerStateInclude) == 0 && len(d.ContainerStateExclude) == 0 {
		d.ContainerStateInclude = []string{"running"}
	}
	stateFilter, err := filter.NewIncludeExcludeFilter(d.ContainerStateInclude, d.ContainerStateExclude)
	if err != nil {
		return fmt.Errorf("creating container state filter failed: %w", err)
	}
	d.stateFilter = stateFilter

	// Allocate maps
	d.lastRecord = make(map[string]time.Time)
	d.containerList = make(map[string]context.CancelFunc)

	return nil
}

// Start is a noop which is required for a *DockerLogs to implement the telegraf.ServiceInput interface
func (*DockerLogs) Start(telegraf.Accumulator) error {
	return nil
}

func (d *DockerLogs) Stop() {
	d.mu.Lock()
	for _, cancel := range d.containerList {
		cancel()
	}
	d.mu.Unlock()
	d.wg.Wait()
}

func (d *DockerLogs) GetState() interface{} {
	d.lastRecordMtx.Lock()
	recordOffsets := make(map[string]time.Time, len(d.lastRecord))
	for k, v := range d.lastRecord {
		recordOffsets[k] = v
	}
	d.lastRecordMtx.Unlock()

	return recordOffsets
}

func (d *DockerLogs) SetState(state interface{}) error {
	recordOffsets, ok := state.(map[string]time.Time)
	if !ok {
		return fmt.Errorf("state has wrong type %T", state)
	}
	d.lastRecordMtx.Lock()
	for k, v := range recordOffsets {
		d.lastRecord[k] = v
	}
	d.lastRecordMtx.Unlock()

	return nil
}

func (d *DockerLogs) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()
	acc.SetPrecision(time.Nanosecond)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.Timeout))
	defer cancel()

	containers, err := d.client.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		return fmt.Errorf("listing containers failed: %w", err)
	}

	for _, cntnr := range containers.Items {
		d.mu.Lock()
		_, found := d.containerList[cntnr.ID]
		d.mu.Unlock()
		if found {
			continue
		}

		containerName := parseContainerName(cntnr.Names)
		if containerName == "" {
			continue
		}

		if !d.containerFilter.Match(containerName) {
			continue
		}

		if !d.stateFilter.Match(string(cntnr.State)) {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		d.mu.Lock()
		d.containerList[cntnr.ID] = cancel
		d.mu.Unlock()

		// Start a new goroutine for every new container that has logs to collect
		d.wg.Add(1)
		go func(container container.Summary) {
			defer d.wg.Done()
			defer func() {
				d.mu.Lock()
				defer d.mu.Unlock()
				delete(d.containerList, container.ID)
			}()

			if err := d.tailContainerLogs(ctx, acc, container, containerName); err != nil {
				if !errors.Is(err, context.Canceled) {
					acc.AddError(err)
				}
			}
		}(cntnr)
	}
	return nil
}

func (d *DockerLogs) hasTTY(ctx context.Context, cntnr container.Summary) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.Timeout))
	defer cancel()

	result, err := d.client.ContainerInspect(ctx, cntnr.ID, client.ContainerInspectOptions{})
	if err != nil {
		return false, fmt.Errorf("inspecting container %q failed: %w", cntnr.ID, err)
	}

	return result.Container.Config.Tty, nil
}

func (d *DockerLogs) tailContainerLogs(
	ctx context.Context,
	acc telegraf.Accumulator,
	cntnr container.Summary,
	name string,
) error {
	imageName, imageVersion := docker.ParseImage(cntnr.Image)
	tags := map[string]string{
		"container_name":    name,
		"container_image":   imageName,
		"container_version": imageVersion,
	}

	if d.IncludeSourceTag {
		tags["source"] = hostnameFromID(cntnr.ID)
	}

	// Add matching container labels as tags
	for k, label := range cntnr.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	hasTTY, err := d.hasTTY(ctx, cntnr)
	if err != nil {
		return err
	}

	since := time.Time{}.Format(time.RFC3339Nano)
	if !d.FromBeginning {
		d.lastRecordMtx.Lock()
		if ts, ok := d.lastRecord[cntnr.ID]; ok {
			since = ts.Format(time.RFC3339Nano)
		}
		d.lastRecordMtx.Unlock()
	}

	logOptions := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Since:      since,
	}

	logReader, err := d.client.ContainerLogs(ctx, cntnr.ID, logOptions)
	if err != nil {
		return err
	}

	// If the container is using a TTY, there is only a single stream
	// (stdout), and data is copied directly from the container output stream,
	// no extra multiplexing or headers.
	//
	// If the container is *not* using a TTY, streams for stdout and stderr are
	// multiplexed.
	var last time.Time
	if hasTTY {
		last, err = tailStream(acc, tags, cntnr.ID, logReader, "tty")
	} else {
		last, err = tailMultiplexed(acc, tags, cntnr.ID, logReader)
	}
	if err != nil {
		return err
	}

	if ts, ok := d.lastRecord[cntnr.ID]; !ok || ts.Before(last) {
		d.lastRecordMtx.Lock()
		d.lastRecord[cntnr.ID] = last
		d.lastRecordMtx.Unlock()
	}

	return nil
}

func init() {
	inputs.Add("docker_log", func() telegraf.Input {
		return &DockerLogs{
			Timeout: config.Duration(time.Second * 5),
		}
	})
}
