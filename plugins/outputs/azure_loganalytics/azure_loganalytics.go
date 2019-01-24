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
  # customer_id = "<Workstation ID>"
  # shared_key = "<Secret>"
`

const (
	baseURL                = "https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01"
	contentType            = "application/json"
	httpMethod             = http.MethodPost
	timeGeneratedFieldName = "DateTime"
	defaultPrefix          = "Telegraf"
	defaultClientTimeout   = 5 * time.Second
)

type AzLogAnalytics struct {
	CustomerID string `toml:"customer_id"`
	SharedKey  string `toml:"shared_key"`

	client *http.Client
}

func (a *AzLogAnalytics) Connect() error {
	ctx := context.Background()
	client, err := a.createClient(ctx)
	if err != nil {
		return err
	}

	a.client = client

	return nil
}

func (a *AzLogAnalytics) Close() error {
	return nil
}

func (a *AzLogAnalytics) Description() string {
	return "A plugin that can transmit metrics to Azure Log Analytics"
}

func (a *AzLogAnalytics) SampleConfig() string {
	return sampleConfig
}

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
	req.Header.Add("Log-Type", defaultPrefix+strings.Title(logType))
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
		Timeout: defaultClientTimeout,
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
	for k, v := range metric.Tags() {
		v := convertField(v)
		if v == nil {
			continue
		}

		if k == "host" {
			m["Computer"] = v
		} else {
			m[underscoreToCaml(k)] = v
		}
	}
	for k, v := range metric.Fields() {
		v := convertField(v)
		if v == nil {
			continue
		}

		m[underscoreToCaml(k)] = v
	}

	return m, metric.Name()
}

// Convert field to a supported type or nil if unconvertible
func convertField(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return v
	case int64:
		return v
	case string:
		return v
	case bool:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case uint64:
		return uint64(v)
	case []byte:
		return string(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case float32:
		return float64(v)
	default:
		return nil
	}
}

func init() {
	outputs.Add("azure_loganalytics", func() telegraf.Output {
		return &AzLogAnalytics{}
	})
}
