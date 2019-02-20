package azure_monitor

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/selfstat"
)

// AzureMonitor allows publishing of metrics to the Azure Monitor custom metrics
// service
type AzureMonitor struct {
	Timeout             internal.Duration
	NamespacePrefix     string `toml:"namespace_prefix"`
	StringsAsDimensions bool   `toml:"strings_as_dimensions"`
	Region              string
	ResourceID          string `toml:"resource_id"`
	EndpointUrl         string `toml:"endpoint_url"`

	url    string
	auth   autorest.Authorizer
	client *http.Client

	cache    map[time.Time]map[uint64]*aggregate
	timeFunc func() time.Time

	MetricOutsideWindow selfstat.Stat
}

type dimension struct {
	name  string
	value string
}

type aggregate struct {
	name       string
	min        float64
	max        float64
	sum        float64
	count      int64
	dimensions []dimension
	updated    bool
}

const (
	defaultRequestTimeout  = time.Second * 5
	defaultNamespacePrefix = "Telegraf/"
	defaultAuthResource    = "https://monitoring.azure.com/"

	vmInstanceMetadataURL = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
	resourceIDTemplate    = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s"
	urlTemplate           = "https://%s.monitoring.azure.com%s/metrics"
	urlOverrideTemplate   = "%s%s/metrics"
	maxRequestBodySize    = 4000000
)

var sampleConfig = `
  ## Timeout for HTTP writes.
  # timeout = "20s"

  ## Set the namespace prefix, defaults to "Telegraf/<input-name>".
  # namespace_prefix = "Telegraf/"

  ## Azure Monitor doesn't have a string value type, so convert string
  ## fields to dimensions (a.k.a. tags) if enabled. Azure Monitor allows
  ## a maximum of 10 dimensions so Telegraf will only send the first 10
  ## alphanumeric dimensions.
  # strings_as_dimensions = false

  ## Both region and resource_id must be set or be available via the
  ## Instance Metadata service on Azure Virtual Machines.
  #
  ## Azure Region to publish metrics against.
  ##   ex: region = "southcentralus"
  # region = ""
  #
  ## The Azure Resource ID against which metric will be logged, e.g.
  ##   ex: resource_id = "/subscriptions/<subscription_id>/resourceGroups/<resource_group>/providers/Microsoft.Compute/virtualMachines/<vm_name>"
  # resource_id = ""

  ## Optionally, if in Azure US Government, China or other sovereign
  ## cloud environment, set appropriate REST endpoint for receiving
  ## metrics. (Note: region may be unused in this context)
  # endpoint_url = "https://monitoring.core.usgovcloudapi.net"
`

// Description provides a description of the plugin
func (a *AzureMonitor) Description() string {
	return "Send aggregate metrics to Azure Monitor"
}

// SampleConfig provides a sample configuration for the plugin
func (a *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Connect initializes the plugin and validates connectivity
func (a *AzureMonitor) Connect() error {
	a.cache = make(map[time.Time]map[uint64]*aggregate, 36)

	if a.Timeout.Duration == 0 {
		a.Timeout.Duration = defaultRequestTimeout
	}

	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: a.Timeout.Duration,
	}

	if a.NamespacePrefix == "" {
		a.NamespacePrefix = defaultNamespacePrefix
	}

	var err error
	var region string
	var resourceID string
	var endpointUrl string

	if a.Region == "" || a.ResourceID == "" {
		// Pull region and resource identifier
		region, resourceID, err = vmInstanceMetadata(a.client)
		if err != nil {
			return err
		}
	}
	if a.Region != "" {
		region = a.Region
	}
	if a.ResourceID != "" {
		resourceID = a.ResourceID
	}
	if a.EndpointUrl != "" {
		endpointUrl = a.EndpointUrl
	}

	if resourceID == "" {
		return fmt.Errorf("no resource ID configured or available via VM instance metadata")
	} else if region == "" {
		return fmt.Errorf("no region configured or available via VM instance metadata")
	}

	if endpointUrl == "" {
		a.url = fmt.Sprintf(urlTemplate, region, resourceID)
	} else {
		a.url = fmt.Sprintf(urlOverrideTemplate, endpointUrl, resourceID)
	}

	log.Printf("D! Writing to Azure Monitor URL: %s", a.url)

	a.auth, err = auth.NewAuthorizerFromEnvironmentWithResource(defaultAuthResource)
	if err != nil {
		return nil
	}

	a.Reset()

	tags := map[string]string{
		"region":      region,
		"resource_id": resourceID,
	}
	a.MetricOutsideWindow = selfstat.Register("azure_monitor", "metric_outside_window", tags)

	return nil
}

