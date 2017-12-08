package libkflow

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/kentik/libkflow/agg"
	"github.com/kentik/libkflow/api"
)

// Config describes the libkflow configuration.
type Config struct {
	email   string
	token   string
	capture Capture
	proxy   *url.URL
	api     *url.URL
	flow    *url.URL
	metrics *url.URL
	verbose int
	timeout time.Duration
	program string
	version string
}

// Capture describes the packet capture settings.
type Capture struct {
	Device  string
	Snaplen int32
	Promisc bool
}

// NewConfig returns a new Config given an API access email and token,
// and the name and version of the program using libkflow.
func NewConfig(email, token, program, version string) *Config {
	return defaultConfig(email, token, program, version)
}

// SetCapture sets the packet capture details.
func (c *Config) SetCapture(capture Capture) {
	c.capture = capture
}

// SetProxy sets the HTTP proxy used for making API requests, sending
// flow, and sending metrics.
func (c *Config) SetProxy(url *url.URL) {
	c.proxy = url
}

// SetServer changes the host and port used for API requests, flow,
// and metrics.
func (c *Config) SetServer(host net.IP, port int) {
	base := "http://" + net.JoinHostPort(host.String(), strconv.Itoa(port))
	var (
		api     = parseURL(base + "/api/v5")
		flow    = parseURL(base + "/chf")
		metrics = parseURL(base + "/tsdb")
	)
	c.OverrideURLs(api, flow, metrics)
}

// SetTimeout sets the HTTP request timeout.
func (c *Config) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Set just the flow server
func (c *Config) SetFlow(server string) {
	c.flow = parseURL(server + "/chf")
}

// SetVerbose sets the verbosity level. Specifying a value greater than
// zero will cause verbose debug messages to be print to stderr.
func (c *Config) SetVerbose(verbose int) {
	c.verbose = verbose
}

// OverrideURLs changes the default endpoint URL for API requests,
// flow, and metrics.
func (c *Config) OverrideURLs(api, flow, metrics *url.URL) {
	c.api = api
	c.flow = flow
	c.metrics = metrics
}

func (c *Config) NewMetrics(dev *api.Device) *Metrics {
	return newMetrics(dev.ClientID(), c.program, c.version)
}

func (c *Config) client() *api.Client {
	return api.NewClient(api.ClientConfig{
		Email:   c.email,
		Token:   c.token,
		Timeout: c.timeout,
		API:     c.api,
		Proxy:   c.proxy,
	})
}

func (c *Config) start(client *api.Client, dev *api.Device, errors chan<- error) (*Sender, error) {
	interval := time.Duration(1) * time.Minute
	metrics := c.NewMetrics(dev)
	metrics.start(c.metrics.String(), c.email, c.token, interval, c.proxy)

	agg, err := agg.NewAgg(time.Second, dev.MaxFlowRate, &metrics.Metrics)
	if err != nil {
		return nil, fmt.Errorf("agg setup error: %s", err)
	}

	sender := newSender(c.flow, c.timeout, c.verbose)
	sender.Errors = errors

	if c.capture.Device != "" {
		nif, err := net.InterfaceByName(c.capture.Device)
		if err != nil {
			return nil, err
		}

		err = client.UpdateInterfaces(dev, nif)
		if err != nil {
			sender.debug("error updating device interfaces: %s", err)
		}
	}

	if err = sender.start(agg, client, dev, 2); err != nil {
		return nil, fmt.Errorf("send startup error: %s", err)
	}

	return sender, nil
}

func defaultConfig(email, token, program, version string) *Config {
	return &Config{
		email:   email,
		token:   token,
		capture: Capture{},
		proxy:   nil,
		api:     parseURL("https://api.kentik.com/api/internal"),
		flow:    parseURL("https://flow.kentik.com/chf"),
		metrics: parseURL("https://flow.kentik.com/tsdb"),
		verbose: 0,
		timeout: 10 * time.Second,
		program: program,
		version: version,
	}
}

func parseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic("invalid URL: " + s)
	}
	return u
}
