//go:generate ../../../tools/readme_config_includer/generator
package prometheus

import (
	"context"
	_ "embed"
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

	"github.com/prometheus/common/expfmt"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/openmetrics"
	parser "github.com/influxdata/telegraf/plugins/parsers/prometheus"
)

//go:embed sample.conf
var sampleConfig string

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

type MonitorMethod string

const (
	MonitorMethodNone                   MonitorMethod = ""
	MonitorMethodAnnotations            MonitorMethod = "annotations"
	MonitorMethodSettings               MonitorMethod = "settings"
	MonitorMethodSettingsAndAnnotations MonitorMethod = "settings+annotations"
)

type PodID string

type Prometheus struct {
	URLs                 []string          `toml:"urls"`
	BearerToken          string            `toml:"bearer_token"`
	BearerTokenString    config.Secret     `toml:"bearer_token_string"`
	Username             config.Secret     `toml:"username"`
	Password             config.Secret     `toml:"password"`
	HTTPHeaders          map[string]string `toml:"http_headers"`
	ContentLengthLimit   config.Size       `toml:"content_length_limit"`
	ContentTypeOverride  string            `toml:"content_type_override"`
	EnableRequestMetrics bool              `toml:"enable_request_metrics"`
	MetricVersion        int               `toml:"metric_version"`
	URLTag               string            `toml:"url_tag"`
	IgnoreTimestamp      bool              `toml:"ignore_timestamp"`

	// Kubernetes service discovery
	MonitorPods                 bool                `toml:"monitor_kubernetes_pods"`
	PodScrapeScope              string              `toml:"pod_scrape_scope"`
	NodeIP                      string              `toml:"node_ip"`
	PodScrapeInterval           int                 `toml:"pod_scrape_interval"`
	PodNamespace                string              `toml:"monitor_kubernetes_pods_namespace"`
	PodNamespaceLabelName       string              `toml:"pod_namespace_label_name"`
	KubernetesServices          []string            `toml:"kubernetes_services"`
	KubeConfig                  string              `toml:"kube_config"`
	KubernetesLabelSelector     string              `toml:"kubernetes_label_selector"`
	KubernetesFieldSelector     string              `toml:"kubernetes_field_selector"`
	MonitorKubernetesPodsMethod MonitorMethod       `toml:"monitor_kubernetes_pods_method"`
	MonitorKubernetesPodsScheme string              `toml:"monitor_kubernetes_pods_scheme"`
	MonitorKubernetesPodsPath   string              `toml:"monitor_kubernetes_pods_path"`
	MonitorKubernetesPodsPort   int                 `toml:"monitor_kubernetes_pods_port"`
	NamespaceAnnotationPass     map[string][]string `toml:"namespace_annotation_pass"`
	NamespaceAnnotationDrop     map[string][]string `toml:"namespace_annotation_drop"`
	PodAnnotationInclude        []string            `toml:"pod_annotation_include"`
	PodAnnotationExclude        []string            `toml:"pod_annotation_exclude"`
	PodLabelInclude             []string            `toml:"pod_label_include"`
	PodLabelExclude             []string            `toml:"pod_label_exclude"`
	CacheRefreshInterval        int                 `toml:"cache_refresh_interval"`

	// Consul discovery
	ConsulConfig ConsulConfig `toml:"consul"`

	Log telegraf.Logger `toml:"-"`
	httpconfig.HTTPClientConfig

	client      *http.Client
	headers     map[string]string
	contentType string

	nsStore          cache.Store
	nsAnnotationPass []models.TagFilter
	nsAnnotationDrop []models.TagFilter

	// Should we scrape Kubernetes services for prometheus annotations
	lock           sync.Mutex
	kubernetesPods map[PodID]URLAndAddress
	cancel         context.CancelFunc
	wg             sync.WaitGroup

	// Only for monitor_kubernetes_pods=true and pod_scrape_scope="node"
	podLabelSelector           labels.Selector
	podFieldSelector           fields.Selector
	isNodeScrapeScope          bool
	podAnnotationIncludeFilter filter.Filter
	podAnnotationExcludeFilter filter.Filter
	podLabelIncludeFilter      filter.Filter
	podLabelExcludeFilter      filter.Filter

	// List of consul services to scrape
	consulServices map[string]URLAndAddress
}

