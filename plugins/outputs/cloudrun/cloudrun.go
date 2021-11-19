package cloudrun

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/influxdata/telegraf/config"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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

  ## Cloud Run service account email address
  ## This is the authorized GCP service account email address from your GCP project
  # cloudrun_email = The authorized service account email

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
	URL string `toml:"url"`
	Timeout         config.Duration `toml:"timeout"`
	Headers         map[string]string `toml:"headers"`
	JSONSecret      string            `toml:"json_file_location"`
	GCPEmailAddress string            `toml:"cloudrun_email"`
	ConvertPaths    bool              `toml:"convert_paths"`
	Method          string
	tls.ClientConfig

	client     *http.Client
	serializer serializers.Serializer
	signedJWT  string
}

// SetSerializer Allows you to use data_format
// TODO: Should I write a test for SetSerializer method? Is there a test case elsewhere? Don't see it in registry where the interface lives...
// 	Nothing serializer/wavefront_test.go. Nor config_test.go.
func (h *CloudRun) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

func (h *CloudRun) createHTTPClient(ctx context.Context) (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(h.Timeout),
	}

	return client, nil
}

func (h *CloudRun) Connect() error {
	if h.Timeout == 0 {
		h.Timeout = config.Duration(defaultClientTimeout)
	}

	ctx := context.Background()
	client, err := h.createHTTPClient(ctx)
	if err != nil {
		return err
	}

	h.client = client

	return nil
}

func (h *CloudRun) Close() error {
	return nil
}

func (h *CloudRun) Description() string {
	return "A plugin that is capable of transmitting metrics over HTTPS to a Cloud Run Wavefront proxy"
}

func (h *CloudRun) SampleConfig() string {
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

func (h *CloudRun) write(reqBody []byte) error {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)
	var err error
	req, err := http.NewRequest(defaultMethod, h.URL, reqBodyBuffer)
	if err != nil {
		return err
	}

	claims := jwt.StandardClaims{}
	_, err = jwt.ParseWithClaims(h.signedJWT, &claims, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})

	if h.signedJWT == "" || claims.VerifyExpiresAt(time.Now().Unix(), true) == false {
		h.signedJWT = gcp.GetToken(h.JSONSecret, h.GCPEmailAddress, h.URL)
	}

	bearerToken := fmt.Sprintf("Bearer %s", h.signedJWT)

	req.Header.Set("User-Agent", internal.ProductToken())
	req.Header.Set("Content-Type", defaultContentType)
	req.Header.Set("Accept", defaultAccept)
	req.Header.Set("Authorization", bearerToken)

	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		}
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d", h.URL, resp.StatusCode)
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
