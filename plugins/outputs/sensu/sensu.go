package sensu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const (
	defaultUrl           = "http://127.0.0.1:3031"
	defaultClientTimeout = 5 * time.Second
	defaultContentType   = "application/json; charset=utf-8"
)

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

type OutputEvent struct {
	Entity    *OutputEntity  `json:"entity,omitempty"`
	Check     *OutputCheck   `json:"check"`
	Metrics   *OutputMetrics `json:"metrics"`
	Timestamp int64          `json:"timestamp"`
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

	EndpointUrl string
	OutEntity   *OutputEntity

	tls.ClientConfig
	client *http.Client

	serializer serializers.Serializer
}

var sampleConfig = `
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
## Below are the event details to be sent to Sensu.  The main portions of the
## event are the check, entity, and metrics specifications. For more information
## on Sensu events and its components, please visit:
## - Events - https://docs.sensu.io/sensu-go/latest/reference/events
## - Checks -  https://docs.sensu.io/sensu-go/latest/reference/checks
## - Entities - https://docs.sensu.io/sensu-go/latest/reference/entities
## - Metrics - https://docs.sensu.io/sensu-go/latest/reference/events#metrics
##
## Check specification
## The check name is the name to give the Sensu check associated with the event
## created.
[outputs.sensu-go.check]
  name = "telegraf"

## Entity specification
## Configure the entity name and namepsace, if necessary.
##
## NOTE: if the output plugin is configured to send events to a
## backend_api_url and entity_name is not set, the value returned by
## os.Hostname() will be used; if the output plugin is configured to send
## events to an agent_api_url, entity_name and entity_namespace are not used.
# [outputs.sensu-go.entity]
#   name = "server-01"
#   namespace = "default"

## Metrics specification
## Configure the tags for the metrics that are sent as part of the Sensu event
# [outputs.sensu-go.tags]
#   source = "telegraf"

## Configure the handler(s) for processing the provided metrics
# [outputs.sensu-go.metrics]
#   handlers = ["influxdb","elasticsearch"]
`

// Description provides a description of the plugin
func (s *Sensu) Description() string {
	return "Send aggregate metrics to Sensu Monitor"
}

// SampleConfig provides a sample configuration for the plugin
func (s *Sensu) SampleConfig() string {
	return sampleConfig
}

func (s *Sensu) createClient(ctx context.Context) (*http.Client, error) {
	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: s.Timeout.Duration,
	}

	return client, nil
}

func (s *Sensu) Connect() error {
	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = defaultClientTimeout
	}

	if len(s.ContentEncoding) != 0 {
		validEncoding := []string{"identity", "gzip"}
		if !choice.Contains(s.ContentEncoding, validEncoding) {
			return fmt.Errorf("Unsupported content_encoding [%q] specified", s.ContentEncoding)
		}
	} else {
		s.ContentEncoding = "identity"
	}

	if s.BackendApiUrl != nil && s.ApiKey == nil {
		return fmt.Errorf("backend_api_url [%q] specified, but no API Key provided", *s.BackendApiUrl)
	}

	err := s.setEndpointUrl()
	if err != nil {
		return err
	}

	err = s.setEntity()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	client, err := s.createClient(ctx)
	if err != nil {
		return err
	}

	s.client = client

	return nil
}

func (s *Sensu) Close() error {
	s.client.CloseIdleConnections()
	return nil
}