func (*Prometheus) SampleConfig() string {
	return sampleConfig
}

func (p *Prometheus) Init() error {
	// Setup content-type override if requested
	switch p.ContentTypeOverride {
	case "": // No override
	case "text":
		p.contentType = string(expfmt.NewFormat(expfmt.TypeTextPlain))
	case "protobuf-delimiter":
		p.contentType = string(expfmt.NewFormat(expfmt.TypeProtoDelim))
	case "protobuf-compact":
		p.contentType = string(expfmt.NewFormat(expfmt.TypeProtoCompact))
	case "protobuf-text":
		p.contentType = string(expfmt.NewFormat(expfmt.TypeProtoText))
	case "openmetrics-text":
		f, err := expfmt.NewOpenMetricsFormat(expfmt.OpenMetricsVersion_1_0_0)
		if err != nil {
			return err
		}
		p.contentType = string(f)
	case "openmetrics-protobuf":
		p.contentType = "application/openmetrics-protobuf;version=1.0.0"
	default:
		return fmt.Errorf("invalid 'content_type_override' setting %q", p.ContentTypeOverride)
	}

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
		p.Log.Infof("Using pod scrape scope at node level to get pod list using cAdvisor.")
	}

	if p.MonitorKubernetesPodsMethod == MonitorMethodNone {
		p.MonitorKubernetesPodsMethod = MonitorMethodAnnotations
	}

	// Parse label and field selectors - will be used to filter pods after cAdvisor call
	var err error
	p.podLabelSelector, err = labels.Parse(p.KubernetesLabelSelector)
	if err != nil {
		return fmt.Errorf("error parsing the specified label selector(s): %w", err)
	}
	p.podFieldSelector, err = fields.ParseSelector(p.KubernetesFieldSelector)
	if err != nil {
		return fmt.Errorf("error parsing the specified field selector(s): %w", err)
	}
	isValid, invalidSelector := fieldSelectorIsSupported(p.podFieldSelector)
	if !isValid {
		return fmt.Errorf("the field selector %q is not supported for pods", invalidSelector)
	}

	p.Log.Infof("Using the label selector: %v and field selector: %v", p.podLabelSelector, p.podFieldSelector)

	for k, vs := range p.NamespaceAnnotationPass {
		tagFilter := models.TagFilter{}
		tagFilter.Name = k
		tagFilter.Values = append(tagFilter.Values, vs...)
		if err := tagFilter.Compile(); err != nil {
			return fmt.Errorf("error compiling 'namespace_annotation_pass', %w", err)
		}
		p.nsAnnotationPass = append(p.nsAnnotationPass, tagFilter)
	}

	for k, vs := range p.NamespaceAnnotationDrop {
		tagFilter := models.TagFilter{}
		tagFilter.Name = k
		tagFilter.Values = append(tagFilter.Values, vs...)
		if err := tagFilter.Compile(); err != nil {
			return fmt.Errorf("error compiling 'namespace_annotation_drop', %w", err)
		}
		p.nsAnnotationDrop = append(p.nsAnnotationDrop, tagFilter)
	}

	if err := p.initFilters(); err != nil {
		return err
	}

	if p.MetricVersion == 0 {
		p.MetricVersion = 1
	}

	ctx := context.Background()

	client, err := p.HTTPClientConfig.CreateClient(ctx, p.Log)
	if err != nil {
		return err
	}
	p.client = client
	if p.HTTPClientConfig.ResponseHeaderTimeout != 0 {
		p.Log.Warn(
			"Config option response_timeout was set to non-zero value. This option's behavior was " +
				"changed in Telegraf 1.30.2 and now controls the HTTP client's header timeout and " +
				"not the Prometheus timeout. Users can ignore this warning if that was the intention. " +
				"Otherwise, please use the timeout config option for the Prometheus timeout.",
		)
	}

	p.headers = map[string]string{
		"User-Agent": internal.ProductToken(),
		"Accept":     acceptHeader,
	}

	p.kubernetesPods = map[PodID]URLAndAddress{}

	return nil
}

