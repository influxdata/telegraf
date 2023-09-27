//go:generate ../../../tools/readme_config_includer/generator
package nebius_cloud_monitoring

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

// NebiusCloudMonitoring allows publishing of metrics to the Nebius Cloud Monitoring custom metrics
// service
type NebiusCloudMonitoring struct {
	Timeout  config.Duration `toml:"timeout"`
	Endpoint string          `toml:"endpoint"`

	Log telegraf.Logger `toml:"-"`

	metadataTokenURL       string
	metadataFolderURL      string
	folderID               string
	iamToken               string
	iamTokenExpirationTime time.Time
	service                string

	client *http.Client

	MetricOutsideWindow selfstat.Stat
}

type nebiusCloudMonitoringMessage struct {
	TS      string                        `json:"ts,omitempty"`
	Labels  map[string]string             `json:"labels,omitempty"`
	Metrics []nebiusCloudMonitoringMetric `json:"metrics"`
}

type nebiusCloudMonitoringMetric struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	MetricType string            `json:"type,omitempty"` // DGAUGE|IGAUGE|COUNTER|RATE. Default: DGAUGE
	TS         string            `json:"ts,omitempty"`
	Value      float64           `json:"value"`
}

type metadataIamToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

const (
	defaultRequestTimeout = time.Second * 20
	defaultEndpoint       = "https://monitoring.api.il.nebius.cloud/monitoring/v2/data/write"
	/*
		There is no DNS for metadata endpoint in Nebius Cloud yet.
		So the only way is to hardcode reserved IP (https://en.wikipedia.org/wiki/Link-local_address)
	*/
	//nolint:gosec // G101: Potential hardcoded credentials - false positive
	defaultMetadataTokenURL  = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
	defaultMetadataFolderURL = "http://169.254.169.254/computeMetadata/v1/yandex/folder-id"
)

func (*NebiusCloudMonitoring) SampleConfig() string {
	return sampleConfig
}

func (a *NebiusCloudMonitoring) Init() error {
	if a.Timeout <= 0 {
		a.Timeout = config.Duration(defaultRequestTimeout)
	}
	if a.Endpoint == "" {
		a.Endpoint = defaultEndpoint
	}
	if a.service == "" {
		a.service = "custom"
	}
	if a.metadataTokenURL == "" {
		a.metadataTokenURL = defaultMetadataTokenURL
	}
	if a.metadataFolderURL == "" {
		a.metadataFolderURL = defaultMetadataFolderURL
	}

	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(a.Timeout),
	}
	tags := map[string]string{}
	a.MetricOutsideWindow = selfstat.Register("nebius_cloud_monitoring", "metric_outside_window", tags)
	return nil
}

// Connect initializes the plugin and validates connectivity
func (a *NebiusCloudMonitoring) Connect() error {
	a.Log.Debugf("Getting folder ID in %s", a.metadataFolderURL)
	body, err := a.getResponseFromMetadata(a.client, a.metadataFolderURL)
	if err != nil {
		return err
	}
	a.folderID = string(body)
	if a.folderID == "" {
		return fmt.Errorf("unable to fetch folder id from URL %s: %w", a.metadataFolderURL, err)
	}
	a.Log.Infof("Writing to Nebius.Cloud Monitoring URL: %s", a.Endpoint)
	a.Log.Infof("FolderID: %s", a.folderID)

	return nil
}

// Close shuts down an any active connections
func (a *NebiusCloudMonitoring) Close() error {
	a.client = nil
	return nil
}

// Write writes metrics to the remote endpoint
func (a *NebiusCloudMonitoring) Write(metrics []telegraf.Metric) error {
	var nebiusCloudMonitoringMetrics []nebiusCloudMonitoringMetric
	for _, m := range metrics {
		for _, field := range m.FieldList() {
			value, err := internal.ToFloat64(field.Value)
			if err != nil {
				a.Log.Errorf("Skipping value: %v", err)
				continue
			}

			nebiusCloudMonitoringMetrics = append(
				nebiusCloudMonitoringMetrics,
				nebiusCloudMonitoringMetric{
					Name:   m.Name() + "_" + field.Key,
					Labels: replaceReservedTagNames(m.Tags()),
					TS:     m.Time().Format(time.RFC3339),
					Value:  value,
				},
			)
		}
	}

	body, err := json.Marshal(
		nebiusCloudMonitoringMessage{
			Metrics: nebiusCloudMonitoringMetrics,
		},
	)
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return a.send(body)
}

func (a *NebiusCloudMonitoring) getResponseFromMetadata(c *http.Client, metadataURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("unable to fetch instance metadata: [%s] %d",
			metadataURL, resp.StatusCode)
	}
	return body, nil
}

func (a *NebiusCloudMonitoring) getIAMTokenFromMetadata() (string, int, error) {
	a.Log.Debugf("Getting new IAM token in %s", a.metadataTokenURL)
	body, err := a.getResponseFromMetadata(a.client, a.metadataTokenURL)
	if err != nil {
		return "", 0, err
	}
	var metadata metadataIamToken
	if err := json.Unmarshal(body, &metadata); err != nil {
		return "", 0, err
	}
	if metadata.AccessToken == "" || metadata.ExpiresIn == 0 {
		return "", 0, fmt.Errorf("unable to fetch authentication credentials %s: %w", a.metadataTokenURL, err)
	}
	return metadata.AccessToken, int(metadata.ExpiresIn), nil
}

func (a *NebiusCloudMonitoring) send(body []byte) error {
	req, err := http.NewRequest("POST", a.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("folderId", a.folderID)
	q.Add("service", a.service)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	isTokenExpired := a.iamTokenExpirationTime.Before(time.Now())
	if a.iamToken == "" || isTokenExpired {
		token, expiresIn, err := a.getIAMTokenFromMetadata()
		if err != nil {
			return err
		}
		a.iamTokenExpirationTime = time.Now().Add(time.Duration(expiresIn) * time.Second)
		a.iamToken = token
	}
	req.Header.Set("Authorization", "Bearer "+a.iamToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to write batch: [%v] %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func init() {
	outputs.Add("nebius_cloud_monitoring", func() telegraf.Output {
		return &NebiusCloudMonitoring{}
	})
}

func replaceReservedTagNames(tagNames map[string]string) map[string]string {
	newTags := make(map[string]string, len(tagNames))
	for tagName, tagValue := range tagNames {
		if tagName == "name" {
			newTags["_name"] = tagValue
		} else {
			newTags[tagName] = tagValue
		}
	}
	return newTags
}
