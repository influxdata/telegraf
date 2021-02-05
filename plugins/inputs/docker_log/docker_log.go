package docker_log

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/docker"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  # endpoint = "unix:///var/run/docker.sock"

  ## When true, container logs are read from the beginning; otherwise
  ## reading begins at the end of the log.
  # from_beginning = false

  ## Timeout for Docker API calls.
  # timeout = "5s"

  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  # container_name_include = []
  # container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  # container_state_include = []
  # container_state_exclude = []

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  # docker_label_include = []
  # docker_label_exclude = []

  ## Set the source tag for the metrics to the container ID hostname, eg first 12 chars
  source_tag = false

  ## Offset flush interval. How often the offset pointer (see below) in the
  ## container log stream is flashed to file. Offset pointer represents the unix time stamp
  ## in nano seconds for the last message read from log stream (default - 3 sec)
  # offset_flush = "3s"

  ## Offset storage path, make sure the user on behalf 
  ## of which the telegraf is running has enough rights to read and write to chosen path.
  ## default value is (cross-platform): "$HOME/telegraf" with fallback to 
  ## "$OS_TEMP/telegraf", if user can't be detected.
  # offset_storage_path = "/var/run/telegraf"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

const (
	defaultEndpoint = "unix:///var/run/docker.sock"

	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024
)

var (
	containerStates = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
	// ensure *DockerLogs implements telegraf.ServiceInput
	_ telegraf.ServiceInput = (*DockerLogs)(nil)
)

type DockerLogs struct {
	Endpoint              string            `toml:"endpoint"`
	FromBeginning         bool              `toml:"from_beginning"`
	Timeout               internal.Duration `toml:"timeout"`
	LabelInclude          []string          `toml:"docker_label_include"`
	LabelExclude          []string          `toml:"docker_label_exclude"`
	ContainerInclude      []string          `toml:"container_name_include"`
	ContainerExclude      []string          `toml:"container_name_exclude"`
	ContainerStateInclude []string          `toml:"container_state_include"`
	ContainerStateExclude []string          `toml:"container_state_exclude"`
	IncludeSourceTag      bool              `toml:"source_tag"`
	OffsetFlush           internal.Duration `toml:"offset_flush"`
	OffsetStoragePath     string            `toml:"offset_storage_path"`

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
	offsetChan      chan offsetData
	wgOffsetFlusher sync.WaitGroup
}

type offsetData struct {
	contID string
	offset int64
}

func (d *DockerLogs) Description() string {
	return "Read logging output from the Docker engine"
}

func (d *DockerLogs) SampleConfig() string {
	return sampleConfig
}

func (d *DockerLogs) Init() error {
	var err error
	var src os.FileInfo

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

	//Get default storage path
	if d.OffsetStoragePath == "" {
		user, err := user.Current()
		if err != nil {
			d.OffsetStoragePath = path.Join(os.TempDir(), "telegraf")
		} else {
			d.OffsetStoragePath = path.Join(user.HomeDir, "telegraf")
		}
		log.Printf("W! Offset storage path set to: %q", d.OffsetStoragePath)
	}
	//Create storage path
	src, err = os.Stat(d.OffsetStoragePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(d.OffsetStoragePath, 0777); err != nil {
			return fmt.Errorf("can't create path '%s' to store offset: %s", d.OffsetStoragePath, err.Error())
		}
	} else if src != nil && src.Mode().IsRegular() {
		return fmt.Errorf("'%s' already exists as a file!", d.OffsetStoragePath)
	}

	//Start offset flusher
	d.wgOffsetFlusher.Add(1)
	go d.flushOffset()

	return nil
}

func (d *DockerLogs) addToContainerList(containerID string, cancel context.CancelFunc) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.containerList[containerID] = cancel
	return nil
}

func (d *DockerLogs) removeFromContainerList(containerID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.containerList, containerID)
	return nil
}

func (d *DockerLogs) containerInContainerList(containerID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.containerList[containerID]
	return ok
}

func (d *DockerLogs) cancelTails() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, cancel := range d.containerList {
		cancel()
	}
	return nil
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

	ctx, cancel := context.WithTimeout(ctx, d.Timeout.Duration)
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
	ctx, cancel := context.WithTimeout(ctx, d.Timeout.Duration)
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

	//Detecting offset
	tailLogsSince, initialOffset := d.loadOffsetFormFs(container.ID)

	if tailLogsSince != "" && d.FromBeginning { // If there is an offset, then it means that we already deliver logs until the offset
		//In this case we can ignore 'fromBeginning' ,and continue from the offset
		log.Printf("D! [inputs.docker_log] Container '%s', 'from_beginning' option ignored, since there is an offset: %s", hostnameFromID(container.ID), tailLogsSince)

	} else if tailLogsSince == "" && !d.FromBeginning { //No offset and not from beginning, means that we ship logs since now.
		tailLogsSince = time.Now().UTC().Format(time.RFC3339Nano)
	}

	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Since:      tailLogsSince,
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
		return tailStream(acc, tags, container.ID, logReader, "tty", d.offsetChan, initialOffset)
	} else {
		return tailMultiplexed(acc, tags, container.ID, logReader, d.offsetChan, initialOffset)
	}
}