func (s *Sensu) Write(metrics []telegraf.Metric) error {
	var points []*OutputMetric
	for _, metric := range metrics {
		// Add tags from config to each metric point
		var tagList []*OutputTag
		for name, value := range s.Tags {
			tag := &OutputTag{
				Name:  name,
				Value: value,
			}
			tagList = append(tagList, tag)
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
			value := getFloat(fieldSet.Value)
			// JSON does not support these special values
			if math.IsInf(value, 1) {
				log.Printf("D! [outputs.sensu-go] metric %s returned positive infity, setting value to %f", key, math.MaxFloat64)
				value = math.MaxFloat64
			}
			if math.IsInf(value, -1) {
				log.Printf("D! [outputs.sensu-go] metric %s returned negative infity, setting value to %f", key, -math.MaxFloat64)
				value = -math.MaxFloat64
			}
			if math.IsNaN(value) {
				log.Printf("D! [outputs.sensu-go] metric %s returned as non a number, skipping", key)
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

	reqBody, err := s.encodeToJson(points)
	if err != nil {
		return err
	}

	return s.write(reqBody)
}

func (s *Sensu) write(reqBody []byte) error {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)
	method := http.MethodPost

	if s.ContentEncoding == "gzip" {
		rc, err := internal.CompressWithGzip(reqBodyBuffer)
		if err != nil {
			return err
		}
		defer rc.Close()
		reqBodyBuffer = rc
	}

	req, err := http.NewRequest(method, s.EndpointUrl, reqBodyBuffer)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", internal.ProductToken())

	req.Header.Set("Content-Type", defaultContentType)
	if s.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	if s.ApiKey != nil {
		req.Header.Set("Authorization", "Key "+*s.ApiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d", s.EndpointUrl, resp.StatusCode)
	}

	return nil
}

// Resolves the event write endpoint
func (s *Sensu) setEndpointUrl() error {
	var endpointUrl string

	if s.BackendApiUrl != nil {
		endpointUrl = *s.BackendApiUrl
	} else if s.AgentApiUrl != nil {
		endpointUrl = *s.AgentApiUrl
	}

	if len(endpointUrl) == 0 {
		log.Printf("D! [outputs.sensu-go] no backend or agent API URL provided, falling back to default agent API URL %s", defaultUrl)
		endpointUrl = defaultUrl
	}

	u, err := url.Parse(endpointUrl)
	if err != nil {
		return err
	}

	var path_suffix string
	if s.BackendApiUrl != nil {
		namespace := "default"
		if s.Entity != nil && s.Entity.Namespace != nil {
			namespace = *s.Entity.Namespace
		}
		path_suffix = "/api/core/v2/namespaces/" + namespace + "/events"
	} else {
		path_suffix = "/events"
	}
	u.Path = path.Join(u.Path, path_suffix)
	s.EndpointUrl = u.String()

	return nil
}

func init() {
	outputs.Add("sensu-go", func() telegraf.Output {
		// Default configuration values

		// make a string from the defaultUrl const
		agentApiUrl := defaultUrl

		return &Sensu{
			AgentApiUrl: &agentApiUrl,
			Timeout:     internal.Duration{Duration: defaultClientTimeout},
		}
	})
}

func (s *Sensu) encodeToJson(metricPoints []*OutputMetric) ([]byte, error) {
	timestamp := time.Now().Unix()

	check, err := s.getCheck(metricPoints)
	if err != nil {
		return []byte{}, err
	}

	output, err := json.Marshal(&OutputEvent{
		Entity: s.OutEntity,
		Check:  check,
		Metrics: &OutputMetrics{
			Handlers: s.getHandlers(),
			Metrics:  metricPoints,
		},
		Timestamp: timestamp,
	})

	return output, err
}

// Constructs the entity payload
// Throws when no entity name is provided and fails resolve to hostname
func (s *Sensu) setEntity() error {
	if s.BackendApiUrl != nil {
		var entityName string
		if s.Entity != nil && s.Entity.Name != nil {
			entityName = *s.Entity.Name
		} else {
			defaultHostname, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("resolving hostname failed: %v", err)
			}
			entityName = defaultHostname
		}

		s.OutEntity = &OutputEntity{
			Metadata: &OutputMetadata{
				Name: entityName,
			},
		}
		return nil
	}
	s.OutEntity = &OutputEntity{}
	return nil
}

// Constructs the check payload
// Throws if check name is not provided
func (s *Sensu) getCheck(metricPoints []*OutputMetric) (*OutputCheck, error) {
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
		OutputMetricHandlers: s.getHandlers(),
	}, nil
}

func (s *Sensu) getHandlers() []string {
	if s.Metrics == nil || s.Metrics.Handlers == nil {
		return []string{}
	}
	return s.Metrics.Handlers
}

func getFloat(unk interface{}) float64 {
	switch i := unk.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int32:
		return float64(i)
	case int:
		return float64(i)
	case uint64:
		return float64(i)
	case uint32:
		return float64(i)
	case uint:
		return float64(i)
	default:
		return math.NaN()
	}
}
