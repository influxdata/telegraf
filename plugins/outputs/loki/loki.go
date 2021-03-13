package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
)

var sampleConfig = `
  ## The domain of Loki
  domain = "https://loki.domain.tld"

  ## Endpoint to write api
  # endpoint = "/loki/api/v1/push"

  ## Connection timeout, defaults to "5s" if not set.
  # timeout = "5s"

  ## Basic auth credential
  # username = "loki"
  # password = "pass"

  ## Additional HTTP headers
  # http_headers = {"X-Scope-OrgID" = "1"}

  ## If the request must be gzip encoded
  # gzip_request = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
`

type Loki struct {
	Domain       string            `toml:"domain"`
	Endpoint     string            `toml:"endpoint"`
	Timeout      internal.Duration `toml:"timeout"`
	Username     string            `toml:"username"`
	Password     string            `toml:"password"`
	Headers      map[string]string `toml:"headers"`
	ClientID     string            `toml:"client_id"`
	ClientSecret string            `toml:"client_secret"`
	TokenURL     string            `toml:"token_url"`
	Scopes       []string          `toml:"scopes"`
	GZipRequest  bool              `toml:"gzip_request"`

	url    string
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
	if l.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	if l.Endpoint == "" {
		l.Endpoint = defaultEndpoint
	}

	l.url = fmt.Sprintf("%s%s", l.Domain, l.Endpoint)

	if l.Timeout.Duration == 0 {
		l.Timeout.Duration = defaultClientTimeout
	}

	ctx := context.Background()
	l.client, err = l.createClient(ctx)
	if err != nil {
		return fmt.Errorf("http client fail: %w", err)
	}

	return
}

func (l *Loki) Close() error {
	l.client.CloseIdleConnections()

	return nil
}

func (l *Loki) Write(metrics []telegraf.Metric) error {
	s := Streams{}

	for _, m := range metrics {
		tags := m.TagList()
		var line string

		for _, f := range m.FieldList() {
			line += fmt.Sprintf("%s=\"%v\" ", f.Key, f.Value)
		}

		s.insertLog(tags, Log{fmt.Sprintf("%d", m.Time().UnixNano()), line})
	}

	return l.write(s)
}

func (l *Loki) write(s Streams) error {
	bs, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	var reqBodyBuffer io.Reader = bytes.NewBuffer(bs)

	if l.GZipRequest {
		rc, err := internal.CompressWithGzip(reqBodyBuffer)
		if err != nil {
			return err
		}
		defer rc.Close()
		reqBodyBuffer = rc
	}

	req, err := http.NewRequest(http.MethodPost, l.url, reqBodyBuffer)
	if err != nil {
		return err
	}

	if l.Username != "" {
		req.SetBasicAuth(l.Username, l.Password)
	}

	for k, v := range l.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		}
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", internal.ProductToken())
	req.Header.Set("Content-Type", "application/json")
	if l.GZipRequest {
		req.Header.Set("Content-Encoding", "gzip")
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d", l.url, resp.StatusCode)
	}

	return nil
}

func init() {
	outputs.Add("loki", func() telegraf.Output {
		return &Loki{}
	})
}
