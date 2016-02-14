package kubernetes

// Plugin inspired from
// https://github.com/prometheus/prom2json/blob/master/main.go

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Kubernetes struct {
	Apiserver         []Apiserver
	Scheduler         []Scheduler
	Controllermanager []Controllermanager
	Kubelet           []Kubelet
	client            HTTPClient
}

type KubeService struct {
	Url      string
	Endpoint string
	Timeout  float64
	Excludes []string
	Includes []string
}

type Apiserver struct {
	KubeService
}

type Scheduler struct {
	KubeService
}

type Controllermanager struct {
	KubeService
}

type Kubelet struct {
	KubeService
}

type HTTPClient interface {
	// Returns the result of an http request
	//
	// Parameters:
	// req: HTTP request object
	//
	// Returns:
	// http.Response:  HTTP respons object
	// error        :  Any error that may have occurred
	MakeRequest(req *http.Request, timeout float64) (*http.Response, error)
}

type RealHTTPClient struct {
	client *http.Client
}

func (c RealHTTPClient) MakeRequest(req *http.Request, timeout float64) (*http.Response, error) {
	c.client.Timeout = time.Duration(timeout) * time.Second
	return c.client.Do(req)
}

var sampleConfig = `
# Get metrics from Kubernetes services    
  [[inputs.kubernetes.apiserver]]
    url = "http://mtl-nvcbladea-15.nuance.com:8080"
    endpoint = "/metrics"
    timeout = 5.0
    # includes only metrics which match one of the
    # following regexp
    includes = ["apiserver_.*"]

  [[inputs.kubernetes.scheduler]]
    url = "http://mtl-nvcbladea-15.nuance.com:10251"
    endpoint = "/metrics"
    timeout = 1.0
    # DO NOT include metrics which match one of the
    # following regexp
    excludes = ["scheduler_.*"]

  [[inputs.kubernetes.controllermanager]]
    url = "http://mtl-nvcbladea-15.nuance.com:10252"

  [[inputs.kubernetes.kubelet]]
    url = "http://mtl-nvcbladea-15.nuance.com:4194"
    # You should increase metric_buffer_limit
    # Because of number of kubelet metrics
    # otherwise you can limit metrics with
    # the following 'excludes' argument
    excludes = ["container_.*"]

`

func (k *Kubernetes) SampleConfig() string {
	return sampleConfig
}

func (k *Kubernetes) Description() string {
	return "Read metrics from Kubernetes services"
}

// Gathers data for all servers.
func (k *Kubernetes) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	errorChannel := make(chan error,
		len(k.Apiserver)+len(k.Scheduler)+
			len(k.Controllermanager)+len(k.Kubelet))

	// Routine to gather metrics and add it to acc
	GatherMetric := func(url string, endpoint string, serviceType string,
		excludes []string, includes []string, timeout float64) {
		defer wg.Done()
		if timeout == 0. {
			timeout = 1.0
		}

		if endpoint == "" {
			endpoint = "/metrics"
		} else if string(endpoint[0]) != "/" {
			endpoint = "/" + endpoint
		}
		url = url + endpoint
		if err := k.gatherServer(acc, url, serviceType, excludes, includes, timeout); err != nil {
			errorChannel <- err
		}
	}

	// Apiservers
	for _, service := range k.Apiserver {
		wg.Add(1)
		serviceType := "apiserver"
		go GatherMetric(service.Url, service.Endpoint, serviceType, service.Excludes, service.Includes, service.Timeout)
	}
	// Schedulers
	for _, service := range k.Scheduler {
		wg.Add(1)
		serviceType := "scheduler"
		go GatherMetric(service.Url, service.Endpoint, serviceType, service.Excludes, service.Includes, service.Timeout)
	}
	// Controllermanager
	for _, service := range k.Controllermanager {
		wg.Add(1)
		serviceType := "controllermanager"
		go GatherMetric(service.Url, service.Endpoint, serviceType, service.Excludes, service.Includes, service.Timeout)
	}
	// Kubelet
	for _, service := range k.Kubelet {
		wg.Add(1)
		serviceType := "kubelet"
		go GatherMetric(service.Url, service.Endpoint, serviceType, service.Excludes, service.Includes, service.Timeout)
	}

	wg.Wait()
	close(errorChannel)

	// Get all errors and return them as one giant error
	errorStrings := []string{}
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorStrings, "\n"))
}

// Gathers data from a particular server
// Parameters:
//     acc          : The telegraf Accumulator to use
//     serverURL    : endpoint to send request to
//     serviceType  : service type (apiserver, kubelet, ...)
//     timeout      : http timeout
//
// Returns:
//     error: Any error that may have occurred
func (k *Kubernetes) gatherServer(
	acc telegraf.Accumulator,
	serviceURL string,
	serviceType string,
	excludes []string,
	includes []string,
	timeout float64,
) error {
	// Get raw data from Kube service
	collectDate := time.Now()
	raw_data, err := k.sendRequest(serviceURL, timeout)
	if err != nil {
		return err
	}
	// Prepare Prometheus parser config
	config := parsers.Config{
		DataFormat: "prometheus",
	}
	// Create Prometheus parser
	promparser, err := parsers.NewParser(&config)
	if err != nil {
		return err
	}
	// Set default tags
	tags := map[string]string{
		"kubeservice": serviceType,
		"serverURL":   serviceURL,
	}
	promparser.SetDefaultTags(tags)

	// Parseing
	metrics, err := promparser.Parse(raw_data)
	if err != nil {
		return err
	}
	// Add (or not) collected metrics
	for _, metric := range metrics {
		if len(includes) > 0 {
			// includes regexp
		IncludeMetric:
			for _, include := range includes {
				r, err := regexp.Compile(include)
				if err == nil {
					if r.MatchString(metric.Name()) {
						acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), collectDate)
						break IncludeMetric
					}
				}
			}
		} else if len(excludes) > 0 {
			// excludes regexp
			includeMetric := true
		ExcludeMetric:
			for _, exclude := range excludes {
				r, err := regexp.Compile(exclude)
				if err == nil {
					if r.MatchString(metric.Name()) {
						includeMetric = false
						break ExcludeMetric
					}
				}
			}
			if includeMetric {
				acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), collectDate)
			}
		} else {
			// no includes/excludes regexp
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}

	return nil
}

// Sends an HTTP request to the server using the Kubernetes object's HTTPClient
// Parameters:
//     serverURL: endpoint to send request to
//     timeout:   request timeout
//
// Returns:
//     []byte: body of the response
//     error : Any error that may have occurred
func (k *Kubernetes) sendRequest(serverURL string, timeout float64) ([]byte, error) {
	// Prepare URL
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("Invalid server URL \"%s\"", serverURL)
	}
	params := url.Values{}
	requestURL.RawQuery = params.Encode()

	// Create request
	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Make request
	resp, err := k.client.MakeRequest(req, timeout)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestURL.String(),
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	return body, err
}

func init() {
	inputs.Add("kubernetes", func() telegraf.Input {
		return &Kubernetes{client: RealHTTPClient{client: &http.Client{}}}
	})
}
