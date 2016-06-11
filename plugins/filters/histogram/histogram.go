package histogram

import (
	"crypto/sha1"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/filters"
	"io"
	"log"
	"sort"
	"strconv"
	"time"
)

const field_sep = "."

type metricID struct {
	Name    string
	TagHash [sha1.Size]byte
}
type Histogram struct {
	inch          chan telegraf.Metric
	outch         chan telegraf.Metric
	FlushInterval string
	interval      time.Duration
	Bucketsize    int
	Metrics       map[string][]float64
	fieldMap      map[metricID]map[string]*Aggregate
	metricTags    map[metricID]map[string]string
}

func (h *Histogram) Description() string {
	return "Histogram: read metrics from inputs and create histogram for output"
}

func (h *Histogram) SampleConfig() string {
	return `
  ## Histogram Filter
  ## This filter can be used to generate
  ## (mean varince percentile count)
  ##  values generated are approxmation please refer to
  ##  Ben-Haim & Yom-Tov's A Streaming Parallel Decision Tree Algorithm 
  ##  http://jmlr.org/papers/volume11/ben-haim10a/ben-haim10a.pdf 
  ##[[filter.histogram]]
  ## bucket size if increase it will increase accuracy 
  ## but it will increase memory usage
  ##  bucketsize = 20
  ##  [filter.histogram.metrics]
  ## if array is empty only count mean 
  ## and variance will be cacluated. _ALL_METRIC special constanct
  ## can be used instead of metric name this will aggregate all the 
  ## the merrtics for this filer
  ##    metric name = [percentiles] 
  ##    tail = [0.90]
`
}

func (h *Histogram) Pipe(in chan telegraf.Metric) chan telegraf.Metric {
	h.inch = in
	h.outch = make(chan telegraf.Metric, 10000)
	return h.outch
}

func (h *Histogram) Start(shutdown chan struct{}) {
	interval, _ := time.ParseDuration(h.FlushInterval)
	ticker := time.NewTicker(interval)
	for {
		select {
		case m := <-h.inch:
			if h.IsEnabled(m.Name()) {
				h.AddMetric(m)
			} else {
				h.outch <- m
			}
		case <-shutdown:
			log.Printf("Shuting down filters, All metric in the queue will be lost.")
			return
		case <-ticker.C:
			h.OutputMetric()
		}
	}
}

func (h *Histogram) hashTags(m map[string]string) (result [sha1.Size]byte) {
	hash := sha1.New()
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, item := range keys {
		io.WriteString(hash, item+m[item])
	}
	copy(result[:], hash.Sum(nil))
	return result
}

func (h *Histogram) AddMetric(metric telegraf.Metric) {
	mID := metricID{
		Name:    metric.Name(),
		TagHash: h.hashTags(metric.Tags()),
	}
	if h.fieldMap[mID] == nil {
		h.fieldMap[mID] = make(map[string]*Aggregate)
	}
	if h.metricTags[mID] == nil {
		h.metricTags[mID] = make(map[string]string)
	}
	h.metricTags[mID] = metric.Tags()
	for key, val := range metric.Fields() {
		switch v := val.(type) {
		case float64:
			if h.fieldMap[mID][key] == nil {
				h.fieldMap[mID][key] = NewAggregate(h.Bucketsize)
			}
			hist := h.fieldMap[mID][key]
			hist.Add(v)
		default:
			log.Printf("When stats enabled all the fields should be of type float64 [field name %s]", key)
		}
	}
}

func (h *Histogram) IsEnabled(name string) bool {
	_, isAllEnabled := h.Metrics["_ALL_METRIC"]
	_, ok := h.Metrics[name]
	return ok || isAllEnabled
}

func (h *Histogram) OutputMetric() {
	all_percentile := h.Metrics["_ALL_METRIC"]
	for mID, fields := range h.fieldMap {
		mFields := make(map[string]interface{})
		for key, val := range fields {
			percentile, ok := h.Metrics[mID.Name]
			if !ok {
				percentile = all_percentile
			}
			for _, perc := range percentile {
				p := strconv.FormatFloat(perc*100, 'f', 0, 64)
				mFields[key+field_sep+"p"+p] = val.Quantile(perc)
			}
			mFields[key+field_sep+"variance"] = val.Variance()
			mFields[key+field_sep+"mean"] = val.Mean()
			mFields[key+field_sep+"count"] = val.Count()
			mFields[key+field_sep+"sum"] = val.Sum()
			mFields[key+field_sep+"max"] = val.Max()
			mFields[key+field_sep+"min"] = val.Min()
		}
		metric, _ := telegraf.NewMetric(mID.Name, h.metricTags[mID], mFields, time.Now().UTC())
		h.outch <- metric
		delete(h.fieldMap, mID)
		delete(h.metricTags, mID)
	}
}

func (h *Histogram) Reset() {
	h.fieldMap = make(map[metricID]map[string]*Aggregate)
	h.metricTags = make(map[metricID]map[string]string)
}

func init() {
	filters.Add("histogram", func() telegraf.Filter {
		return &Histogram{
			fieldMap:   make(map[metricID]map[string]*Aggregate),
			metricTags: make(map[metricID]map[string]string),
		}
	})
}
