package containerapp

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Config struct {
	Name       string
	IP         string
	Values     []map[string]string
	SystemTags map[string]string
}

// ContainerApp struct
type ContainerApp struct {
	ConfigType    string            `toml:"config_type"`
	StartInterval internal.Duration `toml:"start_interval"`
	SyncInterval  internal.Duration `toml:"sync_interval"`
	Tags          []string          `toml:"tags_name"`
	TagsMandatory []string          `toml:"tags_mandatory"`
	TagsPrefix    string            `toml:"tags_prefix"`
	DockerEnv     map[string]string `toml:"docker_env"`
	Kubernetes    map[string]string `toml:"kubernetes"`
	HTTP          map[string]string `toml:"http"`
	HTTPDefaults  map[string]string `toml:"http_defaults"`

	metrics     []Metric
	metricsLock *sync.Mutex
	errors      []error
	clients     map[string]*HTTPGather
	metricsCh   chan Metric
	errCh       chan error
}

func NewContainerApp() *ContainerApp {
	metrics := []Metric{}
	clients := make(map[string]*HTTPGather)
	metricsCh := make(chan Metric)
	errCh := make(chan error)

	return &ContainerApp{
		metrics:   metrics,
		clients:   clients,
		metricsCh: metricsCh,
		StartInterval: internal.Duration{
			Duration: 10 * time.Millisecond,
		},
		SyncInterval: internal.Duration{
			Duration: 10 * time.Minute,
		},
		errCh:       errCh,
		metricsLock: &sync.Mutex{},
	}
}

func checkMandatoryTags(mandatoryValues []string, valuesList []map[string]string) (map[string]string, error) {
OUTER:
	for valuesIndex, values := range valuesList {

		for _, name := range mandatoryValues {
			if _, ok := values[name]; !ok {
				if valuesIndex < len(valuesList) {
					continue OUTER
				}
				return nil, fmt.Errorf("%s is mandatory ", name)
			}
		}

		return values, nil

	}
	return nil, fmt.Errorf("there is no applicable values")
}

func (s *ContainerApp) Add(id string, conf *Config) error {

	metaValues, err := checkMandatoryTags(s.TagsMandatory, conf.Values)

	if err != nil {
		return fmt.Errorf("skip object: %s, %s", conf.Name, err)
	}

	clientcfg, err := CreateHTTPGatherConf(
		conf.Name,
		s.HTTP,
		s.HTTPDefaults,
		metaValues,
	)

	if err != nil {
		return err
	}

	clientcfg.IP = conf.IP

	if clientcfg.Tags == nil {
		clientcfg.Tags = map[string]string{}
	}

	for tagkey, tagvalue := range conf.SystemTags {
		clientcfg.Tags[tagkey] = tagvalue
	}

	if len(s.Tags) > 0 {
		for _, tag := range s.Tags {
			if val, ok := metaValues[tag]; ok {
				clientcfg.Tags[tag] = val
			}
		}
	}
	if len(s.TagsPrefix) > 0 {
		for name, val := range metaValues {
			if strings.HasPrefix(name, s.TagsPrefix) {
				clientcfg.Tags[name] = val
			}
		}
	}

	client, err := NewHTTPGather(s, id, clientcfg)
	if err != nil {
		return err
	}
	client.Run()
	s.clients[client.id] = client
	log.Println("Add", len(s.clients), "clients connected.")
	return nil
}

func (s *ContainerApp) Del(id string) {
	if _, ok := s.clients[id]; ok {
		log.Println("Delete client")
		s.clients[id].Close()
		delete(s.clients, id)
		log.Println("Del", len(s.clients), "clients connected.")
	}
}

func (s *ContainerApp) Error(err error) {
	s.errors = append(s.errors, err)
}

func (s *ContainerApp) Store(msg Metric) {
	s.metricsCh <- msg
}

func (s *ContainerApp) Run() {
	go s.run()
}

func (s *ContainerApp) run() {
	log.Println("I! Init ContainerApp")
	var err error
	if s.ConfigType == "docker_env" {
		docker, err := NewDocker(
			s.DockerEnv["endpoint"],
			s.StartInterval,
			s.SyncInterval,
			s.Add,
			s.Del,
			s.Error,
		)
		if err == nil {
			go docker.Run()
		}
	} else if s.ConfigType == "kubernetes" {
		log.Println("I! Using k8s api as source for ContainerApp")

		kubeconfig, ok := s.Kubernetes["kubeconfig"]
		if !ok {
			kubeconfig = ""
		}

		k8s, err := NewK8s(
			s.Kubernetes["nodename"],
			kubeconfig,
			s.StartInterval,
			s.SyncInterval,
			s.Add,
			s.Del,
			s.Error,
		)
		if err == nil {
			go k8s.Run()
		}
	} else {
		log.Panic("E! Unsupported ContainerApp")
		return
	}

	if err != nil {
		log.Panicf("E! ContainerApp input: %s", err.Error())
		return
	}

	for {
		select {
		case msg := <-s.metricsCh:
			s.metricsLock.Lock()
			s.metrics = append(s.metrics, msg)
			s.metricsLock.Unlock()

		case err := <-s.errCh:
			s.errors = append(s.errors, err)
		}
	}
}

