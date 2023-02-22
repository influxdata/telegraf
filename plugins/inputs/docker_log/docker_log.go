//go:generate ../../../tools/readme_config_includer/generator
package docker_log

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/docker"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultEndpoint = "unix:///var/run/docker.sock"
)

var (
	containerStates = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
	// ensure *DockerLogs implements telegraf.ServiceInput
	_ telegraf.ServiceInput = (*DockerLogs)(nil)
)

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

	tlsint.ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client          Client
	labelFilter     filter.Filter
	containerFilter filter.Filter
	stateFilter     filter.Filter
	opts            types.ContainerListOptions
	wg              sync.WaitGroup
	mu              sync.Mutex
	containerList   map[string]context.CancelFunc
}

func (*DockerLogs) SampleConfig() string {
	return sampleConfig
}

func (d *DockerLogs) Init() error {
	var err error
	if d.Endpoint == "ENV" {
		d.client, err = d.newEnvClient()
		if err != nil {
			return err
		}
	} else {
		tlsConfig, err := d.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		d.client, err = d.newClient(d.Endpoint, tlsConfig)
		if err != nil {
			return err
		}
	}

	// Create filters
	err = d.createLabelFilters()
	if err != nil {
		return err
	}
	err = d.createContainerFilters()
	if err != nil {
		return err
	}
	err = d.createContainerStateFilters()
	if err != nil {
		return err
	}

	filterArgs := filters.NewArgs()
	for _, state := range containerStates {
		if d.stateFilter.Match(state) {
			filterArgs.Add("status", state)
		}
	}

	if filterArgs.Len() != 0 {
		d.opts = types.ContainerListOptions{
			Filters: filterArgs,
		}
	}

	return nil
}

func (d *DockerLogs) addToContainerList(containerID string, cancel context.CancelFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.containerList[containerID] = cancel
}

func (d *DockerLogs) removeFromContainerList(containerID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.containerList, containerID)
}

func (d *DockerLogs) containerInContainerList(containerID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.containerList[containerID]
	return ok
}

func (d *DockerLogs) cancelTails() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, cancel := range d.containerList {
		cancel()
	}
}

func (d *DockerLogs) matchedContainerName(names []string) string {
	// Check if all container names are filtered; in practice I believe
	// this array is always of length 1.
	for _, name := range names {
		trimmedName := strings.TrimPrefix(name, "/")
		match := d.containerFilter.Match(trimmedName)
		if match {
			return trimmedName
		}
	}
	return ""
}

func (d *DockerLogs) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()
	acc.SetPrecision(time.Nanosecond)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.Timeout))
	defer cancel()
	containers, err := d.client.ContainerList(ctx, d.opts)
	if err != nil {
		return err
	}

	for _, container := range containers {
		if d.containerInContainerList(container.ID) {
			continue
		}

		containerName := d.matchedContainerName(container.Names)
		if containerName == "" {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		d.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		d.wg.Add(1)
		go func(container types.Container) {
			defer d.wg.Done()
			defer d.removeFromContainerList(container.ID)

			err = d.tailContainerLogs(ctx, acc, container, containerName)
			if err != nil && err != context.Canceled {
				acc.AddError(err)
			}
		}(container)
	}
	return nil
}

func (d *DockerLogs) hasTTY(ctx context.Context, container types.Container) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.Timeout))
	defer cancel()
	c, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

func (d *DockerLogs) tailContainerLogs(
	ctx context.Context,
	acc telegraf.Accumulator,
	container types.Container,
	containerName string,
) error {
	imageName, imageVersion := docker.ParseImage(container.Image)
	tags := map[string]string{
		"container_name":    containerName,
		"container_image":   imageName,
		"container_version": imageVersion,
	}

	if d.IncludeSourceTag {
		tags["source"] = hostnameFromID(container.ID)
	}

	// Add matching container labels as tags
	for k, label := range container.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	hasTTY, err := d.hasTTY(ctx, container)
	if err != nil {
		return err
	}

	tail := "0"
	if d.FromBeginning {
		tail = "all"
	}

	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Tail:       tail,
	}

	logReader, err := d.client.ContainerLogs(ctx, container.ID, logOptions)
	if err != nil {
		return err
	}

	// If the container is using a TTY, there is only a single stream
	// (stdout), and data is copied directly from the container output stream,
	// no extra multiplexing or headers.
	//
	// If the container is *not* using a TTY, streams for stdout and stderr are
	// multiplexed.
	if hasTTY {
		return tailStream(acc, tags, container.ID, logReader, "tty")
	}
	return tailMultiplexed(acc, tags, container.ID, logReader)
}

