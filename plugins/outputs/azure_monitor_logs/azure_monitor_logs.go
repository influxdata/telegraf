package azure_monitor_logs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
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

  ## Table Namespace Prefix (default: "").
  ## Namespace Prefix is used in "Log-Type" header
  ## Prefix can only contain alphaNumeric characters
  # namespace_prefix = ""
`

const (
	baseURL                = "https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01"
	contentType            = "application/json"
	timeGeneratedFieldName = "DateTime"
	namespacePrefixRegex   = "^[a-zA-Z0-9]*$"
	httpMethod             = http.MethodPost

	defaultClientTimeout = 5 * time.Second
)

// AzLogAnalytics contains information about a azure log analytics service metadata
type AzLogAnalytics struct {
	CustomerID string `toml:"customer_id"`
	SharedKey  string `toml:"shared_key"`

	NamespacePrefix string            `toml:"namespace_prefix"`
	ClientTimeout   internal.Duration `toml:"timeout"`

	serviceURL string
	client     *http.Client
}

// Initializes the plugin
func (a *AzLogAnalytics) Init() error {
	if a.CustomerID == "" {
		return fmt.Errorf("customer_id not configured")
	}

	if a.SharedKey == "" {
		return fmt.Errorf("shared_key not configured")
	}

	if len(a.NamespacePrefix) > 25 {
		return fmt.Errorf("namespace_prefix length is greater than 25 characters")
	}

	if a.ClientTimeout.Duration == 0 {
		a.ClientTimeout.Duration = defaultClientTimeout
	}

	match, err := regexp.MatchString(namespacePrefixRegex, a.NamespacePrefix)
	if err != nil {
		return err
	} else if !match {
		return fmt.Errorf("namespace_prefix contains invalid characters")
	}

	return nil
}

// Connect validates connectivity
func (a *AzLogAnalytics) Connect() error {

	err := a.Init()
	if err != nil {
		return err
	}

	client, err := a.createClient(context.Background())
	if err != nil {
		return err
	}

	a.client = client
	a.serviceURL = fmt.Sprintf(baseURL, a.CustomerID)

	return nil
}

// Close shuts down an any active connections
func (a *AzLogAnalytics) Close() error {
	a.client.CloseIdleConnections()

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
		m, name := a.createObject(metric)
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

	stringToHash := httpMethod + "\n" + 
					strconv.Itoa(utf8.RuneCount(reqBody)) + "\n" +
					contentType + "\n" + "x-ms-date:" +
					dateString + "\n/api/logs"

	signature, err := a.buildSignature(stringToHash)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(httpMethod, a.serviceURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	tableName := a.NamespacePrefix + underscoreToCaml(logType)

	req.Header.Set("User-Agent", "Telegraf/" + internal.ProductToken())
	req.Header.Add("Log-Type", tableName)
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
		return fmt.Errorf(
			"when writing to [%s] received status code: %d",
			a.serviceURL,
			resp.StatusCode)
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

	var buffer bytes.Buffer
	for _, item := range arr {
		buffer.WriteString(strings.Title(item))
	}

	return buffer.String()
}

func (a *AzLogAnalytics) buildSignature(message string) (string, error) {

	keyBytes, err := base64.StdEncoding.DecodeString(a.SharedKey)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	hashedString := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	signature := fmt.Sprintf("SharedKey %s:%s", a.CustomerID, hashedString)

	return signature, nil
}

func (a *AzLogAnalytics) createObject(metric telegraf.Metric) (map[string]interface{}, string) {
	m := make(map[string]interface{}, len(metric.Fields())+len(metric.Tags())+2)
	m["MetricName"] = metric.Name()
	m[timeGeneratedFieldName] = metric.Time().UTC().Format(time.RFC3339)
	for _, tag := range metric.TagList() {
		if tag.Key == "host" {
			m["Computer"] = tag.Value
		} else {
			m[underscoreToCaml(tag.Key)] = tag.Value
		}
	}
	for _, field := range metric.FieldList() {
		m[underscoreToCaml(field.Key)] = field.Value
	}

	return m, metric.Name()
}

func init() {
	outputs.Add("azure_monitor_logs", func() telegraf.Output {
		return &AzLogAnalytics{}
	})
}