func (p *Prometheus) initFilters() error {
	if p.PodAnnotationExclude != nil {
		podAnnotationExclude, err := filter.Compile(p.PodAnnotationExclude)
		if err != nil {
			return fmt.Errorf("error compiling 'pod_annotation_exclude': %w", err)
		}
		p.podAnnotationExcludeFilter = podAnnotationExclude
	}
	if p.PodAnnotationInclude != nil {
		podAnnotationInclude, err := filter.Compile(p.PodAnnotationInclude)
		if err != nil {
			return fmt.Errorf("error compiling 'pod_annotation_include': %w", err)
		}
		p.podAnnotationIncludeFilter = podAnnotationInclude
	}
	if p.PodLabelExclude != nil {
		podLabelExclude, err := filter.Compile(p.PodLabelExclude)
		if err != nil {
			return fmt.Errorf("error compiling 'pod_label_exclude': %w", err)
		}
		p.podLabelExcludeFilter = podLabelExclude
	}
	if p.PodLabelInclude != nil {
		podLabelInclude, err := filter.Compile(p.PodLabelInclude)
		if err != nil {
			return fmt.Errorf("error compiling 'pod_label_include': %w", err)
		}
		p.podLabelIncludeFilter = podLabelInclude
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
	Namespace   string
}

func (p *Prometheus) GetAllURLs() (map[string]URLAndAddress, error) {
	allURLs := make(map[string]URLAndAddress, len(p.URLs)+len(p.consulServices)+len(p.kubernetesPods))
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
	for _, v := range p.kubernetesPods {
		if namespaceAnnotationMatch(v.Namespace, p) {
			allURLs[v.URL.String()] = v
		}
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
	var wg sync.WaitGroup

	allURLs, err := p.GetAllURLs()
	if err != nil {
		return err
	}
	for _, URL := range allURLs {
		wg.Add(1)
		go func(serviceURL URLAndAddress) {
			defer wg.Done()
			requestFields, tags, err := p.gatherURL(serviceURL, acc)
			acc.AddError(err)

			// Add metrics
			if p.EnableRequestMetrics {
				acc.AddFields("prometheus_request", requestFields, tags)
			}
		}(URL)
	}

	wg.Wait()

	return nil
}

func (p *Prometheus) gatherURL(u URLAndAddress, acc telegraf.Accumulator) (map[string]interface{}, map[string]string, error) {
	var req *http.Request
	var uClient *http.Client
	requestFields := make(map[string]interface{})
	tags := map[string]string{}
	if p.URLTag != "" {
		tags[p.URLTag] = u.OriginalURL.String()
	}
	if u.Address != "" {
		tags["address"] = u.Address
	}
	for k, v := range u.Tags {
		tags[k] = v
	}

	if u.URL.Scheme == "unix" {
		path := u.URL.Query().Get("path")
		if path == "" {
			path = "/metrics"
		}

		var err error
		addr := "http://localhost" + path
		req, err = http.NewRequest("GET", addr, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create new request %q: %w", addr, err)
		}

		//nolint:errcheck // ignore error because it's been handled before getting here
		tlsCfg, _ := p.HTTPClientConfig.TLSConfig()
		uClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   tlsCfg,
				DisableKeepAlives: true,
				Dial: func(string, string) (net.Conn, error) {
					c, err := net.Dial("unix", u.URL.Path)
					return c, err
				},
			},
		}
	} else {
		if u.URL.Path == "" {
			u.URL.Path = "/metrics"
		}
		var err error
		req, err = http.NewRequest("GET", u.URL.String(), nil)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create new request %q: %w", u.URL.String(), err)
		}
	}

	p.addHeaders(req)

	if p.BearerToken != "" {
		token, err := os.ReadFile(p.BearerToken)
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Authorization", "Bearer "+string(token))
	} else if !p.BearerTokenString.Empty() {
		token, err := p.BearerTokenString.Get()
		if err != nil {
			return nil, nil, fmt.Errorf("getting token secret failed: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token.String())
		token.Destroy()
	} else if !p.Username.Empty() || !p.Password.Empty() {
		username, err := p.Username.Get()
		if err != nil {
			return nil, nil, fmt.Errorf("getting username secret failed: %w", err)
		}
		password, err := p.Password.Get()
		if err != nil {
			return nil, nil, fmt.Errorf("getting password secret failed: %w", err)
		}
		req.SetBasicAuth(username.String(), password.String())
		username.Destroy()
		password.Destroy()
	}

	for key, value := range p.HTTPHeaders {
		if strings.EqualFold(key, "host") {
			req.Host = value
		} else {
			req.Header.Set(key, value)
		}
	}

	var err error
	var resp *http.Response
	var start time.Time
	if u.URL.Scheme != "unix" {
		start = time.Now()
		//nolint:bodyclose // False positive (because of if-else) - body will be closed in `defer`
		resp, err = p.client.Do(req)
	} else {
		start = time.Now()
		//nolint:bodyclose // False positive (because of if-else) - body will be closed in `defer`
		resp, err = uClient.Do(req)
	}
	end := time.Since(start).Seconds()
	if err != nil {
		return requestFields, tags, fmt.Errorf("error making HTTP request to %q: %w", u.URL, err)
	}
	requestFields["response_time"] = end

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return requestFields, tags, fmt.Errorf("%q returned HTTP status %q", u.URL, resp.Status)
	}

	var body []byte
	if p.ContentLengthLimit != 0 {
		limit := int64(p.ContentLengthLimit)

		// To determine whether io.ReadAll() ended due to EOF or reached the specified limit,
		// read up to the specified limit plus one extra byte, and then make a decision based
		// on the length of the result.
		lr := io.LimitReader(resp.Body, limit+1)

		body, err = io.ReadAll(lr)
		if err != nil {
			return requestFields, tags, fmt.Errorf("error reading body: %w", err)
		}
		if int64(len(body)) > limit {
			p.Log.Infof("skipping %s: content length exceeded maximum body size (%d)", u.URL, limit)
			return requestFields, tags, nil
		}
	} else {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return requestFields, tags, fmt.Errorf("error reading body: %w", err)
		}
	}
	requestFields["content_length"] = len(body)

	// Override the response format if the user requested it
	if p.contentType != "" {
		resp.Header.Set("Content-Type", p.contentType)
	}

	// Parse the metrics
	var metricParser telegraf.Parser
	if openmetrics.AcceptsContent(resp.Header) {
		metricParser = &openmetrics.Parser{
			Header:          resp.Header,
			MetricVersion:   p.MetricVersion,
			IgnoreTimestamp: p.IgnoreTimestamp,
			Log:             p.Log,
		}
	} else {
		metricParser = &parser.Parser{
			Header:          resp.Header,
			MetricVersion:   p.MetricVersion,
			IgnoreTimestamp: p.IgnoreTimestamp,
			Log:             p.Log,
		}
	}
	metrics, err := metricParser.Parse(body)
	if err != nil {
		return requestFields, tags, fmt.Errorf("error reading metrics for %q: %w", u.URL, err)
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

	return requestFields, tags, nil
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

	if p.client != nil {
		p.client.CloseIdleConnections()
	}
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{
			kubernetesPods: map[PodID]URLAndAddress{},
			consulServices: map[string]URLAndAddress{},
			URLTag:         "url",
		}
	})
}
