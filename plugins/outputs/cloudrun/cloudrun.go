package cloudrun

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/influxdata/telegraf/config"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/gcp"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const (
	defaultURL = "http://127.0.0.1:8080/telegraf"
)

var sampleConfig = `
  ## URL is the Cloud Run Wavefront proxy address to send metrics to
  # url = "http://127.0.0.1:8080/telegraf"

  ## Timeout for Cloud Run message, suggested as 30s to account for handshaking
  # timeout = "30s"

  ## Cloud Run JSON file location
  ## This is the location of the JSON file generated from your GCP project that's authorized to send
  ## metrics into CloudRun.
  ## Windows users, note that you need to use forward slashes.
  # json_file_location = "C:/GCP/example-cr.json"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "wavefront"

  ## NOTE: The default headers have already been set that is appropriate to send
  ## metrics which are set to the following so you don't have to.
  ## defaultContentType   = "application/octet-stream"
  ## defaultAccept        = "application/json"
  ## defaultMethod        = http.MethodPost
  ## The token is generated using the URL, json_file_location, and cloudrun_email you set in your conf file
`

const (
	defaultClientTimeout = 5 * time.Second
	defaultContentType   = "application/octet-stream"
	defaultAccept        = "application/json"
	defaultMethod        = http.MethodPost
)

type CloudRun struct {
	URL          string            `toml:"url"`
	Timeout      config.Duration   `toml:"timeout"`
	Headers      map[string]string `toml:"headers"`
	JSONSecret   string            `toml:"json_file_location"`
	ConvertPaths bool              `toml:"convert_paths"`
	Method       string
	tls.ClientConfig

	client      *http.Client
	serializer  serializers.Serializer
	accessToken string
}

func (cr *CloudRun) SetSerializer(serializer serializers.Serializer) {
	cr.serializer = serializer
}

func (cr *CloudRun) createHTTPClient(ctx context.Context) (*http.Client, error) {
	tlsCfg, err := cr.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(cr.Timeout),
	}

	return client, nil
}

func (cr *CloudRun) Connect() error {
	if cr.Timeout == 0 {
		cr.Timeout = config.Duration(defaultClientTimeout)
	}

	ctx := context.Background()
	client, err := cr.createHTTPClient(ctx)
	if err != nil {
		return err
	}

	cr.client = client

	return nil
}

func (cr *CloudRun) Close() error {
	return nil
}

func (cr *CloudRun) Description() string {
	return "A plugin that is capable of transmitting metrics over HTTPS to a Cloud Run Wavefront proxy"
}

func (cr *CloudRun) SampleConfig() string {
	return sampleConfig
}

func (h *CloudRun) Write(metrics []telegraf.Metric) error {
	reqBody, err := h.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	if err := h.write(reqBody); err != nil {
		return err
	}

	return nil
}

func (cr *CloudRun) write(reqBody []byte) error {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)
	var err error
	req, err := http.NewRequest(defaultMethod, cr.URL, reqBodyBuffer)
	if err != nil {
		return err
	}

	claims := jwt.StandardClaims{}
	jwt.ParseWithClaims(cr.accessToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})

	if cr.accessToken == "" || !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		cr.accessToken, err = gcp.GetAccessToken(cr.JSONSecret, cr.URL)
		if err != nil {
			return err
		}
	}

	bearerToken := fmt.Sprintf("Bearer %s", cr.accessToken)

	req.Header.Set("User-Agent", internal.ProductToken())
	req.Header.Set("Content-Type", defaultContentType)
	req.Header.Set("Accept", defaultAccept)
	req.Header.Set("Authorization", bearerToken)

	for k, v := range cr.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		}
		req.Header.Set(k, v)
	}

	resp, err := cr.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d", cr.URL, resp.StatusCode)
	}

	return nil
}

func init() {
	outputs.Add("cloudrun", func() telegraf.Output {
		return &CloudRun{
			Method:  defaultMethod,
			URL:     defaultURL,
			Timeout: config.Duration(defaultClientTimeout),
		}
	})
}
