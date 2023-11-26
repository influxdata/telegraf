package qbittorrent

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//go:embed sample.conf
var sampleConfig string

var globalMainData *MainData

type QBittorrent struct {
	//todo support URLS
	URL      string        `toml:"url"`
	Username config.Secret `toml:"username"`
	Password config.Secret `toml:"password"`

	cookie []*http.Cookie
}

func (q *QBittorrent) Init() error {
	if q.Username.Empty() && q.Password.Empty() {
		return fmt.Errorf("username and password is empty")
	}
	return nil
}

func (*QBittorrent) SampleConfig() string {
	return sampleConfig
}

// Gather ..
func (q *QBittorrent) Gather(acc telegraf.Accumulator) error {
	err := q.getSyncData()
	if err != nil {
		return err
	}
	for k, v := range globalMainData.toMetrics() {
		for i := range v {
			acc.AddFields(k, v[i].Fields(), v[i].Tags())
		}
	}

	return nil
}
func (q *QBittorrent) getSyncData() error {
	var mainData MainData

	param := url.Values{}

	if globalMainData != nil {
		param.Set("rid", strconv.Itoa(int(globalMainData.RID)))
	}
	measure, _, err := q.getMeasure("GET", "/api/v2/sync/maindata", nil, param, nil)
	if err != nil {
		return err
	}

	jErr := json.Unmarshal([]byte(measure), &mainData)
	if err != nil {
		return jErr
	}
	if globalMainData == nil {
		globalMainData = &mainData
	} else {
		//partial update
		globalMainData.partialUpdate(&mainData)
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

func (q *QBittorrent) getMeasure(method string, path string, headers map[string]string, param url.Values, reqBody io.Reader) (string, int, error) {
	if q.cookie == nil || len(q.cookie) == 0 {
		cookie, err := q.login()
		if err != nil {
			return "", -1, err
		}
		q.cookie = cookie
	}

	getURL, err := q.getURL(path)
	if err != nil {
		return "", -1, err
	}

	paramStr := param.Encode()
	reqURL := getURL.String()
	if paramStr != "" {
		reqURL = fmt.Sprintf("%s?%s", reqURL, paramStr)
	}

	// Create + send request
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return "", -1, err
	}
	for c := range q.cookie {
		req.AddCookie(q.cookie[c])
	}

	// Add header parameters
	for k, v := range headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Add(k, v)
		}
	}

	var client = new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return string(respBody), 1, err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response from url %q has status code %d (%s), expected %d (%s)",
			reqURL,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return string(respBody), resp.StatusCode, err
	}

	return string(respBody), resp.StatusCode, nil
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
		var qb = QBittorrent{URL: "http://127.0.0.1:8080", Username: config.NewSecret([]byte("admin")), Password: config.NewSecret([]byte("admin"))}
		return &qb
	})
}
