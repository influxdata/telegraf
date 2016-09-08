package minmax

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type MinMax struct {
	Period internal.Duration

	// metrics waiting to be processed
	metrics  chan telegraf.Metric
	shutdown chan struct{}
	wg       sync.WaitGroup

	// caches for metric fields, names, and tags
	fieldCache map[uint64]map[string]minmax
	nameCache  map[uint64]string
	tagCache   map[uint64]map[string]string

	acc telegraf.Accumulator
}

type minmax struct {
	min interface{}
	max interface{}
}

var sampleConfig = `
  ## TODO doc
  period = "30s"
`

func (m *MinMax) SampleConfig() string {
	return sampleConfig
}

func (m *MinMax) Description() string {
	return "Keep the aggregate min/max of each metric passing through."
}

func (m *MinMax) Apply(in telegraf.Metric) {
	m.metrics <- in
}

func (m *MinMax) apply(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := m.nameCache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		m.nameCache[id] = in.Name()
		m.tagCache[id] = in.Tags()
		m.fieldCache[id] = make(map[string]minmax)
		for k, v := range in.Fields() {
			m.fieldCache[id][k] = minmax{
				min: v,
				max: v,
			}
		}
	} else {
		for k, v := range in.Fields() {
			cmpmin := compare(m.fieldCache[id][k].min, v)
			cmpmax := compare(m.fieldCache[id][k].max, v)
			if cmpmin == 1 {
				tmp := m.fieldCache[id][k]
				tmp.min = v
				m.fieldCache[id][k] = tmp
			}
			if cmpmax == -1 {
				tmp := m.fieldCache[id][k]
				tmp.max = v
				m.fieldCache[id][k] = tmp
			}
		}
	}
}

func (m *MinMax) Start(acc telegraf.Accumulator) error {
	m.metrics = make(chan telegraf.Metric, 10)
	m.shutdown = make(chan struct{})
	m.clearCache()
	m.acc = acc
	m.wg.Add(1)
	if m.Period.Duration > 0 {
		go m.periodHandler()
	} else {
		go m.continuousHandler()
	}
	return nil
}

func (m *MinMax) Stop() {
	close(m.shutdown)
	m.wg.Wait()
}

func (m *MinMax) addfields(id uint64) {
	fields := map[string]interface{}{}
	for k, v := range m.fieldCache[id] {
		fields[k+"_min"] = v.min
		fields[k+"_max"] = v.max
	}
	m.acc.AddFields(m.nameCache[id], fields, m.tagCache[id])
}

func (m *MinMax) clearCache() {
	m.fieldCache = make(map[uint64]map[string]minmax)
	m.nameCache = make(map[uint64]string)
	m.tagCache = make(map[uint64]map[string]string)
}

// periodHandler only adds the aggregate metrics on the configured Period.
//   thus if telegraf's collection interval is 10s, and period is 30s, there
//   will only be one aggregate sent every 3 metrics.
func (m *MinMax) periodHandler() {
	// TODO make this sleep less of a hack!
	time.Sleep(time.Millisecond * 200)
	defer m.wg.Done()
	ticker := time.NewTicker(m.Period.Duration)
	defer ticker.Stop()
	for {
		select {
		case in := <-m.metrics:
			m.apply(in)
		case <-m.shutdown:
			if len(m.metrics) > 0 {
				continue
			}
			return
		case <-ticker.C:
			for id, _ := range m.nameCache {
				m.addfields(id)
			}
			m.clearCache()
		}
	}
}

// continuousHandler sends one metric for every metric that passes through it.
func (m *MinMax) continuousHandler() {
	defer m.wg.Done()
	for {
		select {
		case in := <-m.metrics:
			m.apply(in)
			m.addfields(in.HashID())
		case <-m.shutdown:
			if len(m.metrics) > 0 {
				continue
			}
			return
		}
	}
}

func compare(a, b interface{}) int {
	switch at := a.(type) {
	case int64:
		if bt, ok := b.(int64); ok {
			if at < bt {
				return -1
			} else if at > bt {
				return 1
			}
			return 0
		} else {
			return 0
		}
	case float64:
		if bt, ok := b.(float64); ok {
			if at < bt {
				return -1
			} else if at > bt {
				return 1
			}
			return 0
		} else {
			return 0
		}
	default:
		return 0
	}
}

func init() {
	aggregators.Add("minmax", func() telegraf.Aggregator {
		return &MinMax{}
	})
}
