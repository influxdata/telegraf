package containerapp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	docker "github.com/docker/docker/client"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
)

var (
	version        string
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type Docker struct {
	client        *client.Client
	syncinterval  internal.Duration
	startinterval internal.Duration
	containers    map[string]*Config
	Add           func(id string, conf *Config) error
	Del           func(id string)
	Error         func(err error)
}

func NewDocker(
	host string,
	startinterval internal.Duration,
	syncinterval internal.Duration,
	Add func(id string, conf *Config) error,
	Del func(id string),
	Error func(err error),
) (*Docker, error) {

	tlsclient := tlsint.ClientConfig{}
	tlsConfig, err := tlsclient.TLSConfig()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport}

	cli, err := docker.NewClientWithOpts(
		docker.WithHTTPHeaders(defaultHeaders),
		docker.WithHTTPClient(httpClient),
		docker.WithVersion(version),
		docker.WithHost(host))

	if err != nil {
		return nil, err
	}
	return &Docker{
		client:        cli,
		syncinterval:  syncinterval,
		startinterval: startinterval,
		Add:           Add,
		Del:           Del,
		Error:         Error,
	}, nil
}

type ContainerInfo struct {
	env  map[string]string
	tags map[string]string
	ip   string
}

func (dc *Docker) add(id string) {
	conf, err := dc.сreateConf(id)
	if err != nil {
		return
	}

	err = dc.Add(id, conf)
	if err != nil {
		return
	}

	dc.containers[id] = conf

}
func (dc *Docker) del(id string) {
	delete(dc.containers, id)
	dc.Del(id)
}
func (dc *Docker) error(err error) {
	dc.Error(err)
}

func (dc *Docker) getContainerInfo(
	containerid string,
) (*ContainerInfo, error) {
	ctxInspect, cancelInspect := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelInspect()

	info, err := dc.client.ContainerInspect(ctxInspect, containerid)
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err.Error())
	}
	values := map[string]string{}

	for _, envvar := range info.Config.Env {
		dockEnv := strings.SplitN(envvar, "=", 2)
		values[dockEnv[0]] = dockEnv[1]
	}
	var net *network.EndpointSettings
	for _, network := range info.NetworkSettings.Networks {
		net = network
		break
	}
	if net == nil {
		return nil, fmt.Errorf("Error: can't get network")
	}

	tags := map[string]string{
		"containerid":    containerid,
		"container_name": info.Name,
	}

	return &ContainerInfo{
		env:  values,
		ip:   net.IPAddress,
		tags: tags,
	}, nil
}

func (dc *Docker) сreateConf(
	containerid string,
) (*Config, error) {

	containerInfo, err := dc.getContainerInfo(containerid)
	if err != nil {
		return nil, err
	}

	conf := &Config{
		Name:       containerid,
		IP:         containerInfo.ip,
		Values:     []map[string]string{containerInfo.env},
		SystemTags: containerInfo.tags,
	}

	return conf, nil
}

func (dc *Docker) containerList(ctx context.Context, opts types.ContainerListOptions) ([]types.Container, error) {
	return dc.client.ContainerList(ctx, opts)
}

func (dc *Docker) events(ctx context.Context, opts types.EventsOptions) (<-chan events.Message, <-chan error) {
	return dc.client.Events(ctx, opts)
}

func checkContainerID(containers []types.Container, id string) bool {
	for i := range containers {
		if containers[i].ID == id {
			return true
		}
	}
	return false
}

func (dc *Docker) syncClinets(ctx context.Context, opts types.ContainerListOptions) error {

	containers, err := dc.containerList(ctx, opts)
	if err != nil {
		log.Printf("E! containerapp input: %s", err.Error())
		return err
	}

	for _, container := range containers {
		if _, ok := dc.containers[container.ID]; !ok {
			time.Sleep(dc.startinterval.Duration)
			dc.add(container.ID)
		}
	}

	for containerid := range dc.containers {
		if !checkContainerID(containers, containerid) {
			dc.del(containerid)
		}
	}

	return nil
}

func (dc *Docker) Run() {

	defer func() {
		log.Printf("E! containerapp: docker connect error, restart")
		for containerID := range dc.containers {
			dc.del(containerID)
		}
		time.Sleep(5 * time.Second)
		go dc.Run()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := types.ContainerListOptions{}

	err := dc.syncClinets(ctx, opts)
	if err != nil {
		log.Printf("E! containerapp input: %s", err.Error())
		return
	}

	filters := filters.NewArgs()
	filters.Add("type", "container")
	options := types.EventsOptions{
		Filters: filters,
	}

	eventsCh, errEventsCh := dc.events(ctx, options)
	ticker := time.NewTicker(dc.syncinterval.Duration)

	for {
		select {
		case event := <-eventsCh:
			if event.Status == "start" {
				time.Sleep(dc.startinterval.Duration)
				dc.add(event.ID)
			} else if event.Status == "stop" {
				dc.del(event.ID)
			}

		case err := <-errEventsCh:
			log.Printf("E! containerapp events: %s", err.Error())
			dc.error(err)
			return

		case <-ticker.C:
			err := dc.syncClinets(ctx, opts)
			if err != nil {
				log.Printf("E! containerapp input: %s", err.Error())
				return
			}
		}
	}
}
