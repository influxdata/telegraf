package dcos

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// How long to stayed logged in for
	loginDuration = 65 * time.Minute
)

type Client interface {
	SetToken(token string)

	Login(ctx context.Context, sa *ServiceAccount) (*AuthToken, error)
	GetSummary(ctx context.Context) (*Summary, error)
	GetContainers(ctx context.Context, node string) ([]Container, error)
	GetNodeMetrics(ctx context.Context, node string) (*Metrics, error)
	GetContainerMetrics(ctx context.Context, node, container string) (*Metrics, error)
	GetAppMetrics(ctx context.Context, node, container string) (*Metrics, error)
}

type APIError struct {
	StatusCode  int
	Title       string
	Description string
}

type Login struct {
	UID   string `json:"uid"`
	Exp   int64  `json:"exp"`
	Token string `json:"token"`
}

type LoginError struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type LoginAuth struct {
	Token string `json:"token"`
}

type Slave struct {
	ID string `json:"id"`
}

type Summary struct {
	Cluster string
	Slaves  []Slave
}

type Container struct {
	ID string
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
	token       string
	semaphore   chan struct{}
}

type claims struct {
	UID string `json:"uid"`
	jwt.StandardClaims
}

type AuthToken struct {
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
		clusterURL: clusterURL,
		httpClient: httpClient,
		semaphore:  semaphore,
	}
	return c
}

func (c *client) SetToken(token string) {
	c.token = token
}

func (c *client) Login(ctx context.Context, sa *ServiceAccount) (*AuthToken, error) {
	token, err := c.createLoginToken(sa)
	if err != nil {
		return nil, err
	}

	exp := time.Now().Add(loginDuration)

	body := &Login{
		UID:   sa.AccountID,
		Exp:   exp.Unix(),
		Token: token,
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

	if resp.StatusCode == http.StatusOK {
		auth := &LoginAuth{}
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(auth)
		if err != nil {
			return nil, err
		}

		token := &AuthToken{
			text:   auth.Token,
			expire: exp,
		}
		return token, nil
	}

	loginError := &LoginError{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(loginError)
	if err != nil {
		err := &APIError{
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
		return nil, err
	}

	err = &APIError{
		StatusCode:  resp.StatusCode,
		Title:       loginError.Title,
		Description: loginError.Description,
	}
	return nil, err
}

func (c *client) GetSummary(ctx context.Context) (*Summary, error) {
	summary := &Summary{}
	err := c.doGet(ctx, c.url("/mesos/master/state-summary"), summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (c *client) GetContainers(ctx context.Context, node string) ([]Container, error) {
	list := []string{}

	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers", node)
	err := c.doGet(ctx, c.url(path), &list)
	if err != nil {
		return nil, err
	}

	containers := make([]Container, 0, len(list))
	for _, c := range list {
		containers = append(containers, Container{ID: c})

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
	req, err := createGetRequest(url, c.token)
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
		c.token = ""
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
	url := c.clusterURL
	url.Path = path
	return url.String()
}

func (c *client) createLoginToken(sa *ServiceAccount) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims{
		UID: sa.AccountID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 0,
		},
	})
	return token.SignedString(sa.PrivateKey)
}
