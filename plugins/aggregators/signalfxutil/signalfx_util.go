package signalfxutil

import (
	"log"
	"math"
	"sort"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

// SignalFxUtil -
type SignalFxUtil struct {
	utilizations map[string]Utilization
}

// NewSignalFxUtil -
func NewSignalFxUtil() telegraf.Aggregator {
	ctx := &SignalFxUtil{}
	ctx.utilizations = make(map[string]Utilization)
	ctx.utilizations["cpuUtilization"] = newCPUUtilization()
	ctx.utilizations["memUtilization"] = newMemoryUtilization()
	ctx.utilizations["diskUtilization"] = newDiskUtilization()
	ctx.utilizations["diskTotalUtilization"] = newDiskTotalUtilization()
	ctx.utilizations["networkTotal"] = newNetworkTotal()
	ctx.utilizations["diskTotal"] = newDiskTotal()
	ctx.utilizations["uptime"] = newUpTime()
	ctx.Reset()
	return ctx
}

var sampleConfig = `
  ## SignalFx Utilization Aggregator
  ## Enable this plugin to report utilization metrics
  ## Metrics will report with the plugin name "signalfx-metadata"
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  ## The period must be at least double the collection interval
  ## because this plugin aggregates metrics across two reporting intervals.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## Only pass the following metrics to the utilization plugin
  namepass = ["cpu", "mem", "disk", "diskio", "net", "system"]
`

// SampleConfig -
func (m *SignalFxUtil) SampleConfig() string {
	return sampleConfig
}

// Description -
func (m *SignalFxUtil) Description() string {
	return "Calculate the utilization values for cpu, mem, disk, and network."
}

// Add -
func (m *SignalFxUtil) Add(metric telegraf.Metric) {
	if metric.Name() == "cpu" {
		m.utilizations["cpuUtilization"].addMetric(metric)
	}
	if metric.Name() == "mem" {
		m.utilizations["memUtilization"].addMetric(metric)
	}
	if metric.Name() == "disk" {
		m.utilizations["diskUtilization"].addMetric(metric)
		m.utilizations["diskTotalUtilization"].addMetric(metric)
	}
	if metric.Name() == "net" {
		m.utilizations["networkTotal"].addMetric(metric)
	}
	if metric.Name() == "diskio" {
		m.utilizations["diskTotal"].addMetric(metric)
	}
	if metric.Name() == "system" {
		m.utilizations["uptime"].addMetric(metric)
	}
}

// Push -
func (m *SignalFxUtil) Push(acc telegraf.Accumulator) {
	for _, util := range m.utilizations {
		util.read(acc)
	}
}

// Reset -
func (m *SignalFxUtil) Reset() {
	for _, util := range m.utilizations {
		util.reset()
	}
}

// UpTime -
type UpTime struct {
	metrics []telegraf.Metric
	meta    bool
}

// UpTime.addMetric -
func (u *UpTime) addMetric(metric telegraf.Metric) {
	u.metrics = append(u.metrics, metric)
}

// UpTime.read
func (u *UpTime) read(acc telegraf.Accumulator) {
	max := len(u.metrics)
	for i := 0; i < max; i++ {
		var m = u.metrics[i]
		var fields = make(map[string]interface{})
		var tags = make(map[string]string)
		if m.HasField("uptime") {
			if m.HasTag("host") {
				tags["host"] = m.Tags()["host"]
			}
			fields["host-plugin_uptime"] = m.Fields()["uptime"]
			tags["plugin"] = "signalfx-metadata"
			acc.AddGauge("gauge.sf", fields, tags, m.Time())
		}
	}
}

// Uptime.reset -
func (u *UpTime) reset() {
	u.metrics = make([]telegraf.Metric, 0)
}

func newUpTime() *UpTime {
	var u = new(UpTime)
	u.meta = false
	return u
}

// Utilization -
type Utilization interface {
	addMetric(telegraf.Metric)
	read(telegraf.Accumulator)
	reset()
}

// UtilizationBase -
type UtilizationBase struct {
	metrics     map[string]telegraf.Metric
	metricOrder []string
	lastTime    time.Time
}

// addMetric -
func (u *UtilizationBase) addMetric(metric telegraf.Metric, key string) {
	// Insert the metric
	u.metrics[key] = metric
	u.metricOrder = append(u.metricOrder, key)
}

// Utilization.read -
func (u *UtilizationBase) read(acc telegraf.Accumulator, plugin string, fieldMatch string) {
	for _, t := range u.metricOrder {
		var m = u.metrics[t]
		var fields = make(map[string]interface{})
		if m.Time().After(u.lastTime) {
			for field, value := range m.Fields() {
				if field == fieldMatch {
					fields["utilization"] = value
					u.lastTime = m.Time()
					acc.AddGauge(plugin, fields, m.Tags(), m.Time())
				}
			}
		}
	}
}

func (u *UtilizationBase) reset() {
	u.metrics = make(map[string]telegraf.Metric)
	u.metricOrder = make([]string, 0)
}

// Total -
type Total struct {
	devices        map[string][]telegraf.Metric
	skipped        map[string]bool
	previous       map[string]telegraf.Metric
	instanceTag    string
	inField        string
	outField       string
	pluginInstance string
	pluginName     string
	total          uint64
}

func (t *Total) addMetric(metric telegraf.Metric) {
	if metric.HasField(t.inField) || metric.HasField(t.outField) {

		if _, ok := t.devices[metric.Tags()[t.instanceTag]]; !ok {
			t.devices[metric.Tags()[t.instanceTag]] = make([]telegraf.Metric, 0)
		}

		t.devices[metric.Tags()[t.instanceTag]] = append(t.devices[metric.Tags()[t.instanceTag]], metric)
	}
}

func (t *Total) read(acc telegraf.Accumulator) {
	log.Println("D! Aggregator [signalfx-util] Reading out metrics from Network Total")

	// Removed skipped devices
	t.removeSkipped()

	// Get max value
	max := t.getMax()

	for i := 0; i < max; i++ {
		var time time.Time
		var fields = make(map[string]interface{})
		var tags = make(map[string]string)
		// Iterate for each device pop 1 off the queue
		for key, device := range t.devices {
			if _, ok := t.skipped[key]; !ok {
				var m = device[i]
				if _, ok := t.previous[key]; ok {
					var p = t.previous[key]

					// Take the latest time
					if m.Time().After(time) {
						time = m.Time()
					}

					// Pull out the host tag
					if m.HasTag("host") {
						tags["host"] = m.Tags()["host"]
					}

					// Add to the received and sent total
					var in = t.overflowDiff(m.Fields()[t.inField].(uint64), p.Fields()[t.inField].(uint64))
					var out = t.overflowDiff(m.Fields()[t.outField].(uint64), p.Fields()[t.outField].(uint64))
					t.total = t.overflowAdd(in, t.total)
					t.total = t.overflowAdd(out, t.total)

				}
				t.previous[key] = m
			}
		}

		tags["plugin_instance"] = t.pluginInstance
		tags["plugin"] = "signalfx-metadata"
		fields["total"] = t.total
		acc.AddCounter(t.pluginName, fields, tags, time)
	}
	t.cleanUpDevices(max)
}

func (t *Total) removeSkipped() {
	for key, device := range t.devices {
		if len(device) == 0 {
			// Handle skipped status
			if _, ok := t.skipped[key]; ok {
				// Delete if previously skipped
				delete(t.devices, key)
			} else {
				// If first time zero, skip it
				t.skipped[key] = true
			}
		}
	}
}

func (t *Total) reset() {

}

func (t *Total) getMax() int {
	max := 0
	var lengths = make([]int, 0)
	// Get the lengths of each device
	for _, device := range t.devices {
		// If length is 0 handle it
		if len(device) != 0 {
			lengths = append(lengths, len(device))
		}
	}

	// Sort the lengths
	sort.Ints(lengths)

	// Take the lowest value as max
	if len(lengths) >= 1 {
		max = lengths[0]
	}
	return max
}

func (t *Total) overflowDiff(current uint64, previous uint64) uint64 {
	var response uint64
	if current < previous {
		var partialDelta = math.MaxInt64 - previous
		response = current + partialDelta
	} else {
		response = current - previous
	}
	return response
}

func (t *Total) overflowAdd(current uint64, previous uint64) uint64 {
	var response uint64
	var available = math.MaxUint64 - previous

	if available >= current {
		response = previous + current
	} else {
		response = current - available
	}
	return response
}

func (t *Total) cleanUpDevices(max int) {
	for key, device := range t.devices {
		if _, ok := t.skipped[key]; !ok {
			// Remove the metric once we're done with it
			device = device[max:]
			t.devices[key] = device
		}
	}
}

// DiskTotal -
type DiskTotal struct {
	Total
}

// newDiskTotalUtilization -
func newDiskTotal() *DiskTotal {
	var d = new(DiskTotal)
	d.devices = make(map[string][]telegraf.Metric)
	d.skipped = make(map[string]bool)
	d.previous = make(map[string]telegraf.Metric)
	d.total = uint64(0)
	d.instanceTag = "name"
	d.inField = "reads"
	d.outField = "writes"
	d.pluginInstance = "summation"
	d.pluginName = "disk_ops"
	d.reset()
	return d
}

// NetworkTotal -
type NetworkTotal struct {
	Total
}

// newDiskTotalUtilization -
func newNetworkTotal() *NetworkTotal {
	var n = new(NetworkTotal)
	n.devices = make(map[string][]telegraf.Metric)
	n.skipped = make(map[string]bool)
	n.previous = make(map[string]telegraf.Metric)
	n.total = uint64(0)
	n.instanceTag = "interface"
	n.inField = "bytes_recv"
	n.outField = "bytes_sent"
	n.pluginInstance = "summation"
	n.pluginName = "network"
	n.reset()
	return n
}

// TotalDevice -
type TotalDevice struct {
	skipped bool
	used    []telegraf.Metric
	total   []telegraf.Metric
}

// DiskTotalUtilization -
type DiskTotalUtilization struct {
	UtilizationBase
	devices map[string]*TotalDevice
	host    string
}

// DiskTotalUtilization.read -
func (u *DiskTotalUtilization) read(acc telegraf.Accumulator) {
	log.Println("D! Aggregator [signalfx_util] Reading out metrics from DiskTotalUtilization")

	var total uint64
	var used uint64

	total = 0
	used = 0

	for key, device := range u.devices {

		if len(device.total) > 0 && len(device.used) > 0 {
			total += device.total[len(device.total)-1].Fields()["total"].(uint64)
			used += device.used[len(device.total)-1].Fields()["used"].(uint64)
			device.skipped = false
			// reset the total and used arrays for the given device
			device.total = make([]telegraf.Metric, 0)
			device.used = make([]telegraf.Metric, 0)
		} else if device.skipped {
			// device went 2 cycles without both required metrics
			delete(u.devices, key)
		} else {
			log.Printf("D! Aggregator [signalfx-util] Marking %s as skipped", key)
			device.skipped = true
		}
		u.devices[key] = device
	}
	if total != 0 {
		var fields = make(map[string]interface{})
		var tags = make(map[string]string)
		tags["plugin_instance"] = "utilization"
		fields["summary_utilization"] = 1.0 * float64(used) / float64(total) * 100
		tags["plugin"] = "signalfx-metadata"
		tags["host"] = u.host
		acc.AddGauge("disk", fields, tags, time.Now())
	}
}

// DiskTotalUtilization.addMetric -
func (u *DiskTotalUtilization) addMetric(metric telegraf.Metric) {
	for field := range metric.Fields() {
		var device *TotalDevice
		// Check if device is already in the list
		if _, ok := u.devices[metric.Tags()["device"]]; ok {
			device = u.devices[metric.Tags()["device"]]
		} else {
			device = new(TotalDevice)
			device.skipped = false
		}
		if field == "total" {
			device.total = append(device.total, metric)
		}
		if field == "used" {
			device.used = append(device.used, metric)
		}
		u.devices[metric.Tags()["device"]] = device
	}
	if u.host == "" {
		if metric.HasTag("host") {
			u.host = metric.Tags()["host"]
		}
	}
}

// newDiskTotalUtilization -
func newDiskTotalUtilization() *DiskTotalUtilization {
	var d = new(DiskTotalUtilization)
	d.devices = make(map[string]*TotalDevice)
	d.host = ""
	d.reset()
	return d
}

// DiskUtilization -
type DiskUtilization struct {
	UtilizationBase
	lastTime map[string]time.Time
}

// DiskUtilization.read -
func (u *DiskUtilization) read(acc telegraf.Accumulator) {
	log.Println("D! Aggregator [signalfx-util] Reading out metrics from DiskUtilization")
	var pluginInstance string
	for _, t := range u.metricOrder {
		var m = u.metrics[t]
		var fields = make(map[string]interface{})

		if _, ok := u.lastTime[m.Tags()["device"]]; !ok || m.Time().After(u.lastTime[m.Tags()["device"]]) {
			for field, value := range m.Fields() {
				if field == "used_percent" {
					u.lastTime[m.Tags()["device"]] = m.Time()
					fields["utilization"] = value
					// Add plugin_instance tag for dimensional parity
					if m.HasTag("device") {
						pluginInstance = m.Tags()["device"]
						if pluginInstance == "/" {
							pluginInstance = "root"
						}
						m.AddTag("plugin_instance", pluginInstance)
					}
					m.AddTag("plugin", "signalfx-metadata")
					acc.AddGauge("disk", fields, m.Tags(), m.Time())
				}
			}
		}
	}
}

// DiskUtilization.addMetric -
func (u *DiskUtilization) addMetric(metric telegraf.Metric) {
	var key = metric.Time().String()

	// Check if the metric has a cpu  and include it in the key
	if _, ok := metric.Tags()["device"]; ok {
		key = key + metric.Tags()["device"]
	}

	u.UtilizationBase.addMetric(metric, key)
}

// newDiskUtilization -
func newDiskUtilization() *DiskUtilization {
	var d = new(DiskUtilization)
	d.lastTime = make(map[string]time.Time)
	d.reset()
	return d
}

// MemoryUtilization -
type MemoryUtilization struct {
	UtilizationBase
}

// MemoryUtilization.read -
func (u *MemoryUtilization) read(acc telegraf.Accumulator) {
	log.Println("D! Aggregator [signalfx-util] Reading out metrics from MemoryUtilization")
	u.UtilizationBase.read(acc, "memory", "used_percent")
}

// MemoryUtilization.addMetric -
func (u *MemoryUtilization) addMetric(metric telegraf.Metric) {
	var key = metric.Time().String()
	metric.AddTag("plugin", "signalfx-metadata")
	u.UtilizationBase.addMetric(metric, key)
}

func newMemoryUtilization() *MemoryUtilization {
	var m = new(MemoryUtilization)
	m.reset()
	return m
}

// CPUUtilization -
type CPUUtilization struct {
	UtilizationBase
	lastTime map[string]time.Time
}

// CPUUtilization.addMetric
func (u *CPUUtilization) addMetric(metric telegraf.Metric) {
	var key = metric.Time().String()

	// Check if the metric has a cpu  and include it in the key
	if _, ok := metric.Tags()["cpu"]; ok {
		key = key + metric.Tags()["cpu"]
	}

	u.UtilizationBase.addMetric(metric, key)
}

// CPUUtilization.read -
func (u *CPUUtilization) read(acc telegraf.Accumulator) {
	log.Println("D! Aggregator [signalfx-util] Reading out metrics from CPUUtilization")
	var fieldName string
	for _, t := range u.metricOrder {
		fieldName = "utilization_per_core"
		var m = u.metrics[t]
		var fields = make(map[string]interface{})

		if _, ok := u.lastTime[m.Tags()["cpu"]]; !ok || m.Time().After(u.lastTime[m.Tags()["cpu"]]) {
			if m.Tags()["cpu"] == "cpu-total" {
				fieldName = "utilization"
			}
			for field, value := range m.Fields() {
				if field == "usage_idle" {
					u.lastTime[m.Tags()["cpu"]] = m.Time()
					fields[fieldName] = 100.0 - value.(float64)
					// Add plugin_instance tag for dimensional parity
					m.AddTag("plugin_instance", "utilization")
					m.AddTag("plugin", "signalfx-metadata")
					acc.AddGauge("cpu", fields, m.Tags(), m.Time())
				}
			}
		}
	}
}

// newCPUUtilization -
func newCPUUtilization() *CPUUtilization {
	var c = new(CPUUtilization)
	c.reset()
	c.lastTime = make(map[string]time.Time)
	return c
}

func init() {
	aggregators.Add("signalfx_util", func() telegraf.Aggregator {
		return NewSignalFxUtil()
	})
}
