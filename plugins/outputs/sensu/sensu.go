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

## Timeout for HTTP message
# timeout = "5s"

## HTTP Content-Encoding for write request body, can be set to "gzip" to
## compress body or "identity" to apply no encoding.
# content_encoding = "identity"

## Sensu Event details 
## 
## NOTE: if the output plugin is configured to send events to a 
## backend_api_url and entity_name is not set, the value returned by 
## os.Hostname() will be used; if the output plugin is configured to send
## events to an agent_api_url, entity_name and entity_namespace are not used. 
# [outputs.sensu-go.entity]
#   name = "server-01"
#   namespace = "default"

# [outputs.sensu-go.tags]
#   source = "telegraf"

# [outputs.sensu-go.metrics]
#   handlers = ["elasticsearch","timescaledb"]
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

		// Get all valid numeric values, convert to float64
		for _, fieldSet := range metric.FieldList() {
			key := fieldSet.Key
			// don't need err since math.NaN is returned with err
			value, _ := getFloat(fieldSet.Value)
			// JSON does not support these special values
			if math.IsNaN(value) || math.IsInf(value, 0) {
				continue
			}

			point := &OutputMetric{
				Name:      metric.Name() + "." + key,
				Tags:      tagList,
				Timestamp: metric.Time().Unix(),
				Value:     value,
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
	if s.ContentEncoding == "gzip" {
		rc, err := internal.CompressWithGzip(reqBodyBuffer)
		if err != nil {
			return err
		}
		defer rc.Close()
		reqBodyBuffer = rc
	}

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
	Metrics   *OutputMetrics `json:"metrics"`
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
		Status:               0, // Always OK
		Issued:               time.Now().Unix(),
		Output:               "Telegraf agent processed " + strconv.Itoa(count) + " metrics",
		OutputMetricHandlers: s.GetHandlers(),
	}, nil
}

func (s *Sensu) GetHandlers() []string {
	if s.Metrics == nil || s.Metrics.Handlers == nil {
		return []string{}
	}
	return s.Metrics.Handlers
}

func getFloat(unk interface{}) (float64, error) {
    switch i := unk.(type) {
    case float64:
        return i, nil
    case float32:
        return float64(i), nil
    case int64:
        return float64(i), nil
    case int32:
        return float64(i), nil
    case int:
        return float64(i), nil
    case uint64:
        return float64(i), nil
    case uint32:
        return float64(i), nil
    case uint:
        return float64(i), nil
    default:
        return math.NaN(), fmt.Errorf("Non-numeric type could not be converted to float")
    }
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
	Metadata             *OutputMetadata `json:"metadata"`
	Status               int             `json:"status"`
	Output               string          `json:"output"`
	Issued               int64           `json:"issued"`
	OutputMetricHandlers []string        `json:"output_metric_handlers"`
}

type OutputMetrics struct {
	Handlers []string        `json:"handlers"`
	Metrics  []*OutputMetric `json:"points"`
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
