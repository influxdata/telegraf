//go:generate ../../../tools/readme_config_includer/generator
package azure_monitor

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

const (
	vmInstanceMetadataURL      = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
	resourceIDTemplate         = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s"
	resourceIDScaleSetTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachineScaleSets/%s"
	maxRequestBodySize         = 4000000
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

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

type AzureMonitor struct {
	Timeout              config.Duration `toml:"timeout"`
	NamespacePrefix      string          `toml:"namespace_prefix"`
	StringsAsDimensions  bool            `toml:"strings_as_dimensions"`
	Region               string          `toml:"region"`
	ResourceID           string          `toml:"resource_id"`
	EndpointURL          string          `toml:"endpoint_url"`
	TimestampLimitPast   config.Duration `toml:"timestamp_limit_past"`
	TimestampLimitFuture config.Duration `toml:"timestamp_limit_future"`
	Log                  telegraf.Logger `toml:"-"`

	url      string
	preparer autorest.Preparer
	client   *http.Client

	cache    map[time.Time]map[uint64]*aggregate
	timeFunc func() time.Time

	MetricOutsideWindow selfstat.Stat
}

func (*AzureMonitor) SampleConfig() string {
	return sampleConfig
}

func (a *AzureMonitor) Init() error {
	a.cache = make(map[time.Time]map[uint64]*aggregate, 36)

	authorizer, err := auth.NewAuthorizerFromEnvironmentWithResource("https://monitoring.azure.com/")
	if err != nil {
		return fmt.Errorf("creating authorizer failed: %w", err)
	}
	a.preparer = autorest.CreatePreparer(authorizer.WithAuthorization())

	return nil
}

func (a *AzureMonitor) Connect() error {
	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(a.Timeout),
	}

	// If information is missing try to retrieve it from the Azure VM instance
	if a.Region == "" || a.ResourceID == "" {
		region, resourceID, err := vmInstanceMetadata(a.client)
		if err != nil {
			return fmt.Errorf("getting VM metadata failed: %w", err)
		}

		if a.Region == "" {
			a.Region = region
		}

		if a.ResourceID == "" {
			a.ResourceID = resourceID
		}
	}

	if a.ResourceID == "" {
		return errors.New("no resource ID configured or available via VM instance metadata")
	}

	if a.EndpointURL == "" {
		if a.Region == "" {
			return errors.New("no region configured or available via VM instance metadata")
		}
		a.url = fmt.Sprintf("https://%s.monitoring.azure.com%s/metrics", a.Region, a.ResourceID)
	} else {
		a.url = a.EndpointURL + a.ResourceID + "/metrics"
	}
	a.Log.Debugf("Writing to Azure Monitor URL: %s", a.url)

	a.MetricOutsideWindow = selfstat.Register(
		"azure_monitor",
		"metric_outside_window",
		map[string]string{
			"region":      a.Region,
			"resource_id": a.ResourceID,
		},
	)

	a.Reset()

	return nil
}

// Close shuts down an any active connections
func (a *AzureMonitor) Close() error {
	a.client.CloseIdleConnections()
	a.client = nil
	return nil
}

