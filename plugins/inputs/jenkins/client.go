package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type client struct {
	baseURL       string
	httpClient    *http.Client
	username      string
	password      string
	sessionCookie *http.Cookie
	semaphore     chan struct{}
}

func newClient(httpClient *http.Client, url, username, password string, maxConnections int) *client {
	return &client{
		baseURL:    url,
		httpClient: httpClient,
		username:   username,
		password:   password,
		semaphore:  make(chan struct{}, maxConnections),
	}
}

func (c *client) init() error {
	// get session cookie
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return err
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for _, cc := range resp.Cookies() {
		if strings.Contains(cc.Name, "JSESSIONID") {
			c.sessionCookie = cc
			break
		}
	}
	// first api fetch
	if err := c.doGet(context.Background(), jobPath, new(jobResponse)); err != nil {
		return err
	}
	return nil
}

func (c *client) doGet(ctx context.Context, url string, v interface{}) error {
	req, err := createGetRequest(c.baseURL+url, c.username, c.password, c.sessionCookie)
	if err != nil {
		return err
	}
	select {
	case c.semaphore <- struct{}{}:
		break
	case <-ctx.Done():
		return ctx.Err()
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		<-c.semaphore
		return err
	}
	defer func() {
		resp.Body.Close()
		<-c.semaphore
	}()
	// Clear invalid token if unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		c.sessionCookie = nil
		return APIError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return APIError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}
	if resp.StatusCode == http.StatusNoContent {
		return APIError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}
	if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

type APIError struct {
	URL         string
	StatusCode  int
	Title       string
	Description string
}

func (e APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("[%s] %s: %s", e.URL, e.Title, e.Description)
	}
	return fmt.Sprintf("[%s] %s", e.URL, e.Title)
}

func createGetRequest(url string, username, password string, sessionCookie *http.Cookie) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}
	req.Header.Add("Accept", "application/json")
	return req, nil
}

func (c *client) getJobs(ctx context.Context, jr *jobRequest) (js *jobResponse, err error) {
	js = new(jobResponse)
	url := jobPath
	if jr != nil {
		url = jr.URL()
	}
	err = c.doGet(ctx, url, js)
	return js, err
}

func (c *client) getBuild(ctx context.Context, jr jobRequest, number int64) (b *buildResponse, err error) {
	b = new(buildResponse)
	url := jr.buildURL(number)
	err = c.doGet(ctx, url, b)
	return b, err
}

func (c *client) getAllNodes(ctx context.Context) (nodeResp *nodeResponse, err error) {
	nodeResp = new(nodeResponse)
	err = c.doGet(ctx, nodePath, nodeResp)
	return nodeResp, err
}