var sampleConfig = `
[[inputs.containerapp]]
  ## NOTE This plugin only reads numerical measurements, strings and booleans
  ## will be ignored.

  ## Config source type
  ##   Types: docker_env, kubernetes
  config_type = "docker_env"

  ## Interval with which new docker container collectors start(default 10ms)
  start_interval = "10ms"

  ## Rescan containers list interval(if there is a heavy load on the server, not always the messages come through the event api)
  sync_interval = "10m"

  ## Which config source names should be use as tag
  tags_name = ["MON_INTERVAL", "MON_PATH"]

  ## Prefix of config source names should be use as tag
  tags_prefix = "MON_TAG_"

  ## Mandatory config values, skip if it does not exist
  tags_mandatory = ["mon.db"]

  # docker_env conf
  [inputs.containerapp.docker_env]
     ## Docker Endpoint
     ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
     ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
     endpoint = "unix:///var/run/docker.sock"

  # kubernetes conf
  ## if this section declared plugin will search labels or annotations in k8s pods
  [inputs.containerapp.kubernetes]
     # kubernetes nodename, to be served by this instance 
     nodename = "testnode" 
     # full path to kubernetes config file
     kubeconfig = "/home/user/.kube/config"

  ## Mapping config source -> http
  [inputs.containerapp.http]
     ## Interval with which docker container application metrics gather
     interval  = "MON_INTERVAL"

     name_override  = "MON_NAME_OVERRIDE"

     ## Variable in which additional tags for the container are stored.
     ## Tags for different containers can be very different, therefore some of them 
     ## are stored in a json dump and can be simply populated by external container 
     ## deployment systems
     ##   Example value: MON_CUSTOM_TAGS={"my_tag_1":"my_tag_1"}
     ##   (stored in an config source variable as json dump)
     custom_tags = "MON_CUSTOM_TAGS"

     ## Metrics URL port
     ##   Example value: MON_PORT=8888
     http_port  = "MON_PORT"
	 
     ## Metrics URL path
     ##   Example value: MON_PATH=/mon/
     http_path  = "MON_PATH"

     ## Set response_timeout
     http_response_timeout  = "MON_RESPONSE_TIMEOUT"

     ## HTTP method to use: GET or POST (case-sensitive)
     http_method  = "MON_METHOD"
  
     ## HTTP parameters (all values must be strings).  For "GET" requests, data
     ## will be included in the query.  For "POST" requests, data will be included
     ## in the request body as "x-www-form-urlencoded".
     ##   Example value: MON_PARAMETERS={"my_parameter": "my_parameter"}
     ##   (stored in an config source variable as json dump)
     http_parameters   = "MON_PARAMETERS"
  
     ## HTTP Headers (all values must be strings)
     ##   Example value: MON_HEADERS={"my_header": "my_header"}  
     ##   (stored in an config source variable as json dump)
     http_headers  = "MON_HEADERS"

     ## List of tag names to extract from top-level of JSON server response
     tag_keys_json  = "MON_TAG_KEYS"

  ## HTTP default configuration
  [inputs.containerapp.http_defaults]
     ## Examples:
     interval  = "10s"
     name_override  = "test"
     http_port  = "8080"
     http_path  = "metrics"
     http_response_timeout = "5s"
     http_method  = "metrics"
`

func (h *ContainerApp) SampleConfig() string {
	return sampleConfig
}

func (h *ContainerApp) Description() string {
	return "Read flattened metrics from docker containers JSON HTTP endpoints"
}

// Gathers data for all containers.
func (s *ContainerApp) Gather(acc telegraf.Accumulator) error {
	s.metricsLock.Lock()
	metrics := s.metrics[:]
	s.metrics = s.metrics[:0]
	s.metricsLock.Unlock()
	for _, metric := range metrics {
		tags := metric.tags
		if container, ok := s.clients[metric.containerid]; ok {
			for tagkey, tag := range container.cfg.Tags {
				tags[tagkey] = tag
			}
		}
		switch valueType := metric.valueType; valueType {
		case telegraf.Counter:
			acc.AddCounter(metric.measurement, metric.fields, tags, metric.t)
		case telegraf.Gauge:
			acc.AddGauge(metric.measurement, metric.fields, tags, metric.t)
		case telegraf.Untyped:
			acc.AddFields(metric.measurement, metric.fields, tags, metric.t)
		case telegraf.Summary:
			acc.AddSummary(metric.measurement, metric.fields, tags, metric.t)
		case telegraf.Histogram:
			acc.AddHistogram(metric.measurement, metric.fields, tags, metric.t)
		}
	}

	for _, err := range s.errors {
		acc.AddError(err)
	}
	s.errors = nil
	return nil
}

func init() {
	inputs.Add("containerapp", func() telegraf.Input {
		ContainerApp := NewContainerApp()
		ContainerApp.Run()
		return ContainerApp
	})
}
