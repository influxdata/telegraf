package sensu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultURL           = "http://127.0.0.1:3031"
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
	APIKey        *string           `toml:"api_key"`
	AgentAPIURL   *string           `toml:"agent_api_url"`
	BackendAPIURL *string           `toml:"backend_api_url"`
	Entity        *SensuEntity      `toml:"entity"`
	Tags          map[string]string `toml:"tags"`
	Metrics       *SensuMetrics     `toml:"metrics"`
	Check         *SensuCheck       `toml:"check"`

	Timeout         config.Duration `toml:"timeout"`
	ContentEncoding string          `toml:"content_encoding"`

	EndpointURL string
	OutEntity   *OutputEntity

	Log telegraf.Logger `toml:"-"`

	tls.ClientConfig
	client *http.Client
}

var sampleConfig = `
  ## BACKEND API URL is the Sensu Backend API root URL to send metrics to
  ## (protocol, host, and port only). The output plugin will automatically
  ## append the corresponding backend API path
  ## /api/core/v2/namespaces/:entity_namespace/events/:entity_name/:check_name).
  ##
  ## Backend Events API reference:
  ## https://docs.sensu.io/sensu-go/latest/api/events/
  ##
  ## AGENT API URL is the Sensu Agent API root URL to send metrics to
  ## (protocol, host, and port only). The output plugin will automatically
  ## append the correspeonding agent API path (/events).
  ##
  ## Agent API Events API reference:
  ## https://docs.sensu.io/sensu-go/latest/api/events/
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
  ## created. This maps to check.metatadata.name in the event.
  [outputs.sensu.check]
  name = "telegraf"

  ## Entity specification
  ## Configure the entity name and namespace, if necessary. This will be part of
  ## the entity.metadata in the event.
  ##
  ## NOTE: if the output plugin is configured to send events to a
  ## backend_api_url and entity_name is not set, the value returned by
  ## os.Hostname() will be used; if the output plugin is configured to send
  ## events to an agent_api_url, entity_name and entity_namespace are not used.
  # [outputs.sensu.entity]
  #   name = "server-01"
  #   namespace = "default"

  ## Metrics specification
  ## Configure the tags for the metrics that are sent as part of the Sensu event
  # [outputs.sensu.tags]
  #   source = "telegraf"

  ## Configure the handler(s) for processing the provided metrics
  # [outputs.sensu.metrics]
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

func (s *Sensu) createClient() (*http.Client, error) {
	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(s.Timeout),
	}

	return client, nil
}

func (s *Sensu) Connect() error {
	err := s.setEndpointURL()
	if err != nil {
		return err
	}

	err = s.setEntity()
	if err != nil {
		return err
	}

	client, err := s.createClient()
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
		tagList := make([]*OutputTag, 0, len(s.Tags)+len(metric.TagList()))
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
				s.Log.Debugf("metric %s returned positive infinity, setting value to %f", key, math.MaxFloat64)
				value = math.MaxFloat64
			}
			if math.IsInf(value, -1) {
				s.Log.Debugf("metric %s returned negative infinity, setting value to %f", key, -math.MaxFloat64)
				value = -math.MaxFloat64
			}
			if math.IsNaN(value) {
				s.Log.Debugf("metric %s returned as non a number, skipping", key)
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

	reqBody, err := s.encodeToJSON(points)
	if err != nil {
		return err
	}

	return s.writeMetrics(reqBody)
}

func (s *Sensu) writeMetrics(reqBody []byte) error {
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

	req, err := http.NewRequest(method, s.EndpointURL, reqBodyBuffer)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", internal.ProductToken())

	req.Header.Set("Content-Type", defaultContentType)
	if s.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	if s.APIKey != nil {
		req.Header.Set("Authorization", "Key "+*s.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyData, err := io.ReadAll(resp.Body)
		if err != nil {
			s.Log.Debugf("Couldn't read response body: %v", err)
		}
		s.Log.Debugf("Failed to write, response: %v", string(bodyData))
		if resp.StatusCode < 400 || resp.StatusCode > 499 {
			return fmt.Errorf("when writing to [%s] received status code: %d", s.EndpointURL, resp.StatusCode)
		}
	}

	return nil
}

// Resolves the event write endpoint
func (s *Sensu) setEndpointURL() error {
	var (
		endpointURL string
		pathSuffix  string
	)

	if s.BackendAPIURL != nil {
		endpointURL = *s.BackendAPIURL
		namespace := "default"
		if s.Entity != nil && s.Entity.Namespace != nil {
			namespace = *s.Entity.Namespace
		}
		pathSuffix = "/api/core/v2/namespaces/" + namespace + "/events"
	} else if s.AgentAPIURL != nil {
		endpointURL = *s.AgentAPIURL
		pathSuffix = "/events"
	}

	if len(endpointURL) == 0 {
		s.Log.Debugf("no backend or agent API URL provided, falling back to default agent API URL %s", defaultURL)
		endpointURL = defaultURL
		pathSuffix = "/events"
	}

	u, err := url.Parse(endpointURL)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, pathSuffix)
	s.EndpointURL = u.String()

	return nil
}

func (s *Sensu) Init() error {
	if len(s.ContentEncoding) != 0 {
		validEncoding := []string{"identity", "gzip"}
		if !choice.Contains(s.ContentEncoding, validEncoding) {
			return fmt.Errorf("unsupported content_encoding [%q] specified", s.ContentEncoding)
		}
	}

	if s.BackendAPIURL != nil && s.APIKey == nil {
		return fmt.Errorf("backend_api_url [%q] specified, but no API Key provided", *s.BackendAPIURL)
	}

	return nil
}

func init() {
	outputs.Add("sensu", func() telegraf.Output {
		// Default configuration values

		// make a string from the defaultURL const
		agentAPIURL := defaultURL

		return &Sensu{
			AgentAPIURL:     &agentAPIURL,
			Timeout:         config.Duration(defaultClientTimeout),
			ContentEncoding: "identity",
		}
	})
}

func (s *Sensu) encodeToJSON(metricPoints []*OutputMetric) ([]byte, error) {
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
	if s.BackendAPIURL != nil {
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
