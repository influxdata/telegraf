package sender

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	// defaultConnectTimeout is the maximum amount of time a dial will wait for a connect to complete.
	defaultConnectTimeout = time.Second * 3

	// defaultTimeout specifies a time limit for requests made by this client
	defaultTimeout = time.Second * 10

	headerAgent       = "User-Agent"
	headerContentType = "Content-Type"
	headerProxyAuth   = "Proxy-Authorization"
)

// Config contains sender configuration
type Config struct {
	ProxyURL  *url.URL
	Username  string
	Password  string
	TLSConfig *tls.Config
}

// Sender is a simple wrapper around standard HTTP client
type Sender struct {
	client    *http.Client
	proxyAuth string
}

// NewSender constructs a new HTTP sender that should be used to send requests to Sematext
func NewSender(config *Config) *Sender {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: defaultConnectTimeout,
		}).DialContext,
		TLSHandshakeTimeout: defaultConnectTimeout,
	}

	if config.ProxyURL != nil {
		transport.Proxy = http.ProxyURL(config.ProxyURL)
	}
	if config.TLSConfig != nil {
		transport.TLSClientConfig = config.TLSConfig
	}

	c := &http.Client{
		Timeout:   defaultTimeout,
		Transport: transport,
	}

	return &Sender{client: c, proxyAuth: getProxyHeader(config)}
}

// getProxyHeader creates proxy authentication header based on config settings
func getProxyHeader(config *Config) string {
	var proxyAuth string
	if len(config.Username) > 0 && len(config.Password) > 0 {
		auth := fmt.Sprintf("%s:%s", config.Username, config.Password)
		proxyAuth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
	}
	return proxyAuth
}

// Request emits an HTTP request targeting given HTTP method and URL.
func (s *Sender) Request(method, requestURL, contentType string, body []byte) (*http.Response, error) {
	req, err := s.createRequest(method, requestURL, contentType, body)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// createRequest forms the request object based on method, url, content type and body which should be sent
func (s *Sender) createRequest(method, requestURL, contentType string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerContentType, contentType)
	req.Header.Set(headerAgent, "telegraf")

	if len(s.proxyAuth) > 0 {
		req.Header.Add(headerProxyAuth, s.proxyAuth)
	}

	return req, nil
}

// Close is a method which clears resources used by the Sender, should be invoked once the Sender is not needed anymore
func (s *Sender) Close() {
	s.client = nil
}
