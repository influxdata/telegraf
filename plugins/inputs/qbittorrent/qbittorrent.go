//go:generate ../../../tools/readme_config_includer/generator
package qbittorrent

import (
	"bytes"
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

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type QBittorrent struct {
	URL      string        `toml:"url"`
	Username config.Secret `toml:"username"`
	Password config.Secret `toml:"password"`

	mainData *serverMetric
	cookie   []*http.Cookie
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
		return fmt.Errorf("invalid server URL %q", q.URL)
	}
	return nil
}

func (*QBittorrent) SampleConfig() string {
	return sampleConfig
}

func (q *QBittorrent) Gather(acc telegraf.Accumulator) error {
	err := q.getSyncData()
	if err != nil {
		return err
	}
	for _, m := range q.mainData.toMetrics(q.URL) {
		acc.AddMetric(m)
	}

	return nil
}

func (q *QBittorrent) getSyncData() error {
	param := url.Values{}
	if q.mainData != nil {
		param.Set("rid", strconv.Itoa(int(q.mainData.RID)))
	}
	measure, err := q.getMeasure(param, false)
	if err != nil {
		return err
	}

	var mainData serverMetric
	if err := json.Unmarshal([]byte(measure), &mainData); err != nil {
		return fmt.Errorf("decoding data failed: %w", err)
	}
	if q.mainData == nil {
		q.mainData = &mainData
	} else {
		//partial update
		q.mainData.partialUpdate(&mainData)
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

func (q *QBittorrent) getMeasure(param url.Values, retry bool) (string, error) {
	if q.cookie == nil || len(q.cookie) == 0 {
		cookie, err := q.login()
		if err != nil {
			return "", err
		}
		q.cookie = cookie
	}

	getURL, err := q.getURL("/api/v2/sync/maindata")
	if err != nil {
		return "", err
	}

	getURL.RawQuery = param.Encode()
	reqURL := getURL.String()

	// Create + send request
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	for c := range q.cookie {
		req.AddCookie(q.cookie[c])
	}

	//// Add header parameters
	//for k, v := range headers {
	//	if strings.ToLower(k) == "host" {
	//		req.Host = v
	//	} else {
	//		req.Header.Add(k, v)
	//	}
	//}

	var client = new(http.Client)
	resp, err := client.Do(req)
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
		if resp.StatusCode == http.StatusForbidden && !retry {
			// Reset cookie and retry
			q.cookie = nil
			return q.getMeasure(param, true)
		}
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

func (q *QBittorrent) login() ([]*http.Cookie, error) {
	getURL, err := q.getURL("/api/v2/auth/login")
	if err != nil {
		return nil, err
	}

	username, err := q.Username.Get()
	if err != nil {
		return nil, fmt.Errorf("getting username failed: %w", err)
	}
	defer username.Destroy()

	passwd, err := q.Password.Get()
	if err != nil {
		return nil, fmt.Errorf("getting password failed: %w", err)
	}
	defer passwd.Destroy()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("username", username.String())
	_ = writer.WriteField("password", passwd.String())
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", getURL.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", fmt.Sprintf("%s://%s", getURL.Scheme, getURL.Host))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var client = new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("can not auth,may be server url is not corret")
	}
	cookie := resp.Cookies()
	if len(cookie) == 0 {
		return nil, fmt.Errorf("can not auth,may be username or password is not corret")
	}

	return cookie, nil
}
func init() {
	inputs.Add("qbittorrent", func() telegraf.Input {
		return &QBittorrent{
			URL:      "http://127.0.0.1:8080",
			Username: config.NewSecret([]byte("admin")),
			Password: config.NewSecret([]byte("admin")),
		}
	})
}
