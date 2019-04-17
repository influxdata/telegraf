package docker_log

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"encoding/binary"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type StdType byte

const (
	Stdin StdType = iota
	Stdout
	Stderr
	Systemerr

	stdWriterPrefixLen = 8
	stdWriterFdIndex   = 0
	stdWriterSizeIndex = 4

	startingBufLen = 32*1024 + stdWriterPrefixLen + 1

	ERR_PREFIX = "E! [inputs.docker_log]"
)

type DockerLogs struct {
	Endpoint string

	Timeout internal.Duration

	LabelInclude []string `toml:"docker_label_include"`
	LabelExclude []string `toml:"docker_label_exclude"`

	ContainerInclude []string `toml:"container_name_include"`
	ContainerExclude []string `toml:"container_name_exclude"`

	ContainerStateInclude []string `toml:"container_state_include"`
	ContainerStateExclude []string `toml:"container_state_exclude"`

	tlsint.ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client          Client
	filtersCreated  bool
	labelFilter     filter.Filter
	containerFilter filter.Filter
	stateFilter     filter.Filter
	opts            types.ContainerListOptions
	wg              sync.WaitGroup
	mu              sync.Mutex
	containerList   map[string]io.ReadCloser
}

const (
	defaultEndpoint = "unix:///var/run/docker.sock"
	LOG_BYTES_MAX   = 1000
)

var (
	containerStates = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
)

