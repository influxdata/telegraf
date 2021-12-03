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

	"github.com/golang-jwt/jwt/v4"
)

const (
	// How long to stayed logged in for
	loginDuration = 65 * time.Minute
)

// Client is an interface for communicating with the DC/OS API.
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
	URL         string
	StatusCode  int
	Title       string
	Description string
}

// Login is request data for logging in.
type Login struct {
	UID   string `json:"uid"`
	Exp   int64  `json:"exp"`
	Token string `json:"token"`
}

// LoginError is the response when login fails.
type LoginError struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// LoginAuth is the response to a successful login.
type LoginAuth struct {
	Token string `json:"token"`
}

// Slave is a node in the cluster.
type Slave struct {
	ID string `json:"id"`
}

// Summary provides high level cluster wide information.
type Summary struct {
	Cluster string
	Slaves  []Slave
}

// Container is a container on a node.
type Container struct {
	ID string
}

type DataPoint struct {
	Name  string            `json:"name"`
	Tags  map[string]string `json:"tags"`
	Unit  string            `json:"unit"`
	Value float64           `json:"value"`
}

// Metrics are the DCOS metrics
type Metrics struct {
	Datapoints []DataPoint            `json:"datapoints"`
	Dimensions map[string]interface{} `json:"dimensions"`
}

// AuthToken is the authentication token.
type AuthToken struct {
	Text   string
	Expire time.Time
}

// ClusterClient is a Client that uses the cluster URL.
type ClusterClient struct {
	clusterURL *url.URL
	httpClient *http.Client
	token      string
	semaphore  chan struct{}
}

type claims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

func (e APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("[%s] %s: %s", e.URL, e.Title, e.Description)
	}
	return fmt.Sprintf("[%s] %s", e.URL, e.Title)
}

func NewClusterClient(
	clusterURL *url.URL,
	timeout time.Duration,
	maxConns int,
	tlsConfig *tls.Config,
) *ClusterClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    maxConns,
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}
	semaphore := make(chan struct{}, maxConns)

	c := &ClusterClient{
		clusterURL: clusterURL,
		httpClient: httpClient,
		semaphore:  semaphore,
	}
	return c
}

func (c *ClusterClient) SetToken(token string) {
	c.token = token
}

func (c *ClusterClient) Login(ctx context.Context, sa *ServiceAccount) (*AuthToken, error) {
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

	loc := c.toURL("/acs/api/v1/auth/login")
	req, err := http.NewRequest("POST", loc, bytes.NewBuffer(octets))
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
			Text:   auth.Token,
			Expire: exp,
		}
		return token, nil
	}

	loginError := &LoginError{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(loginError)
	if err != nil {
		err := &APIError{
			URL:        loc,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
		return nil, err
	}

	err = &APIError{
		URL:         loc,
		StatusCode:  resp.StatusCode,
		Title:       loginError.Title,
		Description: loginError.Description,
	}
	return nil, err
}

func (c *ClusterClient) GetSummary(ctx context.Context) (*Summary, error) {
	summary := &Summary{}
	err := c.doGet(ctx, c.toURL("/mesos/master/state-summary"), summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (c *ClusterClient) GetContainers(ctx context.Context, node string) ([]Container, error) {
	list := []string{}

	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers", node)
	err := c.doGet(ctx, c.toURL(path), &list)
	if err != nil {
		return nil, err
	}

	containers := make([]Container, 0, len(list))
	for _, c := range list {
		containers = append(containers, Container{ID: c})
	}

	return containers, nil
}

func (c *ClusterClient) getMetrics(ctx context.Context, address string) (*Metrics, error) {
	metrics := &Metrics{}

	err := c.doGet(ctx, address, metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *ClusterClient) GetNodeMetrics(ctx context.Context, node string) (*Metrics, error) {
	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/node", node)
	return c.getMetrics(ctx, c.toURL(path))
}

func (c *ClusterClient) GetContainerMetrics(ctx context.Context, node, container string) (*Metrics, error) {
	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers/%s", node, container)
	return c.getMetrics(ctx, c.toURL(path))
}

func (c *ClusterClient) GetAppMetrics(ctx context.Context, node, container string) (*Metrics, error) {
	path := fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers/%s/app", node, container)
	return c.getMetrics(ctx, c.toURL(path))
}

func createGetRequest(address string, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Add("Authorization", "token="+token)
	}
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (c *ClusterClient) doGet(ctx context.Context, address string, v interface{}) error {
	req, err := createGetRequest(address, c.token)
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
		//nolint:errcheck,revive // we cannot do anything if the closing fails
		resp.Body.Close()
		<-c.semaphore
	}()

	// Clear invalid token if unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		c.token = ""
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			URL:        address,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	return err
}

func (c *ClusterClient) toURL(path string) string {
	clusterURL := *c.clusterURL
	clusterURL.Path = path
	return clusterURL.String()
}

func (c *ClusterClient) createLoginToken(sa *ServiceAccount) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims{
		UID: sa.AccountID,
		RegisteredClaims: jwt.RegisteredClaims{
			// How long we have to login with this token
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
		},
	})
	return token.SignedString(sa.PrivateKey)
}