func parseLine(line []byte) (time.Time, string, error) {
	parts := bytes.SplitN(line, []byte(" "), 2)

	switch len(parts) {
	case 1:
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
	offsetChan chan<- offsetData,
	initialOffset int64,
) error {

	messageOffset := initialOffset

	defer func() {
		reader.Close()
		//Sending last offset to flusher in a blocking manner
		if messageOffset != initialOffset {
			offsetChan <- offsetData{containerID, messageOffset}
		}
	}()

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

				messageOffset = ts.UTC().UnixNano() + 1 //+1 (ns) here prevents to include current message

				//Send offset to flusher (in a non blocking manner)
				select {
				case offsetChan <- offsetData{containerID, messageOffset}:
				default:
				}
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
	offsetChan chan<- offsetData,
	initialOffset int64,
) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(acc, tags, containerID, outReader, "stdout", offsetChan, initialOffset)
		if err != nil {
			acc.AddError(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(acc, tags, containerID, errReader, "stderr", offsetChan, initialOffset)
		if err != nil {
			acc.AddError(err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)
	outWriter.Close()
	errWriter.Close()
	src.Close()
	wg.Wait()
	return err
}

func (d *DockerLogs) flushOffsetToFs(contID string, offset int64) error {
	fileName := path.Join(d.OffsetStoragePath, contID)
	offsetByte := []byte(strconv.FormatInt(offset, 10))
	err := ioutil.WriteFile(fileName, offsetByte, 0777)
	if err != nil {
		log.Printf("D! [inputs.docker_log] Can't write message offset to file %q: %v", fileName, err)
	}
	return err
}
func (d *DockerLogs) loadOffsetFormFs(contID string) (offsetString string, offsetInt int64) {
	var (
		err  error
		data []byte
	)
	fileName := path.Join(d.OffsetStoragePath, contID)

	if _, err = os.Stat(fileName); !os.IsNotExist(err) {
		data, err = ioutil.ReadFile(fileName)
		if err != nil {
			log.Printf("E! [inputs.docker_log] Can't read message offset from file %q: %v", fileName, err)
			return "", 0
		}

		offsetInt, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			log.Printf("E! [inputs.docker_log] Can't parse integer from offset file %q content %q: %v", fileName, string(data), err)
			return "", 0
		}
		offsetString = time.Unix(0, offsetInt).UTC().Format(time.RFC3339Nano)

		//log.Printf("D! [inputs.docker_log] Parsed offset from '%s'\nvalue: %s, %s",
		//	fileName, string(data), offsetString)
		return offsetString, offsetInt
	}
	return "", 0
}

func (d *DockerLogs) flushOffset() {
	var containerOffsets = map[string]int64{}
	var ticker = time.NewTicker(d.OffsetFlush.Duration)

	defer ticker.Stop()
	defer d.wgOffsetFlusher.Done()

	for offset := range d.offsetChan {
		containerOffsets[offset.contID] = offset.offset
		select {
		case <-ticker.C:
			//flushing to filesystem
			for contID, offsetInt := range containerOffsets {
				if d.flushOffsetToFs(contID, offsetInt) == nil {
					delete(containerOffsets, contID)
				}
			}
		default:
		}
	}
	//Check if there is smth. that is not flushed:
	for contID, offsetInt := range containerOffsets {
		d.flushOffsetToFs(contID, offsetInt)
	}
}

// Start is a noop which is required for a *DockerLogs to implement
// the telegraf.ServiceInput interface
func (d *DockerLogs) Start(telegraf.Accumulator) error {
	return nil
}

func (d *DockerLogs) Stop() {
	d.cancelTails()
	d.wg.Wait()

	//Stop offset flushing
	close(d.offsetChan)
	d.wgOffsetFlusher.Wait()
}

// Following few functions have been inherited from telegraf docker input plugin
func (d *DockerLogs) createContainerFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return err
	}
	d.containerFilter = filter
	return nil
}

func (d *DockerLogs) createLabelFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(d.LabelInclude, d.LabelExclude)
	if err != nil {
		return err
	}
	d.labelFilter = filter
	return nil
}

func (d *DockerLogs) createContainerStateFilters() error {
	if len(d.ContainerStateInclude) == 0 && len(d.ContainerStateExclude) == 0 {
		d.ContainerStateInclude = []string{"running"}
	}
	filter, err := filter.NewIncludeExcludeFilter(d.ContainerStateInclude, d.ContainerStateExclude)
	if err != nil {
		return err
	}
	d.stateFilter = filter
	return nil
}

func init() {
	inputs.Add("docker_log", func() telegraf.Input {
		return &DockerLogs{
			Timeout:       internal.Duration{Duration: time.Second * 5},
			Endpoint:      defaultEndpoint,
			newEnvClient:  NewEnvClient,
			newClient:     NewClient,
			containerList: make(map[string]context.CancelFunc),
			offsetChan:    make(chan offsetData),
			OffsetFlush:   internal.Duration{Duration: 3 * time.Second},
		}
	})
}

func hostnameFromID(id string) string {
	if len(id) > 12 {
		return id[0:12]
	}
	return id
}
