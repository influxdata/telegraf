package sensu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const (
	defaultUrl           = "http://localhost:3031"
	defaultClientTimeout = 5 * time.Second
	defaultContentType   = "application/json; charset=utf-8"
)

type Sensu struct {
	ApiKey        *string           `toml:"api_key"`
	AgentApiUrl   *string           `toml:"agent_api_url"`
	BackendApiUrl *string           `toml:"backend_api_url"`
	Entity        *SensuEntity      `toml:"entity"`
	Tags          map[string]string `toml:"tags"`
	Metrics       *SensuMetrics     `toml:"metrics"`
	Check         *SensuCheck       `toml:"check"`

	Timeout         internal.Duration `toml:"timeout"`
	ContentEncoding string            `toml:"content_encoding"`
	Headers         map[string]string `toml:"headers"`

	tls.ClientConfig
	client *http.Client

	serializer serializers.Serializer
}

var sampleConfig = `
	## Configure check configurations
	[outputs.sensu-go.check]
		name = "telegraf"

	## BACKEND API URL is the Sensu Backend API root URL to send metrics to 
	## (protocol, host, and port only). The output plugin will automatically 
	## append the corresponding backend or agent API path (e.g. /events or 
	## /api/core/v2/namespaces/:entity_namespace/events/:entity_name/:check_name).
	## 
	## NOTE: if backend_api_url and agent_api_url and api_key are set, the output 
	## plugin will use backend_api_url. If backend_api_url and agent_api_url are 
	## not provided, the output plugin will default to use an agent_api_url of 
	## http://127.0.0.1:3031
	## 
  # backend_api_url = "http://127.0.0.1:8080"
	# agent_api_url = "http://127.0.0.1:3031"

	## API KEY is the Sensu Backend API token 
  ## Generate a new API token via: 
  ## 
  ## $ sensuctl cluster-role create telegraf --verb create --resource events,entities
  ## $ sensuctl cluster-role-binding create telegraf --cluster-role telegraf --group telegraf
  ## $ sensuctl user create telegraf --group telegraf --password REDACTED 
  ## $ sensuctl api-key grant telegraf
  ##
  ## For more information on Sensu RBAC profiles & API tokens, please visit: 
  ## - https://docs.sensu.io/sensu-go/latest/reference/rbac/
  ## - https://docs.sensu.io/sensu-go/latest/reference/apikeys/ 
  ## 
	# api_key = "${SENSU_API_KEY}"
	
  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
	# insecure_skip_verify = false

	# Optional
	## Set whether publishing to agent or directly to backend.
	# [[outputs.sensu-go]]
	#   api_key = "$SENSU_API_KEY"
	#   url = "http://127.0.0.1:3031"
	#
	#   [outputs.sensu.check]
	#     metadata = { name: "telegraf" }
	#     
	#   [outputs.sensu.entity]
	#     metadata = { name: "server-01", namespace: "default" }
	#
	#   [outputs.sensu.metrics]
	#     handlers = ["elasticsearch", "timescaledb"]
	#
	## Timeout for HTTP writes.
	# timeout = "5s"
`

// Description provides a description of the plugin
func (s *Sensu) Description() string {
	return "Send aggregate metrics to Sensu Monitor"
}

// SampleConfig provides a sample configuration for the plugin
func (s *Sensu) SampleConfig() string {
	return sampleConfig
}

func (s *Sensu) SetSerializer(serializer serializers.Serializer) {
	s.serializer = serializer
}

func (s *Sensu) createClient(ctx context.Context) (*http.Client, error) {
	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: s.Timeout.Duration,
	}

	return client, nil
}

func (s *Sensu) Connect() error {
	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = defaultClientTimeout
	}

	ctx := context.Background()
	client, err := s.createClient(ctx)
	if err != nil {
		return err
	}

	s.client = client

	return nil
}

func (s *Sensu) Close() error {
	return nil
}

func (s *Sensu) Write(metrics []telegraf.Metric) error {
	var points []*OutputMetric
	for _, metric := range metrics {
		// Add tags from config to each metric point
		var tagList []*OutputTag
		for name, value := range s.Tags {
			metric.AddTag(name, value)
		}
		for _, tagSet := range metric.TagList() {
			tag := &OutputTag{
				Name:  tagSet.Key,
				Value: tagSet.Value,
			}
			tagList = append(tagList, tag)
		}

		// Filter valid float64 values
		for _, fieldSet := range metric.FieldList() {
			key := fieldSet.Key
			switch fv := fieldSet.Value.(type) {
			// Only support float64 values
			case float64:
				// JSON does not support these special values
				if math.IsNaN(fv) || math.IsInf(fv, 0) {
					continue
				}
			default:
				// Ignore unsupported value types
				continue
			}

			point := &OutputMetric{
				Name:      metric.Name() + "." + key,
				Tags:      tagList,
				Timestamp: metric.Time().Unix(),
				Value:     fieldSet.Value,
			}
			points = append(points, point)
		}
	}

	reqBody, err := s.Wrapper(points)

	if err != nil {
		return err
	}

	if err := s.write(reqBody); err != nil {
		return err
	}

	return nil
}

