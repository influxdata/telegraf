package loki

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultEndpoint      = "/loki/api/v1/push"
	defaultClientTimeout = 5 * time.Second
	defaultFieldLine     = "line"
)

var sampleConfig = `
  ## Connection timeout, defaults to "5s" if not set.
  timeout = "5s"

  ## The URL of Loki
  # url = "https://loki.domain.tld"

  ## Basic auth credential
  # username = "loki"
  # password = "pass"

  ## Additional HTTP headers
  # http_headers = {"X-Scope-OrgID" = "1"}

  ## The field containing the log
  # field_line = "log"

  ## If the request must be gzip encoded
  # gzip_request = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
`

type Loki struct {
	URL          string            `toml:"url"`
	Endpoint     string            `toml:"endpoint"`
	Timeout      internal.Duration `toml:"timeout"`
	Username     string            `toml:"username"`
	Password     string            `toml:"password"`
	Headers      map[string]string `toml:"headers"`
	ClientID     string            `toml:"client_id"`
	ClientSecret string            `toml:"client_secret"`
	TokenURL     string            `toml:"token_url"`
	Scopes       []string          `toml:"scopes"`
	FieldLine    string            `toml:"field_line"`
	GZipRequest  bool              `toml:"gzip_request"`

	client *http.Client
	tls.ClientConfig
}

func (l *Loki) SampleConfig() string {
	return sampleConfig
}

func (l *Loki) Description() string {
	return "Send logs to Loki"
}

func (l *Loki) createClient(ctx context.Context) (*http.Client, error) {
	tlsCfg, err := l.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("tls config fail: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: l.Timeout.Duration,
	}

	if l.ClientID != "" && l.ClientSecret != "" && l.TokenURL != "" {
		oauthConfig := clientcredentials.Config{
			ClientID:     l.ClientID,
			ClientSecret: l.ClientSecret,
			TokenURL:     l.TokenURL,
			Scopes:       l.Scopes,
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
		client = oauthConfig.Client(ctx)
	}

	return client, nil
}

func (l *Loki) Connect() (err error) {
	if l.URL == "" {
		return fmt.Errorf("url is required")
	}

	if l.Endpoint == "" {
		l.Endpoint = defaultEndpoint
	}

	if l.Timeout.Duration == 0 {
		l.Timeout.Duration = defaultClientTimeout
	}

	if l.FieldLine == "" {
		l.FieldLine = defaultFieldLine
	}

	ctx := context.Background()
	l.client, err = l.createClient(ctx)
	if err != nil {
		return fmt.Errorf("http client fail: %w", err)
	}

	return
}

func (l *Loki) Close() error {
	return nil
}

func (l *Loki) Write(metrics []telegraf.Metric) error {
	s := Streams{
		Streams: make([]Stream, 0, len(metrics)),
	}

	for _, m := range metrics {
		tags := m.TagList()
		line, ok := m.GetField(l.FieldLine)

		if !ok {
			continue
		}

		s.insertLog(tags, Log{
			Timestamp: m.Time().UnixNano(),
			Line:      line.(string),
		})
	}

	return l.write(s)
}

func (l *Loki) write(s Streams) (err error) {
	var (
		buf     bytes.Buffer
		encoder *json.Encoder
	)

	if l.GZipRequest {
		gz := gzip.NewWriter(&buf)
		defer gz.Close()

		encoder = json.NewEncoder(gz)
	} else {
		encoder = json.NewEncoder(&buf)
	}

	if err := encoder.Encode(s); err != nil {
		return fmt.Errorf("new json encoder fail: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", l.URL, l.Endpoint), &buf)
	if err != nil {
		return fmt.Errorf("new request fail: %w", err)
	}

	for k, v := range l.Headers {
		req.Header.Add(k, v)
	}

	req.SetBasicAuth(l.Username, l.Password)
	req.Header.Add("Content-Type", "application/json")
	if l.GZipRequest {
		req.Header.Add("Content-Encoding", "gzip")
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("client do fail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("received bad status code %d", resp.StatusCode)
	}

	return
}

func init() {
	outputs.Add("loki", func() telegraf.Output {
		return &Loki{
			Timeout: internal.Duration{Duration: defaultClientTimeout},
		}
	})
}
