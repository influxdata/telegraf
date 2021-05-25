package prometheus

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	parser_v2 "github.com/influxdata/telegraf/plugins/parsers/prometheus"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,*/*;q=0.1`

type Prometheus struct {
	// An array of urls to scrape metrics from.
	URLs []string `toml:"urls"`

	// An array of Kubernetes services to scrape metrics from.
	KubernetesServices []string

	// Location of kubernetes config file
	KubeConfig string

	// Label Selector/s for Kubernetes
	KubernetesLabelSelector string `toml:"kubernetes_label_selector"`

	// Field Selector/s for Kubernetes
	KubernetesFieldSelector string `toml:"kubernetes_field_selector"`

	// Bearer Token authorization file path
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`

	// Basic authentication credentials
	Username string `toml:"username"`
	Password string `toml:"password"`

	ResponseTimeout config.Duration `toml:"response_timeout"`

	MetricVersion int `toml:"metric_version"`

	URLTag string `toml:"url_tag"`

	tls.ClientConfig

	Log telegraf.Logger

	client  *http.Client
	headers map[string]string

	// Should we scrape Kubernetes services for prometheus annotations
	MonitorPods       bool   `toml:"monitor_kubernetes_pods"`
	PodScrapeScope    string `toml:"pod_scrape_scope"`
	NodeIP            string `toml:"node_ip"`
	PodScrapeInterval int    `toml:"pod_scrape_interval"`
	PodNamespace      string `toml:"monitor_kubernetes_pods_namespace"`
	lock              sync.Mutex
	kubernetesPods    map[string]URLAndAddress
	cancel            context.CancelFunc
	wg                sync.WaitGroup

	// Only for monitor_kubernetes_pods=true and pod_scrape_scope="node"
	podLabelSelector  labels.Selector
	podFieldSelector  fields.Selector
	isNodeScrapeScope bool
}

var sampleConfig = `
  ## An array of urls to scrape metrics from.
  urls = ["http://localhost:9100/metrics"]

  ## Metric version controls the mapping from Prometheus metrics into
  ## Telegraf metrics.  When using the prometheus_client output, use the same
  ## value in both plugins to ensure metrics are round-tripped without
  ## modification.
  ##
  ##   example: metric_version = 1; 
  ##            metric_version = 2; recommended version
  # metric_version = 1

  ## Url tag name (tag containing scrapped url. optional, default is "url")
  # url_tag = "scrapeUrl"

  ## An array of Kubernetes services to scrape metrics from.
  # kubernetes_services = ["http://my-service-dns.my-namespace:9100/metrics"]

  ## Kubernetes config file to create client from.
  # kube_config = "/path/to/kubernetes.config"

  ## Scrape Kubernetes pods for the following prometheus annotations:
  ## - prometheus.io/scrape: Enable scraping for this pod
  ## - prometheus.io/scheme: If the metrics endpoint is secured then you will need to
  ##     set this to 'https' & most likely set the tls config.
  ## - prometheus.io/path: If the metrics path is not /metrics, define it with this annotation.
  ## - prometheus.io/port: If port is not 9102 use this annotation
  # monitor_kubernetes_pods = true
  ## Get the list of pods to scrape with either the scope of
  ## - cluster: the kubernetes watch api (default, no need to specify)
  ## - node: the local cadvisor api; for scalability. Note that the config node_ip or the environment variable NODE_IP must be set to the host IP.
  # pod_scrape_scope = "cluster"
  ## Only for node scrape scope: node IP of the node that telegraf is running on.
  ## Either this config or the environment variable NODE_IP must be set.
  # node_ip = "10.180.1.1"
	## Only for node scrape scope: interval in seconds for how often to get updated pod list for scraping.
	## Default is 60 seconds.
	# pod_scrape_interval = 60
  ## Restricts Kubernetes monitoring to a single namespace
  ##   ex: monitor_kubernetes_pods_namespace = "default"
  # monitor_kubernetes_pods_namespace = ""
  # label selector to target pods which have the label
  # kubernetes_label_selector = "env=dev,app=nginx"
  # field selector to target pods
  # eg. To scrape pods on a specific node
  # kubernetes_field_selector = "spec.nodeName=$HOSTNAME"

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## HTTP Basic Authentication username and password. ('bearer_token' and
  ## 'bearer_token_string' take priority)
  # username = ""
  # password = ""

  ## Specify timeout duration for slower prometheus clients (default is 3s)
  # response_timeout = "3s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (p *Prometheus) SampleConfig() string {
	return sampleConfig
}

func (p *Prometheus) Description() string {
	return "Read metrics from one or many prometheus clients"
}

func (p *Prometheus) Init() error {

	// Config proccessing for node scrape scope for monitor_kubernetes_pods
	p.isNodeScrapeScope = strings.EqualFold(p.PodScrapeScope, "node")
	if p.isNodeScrapeScope {
		// Need node IP to make cAdvisor call for pod list. Check if set in config and valid IP address
		if p.NodeIP == "" || net.ParseIP(p.NodeIP) == nil {
			p.Log.Infof("The config node_ip is empty or invalid. Using NODE_IP env var as default.")

			// Check if set as env var and is valid IP address
			envVarNodeIP := os.Getenv("NODE_IP")
			if envVarNodeIP == "" || net.ParseIP(envVarNodeIP) == nil {
				return errors.New("the node_ip config and the environment variable NODE_IP are not set or invalid; " +
					"cannot get pod list for monitor_kubernetes_pods using node scrape scope")
			}

			p.NodeIP = envVarNodeIP
		}

		// Parse label and field selectors - will be used to filter pods after cAdvisor call
		var err error
		p.podLabelSelector, err = labels.Parse(p.KubernetesLabelSelector)
		if err != nil {
			return fmt.Errorf("error parsing the specified label selector(s): %s", err.Error())
		}
		p.podFieldSelector, err = fields.ParseSelector(p.KubernetesFieldSelector)
		if err != nil {
			return fmt.Errorf("error parsing the specified field selector(s): %s", err.Error())
		}
		isValid, invalidSelector := fieldSelectorIsSupported(p.podFieldSelector)
		if !isValid {
			return fmt.Errorf("the field selector %s is not supported for pods", invalidSelector)
		}

		p.Log.Infof("Using pod scrape scope at node level to get pod list using cAdvisor.")
		p.Log.Infof("Using the label selector: %v and field selector: %v", p.podLabelSelector, p.podFieldSelector)
	}

	return nil
}

var ErrProtocolError = errors.New("prometheus protocol error")

func (p *Prometheus) AddressToURL(u *url.URL, address string) *url.URL {
	host := address
	if u.Port() != "" {
		host = address + ":" + u.Port()
	}
	reconstructedURL := &url.URL{
		Scheme:     u.Scheme,
		Opaque:     u.Opaque,
		User:       u.User,
		Path:       u.Path,
		RawPath:    u.RawPath,
		ForceQuery: u.ForceQuery,
		RawQuery:   u.RawQuery,
		Fragment:   u.Fragment,
		Host:       host,
	}
	return reconstructedURL
}

type URLAndAddress struct {
	OriginalURL *url.URL
	URL         *url.URL
	Address     string
	Tags        map[string]string
}

func (p *Prometheus) GetAllURLs() (map[string]URLAndAddress, error) {
	allURLs := make(map[string]URLAndAddress)
	for _, u := range p.URLs {
		URL, err := url.Parse(u)
		if err != nil {
			p.Log.Errorf("Could not parse %q, skipping it. Error: %s", u, err.Error())
			continue
		}
		allURLs[URL.String()] = URLAndAddress{URL: URL, OriginalURL: URL}
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	// loop through all pods scraped via the prometheus annotation on the pods
	for k, v := range p.kubernetesPods {
		allURLs[k] = v
	}

	for _, service := range p.KubernetesServices {
		URL, err := url.Parse(service)
		if err != nil {
			return nil, err
		}

		resolvedAddresses, err := net.LookupHost(URL.Hostname())
		if err != nil {
			p.Log.Errorf("Could not resolve %q, skipping it. Error: %s", URL.Host, err.Error())
			continue
		}
		for _, resolved := range resolvedAddresses {
			serviceURL := p.AddressToURL(URL, resolved)
			allURLs[serviceURL.String()] = URLAndAddress{
				URL:         serviceURL,
				Address:     resolved,
				OriginalURL: URL,
			}
		}
	}
	return allURLs, nil
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (p *Prometheus) Gather(acc telegraf.Accumulator) error {
	if p.client == nil {
		client, err := p.createHTTPClient()
		if err != nil {
			return err
		}
		p.client = client
		p.headers = map[string]string{
			"User-Agent": internal.ProductToken(),
			"Accept":     acceptHeader,
		}
	}

	var wg sync.WaitGroup

	allURLs, err := p.GetAllURLs()
	if err != nil {
		return err
	}
	for _, URL := range allURLs {
		wg.Add(1)
		go func(serviceURL URLAndAddress) {
			defer wg.Done()
			acc.AddError(p.gatherURL(serviceURL, acc))
		}(URL)
	}

	wg.Wait()

	return nil
}

func (p *Prometheus) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := p.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsCfg,
			DisableKeepAlives: true,
		},
		Timeout: time.Duration(p.ResponseTimeout),
	}

	return client, nil
}

func (p *Prometheus) gatherURL(u URLAndAddress, acc telegraf.Accumulator) error {
	var req *http.Request
	var err error
	var uClient *http.Client
	var metrics []telegraf.Metric
	if u.URL.Scheme == "unix" {
		path := u.URL.Query().Get("path")
		if path == "" {
			path = "/metrics"
		}
		addr := "http://localhost" + path
		req, err = http.NewRequest("GET", addr, nil)
		if err != nil {
			return fmt.Errorf("unable to create new request '%s': %s", addr, err)
		}

		// ignore error because it's been handled before getting here
		tlsCfg, _ := p.ClientConfig.TLSConfig()
		uClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   tlsCfg,
				DisableKeepAlives: true,
				Dial: func(network, addr string) (net.Conn, error) {
					c, err := net.Dial("unix", u.URL.Path)
					return c, err
				},
			},
			Timeout: time.Duration(p.ResponseTimeout),
		}
	} else {
		if u.URL.Path == "" {
			u.URL.Path = "/metrics"
		}
		req, err = http.NewRequest("GET", u.URL.String(), nil)
		if err != nil {
			return fmt.Errorf("unable to create new request '%s': %s", u.URL.String(), err)
		}
	}

	p.addHeaders(req)

	if p.BearerToken != "" {
		token, err := ioutil.ReadFile(p.BearerToken)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+string(token))
	} else if p.BearerTokenString != "" {
		req.Header.Set("Authorization", "Bearer "+p.BearerTokenString)
	} else if p.Username != "" || p.Password != "" {
		req.SetBasicAuth(p.Username, p.Password)
	}

	var resp *http.Response
	if u.URL.Scheme != "unix" {
		resp, err = p.client.Do(req)
	} else {
		resp, err = uClient.Do(req)
	}
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", u.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", u.URL, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	if p.MetricVersion == 2 {
		parser := parser_v2.Parser{Header: resp.Header}
		metrics, err = parser.Parse(body)
	} else {
		metrics, err = Parse(body, resp.Header)
	}

	if err != nil {
		return fmt.Errorf("error reading metrics for %s: %s",
			u.URL, err)
	}

	for _, metric := range metrics {
		tags := metric.Tags()
		// strip user and password from URL
		u.OriginalURL.User = nil
		if p.URLTag != "" {
			tags[p.URLTag] = u.OriginalURL.String()
		}
		if u.Address != "" {
			tags["address"] = u.Address
		}
		for k, v := range u.Tags {
			tags[k] = v
		}

		switch metric.Type() {
		case telegraf.Counter:
			acc.AddCounter(metric.Name(), metric.Fields(), tags, metric.Time())
		case telegraf.Gauge:
			acc.AddGauge(metric.Name(), metric.Fields(), tags, metric.Time())
		case telegraf.Summary:
			acc.AddSummary(metric.Name(), metric.Fields(), tags, metric.Time())
		case telegraf.Histogram:
			acc.AddHistogram(metric.Name(), metric.Fields(), tags, metric.Time())
		default:
			acc.AddFields(metric.Name(), metric.Fields(), tags, metric.Time())
		}
	}

	return nil
}

func (p *Prometheus) addHeaders(req *http.Request) {
	for header, value := range p.headers {
		req.Header.Add(header, value)
	}
}

/* Check if the field selector specified is valid.
 * See ToSelectableFields() for list of fields that are selectable:
 * https://github.com/kubernetes/kubernetes/release-1.20/pkg/registry/core/pod/strategy.go
 */
func fieldSelectorIsSupported(fieldSelector fields.Selector) (bool, string) {
	supportedFieldsToSelect := map[string]bool{
		"spec.nodeName":            true,
		"spec.restartPolicy":       true,
		"spec.schedulerName":       true,
		"spec.serviceAccountName":  true,
		"status.phase":             true,
		"status.podIP":             true,
		"status.nominatedNodeName": true,
	}

	for _, requirement := range fieldSelector.Requirements() {
		if !supportedFieldsToSelect[requirement.Field] {
			return false, requirement.Field
		}
	}

	return true, ""
}

// Start will start the Kubernetes scraping if enabled in the configuration
func (p *Prometheus) Start(_ telegraf.Accumulator) error {
	if p.MonitorPods {
		var ctx context.Context
		ctx, p.cancel = context.WithCancel(context.Background())
		return p.start(ctx)
	}
	return nil
}

func (p *Prometheus) Stop() {
	if p.MonitorPods {
		p.cancel()
	}
	p.wg.Wait()
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{
			ResponseTimeout: config.Duration(time.Second * 3),
			kubernetesPods:  map[string]URLAndAddress{},
			URLTag:          "url",
		}
	})
}