var sampleConfig = `
  ## Docker Endpoint
  ## To use TCP, set endpoint = "tcp://[ip]:[port]"
  ## To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"
  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  container_name_include = []
  container_name_exclude = []
  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  # container_state_include = []
  # container_state_exclude = []

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  docker_label_include = []
  docker_label_exclude = []

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (d *DockerLogs) Description() string {
	return "Plugin to get docker logs"
}

func (d *DockerLogs) SampleConfig() string {
	return sampleConfig
}

func (d *DockerLogs) Gather(acc telegraf.Accumulator) error {
	/*Check to see if any new containers have been created since last time*/
	d.containerListUpdate(acc)
	return nil
}

/*Following few functions have been inherited from telegraf docker input plugin*/
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

func (d *DockerLogs) addToContainerList(containerId string, logReader io.ReadCloser) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.containerList[containerId] = logReader
	return nil
}

func (d *DockerLogs) removeFromContainerList(containerId string) error {
	delete(d.containerList, containerId)
	return nil
}

func (d *DockerLogs) containerInContainerList(containerId string) bool {
	if _, ok := d.containerList[containerId]; ok {
		return true
	}
	return false
}

func (d *DockerLogs) stopAllReaders() error {
	for _, container := range d.containerList {
		container.Close()
	}
	return nil
}

func (d *DockerLogs) containerListUpdate(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()
	if d.client == nil {
		log.Printf("%s : Dock client is null", ERR_PREFIX)
		return nil
	}
	containers, err := d.client.ContainerList(ctx, d.opts)
	if err != nil {
		log.Printf("%s : %s ", ERR_PREFIX, err.Error())
		return err
	}
	for _, container := range containers {
		if d.containerInContainerList(container.ID) {
			continue
		}
		d.wg.Add(1)
		/*Start a new goroutine for every new container that has logs to collect*/
		go func(c types.Container) {
			defer d.wg.Done()
			err := d.getContainerLogs(c, acc)
			if err != nil {
				log.Printf("%s : %s ", ERR_PREFIX, err.Error())
			}
			d.removeFromContainerList(c.ID)
			return
		}(container)
	}
	return nil
}

func (d *DockerLogs) getContainerLogs(
	container types.Container,
	acc telegraf.Accumulator,
) error {
	c, err := d.client.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		log.Printf("%s : %s", ERR_PREFIX, err.Error())
		return err
	}
	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: false,
		Details:    true,
		Follow:     true,
		Tail:       "0",
	}
	logReader, err := d.client.ContainerLogs(context.Background(), container.ID, logOptions)
	d.addToContainerList(container.ID, logReader)
	if err != nil {
		log.Printf("%s : %s", ERR_PREFIX, err.Error())
		return err
	}
	/* Parse container name */
	cname := "unknown"
	if len(container.Names) > 0 {
		/*Pick first container name*/
		cname = strings.TrimPrefix(container.Names[0], "/")
	}

	tags := map[string]string{
		"containerId":   container.ID,
		"containerName": cname,
	}
	// Add labels to tags
	for k, label := range container.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}
	if c.Config.Tty {
		_, err = pushTtyLogs(acc, tags, logReader)
	} else {
		_, err = pushLogs(acc, tags, logReader)
	}
	if err != nil {
		log.Printf("%s : %s", ERR_PREFIX, err.Error())
		return err
	}
	return nil
}
func pushTtyLogs(acc telegraf.Accumulator, tags map[string]string, src io.Reader) (written int64, err error) {
	tags["logType"] = "unknown" //in tty mode we wont be able to differentiate b/w stdout and stderr hence unknown
	fields := map[string]interface{}{}
	data := make([]byte, LOG_BYTES_MAX)
	for {
		num, err := src.Read(data)
		if err != nil {
			if err == io.EOF {
				written += int64(num)
				fields["log"] = data[1:num]
				acc.AddFields("docker_log", fields, tags)
				return written, nil
			}
			return written, err
		}
		written += int64(num)
		if len(data) > 0 {
			fields["log"] = data[1:num]
			acc.AddFields("docker_log", fields, tags)
		}
	}
}

/* Inspired from https://github.com/moby/moby/blob/master/pkg/stdcopy/stdcopy.go */
func pushLogs(acc telegraf.Accumulator, tags map[string]string, src io.Reader) (written int64, err error) {
	var (
		buf       = make([]byte, startingBufLen)
		bufLen    = len(buf)
		nr        int
		er        error
		frameSize int
	)
	for {
		// Make sure we have at least a full header
		for nr < stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}
		stream := StdType(buf[stdWriterFdIndex])
		// Check the first byte to know where to write
		var logType string
		switch stream {
		case Stdin:
			logType = "stdin"
			break
		case Stdout:
			logType = "stdout"
			break
		case Stderr:
			logType = "stderr"
			break
		case Systemerr:
			fallthrough
		default:
			return 0, fmt.Errorf("Unrecognized input header: %d", buf[stdWriterFdIndex])
		}
		// Retrieve the size of the frame
		frameSize = int(binary.BigEndian.Uint32(buf[stdWriterSizeIndex : stdWriterSizeIndex+4]))

		// Check if the buffer is big enough to read the frame.
		// Extend it if necessary.
		if frameSize+stdWriterPrefixLen > bufLen {
			buf = append(buf, make([]byte, frameSize+stdWriterPrefixLen-bufLen+1)...)
			bufLen = len(buf)
		}

		// While the amount of bytes read is less than the size of the frame + header, we keep reading
		for nr < frameSize+stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < frameSize+stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}

		// we might have an error from the source mixed up in our multiplexed
		// stream. if we do, return it.
		if stream == Systemerr {
			return written, fmt.Errorf("error from daemon in stream: %s", string(buf[stdWriterPrefixLen:frameSize+stdWriterPrefixLen]))
		}

		tags["logType"] = logType
		fields := map[string]interface{}{}
		fields["log"] = buf[stdWriterPrefixLen+1 : frameSize+stdWriterPrefixLen]
		acc.AddFields("docker_log", fields, tags)
		written += int64(frameSize)

		// Move the rest of the buffer to the beginning
		copy(buf, buf[frameSize+stdWriterPrefixLen:])
		// Move the index
		nr -= frameSize + stdWriterPrefixLen
	}
}

func (d *DockerLogs) Start(acc telegraf.Accumulator) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	var c Client
	var err error
	if d.Endpoint == "ENV" {
		c, err = d.newEnvClient()
	} else {
		tlsConfig, err := d.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		c, err = d.newClient(d.Endpoint, tlsConfig)
	}
	if err != nil {
		log.Printf("%s : %s", ERR_PREFIX, err.Error())
		return err
	}
	d.client = c
	// Create label filters if not already created
	if !d.filtersCreated {
		err := d.createLabelFilters()
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
		d.filtersCreated = true
	}
	filterArgs := filters.NewArgs()
	for _, state := range containerStates {
		if d.stateFilter.Match(state) {
			filterArgs.Add("status", state)
		}
	}

	// All container states were excluded
	if filterArgs.Len() == 0 {
		return nil
	}

	d.opts = types.ContainerListOptions{
		Filters: filterArgs,
	}
	return nil
}

func (d *DockerLogs) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopAllReaders()
	d.wg.Wait()
}

func init() {
	inputs.Add("docker_log", func() telegraf.Input {
		return &DockerLogs{
			Timeout:        internal.Duration{Duration: time.Second * 5},
			Endpoint:       defaultEndpoint,
			newEnvClient:   NewEnvClient,
			newClient:      NewClient,
			filtersCreated: false,
			containerList:  make(map[string]io.ReadCloser),
		}
	})
}
