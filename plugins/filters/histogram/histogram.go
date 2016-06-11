package histogram

import (
	"crypto/sha1"
	"github.com/gobwas/glob"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/filters"
	"io"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const field_sep = "."

type metricID struct {
	Name    string
	TagHash [sha1.Size]byte
}

type rollup struct {
	Name         string
	Measurements []string
	Tags         map[string]string
	Functions    []string
	Pass         bool
}
type Histogram struct {
	inch          chan telegraf.Metric
	outch         chan telegraf.Metric
	FlushInterval string
	interval      time.Duration
	Bucketsize    int
	Rollup        []string
	rules         []rollup
	rollupMap     map[metricID]*rollup
	fieldMap      map[metricID]map[string]*Aggregate
	metricTags    map[metricID]map[string]string
	matchGlobs    map[string]glob.Glob
}

func (h *Histogram) Description() string {
	return "Histogram: read metrics from inputs and create histogram for output"
}

func (h *Histogram) SampleConfig() string {
	return `
  ## Histogram Filter
  ## This filter can be used to generate
  ## values generated are approxmation please refer to
  ## Ben-Haim & Yom-Tov's A Streaming Parallel Decision Tree Algorithm 
  ## http://jmlr.org/papers/volume11/ben-haim10a/ben-haim10a.pdf 
  [[filter.histogram]]
  ## bucket size if increase it will increase accuracy of mean, variance and percentile
  ## but it will increase memory usage
    bucketsize = 20
  ## rollup to create 
  ## Name is the name of rollup.
  ## Tag is list of tags to match against metric  ex (Tag key1 value1) (Tag key2 value2) 
  ## Tag value does support glob matching 
  ## Measurements list of mesurment to match against metrics (Measurements cpu* en*)
  ## Functions list to be applied on matched metrics
  ## Pass not to  drop the original metric default false ex (Pass)
  ## supported functions sum, min, max, mean, variance, numbers for percentile ex 0.90
    rollup = [
      "(Name new) (Tag interface en*) (Functions mean 0.90)",
      "(Name cpu_value) (Measurements cpu) (Functions mean sum)",
    ]
`
}

func (h *Histogram) Pipe(in chan telegraf.Metric) chan telegraf.Metric {
	h.inch = in
	h.outch = make(chan telegraf.Metric, 10000)
	return h.outch
}

func (h *Histogram) processLine(list []string) {
	var r rollup
	re := regexp.MustCompile("\\(|\\)")
	for _, item := range list {
		item = re.ReplaceAllString(item, "")
		item = strings.Trim(item, " ")
		match := strings.Split(item, " ")
		if len(match) == 0 {
			match = []string{item}
		}
		switch match[0] {
		case "Name":
			r.Name = match[1]
		case "Tag":
			if len(match) < 3 {
				log.Printf("Each Tag should consist of a name and a value (Tag tag-name tag-value)")
				continue
			}
			if r.Tags == nil {
				r.Tags = make(map[string]string)
			}
			r.Tags[match[1]] = match[2]
		case "Measurements":
			r.Measurements = match[1:]
		case "Functions":
			r.Functions = match[1:]
		case "Pass":
			r.Pass = true
		default:
			log.Printf("Unkown command (%s)", match[0])
		}
	}
	if r.Name != "" {
		h.rules = append(h.rules, r)
	} else {
		log.Printf("Each rollup should have a name (Name [rollup name])")
	}
}

func (h *Histogram) parseRollup() {
	re := regexp.MustCompile("([^)]+)")
	for _, item := range h.Rollup {
		list := re.FindAllString(item, -1)
		if list == nil {
			log.Printf("Please make sure that rollup well formated (%s).", item)
			continue
		}
		h.processLine(list)
	}
}

func (h *Histogram) Start(shutdown chan struct{}) {
	interval, _ := time.ParseDuration(h.FlushInterval)
	ticker := time.NewTicker(interval)
	h.parseRollup()
	for {
		select {
		case m := <-h.inch:
			r, ok := h.matchMetric(m)
			if ok {
				h.AddMetric(m, r)
				if r.Pass {
					h.outch <- m
				}
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

func (h *Histogram) AddMetric(metric telegraf.Metric, r *rollup) {
	mID := metricID{
		Name:    r.Name,
		TagHash: h.hashTags(metric.Tags()),
	}
	h.rollupMap[mID] = r
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
		case int:
			if h.fieldMap[mID][key] == nil {
				h.fieldMap[mID][key] = NewAggregate(h.Bucketsize)
			}
			hist := h.fieldMap[mID][key]
			hist.Add(float64(v))
		case int64:
			if h.fieldMap[mID][key] == nil {
				h.fieldMap[mID][key] = NewAggregate(h.Bucketsize)
			}
			hist := h.fieldMap[mID][key]
			hist.Add(float64(v))
		default:
			log.Printf("When Histogram enabled all the fields should be of type float64 [field name %s]", key)
		}
	}
}

func (h *Histogram) matchMetric(metric telegraf.Metric) (*rollup, bool) {
	glob := func(toMatch string, matchTo []string) bool {
		for _, item := range matchTo {
			g, ok := h.matchGlobs[item]
			if !ok {
				h.matchGlobs[item] = glob.MustCompile(item)
				g = h.matchGlobs[item]
			}
			if g.Match(toMatch) {
				return true
			}
		}
		return false
	}
	for _, y := range h.rules {
		mTags := metric.Tags()
		for k, v := range y.Tags {
			tValue, ok := mTags[k]
			if ok {
				l := []string{v}
				if glob(tValue, l) {
					return &y, true
				}
			}
		}
		if glob(metric.Name(), y.Measurements) {
			return &y, true
		}
	}
	return nil, false
}
func (h *Histogram) OutputMetric() {
	for mID, fields := range h.fieldMap {
		mFields := make(map[string]interface{})
		for key, val := range fields {
			for _, x := range h.rollupMap[mID].Functions {
				p, err := strconv.ParseFloat(x, 64)
				if err == nil {
					mFields[key+field_sep+"p"+x] = val.Quantile(p)
					continue
				}
				switch x {
				case "variance":
					mFields[key+field_sep+"variance"] = val.Variance()
				case "mean":
					mFields[key+field_sep+"mean"] = val.Mean()
				case "count":
					mFields[key+field_sep+"count"] = val.Count()
				case "sum":
					mFields[key+field_sep+"sum"] = val.Sum()
				case "max":
					mFields[key+field_sep+"max"] = val.Max()
				case "min":
					mFields[key+field_sep+"min"] = val.Min()
				}
			}
		}
		metric, _ := telegraf.NewMetric(mID.Name, h.metricTags[mID], mFields, time.Now().UTC())
		h.outch <- metric
		delete(h.rollupMap, mID)
		delete(h.fieldMap, mID)
		delete(h.metricTags, mID)
	}
}

func (h *Histogram) Reset() {
	h.fieldMap = make(map[metricID]map[string]*Aggregate)
	h.metricTags = make(map[metricID]map[string]string)
	h.rollupMap = make(map[metricID]*rollup)
}

func init() {
	filters.Add("histogram", func() telegraf.Filter {
		return &Histogram{
			fieldMap:   make(map[metricID]map[string]*Aggregate),
			metricTags: make(map[metricID]map[string]string),
			rollupMap:  make(map[metricID]*rollup),
			matchGlobs: make(map[string]glob.Glob),
		}
	})
}
