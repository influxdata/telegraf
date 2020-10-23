package yandex_cloud_monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/selfstat"
)

// YandexCloudMonitoring allows publishing of metrics to the Yandex Cloud Monitoring custom metrics
// service
type YandexCloudMonitoring struct {
	Timeout            internal.Duration `toml:"timeout"`
	EndpointUrl        string            `toml:"endpoint_url"`
	MetadataTokenUrl   string            `toml:"metadata_token_url"`
	MetadataFolderUrl  string            `toml:"metadata_folder_url"`
	Service            string            `toml:"service"`
	FolderID           string            `toml:"folder_id"`
	IAMTokenFromConfig string            `toml:"iam_token"`

	IAMToken               string
	IamTokenExpirationTime time.Time

	client *http.Client

	timeFunc func() time.Time

	MetricOutsideWindow selfstat.Stat
}

type yandexCloudMonitoringMessage struct {
	TS      string                        `json:"ts,omitempty"`
	Labels  map[string]string             `json:"labels,omitempty"`
	Metrics []yandexCloudMonitoringMetric `json:"metrics"`
}

type yandexCloudMonitoringMetric struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	MetricType string            `json:"type,omitempty"` // DGAUGE|IGAUGE|COUNTER|RATE. Default: DGAUGE
	TS         string            `json:"ts,omitempty"`
	Value      float64           `json:"value"`
}

type MetadataIamToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

const (
	defaultRequestTimeout    = time.Second * 5
	defaultEndpointUrl       = "https://monitoring.api.cloud.yandex.net/monitoring/v2/data/write"
	defaultMetadataTokenUrl  = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
	defaultMetadataFolderUrl = "http://169.254.169.254/computeMetadata/v1/instance/attributes/folder-id"
	maxRequestBodySize       = 4000000
)

var sampleConfig = `
  ## Timeout for HTTP writes.
  # timeout = "20s"

  ## Normally should not be changed
  # endpoint_url = "https://monitoring.api.cloud.yandex.net/monitoring/v2/data/write"

  ## Normally folder ID is taken from Compute instance metadata
  # folder_id = "..."

  ## Can be set explicitly for authentification debugging purposes 
  # iam_token = "..."  
`

// Description provides a description of the plugin
func (a *YandexCloudMonitoring) Description() string {
	return "Send aggregated metrics to Yandex.Cloud Monitoring"
}

// SampleConfig provides a sample configuration for the plugin
func (a *YandexCloudMonitoring) SampleConfig() string {
	return sampleConfig
}

// Connect initializes the plugin and validates connectivity
func (a *YandexCloudMonitoring) Connect() error {
	if a.Timeout.Duration <= 0 {
		a.Timeout.Duration = defaultRequestTimeout
	}

	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: a.Timeout.Duration,
	}

	if a.EndpointUrl == "" {
		a.EndpointUrl = defaultEndpointUrl
	}
	if a.MetadataTokenUrl == "" {
		a.MetadataTokenUrl = defaultMetadataTokenUrl
	}
	if a.MetadataFolderUrl == "" {
		a.MetadataFolderUrl = defaultMetadataFolderUrl
	}
	if a.FolderID == "" {
		folderID, err := getFolderIDFromMetadata(a.client, a.MetadataFolderUrl)
		if err != nil {
			return err
		}
		a.FolderID = folderID
	}
	if a.Service == "" {
		a.Service = "custom"
	}

	log.Printf("D! Writing to Yandex.Cloud Monitoring URL: %s", a.EndpointUrl)

	tags := map[string]string{}
	a.MetricOutsideWindow = selfstat.Register("yandex_cloud_monitoring", "metric_outside_window", tags)

	return nil
}

// Close shuts down an any active connections
func (a *YandexCloudMonitoring) Close() error {
	a.client = nil
	return nil
}

// Write writes metrics to the remote endpoint
func (a *YandexCloudMonitoring) Write(metrics []telegraf.Metric) error {
	var yandexCloudMonitoringMetrics []yandexCloudMonitoringMetric
	for _, m := range metrics {
		for _, field := range m.FieldList() {
			yandexCloudMonitoringMetrics = append(
				yandexCloudMonitoringMetrics,
				yandexCloudMonitoringMetric{
					Name:   field.Key,
					Labels: m.Tags(),
					TS:     fmt.Sprint(m.Time().Format(time.RFC3339)),
					Value:  field.Value.(float64),
				},
			)
		}
	}

	var body []byte
	jsonBytes, err := json.Marshal(
		yandexCloudMonitoringMessage{
			Metrics: yandexCloudMonitoringMetrics,
		},
	)

	if err != nil {
		return err
	}
	// Send batches that exceed this size via separate write requests.
	if (len(body) + len(jsonBytes) + 1) > maxRequestBodySize {
		err := a.send(body)
		if err != nil {
			return err
		}
		body = nil
	}
	body = append(body, jsonBytes...)
	body = append(body, '\n')

	return a.send(body)
}

func getResponseFromMetadata(c *http.Client, metadataUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", metadataUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("unable to fetch instance metadata: [%s] %d",
			metadataUrl, resp.StatusCode)
	}
	return body, nil
}

func getFolderIDFromMetadata(c *http.Client, metadataUrl string) (string, error) {
	log.Printf("!D getting folder ID in %s", metadataUrl)
	body, err := getResponseFromMetadata(c, metadataUrl)
	if err != nil {
		return "", err
	}
	folderID := string(body)
	if folderID == "" {
		return "", fmt.Errorf("unable to fetch folder id from URL %s: %v", metadataUrl, err)
	}
	return folderID, nil
}

func getIAMTokenFromMetadata(c *http.Client, metadataUrl string) (string, int, error) {
	log.Printf("!D getting new IAM token in %s", metadataUrl)
	body, err := getResponseFromMetadata(c, metadataUrl)
	if err != nil {
		return "", 0, err
	}
	var metadata MetadataIamToken
	if err := json.Unmarshal(body, &metadata); err != nil {
		return "", 0, err
	}
	if metadata.AccessToken == "" || metadata.ExpiresIn == 0 {
		return "", 0, fmt.Errorf("unable to fetch authentication credentials: %v", err)
	}
	return metadata.AccessToken, int(metadata.ExpiresIn), nil
}

func (a *YandexCloudMonitoring) send(body []byte) error {
	req, err := http.NewRequest("POST", a.EndpointUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("folderId", a.FolderID)
	q.Add("service", a.Service)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	isTokenExpired := !a.IamTokenExpirationTime.After(time.Now())
	if a.IAMTokenFromConfig != "" {
		a.IAMToken = a.IAMTokenFromConfig
	} else if isTokenExpired {
		token, expiresIn, err := getIAMTokenFromMetadata(a.client, a.MetadataTokenUrl)
		if err != nil {
			return err
		}
		a.IamTokenExpirationTime = time.Now().Add(time.Duration(expiresIn) * time.Second)
		a.IAMToken = token
	}
	req.Header.Set("Authorization", "Bearer "+a.IAMToken)

	log.Printf("!D sending metrics to %s", req.URL.String())
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to write batch: [%v] %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func init() {
	outputs.Add("yandex_cloud_monitoring", func() telegraf.Output {
		return &YandexCloudMonitoring{
			timeFunc: time.Now,
		}
	})
}
