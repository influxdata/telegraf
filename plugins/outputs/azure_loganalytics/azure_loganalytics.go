package http

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"unicode/utf8"
	"strconv"
	"strings"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  # customer_id = "<Workstation ID>"
  # shared_key = "<Secret>"
`

const (
	defaultClientTimeout = 5 * time.Second
	defaultContentType   = "application/json"
	defaultMethod        = http.MethodPost
	defaultUrl           = "https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01"
)

type AzLogAnalytics struct {
	CustomerId      string            `toml:"customer_id"`
	SharedKey       string            `toml:"shared_key"`

	client     *http.Client
}

func (a *AzLogAnalytics) createClient(ctx context.Context) (*http.Client, error) {
	client := &http.Client{
		Transport: &http.Transport {
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: defaultClientTimeout,
	}

	return client, nil
}

func BuildSignature(message, secret string) (string, error) {

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
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

func createObject(metric telegraf.Metric) map[string]interface{} {
	timeUnit, _ := time.ParseDuration("1s")

	m := make(map[string]interface{}, len(metric.Fields()) + len(metric.Tags()) + 2)
	m["timestamp"] = metric.Time().UnixNano() / int64(timeUnit)
	m["name"] = metric.Name()
	for k, v := range metric.Tags() {
		v := convertField(v)
		if v == nil {
			continue
		}

		if k == "host" {
			m["Computer"] = v
		} else {
			m["f_" + k] = v
		}
	}
	for k, v := range metric.Fields() {
		v := convertField(v)
		if v == nil {
			continue
		}
		m["t_" + k] = v
	}

	return m
}

func (a *AzLogAnalytics) Write(metrics []telegraf.Metric) error {

	objects := make([]interface{}, 0, len(metrics))
	for _, metric := range metrics {
		m := createObject(metric)
		objects = append(objects, m)
	}

	serialized, err := json.Marshal(objects)
	if err != nil {
		return err
	}

	if err := a.write(serialized); err != nil {
		return err
	}

	return nil
}

func (a *AzLogAnalytics) write(reqBody []byte) error {

	dateString := time.Now().UTC().Format(time.RFC1123)
	dateString = strings.Replace(dateString, "UTC", "GMT", -1)

	stringToHash := defaultMethod + "\n" + strconv.Itoa(utf8.RuneCount(reqBody)) + "\n" + defaultContentType + "\n" + "x-ms-date:" + dateString + "\n/api/logs"
	hashedString, err := BuildSignature(stringToHash, a.SharedKey)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	signature := "SharedKey " + a.CustomerId + ":" + hashedString

	url := fmt.Sprintf(defaultUrl, a.CustomerId)
	req, err := http.NewRequest(defaultMethod, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Telegraf/"+internal.Version())
	req.Header.Add("Log-Type", "Telegraf")
	req.Header.Add("Authorization", signature)
	req.Header.Add("Content-Type", defaultContentType)
	req.Header.Add("x-ms-date", dateString)
	req.Header.Add("time-generated-field", "timestamp")

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

func init() {
	outputs.Add("azure_loganalytics", func() telegraf.Output {
		return &AzLogAnalytics {
		}
	})
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