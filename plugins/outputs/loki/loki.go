//go:generate ../../../tools/readme_config_includer/generator
package loki

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultEndpoint      = "/loki/api/v1/push"
	defaultClientTimeout = 5 * time.Second
)

type Loki struct {
	Domain             string            `toml:"domain"`
	Endpoint           string            `toml:"endpoint"`
	Timeout            config.Duration   `toml:"timeout"`
	Username           config.Secret     `toml:"username"`
	Password           config.Secret     `toml:"password"`
	Headers            map[string]string `toml:"http_headers"`
	ClientID           string            `toml:"client_id"`
	ClientSecret       string            `toml:"client_secret"`
	TokenURL           string            `toml:"token_url"`
	Scopes             []string          `toml:"scopes"`
	GZipRequest        bool              `toml:"gzip_request"`
	MetricNameLabel    string            `toml:"metric_name_label"`
	SanitizeLabelNames bool              `toml:"sanitize_label_names"`

	url    string
	client *http.Client
	tls.ClientConfig
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
		Timeout: time.Duration(l.Timeout),
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

func (*Loki) SampleConfig() string {
	return sampleConfig
}

func (l *Loki) Connect() (err error) {
	if l.Domain == "" {
		return errors.New("domain is required")
	}

	if l.Endpoint == "" {
		l.Endpoint = defaultEndpoint
	}

	l.url = fmt.Sprintf("%s%s", l.Domain, l.Endpoint)

	if l.Timeout == 0 {
		l.Timeout = config.Duration(defaultClientTimeout)
	}

	ctx := context.Background()
	l.client, err = l.createClient(ctx)
	if err != nil {
		return fmt.Errorf("http client fail: %w", err)
	}

	return nil
}

func (l *Loki) Close() error {
	l.client.CloseIdleConnections()

	return nil
}

func (l *Loki) Write(metrics []telegraf.Metric) error {
	s := Streams{}

	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].Time().Before(metrics[j].Time())
	})

	for _, m := range metrics {
		if l.MetricNameLabel != "" {
			m.AddTag(l.MetricNameLabel, m.Name())
		}

		tags := m.TagList()
		if l.SanitizeLabelNames {
			for _, t := range tags {
				t.Key = sanitizeLabelName(t.Key)
			}
		}

		var line string
		for _, f := range m.FieldList() {
			line += fmt.Sprintf("%s=\"%v\" ", f.Key, f.Value)
		}

		s.insertLog(tags, Log{strconv.FormatInt(m.Time().UnixNano(), 10), line})
	}

	return l.writeMetrics(s)
}

func (l *Loki) writeMetrics(s Streams) error {
	bs, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	var reqBodyBuffer io.Reader = bytes.NewBuffer(bs)

	if l.GZipRequest {
		rc := internal.CompressWithGzip(reqBodyBuffer)
		defer rc.Close()
		reqBodyBuffer = rc
	}

	req, err := http.NewRequest(http.MethodPost, l.url, reqBodyBuffer)
	if err != nil {
		return err
	}

	if !l.Username.Empty() {
		username, err := l.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		password, err := l.Password.Get()
		if err != nil {
			username.Destroy()
			return fmt.Errorf("getting password failed: %w", err)
		}
		req.SetBasicAuth(username.String(), password.String())
		username.Destroy()
		password.Destroy()
	}

	for k, v := range l.Headers {
		if strings.EqualFold(k, "host") {
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
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		//nolint:errcheck // err can be ignored since it is just for logging
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("when writing to [%s] received status code, %d: %s", l.url, resp.StatusCode, body)
	}

	return nil
}

// Verify the label name matches the regex [a-zA-Z_][a-zA-Z0-9_]*
func sanitizeLabelName(name string) string {
	re := regexp.MustCompile(`^[^a-zA-Z_]`)
	result := re.ReplaceAllString(name, "_")

	re = regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(result, "_")
}

func init() {
	outputs.Add("loki", func() telegraf.Output {
		return &Loki{
			MetricNameLabel: "__name",
		}
	})
}
