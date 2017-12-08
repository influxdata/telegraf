package libkflow_test

import (
	"net/url"
	"testing"

	"github.com/kentik/libkflow"
	"github.com/kentik/libkflow/api"
	"github.com/kentik/libkflow/api/test"
	"github.com/stretchr/testify/assert"
)

func TestNewSenderWithDeviceID(t *testing.T) {
	dev, assert := setupLibTest(t)

	errors := make(chan error, 100)
	config := libkflow.NewConfig(email, token, "test", "0.0.1")
	config.OverrideURLs(apiurl, flowurl, metricsurl)

	s, err := libkflow.NewSenderWithDeviceID(dev.ID, errors, config)

	assert.NotNil(s)
	assert.Nil(err)
}

func TestNewSenderWithDeviceIP(t *testing.T) {
	dev, assert := setupLibTest(t)

	errors := make(chan error, 100)
	config := libkflow.NewConfig(email, token, "test", "0.0.1")
	config.OverrideURLs(apiurl, flowurl, metricsurl)

	s, err := libkflow.NewSenderWithDeviceIP(dev.IP, errors, config)

	assert.NotNil(s)
	assert.Nil(err)
}

func TestMetricsConfig(t *testing.T) {
	dev, assert := setupLibTest(t)

	program := "test"
	version := "0.0.1"

	config := libkflow.NewConfig(email, token, "test", "0.0.1")
	metrics := config.NewMetrics(dev)

	assert.Equal(program+"-"+version, metrics.Extra["ver"])
	assert.Equal(program, metrics.Extra["ft"])
	assert.Equal("libkflow", metrics.Extra["dt"])
	assert.Equal("primary", metrics.Extra["level"])
}

func setupLibTest(t *testing.T) (*api.Device, *assert.Assertions) {
	client, server, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)

	apiurl = server.URL(test.API)
	flowurl = server.URL(test.FLOW)
	metricsurl = server.URL(test.TSDB)

	email = client.Email
	token = client.Token

	return device, assert
}

var (
	apiurl     *url.URL
	flowurl    *url.URL
	metricsurl *url.URL
	email      string
	token      string
)