// Add will append a metric to the output aggregate
func (a *AzureMonitor) Add(m telegraf.Metric) {
	// Azure Monitor only supports aggregates 30 minutes into the past and 4
	// minutes into the future. Future metrics are dropped when pushed.
	tbucket := m.Time().Truncate(time.Minute)
	if tbucket.Before(a.timeFunc().Add(-time.Duration(a.TimestampLimitPast))) {
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
		fv, err := internal.ToFloat64(f.Value)
		if err != nil {
			continue
		}

		// Azure Monitor does not support fields so the field name is appended
		// to the metric name.
		sanitizeKey := invalidNameCharRE.ReplaceAllString(f.Key, "_")
		name := m.Name() + "-" + sanitizeKey
		id := hashIDWithField(m.HashID(), f.Key)

		// Create the time bucket if doesn't exist
		if _, ok := a.cache[tbucket]; !ok {
			a.cache[tbucket] = make(map[uint64]*aggregate)
		}

		// Fetch existing aggregate
		agg, ok := a.cache[tbucket][id]
		if !ok {
			dimensions := make([]dimension, 0, len(m.TagList()))
			for _, tag := range m.TagList() {
				dimensions = append(dimensions, dimension{
					name:  tag.Key,
					value: tag.Value,
				})
			}
			a.cache[tbucket][id] = &aggregate{
				name:       name,
				dimensions: dimensions,
				min:        fv,
				max:        fv,
				sum:        fv,
				count:      1,
				updated:    true,
			}
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

// Push sends metrics to the output metric buffer
func (a *AzureMonitor) Push() []telegraf.Metric {
	var metrics []telegraf.Metric
	for tbucket, aggs := range a.cache {
		// Do not send metrics early
		if tbucket.After(a.timeFunc().Add(time.Duration(a.TimestampLimitFuture))) {
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

			m := metric.New(agg.name,
				tags,
				map[string]interface{}{
					"min":   agg.min,
					"max":   agg.max,
					"sum":   agg.sum,
					"count": agg.count,
				},
				tbucket,
			)

			metrics = append(metrics, m)
		}
	}
	return metrics
}

// Reset clears the cache of aggregate metrics
func (a *AzureMonitor) Reset() {
	for tbucket := range a.cache {
		// Remove aggregates older than 30 minutes
		if tbucket.Before(a.timeFunc().Add(-time.Duration(a.TimestampLimitPast))) {
			delete(a.cache, tbucket)
			continue
		}
		// Metrics updated within the latest 1m have not been pushed and should
		// not be cleared.
		if tbucket.After(a.timeFunc().Add(time.Duration(a.TimestampLimitFuture))) {
			continue
		}
		for id := range a.cache[tbucket] {
			a.cache[tbucket][id].updated = false
		}
	}
}

// Write writes metrics to the remote endpoint
func (a *AzureMonitor) Write(metrics []telegraf.Metric) error {
	now := a.timeFunc()
	tsEarliest := now.Add(-time.Duration(a.TimestampLimitPast))
	tsLatest := now.Add(time.Duration(a.TimestampLimitFuture))

	writeErr := &internal.PartialWriteError{
		MetricsAccept: make([]int, 0, len(metrics)),
	}
	azmetrics := make(map[uint64]*azureMonitorMetric, len(metrics))
	for i, m := range metrics {
		// Skip metrics that our outside of the valid timespan
		if m.Time().Before(tsEarliest) || m.Time().After(tsLatest) {
			a.Log.Tracef("Metric outside acceptable time window: %v", m)
			a.MetricOutsideWindow.Incr(1)
			writeErr.Err = errors.New("metric(s) outside of acceptable time window")
			writeErr.MetricsReject = append(writeErr.MetricsReject, i)
			continue
		}

		amm, err := translate(m, a.NamespacePrefix)
		if err != nil {
			a.Log.Errorf("Could not create azure metric for %q; discarding point", m.Name())
			if writeErr.Err == nil {
				writeErr.Err = errors.New("translating metric(s) failed")
			}
			writeErr.MetricsReject = append(writeErr.MetricsReject, i)
			continue
		}

		id := hashIDWithTagKeysOnly(m)
		if azm, ok := azmetrics[id]; !ok {
			azmetrics[id] = amm
			azmetrics[id].index = i
		} else {
			azmetrics[id].Data.BaseData.Series = append(
				azm.Data.BaseData.Series,
				amm.Data.BaseData.Series...,
			)
			azmetrics[id].index = i
		}
	}

	if len(azmetrics) == 0 {
		if writeErr.Err == nil {
			return nil
		}
		return writeErr
	}

	var buffer bytes.Buffer
	buffer.Grow(maxRequestBodySize)
	batchIndices := make([]int, 0, len(azmetrics))
	for _, m := range azmetrics {
		// Azure Monitor accepts new batches of points in new-line delimited
		// JSON, following RFC 4288 (see https://github.com/ndjson/ndjson-spec).
		buf, err := json.Marshal(m)
		if err != nil {
			writeErr.MetricsReject = append(writeErr.MetricsReject, m.index)
			writeErr.Err = err
			continue
		}
		batchIndices = append(batchIndices, m.index)

		// Azure Monitor's maximum request body size of 4MB. Send batches that
		// exceed this size via separate write requests.
		if buffer.Len()+len(buf)+1 > maxRequestBodySize {
			if retryable, err := a.send(buffer.Bytes()); err != nil {
				writeErr.Err = err
				if !retryable {
					writeErr.MetricsReject = append(writeErr.MetricsAccept, batchIndices...)
				}
				return writeErr
			}
			writeErr.MetricsAccept = append(writeErr.MetricsAccept, batchIndices...)
			batchIndices = make([]int, 0, len(azmetrics))
			buffer.Reset()
		}
		if _, err := buffer.Write(buf); err != nil {
			return fmt.Errorf("writing to buffer failed: %w", err)
		}
		if err := buffer.WriteByte('\n'); err != nil {
			return fmt.Errorf("writing to buffer failed: %w", err)
		}
	}

	if retryable, err := a.send(buffer.Bytes()); err != nil {
		writeErr.Err = err
		if !retryable {
			writeErr.MetricsReject = append(writeErr.MetricsAccept, batchIndices...)
		}
		return writeErr
	}
	writeErr.MetricsAccept = append(writeErr.MetricsAccept, batchIndices...)

	if writeErr.Err == nil {
		return nil
	}

	return writeErr
}

func (a *AzureMonitor) send(body []byte) (bool, error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(body); err != nil {
		return false, fmt.Errorf("zipping content failed: %w", err)
	}
	if err := g.Close(); err != nil {
		return false, fmt.Errorf("closing gzip writer failed: %w", err)
	}

	req, err := http.NewRequest("POST", a.url, &buf)
	if err != nil {
		return false, fmt.Errorf("creating request failed: %w", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-ndjson")

	// Add the authorization header. WithAuthorization will automatically
	// refresh the token if needed.
	req, err = a.preparer.Prepare(req)
	if err != nil {
		return false, fmt.Errorf("unable to fetch authentication credentials: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			a.client.CloseIdleConnections()
			a.client = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
				},
				Timeout: time.Duration(a.Timeout),
			}
		}
		return true, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return false, nil
	}

	retryable := resp.StatusCode != 400
	if respbody, err := io.ReadAll(resp.Body); err == nil {
		return retryable, fmt.Errorf("failed to write batch: [%d] %s: %s", resp.StatusCode, resp.Status, string(respbody))
	}

	return retryable, fmt.Errorf("failed to write batch: [%d] %s", resp.StatusCode, resp.Status)
}

// vmMetadata retrieves metadata about the current Azure VM
func vmInstanceMetadata(c *http.Client) (region, resourceID string, err error) {
	req, err := http.NewRequest("GET", vmInstanceMetadataURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Metadata", "true")

	resp, err := c.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return "", "", fmt.Errorf("unable to fetch instance metadata: [%s] %d",
			vmInstanceMetadataURL, resp.StatusCode)
	}

	var metadata virtualMachineMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return "", "", err
	}

	region = metadata.Compute.Location
	resourceID = metadata.ResourceID()

	return region, resourceID, nil
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
	dimensionNames := make([]string, 0, len(m.TagList()))
	dimensionValues := make([]string, 0, len(m.TagList()))
	for _, tag := range m.TagList() {
		// Azure custom metrics service supports up to 10 dimensions
		if len(dimensionNames) >= 10 {
			continue
		}

		if tag.Key == "" || tag.Value == "" {
			continue
		}

		dimensionNames = append(dimensionNames, tag.Key)
		dimensionValues = append(dimensionValues, tag.Value)
	}

	vmin, err := getFloatField(m, "min")
	if err != nil {
		return nil, err
	}
	vmax, err := getFloatField(m, "max")
	if err != nil {
		return nil, err
	}
	vsum, err := getFloatField(m, "sum")
	if err != nil {
		return nil, err
	}
	vcount, err := getIntField(m, "count")
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
						Min:             vmin,
						Max:             vmax,
						Sum:             vsum,
						Count:           vcount,
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

func init() {
	outputs.Add("azure_monitor", func() telegraf.Output {
		return &AzureMonitor{
			NamespacePrefix:      "Telegraf/",
			TimestampLimitPast:   config.Duration(20 * time.Minute),
			TimestampLimitFuture: config.Duration(-1 * time.Minute),
			Timeout:              config.Duration(5 * time.Second),
			timeFunc:             time.Now,
		}
	})
}
