package azure_loganalytics

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  ## Customer ID (Workstation ID) and Key for Azure Log Analytics resource.
  # customer_id = "<Workstation ID>"
  # shared_key = "<Secret>"

  ## Timeout for closing (default: 5s).
  # timeout = "5s"

  ## Table Namespace Prefix (default: "Telegraf").
  ## Namespace Prefex is used in "Log-Type" header
  ## Restrictions can be found here (https://docs.microsoft.com/en-us/azure/azure-monitor/platform/data-collector-api#request-headers)
  # namespace_prefix = "Telegraf"
`

const (
	baseURL                = "https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01"
	contentType            = "application/json"
	httpMethod             = http.MethodPost
	timeGeneratedFieldName = "DateTime"

	defaultClientTimeout   = 5 * time.Second
	defaultNamespacePrefix = "Telegraf"
)

// AzLogAnalytics contains information about a azure log analytics service metadata
type AzLogAnalytics struct {
	CustomerID string `toml:"customer_id"`
	SharedKey  string `toml:"shared_key"`

	NamespacePrefix string `toml:"namespace_prefix"`
	ClientTimeout   internal.Duration

	client *http.Client
}

// Connect initializes the plugin and validates connectivity
func (a *AzLogAnalytics) Connect() error {
	ctx := context.Background()
	client, err := a.createClient(ctx)
	if err != nil {
		return err
	}

	a.client = client

	return nil
}

// Close shuts down an any active connections
func (a *AzLogAnalytics) Close() error {
	return nil
}

// Description provides a description of the plugin
func (a *AzLogAnalytics) Description() string {
	return "A plugin that can transmit metrics to Azure Log Analytics"
}

// SampleConfig provides a sample configuration for the plugin
func (a *AzLogAnalytics) SampleConfig() string {
	return sampleConfig
}

// Write writes metrics to the remote endpoint
func (a *AzLogAnalytics) Write(metrics []telegraf.Metric) error {

	objects := make(map[string][]interface{}, len(metrics))
	for _, metric := range metrics {
		m, name := createObject(metric)
		objects[name] = append(objects[name], m)
	}

	for name, data := range objects {
		serialized, err := json.Marshal(data)
		if err != nil {
			return err
		}

		if err := a.write(name, serialized); err != nil {
			return err
		}
	}

	return nil
}

func (a *AzLogAnalytics) write(logType string, reqBody []byte) error {

	dateString := time.Now().UTC().Format(time.RFC1123)
	dateString = strings.Replace(dateString, "UTC", "GMT", -1)

	stringToHash := httpMethod + "\n" + strconv.Itoa(utf8.RuneCount(reqBody)) + "\n" + contentType + "\n" + "x-ms-date:" + dateString + "\n/api/logs"
	signature, err := a.buildSignature(stringToHash)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	url := fmt.Sprintf(baseURL, a.CustomerID)
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Telegraf/"+internal.Version())
	req.Header.Add("Log-Type", a.NamespacePrefix+strings.Title(logType))
	req.Header.Add("Authorization", signature)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("x-ms-date", dateString)
	req.Header.Add("time-generated-field", timeGeneratedFieldName)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d", url, resp.StatusCode)
	}

	return nil
}

func (a *AzLogAnalytics) createClient(ctx context.Context) (*http.Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: a.ClientTimeout.Duration,
	}

	return client, nil
}

func underscoreToCaml(s string) string {

	arr := strings.Split(s, "_")

	var sb strings.Builder
	for _, item := range arr {
		sb.WriteString(strings.Title(item))
	}

	return sb.String()
}

func (a *AzLogAnalytics) buildSignature(message string) (string, error) {

	keyBytes, err := base64.StdEncoding.DecodeString(a.SharedKey)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	hashedString := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	signiture := fmt.Sprintf("SharedKey %s:%s", a.CustomerID, hashedString)

	return signiture, nil
}

func createObject(metric telegraf.Metric) (map[string]interface{}, string) {
	m := make(map[string]interface{}, len(metric.Fields())+len(metric.Tags())+1)
	m[timeGeneratedFieldName] = metric.Time().UTC().Format(time.RFC3339)
	fmt.Println(m[timeGeneratedFieldName])
	for k, v := range metric.Tags() {
		if k == "host" {
			m["Computer"] = v
		} else {
			m[underscoreToCaml(k)] = v
		}
	}
	for k, v := range metric.Fields() {
		m[underscoreToCaml(k)] = v
	}

	return m, metric.Name()
}

func init() {
	outputs.Add("azure_loganalytics", func() telegraf.Output {
		return &AzLogAnalytics{
			ClientTimeout:   internal.Duration{Duration: defaultClientTimeout},
			NamespacePrefix: defaultNamespacePrefix,
		}
	})
}
