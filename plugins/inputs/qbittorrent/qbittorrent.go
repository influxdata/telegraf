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
	Host     string        `toml:"host"`
	Port     int           `toml:"port"`
	Username config.Secret `toml:"username"`
	Password config.Secret `toml:"password"`
	Tls      bool          `toml:"tls"`
	Cookie   []*http.Cookie
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
	for k, v := range globalMainData.toFields() {
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
		param.Set("rid", strconv.Itoa(globalMainData.RID))
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

// getUrl returns a URL object constructed from the given path and the QBittorrent
// configuration. The returned URL is constructed using the scheme (http or https)
// specified in the configuration. The path is appended to the base URL constructed
// using the host and port from the configuration. If the URL is invalid, an error is
// returned.
//
// path: a string representing the path to be appended to the base URL.
// Returns a URL object and an error.
func (q *QBittorrent) getUrl(path string) (*url.URL, error) {
	var scheme string
	if q.Tls {
		scheme = "https"
	} else {
		scheme = "http"
	}
	strUrl := fmt.Sprintf("%s://%s:%d/%s", scheme, q.Host, q.Port, strings.TrimLeft(path, "/"))
	parseUrl, err := url.Parse(strUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL %q", strUrl)
	}
	return parseUrl, nil
}

func (q *QBittorrent) getMeasure(method string, path string, headers map[string]string, param url.Values, reqBody io.Reader) (string, int, error) {
	if q.Cookie == nil || len(q.Cookie) == 0 {
		cookie, err := q.login()
		if err != nil {
			return "", -1, err
		}
		q.Cookie = cookie
	}

	getUrl, err := q.getUrl(path)
	if err != nil {
		return "", -1, err
	}

	paramStr := param.Encode()
	reqUrl := getUrl.String()
	if paramStr != "" {
		reqUrl = fmt.Sprintf("%s?%s", reqUrl, paramStr)
	}

	// Create + send request
	req, err := http.NewRequest(method, reqUrl, reqBody)
	if err != nil {
		return "", -1, err
	}
	for c := range q.Cookie {
		req.AddCookie(q.Cookie[c])
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
			reqUrl,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return string(respBody), resp.StatusCode, err
	}

	return string(respBody), resp.StatusCode, nil
}

func (q *QBittorrent) login() ([]*http.Cookie, error) {
	getUrl, err := q.getUrl("/api/v2/auth/login")
	if err != nil {
		return nil, err
	}

	if q.Username.Empty() && q.Password.Empty() {
		return nil, fmt.Errorf("username and password is empty")
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

	req, err := http.NewRequest("POST", getUrl.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", fmt.Sprintf("%s://%s", getUrl.Scheme, getUrl.Host))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var client = new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("can not auth,may be server url is not corret")
	}
	cookie := resp.Cookies()
	if cookie == nil || len(cookie) == 0 {
		return nil, fmt.Errorf("can not auth,may be username or password is not corret")
	}

	return cookie, nil
}
func init() {
	inputs.Add("qbittorrent", func() telegraf.Input {
		var qb = QBittorrent{Host: "127.0.0.1", Port: 8080, Username: config.NewSecret([]byte("admin")), Password: config.NewSecret([]byte("admin"))}
		return &qb
	})
}