func (s *Sensu) write(reqBody []byte) error {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)
	method := http.MethodPost

	var err error
	endpointUrl, err := s.GetEndpointUrl()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, endpointUrl, reqBodyBuffer)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", internal.ProductToken())

	req.Header.Set("Content-Type", defaultContentType)
	if s.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	if s.ApiKey != nil {
		bearerType := "Bearer"
		if s.BackendApiUrl != nil {
			bearerType = "Key"
		}
		req.Header.Set("Authorization", bearerType+" "+*s.ApiKey)
	}

	for k, v := range s.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		}
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d", endpointUrl, resp.StatusCode)
	}

	return nil
}

// Resolves the event write endpoint
func (s *Sensu) GetEndpointUrl() (string, error) {
	var endpointUrl string

	if s.BackendApiUrl != nil {
		endpointUrl = *s.BackendApiUrl
	} else if s.AgentApiUrl != nil {
		endpointUrl = *s.AgentApiUrl
	}

	u, err := url.Parse(endpointUrl)
	if err != nil {
		return endpointUrl, err
	}

	if s.BackendApiUrl != nil {
		namespace := "default"
		if s.Entity != nil && s.Entity.Namespace != nil {
			namespace = *s.Entity.Namespace
		}
		u.Path = path.Join(u.Path, "/api/core/v2/namespaces", namespace, "events")
	} else {
		u.Path = path.Join(u.Path, "/events")
	}
	endpointUrl = u.String()

	return endpointUrl, nil
}

func init() {
	outputs.Add("sensu-go", func() telegraf.Output {
		// Default configuration values
		defaultUrl := defaultUrl

		return &Sensu{
			AgentApiUrl: &defaultUrl,
			Timeout:     internal.Duration{Duration: defaultClientTimeout},
		}
	})
}

type OutputEvent struct {
	Entity    *OutputEntity  `json:"entity,omitempty"`
	Check     *OutputCheck   `json:"check"`
	Metrics   *OutputMetrics `json:"name"`
	Timestamp int64          `json:"timestamp"`
}

func (s *Sensu) Wrapper(metricPoints []*OutputMetric) ([]byte, error) {
	timestamp := time.Now().Unix()

	var err error
	entity, err := s.GetEntity()
	if err != nil {
		return []byte{}, err
	}

	check, err := s.GetCheck(metricPoints)
	if err != nil {
		return []byte{}, err
	}

	output, err := json.Marshal(&OutputEvent{
		Entity: entity,
		Check:  check,
		Metrics: &OutputMetrics{
			Handlers: s.GetHandlers(),
			Metrics:  metricPoints,
		},
		Timestamp: timestamp,
	})

	return output, err
}

// Constructs the entity payload
// Throws when no entity name is provided and fails resolve to hostname
func (s *Sensu) GetEntity() (*OutputEntity, error) {
	if s.BackendApiUrl != nil {
		var entityName string
		if s.Entity != nil && s.Entity.Name != nil {
			entityName = *s.Entity.Name
		} else {
			defaultHostname, err := os.Hostname()

			if err != nil {
				return &OutputEntity{}, fmt.Errorf("when resolving hostname")
			}
			entityName = defaultHostname
		}

		return &OutputEntity{
			Metadata: &OutputMetadata{
				Name: entityName,
			},
		}, nil
	}
	return nil, nil
}

// Constructs the check payload
// Throws if check name is not provided
func (s *Sensu) GetCheck(metricPoints []*OutputMetric) (*OutputCheck, error) {
	count := len(metricPoints)

	if s.Check == nil || s.Check.Name == nil {
		return &OutputCheck{}, fmt.Errorf("missing check name")
	}

	return &OutputCheck{
		Metadata: &OutputMetadata{
			Name: *s.Check.Name,
		},
		Status: 0, // Always OK
		Issued: time.Now().Unix(),
		Output: "Telegraf agent processed " + strconv.Itoa(count) + " metrics",
	}, nil
}

func (s *Sensu) GetHandlers() []string {
	if s.Metrics == nil || s.Metrics.Handlers == nil {
		return []string{}
	}
	return s.Metrics.Handlers
}

type SensuEntity struct {
	Name      *string `toml:"name"`
	Namespace *string `toml:"namespace"`
}

type SensuCheck struct {
	Name *string `toml:"name"`
}

type SensuMetrics struct {
	Handlers []string `toml:"handlers"`
}

type OutputMetadata struct {
	Name string `json:"name"`
}

type OutputEntity struct {
	Metadata *OutputMetadata `json:"metadata"`
}

type OutputCheck struct {
	Metadata *OutputMetadata `json:"metadata"`
	Status   int             `json:"status"`
	Output   string          `json:"output"`
	Issued   int64           `json:"issued"`
}

type OutputMetrics struct {
	Handlers []string        `json:"handlers"`
	Metrics  []*OutputMetric `json:"metrics"`
}

type OutputMetric struct {
	Name      string       `json:"name"`
	Tags      []*OutputTag `json:"tags"`
	Value     interface{}  `json:"value"`
	Timestamp int64        `json:"timestamp"`
}

type OutputTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
