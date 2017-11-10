package dcos

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// How long to stayed logged in for
	loginDuration = 65 * time.Minute

	// How long before expiration to renew token
	relogDuration = 5 * time.Minute
)

type Client interface {
	Token() string
	EnsureAuth(ctx context.Context) error
	GetSummary(ctx context.Context) (*Summary, error)
	GetContainers(ctx context.Context, node string) ([]string, error)
	GetNodeMetrics(ctx context.Context, node string) (*Metrics, error)
	GetContainerMetrics(ctx context.Context, node, container string) (*Metrics, error)
	GetAppMetrics(ctx context.Context, node, container string) (*Metrics, error)
}

type Credentials struct {
	Username   string
	PrivateKey *rsa.PrivateKey
	TokenFile  string
}

type APIError struct {
	StatusCode  int
	Title       string
	Description string
}
type Login struct {
	Token       string
	Title       string
	Description string
}

type Slave struct {
	ID string `json:"id"`
}

type Summary struct {
	Cluster string
	Slaves  []Slave
}

type DataPoint struct {
	Name  string
	Tags  map[string]string
	Unit  string
	Value float64
}

type Metrics struct {
	Datapoints []DataPoint
	Dimensions map[string]interface{}
}

type client struct {
	clusterURL  *url.URL
	httpClient  *http.Client
	credentials *Credentials
	token       *authToken
	semaphore   chan struct{}
}

type claims struct {
	Uid string `json:"uid"`
	jwt.StandardClaims
}

type authToken struct {
	text   string
	expire time.Time
}

func (e APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Title, e.Description)
	}
	return e.Title
}

func NewClient(
	clusterURL *url.URL,
	creds *Credentials,
	timeout time.Duration,
	maxConns int,
	tlsConfig *tls.Config,
) *client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    maxConns,
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}
	semaphore := make(chan struct{}, maxConns)

	c := &client{
		clusterURL:  clusterURL,
		httpClient:  httpClient,
		credentials: creds,
		semaphore:   semaphore,
	}
	return c
}

func (c *client) Token() string {
	if c.token == nil {
		return ""
	}
	return c.token.text
}

func (c *client) EnsureAuth(ctx context.Context) error {
	if c.credentials == nil {
		return nil
	}

	if c.credentials.TokenFile != "" {
		tf := c.credentials.TokenFile
		tokenData, err := ioutil.ReadFile(tf)
		if err != nil {
			return fmt.Errorf("Error opening token_file %q: %s", tf, err)
		}
		if !utf8.Valid(tokenData) {
			return fmt.Errorf("Token file does not contain utf-8 encoded text: %s", tf)
		}
		token := strings.TrimSpace(string(tokenData))
		c.token = &authToken{text: token}
	}

	if c.token == nil || c.token.expire.Add(relogDuration).After(time.Now()) {
		token, err := c.login(ctx)
		if err != nil {
			return err
		}
		c.token = token
	}
	return nil
}

func (c *client) GetSummary(ctx context.Context) (*Summary, error) {
	summary := &Summary{}
	err := c.doGet(ctx, c.url("/mesos/master/state-summary"), summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (c *client) GetContainers(ctx context.Context, node string) ([]string, error) {
	containers := []string{}

	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers", node)
	err := c.doGet(ctx, c.url(path), &containers)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (c *client) GetNodeMetrics(ctx context.Context, node string) (*Metrics, error) {
	metrics := &Metrics{}

	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/node", node)
	err := c.doGet(ctx, c.url(path), metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *client) GetContainerMetrics(ctx context.Context, node, container string) (*Metrics, error) {
	metrics := &Metrics{}

	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers/%s", node, container)
	err := c.doGet(ctx, c.url(path), metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *client) GetAppMetrics(ctx context.Context, node, container string) (*Metrics, error) {
	metrics := &Metrics{}

	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers/%s/app", node, container)
	err := c.doGet(ctx, c.url(path), metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func createGetRequest(url string, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Add("Authorization", "token="+token)
	}
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (c *client) doGet(ctx context.Context, url string, v interface{}) error {
	req, err := createGetRequest(url, c.Token())
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
	defer resp.Body.Close()

	// Clear invalid token if unauthorized
	if resp.StatusCode == 401 {
		c.token = nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		<-c.semaphore
		return &APIError{
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}

	if resp.StatusCode == 204 {
		<-c.semaphore
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	<-c.semaphore
	return err
}
func (c *client) url(path string) string {
	c.clusterURL.Path = path
	return c.clusterURL.String()
}

func (c *client) login(ctx context.Context) (*authToken, error) {
	token, err := c.createLoginToken()
	if err != nil {
		return nil, err
	}

	exp := time.Now().Add(loginDuration)

	body := map[string]interface{}{
		"uid":   c.credentials.Username,
		"exp":   exp.Unix(),
		"token": token,
	}

	octets, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url("/acs/api/v1/auth/login"), bytes.NewBuffer(octets))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	login := Login{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&login)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 || login.Token == "" {
		return nil, &APIError{resp.StatusCode, login.Title, login.Description}
	}

	authToken := &authToken{
		text:   login.Token,
		expire: exp,
	}

	return authToken, err
}

func (c *client) createLoginToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims{
		Uid: c.credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 0,
		},
	})
	ss, err := token.SignedString(c.credentials.PrivateKey)
	return ss, err
}