func parseLine(line []byte) (time.Time, string, error) {
	parts := bytes.SplitN(line, []byte(" "), 2)

	if len(parts) == 1 {
		parts = append(parts, []byte(""))
	}

	tsString := string(parts[0])

	// Keep any leading space, but remove whitespace from end of line.
	// This preserves space in, for example, stacktraces, while removing
	// annoying end of line characters and is similar to how other logging
	// plugins such as syslog behave.
	message := bytes.TrimRightFunc(parts[1], unicode.IsSpace)

	ts, err := time.Parse(time.RFC3339Nano, tsString)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("error parsing timestamp %q: %v", tsString, err)
	}

	return ts, string(message), nil
}

func tailStream(
	acc telegraf.Accumulator,
	baseTags map[string]string,
	containerID string,
	reader io.ReadCloser,
	stream string,
) error {
	defer reader.Close()

	tags := make(map[string]string, len(baseTags)+1)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["stream"] = stream

	r := bufio.NewReaderSize(reader, 64*1024)

	for {
		line, err := r.ReadBytes('\n')

		if len(line) != 0 {
			ts, message, err := parseLine(line)
			if err != nil {
				acc.AddError(err)
			} else {
				acc.AddFields("docker_log", map[string]interface{}{
					"container_id": containerID,
					"message":      message,
				}, tags, ts)
			}
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func tailMultiplexed(
	acc telegraf.Accumulator,
	tags map[string]string,
	containerID string,
	src io.ReadCloser,
) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(acc, tags, containerID, outReader, "stdout")
		if err != nil {
			acc.AddError(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(acc, tags, containerID, errReader, "stderr")
		if err != nil {
			acc.AddError(err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)
	outWriter.Close() //nolint:revive // we cannot do anything if the closing fails
	errWriter.Close() //nolint:revive // we cannot do anything if the closing fails
	src.Close()       //nolint:revive // we cannot do anything if the closing fails
	wg.Wait()
	return err
}

// Start is a noop which is required for a *DockerLogs to implement
// the telegraf.ServiceInput interface
func (d *DockerLogs) Start(telegraf.Accumulator) error {
	return nil
}

func (d *DockerLogs) Stop() {
	d.cancelTails()
	d.wg.Wait()
}

// Following few functions have been inherited from telegraf docker input plugin
func (d *DockerLogs) createContainerFilters() error {
	containerFilter, err := filter.NewIncludeExcludeFilter(d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return err
	}
	d.containerFilter = containerFilter
	return nil
}

func (d *DockerLogs) createLabelFilters() error {
	labelFilter, err := filter.NewIncludeExcludeFilter(d.LabelInclude, d.LabelExclude)
	if err != nil {
		return err
	}
	d.labelFilter = labelFilter
	return nil
}

func (d *DockerLogs) createContainerStateFilters() error {
	if len(d.ContainerStateInclude) == 0 && len(d.ContainerStateExclude) == 0 {
		d.ContainerStateInclude = []string{"running"}
	}
	stateFilter, err := filter.NewIncludeExcludeFilter(d.ContainerStateInclude, d.ContainerStateExclude)
	if err != nil {
		return err
	}
	d.stateFilter = stateFilter
	return nil
}

func init() {
	inputs.Add("docker_log", func() telegraf.Input {
		return &DockerLogs{
			Timeout:       config.Duration(time.Second * 5),
			Endpoint:      defaultEndpoint,
			newEnvClient:  NewEnvClient,
			newClient:     NewClient,
			containerList: make(map[string]context.CancelFunc),
		}
	})
}

func hostnameFromID(id string) string {
	if len(id) > 12 {
		return id[0:12]
	}
	return id
}
