package minmax

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type MinMax struct {
	// caches for metric fields, names, and tags
	fieldCache map[uint64]map[string]minmax
	nameCache  map[uint64]string
	tagCache   map[uint64]map[string]string
}

func NewMinMax() telegraf.Aggregator {
	mm := &MinMax{}
	mm.Reset()
	return mm
}

type minmax struct {
	min float64
	max float64
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

func (m *MinMax) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := m.nameCache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		m.nameCache[id] = in.Name()
		m.tagCache[id] = in.Tags()
		m.fieldCache[id] = make(map[string]minmax)
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				m.fieldCache[id][k] = minmax{
					min: fv,
					max: fv,
				}
			}
		}
	} else {
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				if _, ok := m.fieldCache[id][k]; !ok {
					// hit an uncached field of a cached metric
					m.fieldCache[id][k] = minmax{
						min: fv,
						max: fv,
					}
					continue
				}
				cmpmin := compare(m.fieldCache[id][k].min, fv)
				cmpmax := compare(m.fieldCache[id][k].max, fv)
				if cmpmin == 1 {
					tmp := m.fieldCache[id][k]
					tmp.min = fv
					m.fieldCache[id][k] = tmp
				}
				if cmpmax == -1 {
					tmp := m.fieldCache[id][k]
					tmp.max = fv
					m.fieldCache[id][k] = tmp
				}
			}
		}
	}
}

func (m *MinMax) Push(acc telegraf.Accumulator) {
	for id, _ := range m.nameCache {
		fields := map[string]interface{}{}
		for k, v := range m.fieldCache[id] {
			fields[k+"_min"] = v.min
			fields[k+"_max"] = v.max
		}
		acc.AddFields(m.nameCache[id], fields, m.tagCache[id])
	}
}

func (m *MinMax) Reset() {
	m.fieldCache = make(map[uint64]map[string]minmax)
	m.nameCache = make(map[uint64]string)
	m.tagCache = make(map[uint64]map[string]string)
}

func compare(a, b float64) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func init() {
	aggregators.Add("minmax", func() telegraf.Aggregator {
		return NewMinMax()
	})
}
