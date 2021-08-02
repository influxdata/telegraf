package sender

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestGetProxyHeaderNone(t *testing.T) {
	config := &Config{}
	assert.Equal(t, "", getProxyHeader(config))
}

func TestGetProxyHeaderNoUsernameOrPassword(t *testing.T) {
	proxyURL, _ := url.Parse("https://proxy.sematext.com:1234")
	config := &Config{
		ProxyURL: proxyURL,
	}
	assert.Equal(t, "", getProxyHeader(config))
}

func TestGetProxyHeaderWithAuth(t *testing.T) {
	proxyURL, _ := url.Parse("https://proxy.sematext.com:1234")
	config := &Config{
		ProxyURL: proxyURL,
		Username: "user",
		Password: "password",
	}
	assert.Equal(t, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("user:password"))),
		getProxyHeader(config))
}

func TestCreateRequestNoProxy(t *testing.T) {
	config := &Config{}
	sender := NewSender(config)
	req, err := sender.createRequest("POST", "www.sematext.com", "text/plain; charset=utf-8",
		[]byte("metrics"))

	assert.Nil(t, err)

	assert.Equal(t, len(req.Header), 2)
	assert.Equal(t, req.Header.Get(headerContentType), "text/plain; charset=utf-8")
	assert.Equal(t, req.Header.Get(headerAgent), "telegraf")
}

func TestCreateRequestWithProxy(t *testing.T) {
	proxyURL, _ := url.Parse("https://proxy.sematext.com:1234")
	config := &Config{
		ProxyURL: proxyURL,
		Username: "username",
		Password: "password",
	}
	sender := NewSender(config)

	assert.NotNil(t, sender.client.Transport.(*http.Transport).Proxy)

	req, err := sender.createRequest("POST", "www.sematext.com", "text/plain; charset=utf-8",
		[]byte("metrics"))

	assert.Nil(t, err)

	assert.Equal(t, len(req.Header), 3)
	assert.Equal(t, req.Header.Get(headerContentType), "text/plain; charset=utf-8")
	assert.Equal(t, req.Header.Get(headerAgent), "telegraf")
	assert.Equal(t, req.Header.Get(headerProxyAuth), getProxyHeader(config))
}