// vmMetadata retrieves metadata about the current Azure VM
func vmInstanceMetadata(c *http.Client) (string, string, error) {
	req, err := http.NewRequest("GET", vmInstanceMetadataURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Metadata", "true")

	resp, err := c.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return "", "", fmt.Errorf("unable to fetch instance metadata: [%v] %s", resp.StatusCode, body)
	}

	// VirtualMachineMetadata contains information about a VM from the metadata service
	type VirtualMachineMetadata struct {
		Compute struct {
			Location          string `json:"location"`
			Name              string `json:"name"`
			ResourceGroupName string `json:"resourceGroupName"`
			SubscriptionID    string `json:"subscriptionId"`
		} `json:"compute"`
	}

	var metadata VirtualMachineMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return "", "", err
	}

	region := metadata.Compute.Location
	resourceID := fmt.Sprintf(
		resourceIDTemplate,
		metadata.Compute.SubscriptionID,
		metadata.Compute.ResourceGroupName,
		metadata.Compute.Name,
	)

	return region, resourceID, nil
}

// Close shuts down an any active connections
func (a *AzureMonitor) Close() error {
	a.client = nil
	return nil
}

type azureMonitorMetric struct {
	Time time.Time         `json:"time"`
	Data *azureMonitorData `json:"data"`
}

type azureMonitorData struct {
	BaseData *azureMonitorBaseData `json:"baseData"`
}

type azureMonitorBaseData struct {
	Metric         string                `json:"metric"`
	Namespace      string                `json:"namespace"`
	DimensionNames []string              `json:"dimNames"`
	Series         []*azureMonitorSeries `json:"series"`
}

type azureMonitorSeries struct {
	DimensionValues []string `json:"dimValues"`
	Min             float64  `json:"min"`
	Max             float64  `json:"max"`
	Sum             float64  `json:"sum"`
	Count           int64    `json:"count"`
}

// Write writes metrics to the remote endpoint
func (a *AzureMonitor) Write(metrics []telegraf.Metric) error {
	azmetrics := make(map[uint64]*azureMonitorMetric, len(metrics))
	for _, m := range metrics {
		id := hashIDWithTagKeysOnly(m)
		if azm, ok := azmetrics[id]; !ok {
			amm, err := translate(m, a.NamespacePrefix)
			if err != nil {
				log.Printf("E! [outputs.azure_monitor]: could not create azure metric for %q; discarding point", m.Name())
				continue
			}
			azmetrics[id] = amm
		} else {
			amm, err := translate(m, a.NamespacePrefix)
			if err != nil {
				log.Printf("E! [outputs.azure_monitor]: could not create azure metric for %q; discarding point", m.Name())
				continue
			}

			azmetrics[id].Data.BaseData.Series = append(
				azm.Data.BaseData.Series,
				amm.Data.BaseData.Series...,
			)
		}
	}

	if len(azmetrics) == 0 {
		return nil
	}

	var body []byte
	for _, m := range azmetrics {
		// Azure Monitor accepts new batches of points in new-line delimited
		// JSON, following RFC 4288 (see https://github.com/ndjson/ndjson-spec).
		jsonBytes, err := json.Marshal(&m)
		if err != nil {
			return err
		}
		// Azure Monitor's maximum request body size of 4MB. Send batches that
		// exceed this size via separate write requests.
		if (len(body) + len(jsonBytes) + 1) > maxRequestBodySize {
			err := a.send(body)
			if err != nil {
				return err
			}
			body = nil
		}
		body = append(body, jsonBytes...)
		body = append(body, '\n')
	}

	return a.send(body)
}

