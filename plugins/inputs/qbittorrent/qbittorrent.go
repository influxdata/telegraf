//go:generate ../../../tools/readme_config_includer/generator
package qbittorrent

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/cookie"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type QBittorrent struct {
	URL      string        `toml:"url"`
	Username config.Secret `toml:"username"`
	Password config.Secret `toml:"password"`

	Log telegraf.Logger `toml:"-"`

	serverMetric *serverMetric
	httpconfig.HTTPClientConfig
	client *http.Client
}

func (q *QBittorrent) Init() error {
	if q.Username.Empty() {
		return errors.New("non-empty username required")
	}
	if q.Password.Empty() {
		return errors.New("non-empty password required")
	}
	_, err := url.Parse(q.URL)
	if err != nil {
		return fmt.Errorf("invalid server URL %q: %w", q.URL, err)
	}

	getURL, err := q.getURL("/api/v2/auth/login")
	if err != nil {
		return err
	}

	q.HTTPClientConfig.URL = getURL.String()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	username, err := q.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	_ = writer.WriteField("username", username.String())
	password, err := q.Password.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	_ = writer.WriteField("password", password.String())
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("writer close err: %w", err)
	}

	requester, err := io.ReadAll(payload)
	if err != nil {
		return fmt.Errorf("reading requset err:%w", err)
	}
	q.HTTPClientConfig.Body = string(requester)

	if q.HTTPClientConfig.Headers == nil {
		q.HTTPClientConfig.Headers = map[string]string{"Content-Type": writer.FormDataContentType()}
	} else {
		q.HTTPClientConfig.Headers["Content-Type"] = writer.FormDataContentType()
	}

	client, err := q.HTTPClientConfig.CreateClient(context.Background(), q.Log)
	if err != nil {
		return fmt.Errorf("create client err: %w", err)
	}
	q.client = client
	return nil
}

func (*QBittorrent) SampleConfig() string {
	return sampleConfig
}

func (q *QBittorrent) Gather(acc telegraf.Accumulator) error {
	measure, err := q.getMeasure()
	if err != nil {
		return err
	}

	var mainData serverMetric
	if err := json.Unmarshal([]byte(measure), &mainData); err != nil {
		return fmt.Errorf("decoding data failed: %w", err)
	}
	if q.serverMetric == nil {
		q.serverMetric = &mainData
	} else {
		//partial update
		q.serverMetric.partialUpdate(&mainData)
	}
	for _, m := range q.serverMetric.toMetrics(q.URL) {
		acc.AddMetric(m)
	}

	return nil
}

// getURL returns a URL object constructed from the given path and the QBittorrent
// configuration. The path is appended to the base URL constructed using the url
// from the configuration. If the URL is invalid, an error is returned.
//
// path: a string representing the path to be appended to the base URL.
// Returns a URL object and an error.
func (q *QBittorrent) getURL(path string) (*url.URL, error) {
	strURL := fmt.Sprintf("%s/%s", q.URL, strings.TrimLeft(path, "/"))
	parseURL, err := url.Parse(strURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL %q", strURL)
	}
	return parseURL, nil
}

func (q *QBittorrent) getMeasure() (string, error) {
	getURL, err := q.getURL("/api/v2/sync/maindata")
	if err != nil {
		return "", err
	}

	param := url.Values{}
	if q.serverMetric != nil {
		param.Set("rid", strconv.Itoa(int(q.serverMetric.RID)))
	}
	getURL.RawQuery = param.Encode()
	reqURL := getURL.String()

	resp, err := q.client.Get(reqURL)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return string(respBody), err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response from url %q has status code %d (%s), expected %d (%s)",
			reqURL,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return string(respBody), err
	}

	return string(respBody), nil
}

func init() {
	inputs.Add("qbittorrent", func() telegraf.Input {
		return &QBittorrent{
			URL:      "http://127.0.0.1:8080",
			Username: config.NewSecret([]byte("admin")),
			Password: config.NewSecret([]byte("admin")),
			HTTPClientConfig: httpconfig.HTTPClientConfig{
				Timeout: config.Duration(5 * time.Second),
				CookieAuthConfig: cookie.CookieAuthConfig{
					Renewal: config.Duration(3600 * time.Second),
				},
			},
		}
	})
}
