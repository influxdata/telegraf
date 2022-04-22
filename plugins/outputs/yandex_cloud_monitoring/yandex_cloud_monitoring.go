package yandex_cloud_monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/selfstat"
)

// YandexCloudMonitoring allows publishing of metrics to the Yandex Cloud Monitoring custom metrics
// service
type YandexCloudMonitoring struct {
	Timeout     config.Duration `toml:"timeout"`
	EndpointURL string          `toml:"endpoint_url"`
	Service     string          `toml:"service"`

	Log telegraf.Logger

	MetadataTokenURL       string
	MetadataFolderURL      string
	FolderID               string
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
	defaultRequestTimeout    = time.Second * 20
	defaultEndpointURL       = "https://monitoring.api.cloud.yandex.net/monitoring/v2/data/write"
	defaultMetadataTokenURL  = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
	defaultMetadataFolderURL = "http://169.254.169.254/computeMetadata/v1/yandex/folder-id"
)

// Connect initializes the plugin and validates connectivity
func (a *YandexCloudMonitoring) Connect() error {
	if a.Timeout <= 0 {
		a.Timeout = config.Duration(defaultRequestTimeout)
	}
	if a.EndpointURL == "" {
		a.EndpointURL = defaultEndpointURL
	}
	if a.Service == "" {
		a.Service = "custom"
	}
	if a.MetadataTokenURL == "" {
		a.MetadataTokenURL = defaultMetadataTokenURL
	}
	if a.MetadataFolderURL == "" {
		a.MetadataFolderURL = defaultMetadataFolderURL
	}

	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(a.Timeout),
	}

	var err error
	a.FolderID, err = a.getFolderIDFromMetadata()
	if err != nil {
		return err
	}

	a.Log.Infof("Writing to Yandex.Cloud Monitoring URL: %s", a.EndpointURL)

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
	body = append(jsonBytes, '\n')
	return a.send(body)
}

func getResponseFromMetadata(c *http.Client, metadataURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
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

func (a *YandexCloudMonitoring) getFolderIDFromMetadata() (string, error) {
	a.Log.Infof("getting folder ID in %s", a.MetadataFolderURL)
	body, err := getResponseFromMetadata(a.client, a.MetadataFolderURL)
	if err != nil {
		return "", err
	}
	folderID := string(body)
	if folderID == "" {
		return "", fmt.Errorf("unable to fetch folder id from URL %s: %v", a.MetadataFolderURL, err)
	}
	return folderID, nil
}

func (a *YandexCloudMonitoring) getIAMTokenFromMetadata() (string, int, error) {
	a.Log.Debugf("getting new IAM token in %s", a.MetadataTokenURL)
	body, err := getResponseFromMetadata(a.client, a.MetadataTokenURL)
	if err != nil {
		return "", 0, err
	}
	var metadata MetadataIamToken
	if err := json.Unmarshal(body, &metadata); err != nil {
		return "", 0, err
	}
	if metadata.AccessToken == "" || metadata.ExpiresIn == 0 {
		return "", 0, fmt.Errorf("unable to fetch authentication credentials %s: %v", a.MetadataTokenURL, err)
	}
	return metadata.AccessToken, int(metadata.ExpiresIn), nil
}

func (a *YandexCloudMonitoring) send(body []byte) error {
	req, err := http.NewRequest("POST", a.EndpointURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("folderId", a.FolderID)
	q.Add("service", a.Service)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	isTokenExpired := !a.IamTokenExpirationTime.After(time.Now())
	if a.IAMToken == "" || isTokenExpired {
		token, expiresIn, err := a.getIAMTokenFromMetadata()
		if err != nil {
			return err
		}
		a.IamTokenExpirationTime = time.Now().Add(time.Duration(expiresIn) * time.Second)
		a.IAMToken = token
	}
	req.Header.Set("Authorization", "Bearer "+a.IAMToken)

	a.Log.Debugf("sending metrics to %s", req.URL.String())
	a.Log.Debugf("body: %s", body)
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
	outputs.Add("yandex_cloud_monitoring", func() telegraf.Output {
		return &YandexCloudMonitoring{
			timeFunc: time.Now,
		}
	})
}