func (a *AzureMonitor) send(body []byte) error {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(body); err != nil {
		return err
	}
	if err := g.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", a.url, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-ndjson")

	// Add the authorization header. WithAuthorization will automatically
	// refresh the token if needed.
	req, err = autorest.CreatePreparer(a.auth.WithAuthorization()).Prepare(req)
	if err != nil {
		return fmt.Errorf("unable to fetch authentication credentials: %v", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to write batch: [%v] %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func hashIDWithTagKeysOnly(m telegraf.Metric) uint64 {
	h := fnv.New64a()
	h.Write([]byte(m.Name()))
	h.Write([]byte("\n"))
	for _, tag := range m.TagList() {
		if tag.Key == "" || tag.Value == "" {
			continue
		}

		h.Write([]byte(tag.Key))
		h.Write([]byte("\n"))
	}
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, uint64(m.Time().UnixNano()))
	h.Write(b[:n])
	h.Write([]byte("\n"))
	return h.Sum64()
}

func translate(m telegraf.Metric, prefix string) (*azureMonitorMetric, error) {
	var dimensionNames []string
	var dimensionValues []string
	for _, tag := range m.TagList() {
		// Azure custom metrics service supports up to 10 dimensions
		if len(dimensionNames) > 10 {
			continue
		}

		if tag.Key == "" || tag.Value == "" {
			continue
		}

		dimensionNames = append(dimensionNames, tag.Key)
		dimensionValues = append(dimensionValues, tag.Value)
	}

	min, err := getFloatField(m, "min")
	if err != nil {
		return nil, err
	}
	max, err := getFloatField(m, "max")
	if err != nil {
		return nil, err
	}
	sum, err := getFloatField(m, "sum")
	if err != nil {
		return nil, err
	}
	count, err := getIntField(m, "count")
	if err != nil {
		return nil, err
	}

	mn, ns := "Missing", "Missing"
	names := strings.SplitN(m.Name(), "-", 2)
	if len(names) > 1 {
		mn = names[1]
	}
	if len(names) > 0 {
		ns = names[0]
	}
	ns = prefix + ns

	return &azureMonitorMetric{
		Time: m.Time(),
		Data: &azureMonitorData{
			BaseData: &azureMonitorBaseData{
				Metric:         mn,
				Namespace:      ns,
				DimensionNames: dimensionNames,
				Series: []*azureMonitorSeries{
					{
						DimensionValues: dimensionValues,
						Min:             min,
						Max:             max,
						Sum:             sum,
						Count:           count,
					},
				},
			},
		},
	}, nil
}

func getFloatField(m telegraf.Metric, key string) (float64, error) {
	fv, ok := m.GetField(key)
	if !ok {
		return 0, fmt.Errorf("missing field: %s", key)
	}

	if value, ok := fv.(float64); ok {
		return value, nil
	}
	return 0, fmt.Errorf("unexpected type: %s: %T", key, fv)
}

func getIntField(m telegraf.Metric, key string) (int64, error) {
	fv, ok := m.GetField(key)
	if !ok {
		return 0, fmt.Errorf("missing field: %s", key)
	}

	if value, ok := fv.(int64); ok {
		return value, nil
	}
	return 0, fmt.Errorf("unexpected type: %s: %T", key, fv)
}

// Add will append a metric to the output aggregate
func (a *AzureMonitor) Add(m telegraf.Metric) {
	// Azure Monitor only supports aggregates 30 minutes into the past and 4
	// minutes into the future. Future metrics are dropped when pushed.
	t := m.Time()
	tbucket := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	if tbucket.Before(a.timeFunc().Add(-time.Minute * 30)) {
		a.MetricOutsideWindow.Incr(1)
		return
	}

	// Azure Monitor doesn't have a string value type, so convert string fields
	// to dimensions (a.k.a. tags) if enabled.
	if a.StringsAsDimensions {
		for _, f := range m.FieldList() {
			if v, ok := f.Value.(string); ok {
				m.AddTag(f.Key, v)
			}
		}
	}

	for _, f := range m.FieldList() {
		fv, ok := convert(f.Value)
		if !ok {
			continue
		}

		// Azure Monitor does not support fields so the field name is appended
		// to the metric name.
		name := m.Name() + "-" + sanitize(f.Key)
		id := hashIDWithField(m.HashID(), f.Key)

		_, ok = a.cache[tbucket]
		if !ok {
			// Time bucket does not exist and needs to be created.
			a.cache[tbucket] = make(map[uint64]*aggregate)
		}

		// Fetch existing aggregate
		var agg *aggregate
		agg, ok = a.cache[tbucket][id]
		if !ok {
			agg := &aggregate{
				name:  name,
				min:   fv,
				max:   fv,
				sum:   fv,
				count: 1,
			}
			for _, tag := range m.TagList() {
				dim := dimension{
					name:  tag.Key,
					value: tag.Value,
				}
				agg.dimensions = append(agg.dimensions, dim)
			}
			agg.updated = true
			a.cache[tbucket][id] = agg
			continue
		}

		if fv < agg.min {
			agg.min = fv
		}
		if fv > agg.max {
			agg.max = fv
		}
		agg.sum += fv
		agg.count++
		agg.updated = true
	}
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float64:
		return v, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitize(value string) string {
	return invalidNameCharRE.ReplaceAllString(value, "_")
}

func hashIDWithField(id uint64, fk string) uint64 {
	h := fnv.New64a()
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, id)
	h.Write(b[:n])
	h.Write([]byte("\n"))
	h.Write([]byte(fk))
	h.Write([]byte("\n"))
	return h.Sum64()
}

// Push sends metrics to the output metric buffer
func (a *AzureMonitor) Push() []telegraf.Metric {
	var metrics []telegraf.Metric
	for tbucket, aggs := range a.cache {
		// Do not send metrics early
		if tbucket.After(a.timeFunc().Add(-time.Minute)) {
			continue
		}
		for _, agg := range aggs {
			// Only send aggregates that have had an update since the last push.
			if !agg.updated {
				continue
			}

			tags := make(map[string]string, len(agg.dimensions))
			for _, tag := range agg.dimensions {
				tags[tag.name] = tag.value
			}

			m, err := metric.New(agg.name,
				tags,
				map[string]interface{}{
					"min":   agg.min,
					"max":   agg.max,
					"sum":   agg.sum,
					"count": agg.count,
				},
				tbucket,
			)

			if err != nil {
				log.Printf("E! [outputs.azure_monitor]: could not create metric for aggregation %q; discarding point", agg.name)
			}

			metrics = append(metrics, m)
		}
	}
	return metrics
}

// Reset clears the cache of aggregate metrics
func (a *AzureMonitor) Reset() {
	for tbucket := range a.cache {
		// Remove aggregates older than 30 minutes
		if tbucket.Before(a.timeFunc().Add(-time.Minute * 30)) {
			delete(a.cache, tbucket)
			continue
		}
		// Metrics updated within the latest 1m have not been pushed and should
		// not be cleared.
		if tbucket.After(a.timeFunc().Add(-time.Minute)) {
			continue
		}
		for id := range a.cache[tbucket] {
			a.cache[tbucket][id].updated = false
		}
	}
}

func init() {
	outputs.Add("azure_monitor", func() telegraf.Output {
		return &AzureMonitor{
			timeFunc: time.Now,
		}
	})
}
