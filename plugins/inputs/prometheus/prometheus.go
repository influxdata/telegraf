package prometheus

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	parserV2 "github.com/influxdata/telegraf/plugins/parsers/prometheus"
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

	// Consul SD configuration
	ConsulConfig ConsulConfig `toml:"consul"`

	// Bearer Token authorization file path
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`

	// Basic authentication credentials
	Username string `toml:"username"`
	Password string `toml:"password"`

	ResponseTimeout config.Duration `toml:"response_timeout"`

	MetricVersion int `toml:"metric_version"`

	URLTag string `toml:"url_tag"`

	IgnoreTimestamp bool `toml:"ignore_timestamp"`

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

	// Only for monitor_kubernetes_pods=true
	CacheRefreshInterval int `toml:"cache_refresh_interval"`

	// List of consul services to scrape
	consulServices map[string]URLAndAddress
}

func (p *Prometheus) Init() error {
	// Config processing for node scrape scope for monitor_kubernetes_pods
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
		address, err := url.Parse(u)
		if err != nil {
			p.Log.Errorf("Could not parse %q, skipping it. Error: %s", u, err.Error())
			continue
		}
		allURLs[address.String()] = URLAndAddress{URL: address, OriginalURL: address}
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	// add all services collected from consul
	for k, v := range p.consulServices {
		allURLs[k] = v
	}
	// loop through all pods scraped via the prometheus annotation on the pods
	for k, v := range p.kubernetesPods {
		allURLs[k] = v
	}

	for _, service := range p.KubernetesServices {
		address, err := url.Parse(service)
		if err != nil {
			return nil, err
		}

		resolvedAddresses, err := net.LookupHost(address.Hostname())
		if err != nil {
			p.Log.Errorf("Could not resolve %q, skipping it. Error: %s", address.Host, err.Error())
			continue
		}
		for _, resolved := range resolvedAddresses {
			serviceURL := p.AddressToURL(address, resolved)
			allURLs[serviceURL.String()] = URLAndAddress{
				URL:         serviceURL,
				Address:     resolved,
				OriginalURL: address,
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
		token, err := os.ReadFile(p.BearerToken)
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
		//nolint:bodyclose // False positive (because of if-else) - body will be closed in `defer`
		resp, err = p.client.Do(req)
	} else {
		//nolint:bodyclose // False positive (because of if-else) - body will be closed in `defer`
		resp, err = uClient.Do(req)
	}
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", u.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", u.URL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	if p.MetricVersion == 2 {
		parser := parserV2.Parser{
			Header:          resp.Header,
			IgnoreTimestamp: p.IgnoreTimestamp,
		}
		metrics, err = parser.Parse(body)
	} else {
		metrics, err = Parse(body, resp.Header, p.IgnoreTimestamp)
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

// Start will start the Kubernetes and/or Consul scraping if enabled in the configuration
func (p *Prometheus) Start(_ telegraf.Accumulator) error {
	var ctx context.Context
	p.wg = sync.WaitGroup{}
	ctx, p.cancel = context.WithCancel(context.Background())

	if p.ConsulConfig.Enabled && len(p.ConsulConfig.Queries) > 0 {
		if err := p.startConsul(ctx); err != nil {
			return err
		}
	}
	if p.MonitorPods {
		if err := p.startK8s(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Prometheus) Stop() {
	p.cancel()
	p.wg.Wait()
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{
			ResponseTimeout: config.Duration(time.Second * 3),
			kubernetesPods:  map[string]URLAndAddress{},
			consulServices:  map[string]URLAndAddress{},
			URLTag:          "url",
		}
	})
}
