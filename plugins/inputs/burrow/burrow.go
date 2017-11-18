package burrow

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const configSample = `
  ## Burrow endpoints in format "sheme://[user:password@]host:port"
  ## e.g.
  ##   servers = ["http://localhost:8080"]
  ##   servers = ["https://example.com:8000"]
  ##   servers = ["http://user:pass@example.com:8000"]
  ##
  servers = [ "http://127.0.0.1:8000" ]

  ## Prefix all HTTP API requests.
  #api_prefix = "/v2/kafka"

  ## Maximum time to receive response.
  #timeout = "5s"

  ## Optional, gather info only about specific clusters.
  ## Default is gather all.
  #clusters = ["clustername1"]

  ## Optional, gather stats only about specific groups.
  ## Default is gather all.
  #groups = ["group1"]

  ## Optional, gather info only about specific topics.
  ## Default is gather all
  #topics = ["topicA"]

  ## Concurrent connections limit (per server), default is 4.
  #max_concurrent_connections = 10

  ## Internal working queue adjustments (per measurement, per server), default is 4.
  #worker_queue_length = 5

  ## Credentials for basic HTTP authentication.
  #username = ""
  #password = ""

  ## Optional SSL config
  #ssl_ca = "/etc/telegraf/ca.pem"
  #ssl_cert = "/etc/telegraf/cert.pem"
  #ssl_key = "/etc/telegraf/key.pem"

  ## Use SSL but skip chain & host verification
  #insecure_skip_verify = false
`

type (
	// burrow plugin
	burrow struct {
		Servers []string

		Username string
		Password string
		Timeout  internal.Duration

		APIPrefix string `toml:"api_prefix"`

		Clusters []string
		Groups   []string
		Topics   []string

		MaxConcurrentConnections int `toml:"max_concurrent_connections"`
		WorkerQueueLength        int `toml:"worker_queue_length"`

		// Path to CA file
		SSLCA string `toml:"ssl_ca"`
		// Path to host cert file
		SSLCert string `toml:"ssl_cert"`
		// Path to cert key file
		SSLKey string `toml:"ssl_key"`
		// Use SSL but skip chain & host verification
		InsecureSkipVerify bool
	}

	// function prototype for worker spawning helper
	resolverFn func(api apiClient, res apiResponse, uri string)
)

var (
	statusMapping = map[string]int{
		"OK":        1,
		"NOT_FOUND": 2,
		"WARN":      3,
		"ERR":       4,
		"STOP":      5,
		"STALL":     6,
	}
)

func init() {
	inputs.Add("burrow", func() telegraf.Input {
		return &burrow{}
	})
}

func (b *burrow) SampleConfig() string {
	return configSample
}

func (b *burrow) Description() string {
	return "Collect Kafka topics and consumers status from Burrow HTTP API."
}

// Gather Burrow stats
func (b *burrow) Gather(acc telegraf.Accumulator) error {
	var workers sync.WaitGroup

	errorChan := b.getErrorChannel(acc)
	for _, addr := range b.Servers {
		c, err := b.getClient(acc, addr, errorChan)
		if err != nil {
			errorChan <- err
			continue
		}

		endpointChan := make(chan string)
		workers.Add(2) // will spawn two workers

		go withAPICall(c, endpointChan, nil, func(api apiClient, res apiResponse, endpoint string) {
			clusters := whitelistSlice(res.Clusters, api.limitClusters)

			go gatherTopicStats(api, clusters, &workers)
			go gatherGroupStats(api, clusters, &workers)
		})

		endpointChan <- c.apiPrefix
		close(endpointChan)
	}

	workers.Wait()
	close(errorChan)

	return nil
}

// Error collector / register
func (b *burrow) getErrorChannel(acc telegraf.Accumulator) chan error {
	errorChan := make(chan error)
	go func(acc telegraf.Accumulator) {
		for {
			err := <-errorChan
			if err != nil {
				acc.AddError(err)
			} else {
				break
			}
		}
	}(acc)

	return errorChan
}

// API client construction
func (b *burrow) getClient(acc telegraf.Accumulator, addr string, errorChan chan<- error) (apiClient, error) {
	var c apiClient

	u, err := url.Parse(addr)
	if err != nil {
		return c, err
	}

	// override global credentials (if endpoint contains auth credentials)
	requestUser := b.Username
	requestPass := b.Password
	if u.User != nil {
		requestUser = u.User.Username()
		requestPass, _ = u.User.Password()
	}

	// enable SSL configuration (if provided by configuration)
	tlsCfg, err := internal.GetTLSConfig(b.SSLCert, b.SSLKey, b.SSLCA, b.InsecureSkipVerify)
	if err != nil {
		return c, err
	}

	if b.APIPrefix == "" {
		b.APIPrefix = "/v2/kafka"
	}

	if b.MaxConcurrentConnections < 1 {
		b.MaxConcurrentConnections = 10
	}

	if b.WorkerQueueLength < 1 {
		b.WorkerQueueLength = 5
	}

	if b.Timeout.Duration < time.Second {
		b.Timeout.Duration = time.Second * 5
	}

	c = apiClient{
		client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
			},
			Timeout: b.Timeout.Duration,
		},

		acc:       acc,
		apiPrefix: b.APIPrefix,
		baseURL:   fmt.Sprintf("%s://%s", u.Scheme, u.Host),

		limitClusters: b.Clusters,
		limitGroups:   b.Groups,
		limitTopics:   b.Topics,

		guardChan:   make(chan bool, b.MaxConcurrentConnections),
		errorChan:   errorChan,
		workerCount: b.WorkerQueueLength,

		requestUser: requestUser,
		requestPass: requestPass,
	}

	return c, nil
}

func remapStatus(src string) int {
	if status, ok := statusMapping[src]; ok {
		return status
	}

	return 0
}

// whitelist function
func whitelistSlice(src, items []string) []string {
	var result []string

	if len(items) == 0 {
		return src
	}

	for _, w := range items {
		for _, s := range src {
			if w == s {
				result = append(result, s)
				break
			}
		}
	}

	return result
}

// worker spawn helper function
func withAPICall(api apiClient, producer <-chan string, done chan<- bool, resolver resolverFn) {
	for {
		uri := <-producer
		if uri == "" {
			break
		}

		res, err := api.call(uri)
		if err != nil {
			api.errorChan <- err
		}

		resolver(api, res, uri)
		if done != nil {
			done <- true
		}
	}
}
