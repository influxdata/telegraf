package prometheus_http

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/araddon/dateparse"
	"gopkg.in/yaml.v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"

	toolsRender "github.com/devopsext/tools/render"
	utils "github.com/devopsext/utils"
)

type PrometheusHttpTextTemplate struct {
	template *toolsRender.TextTemplate
	input    *PrometheusHttp
	name     string
	hash     uint64
}

// PrometheusHttpMetric struct
type PrometheusHttpMetric struct {
	Name        string `toml:"name"`
	Query       string `toml:"query"`
	Transform   string `toml:"transform"`
	template    *toolsRender.TextTemplate
	Duration    config.Duration   `toml:"duration"`
	From        string            `toml:"from"`
	Step        string            `toml:"step"`
	Params      string            `toml:"params"`
	Timeout     config.Duration   `toml:"timeout"`
	Interval    config.Duration   `toml:"interval"`
	Tags        map[string]string `toml:"tags"`
	UniqueBy    []string          `toml:"unique_by"`
	templates   map[string]*toolsRender.TextTemplate
	dependecies map[string][]string
	only        map[string]string
	uniques     map[uint64]bool
}

// PrometheusHttpFile
type PrometheusHttpFile struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
	Type string `toml:"type"`
}

// PrometheusHttpPeriod struct
type PrometheusHttpPeriod struct {
	duration config.Duration
	from     string
}

// PrometheusHttp struct
type PrometheusHttp struct {
	Name          string                  `toml:"name"`
	URL           string                  `toml:"url"`
	User          string                  `toml:"user"`
	Password      string                  `toml:"password"`
	Metrics       []*PrometheusHttpMetric `toml:"metric"`
	Duration      config.Duration         `toml:"duration"`
	Interval      config.Duration         `toml:"interval"`
	From          string                  `toml:"from"`
	Timeout       config.Duration         `toml:"timeout"`
	Version       string                  `toml:"version"`
	Step          string                  `toml:"step"`
	Params        string                  `toml:"params"`
	Prefix        string                  `toml:"prefix"`
	SkipEmptyTags bool                    `toml:"skip_empty_tags"`
	//Availability  config.Duration         `toml:"availability,omitempty"`
	Files []*PrometheusHttpFile `toml:"file"`

	Log telegraf.Logger `toml:"-"`
	acc telegraf.Accumulator

	requests *RateCounter
	errors   *RateCounter
	cache    map[uint64]map[string]interface{}
}

type PrometheusHttpPushFunc = func(when time.Time, tags map[string]string, stamp time.Time, value float64)

type PrometheusHttpDatasource interface {
	GetData(query string, period *PrometheusHttpPeriod, push PrometheusHttpPushFunc) error
}

var description = "Collect data from Prometheus http api"
var globalFiles = sync.Map{}

const pluginName = "prometheus_http"

// Description will return a short string to explain what the plugin does.
func (*PrometheusHttp) Description() string {
	return description
}

var sampleConfig = `
#
`

func (p *PrometheusHttpPeriod) Duration() config.Duration {
	return p.duration
}

func (p *PrometheusHttpPeriod) DurationHuman() string {
	return time.Duration(p.duration).String()
}

func (p *PrometheusHttpPeriod) From() time.Time {
	t, err := dateparse.ParseAny(p.from)
	if err == nil {
		return t
	}
	return time.Now()
}

func (p *PrometheusHttpPeriod) StartEnd() (time.Time, time.Time) {

	start := p.From()
	end := p.From().Add(time.Duration(p.Duration()))

	if start.UnixNano() > end.UnixNano() {
		t := end
		end = start
		start = t
	}

	return start, end
}

// SampleConfig will return a complete configuration example with details about each field.
func (*PrometheusHttp) SampleConfig() string {
	return sampleConfig
}

func (p *PrometheusHttp) Info(obj interface{}, args ...interface{}) {
	s, ok := obj.(string)
	if !ok {
		return
	}
	p.Log.Infof(s, args)
}

func (p *PrometheusHttp) Warn(obj interface{}, args ...interface{}) {
	s, ok := obj.(string)
	if !ok {
		return
	}
	p.Log.Warnf(s, args)
}

func (p *PrometheusHttp) Debug(obj interface{}, args ...interface{}) {
	s, ok := obj.(string)
	if !ok {
		return
	}
	p.Log.Debugf(s, args)
}

func (p *PrometheusHttp) Error(obj interface{}, args ...interface{}) {
	s, ok := obj.(string)
	if !ok {
		return
	}
	p.Log.Errorf(s, args)
}

func (p *PrometheusHttp) getMetricPeriod(m *PrometheusHttpMetric) *PrometheusHttpPeriod {

	duration := p.Duration
	if m != nil && m.Duration > 0 {
		duration = m.Duration
	}

	from := p.From
	if m != nil && m.From != "" {
		from = m.From
	}

	return &PrometheusHttpPeriod{
		duration: duration,
		from:     from,
	}
}

func (p *PrometheusHttp) getTemplateValue(t *toolsRender.TextTemplate, value float64) (float64, error) {

	if t == nil {
		return value, nil
	}

	b, err := t.RenderObject(value)
	if err != nil {
		return value, err
	}
	v := strings.TrimSpace(string(b))

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return value, err
	}
	return f, nil
}

func (p *PrometheusHttp) fRenderMetricTag(template string, obj interface{}) interface{} {

	t, err := toolsRender.NewTextTemplate(toolsRender.TemplateOptions{
		Content:     template,
		FilterFuncs: true,
	}, p)
	if err != nil {
		p.Log.Error(err)
		return err
	}

	b, err := t.RenderObject(obj)
	if err != nil {
		p.Log.Error(err)
		return err
	}
	return string(b)
}

func (p *PrometheusHttp) getAllTags(values, metricTags, metricVars map[string]string) map[string]interface{} {

	tgs := make(map[string]interface{})
	for k, v := range values {
		tgs[k] = v
	}

	m := tgs
	m["values"] = values
	m["tags"] = metricTags
	m["vars"] = metricVars
	files := make(map[string]interface{})
	globalFiles.Range(func(key, value interface{}) bool {
		files[fmt.Sprint(key)] = value
		return true
	})
	m["files"] = files
	return m
}

func (p *PrometheusHttp) setExtraMetricTag(gid uint64, t *toolsRender.TextTemplate, values, metricTags, metricVars map[string]string) (string, error) {

	m := p.getAllTags(values, metricTags, metricVars)
	b, err := t.RenderObject(&m)
	if err != nil {
		p.Log.Errorf("[%d] %s failed to execute template: %v", gid, p.Name, err)
		return "", err
	}
	r := strings.TrimSpace(string(b))
	// simplify <no value> => empty string
	return strings.ReplaceAll(r, "<no value>", ""), nil
}

func (p *PrometheusHttp) getOnly(gid uint64, only string, values, metricTags, metricVars map[string]string) string {

	e := "error"
	m := p.getAllTags(values, metricTags, metricVars)

	arr := strings.FieldsFunc(only, func(c rune) bool {
		return c == '.'
	})

	l := len(arr)
	switch l {
	case 0:
		p.Log.Errorf("[%d] %s no dots for: %s", gid, p.Name, only)
		return e
	case 1:
		v, ok := m["values"].(map[string]string)
		if ok {
			return v[arr[0]]
		}
	case 2:
		v, ok := m[arr[0]].(map[string]string)
		if ok {
			return v[arr[1]]
		}
	default:
		p.Log.Errorf("[%d] %s dots are more than two: %s", gid, p.Name, only)
		return e
	}
	return e
}

func (p *PrometheusHttp) getKeys(arr map[string][]string) []string {
	var keys []string
	for k := range arr {
		keys = append(keys, k)
	}
	return keys
}

func (p *PrometheusHttp) countDependecies(tags []string, deps []string) int {

	cnt := 0
	for _, k := range deps {
		if !utils.Contains(tags, k) {
			cnt++
		}
	}
	return cnt
}

func (p *PrometheusHttp) sortMetricTags(m *PrometheusHttpMetric) []string {

	var tags []string
	mm := make(map[string][]string)
	for k := range m.Tags {
		if m.dependecies[k] == nil {
			if !utils.Contains(tags, k) {
				tags = append(tags, k)
			}
		} else {
			if len(m.dependecies[k]) > 0 {
				mm[k] = m.dependecies[k]
			}
		}
	}
	keys := p.getKeys(mm)
	// make it ordered
	sort.SliceStable(keys, func(i, j int) bool {

		l1 := p.countDependecies(tags, mm[keys[i]])
		l2 := p.countDependecies(tags, mm[keys[j]])

		return l1 < l2
	})

	for _, k := range keys {
		if utils.Contains(tags, k) {
			continue
		}
		tags = append(tags, k)
	}

	return tags
}

func (p *PrometheusHttp) getExtraMetricTags(gid uint64, values map[string]string, m *PrometheusHttpMetric) map[string]string {

	if m.templates == nil {
		return values
	}

	vars := make(map[string]string)
	mTags := p.sortMetricTags(m)
	for _, k := range mTags {

		tpl := m.templates[k]
		if tpl != nil {
			vk, err := p.setExtraMetricTag(gid, tpl, values, m.Tags, vars)
			if err != nil {
				vars[k] = "error"
				continue
			}
			vars[k] = vk
			if !p.SkipEmptyTags && vars[k] == "" {
				vars[k] = m.Tags[k]
			}
		} else {
			only := m.only[k]
			if only != "" {
				vars[k] = p.getOnly(gid, only, values, m.Tags, vars)
			} else {
				vars[k] = m.Tags[k]
			}
		}
	}
	return vars
}

func byteHash64(b []byte) uint64 {
	h := fnv.New64()
	h.Write(b)
	return h.Sum64()
}

func byteSha512(b []byte) []byte {
	hasher := sha512.New()
	hasher.Write(b)
	return hasher.Sum(nil)
}

func (p *PrometheusHttp) uniqueHash(pm *PrometheusHttpMetric, tgs map[string]string, stamp time.Time) uint64 {

	if len(pm.UniqueBy) == 0 {
		return 0
	}

	if len(tgs) == 0 {
		return 0
	}

	s := ""
	for _, t := range pm.UniqueBy {
		v1 := ""
		flag := false
		for k, v := range tgs {
			if k == t {
				v1 = v
				flag = true
				break
			}
		}
		if flag {
			s1 := fmt.Sprintf("%s=%s", t, v1)
			if s == "" {
				s = s1
			} else {
				s = fmt.Sprintf("%s,%s", s, s1)
			}
		}
	}

	if s == "" {
		return 0
	}

	hash := fmt.Sprintf("%d:%s", stamp.UnixNano(), s)
	return byteHash64(byteSha512([]byte(hash)))
}

func (p *PrometheusHttp) addFields(name string, value interface{}) map[string]interface{} {

	m := make(map[string]interface{})
	m[name] = value
	return m
}

func (p *PrometheusHttp) setMetrics(w *sync.WaitGroup, pm *PrometheusHttpMetric,
	ds PrometheusHttpDatasource, callback func(err error)) {

	gid := utils.GoRoutineID()
	p.Log.Debugf("[%d] %s start gathering %s...", gid, p.Name, pm.Name)

	timeout := pm.Timeout
	if timeout == 0 {
		timeout = p.Timeout
	}

	step := pm.Step
	if step == "" {
		step = p.Step
	}

	params := pm.Params
	if params == "" {
		params = p.Params
	}

	defer w.Done()
	var push = func(when time.Time, tgs map[string]string, stamp time.Time, value float64) {

		hash := p.uniqueHash(pm, tgs, stamp)
		if hash > 0 {
			if pm.uniques[hash] {
				return
			} else {
				pm.uniques[hash] = true
			}
		}

		v, err := p.getTemplateValue(pm.template, value)
		if err != nil {
			p.Log.Error(err)
			return
		}

		millis := when.UTC().UnixMilli()
		tags := make(map[string]string)
		tags["timestamp"] = strconv.Itoa(int(millis))
		tags["duration_ms"] = strconv.Itoa(int(time.Now().UTC().UnixMilli()) - int(millis))

		for k, t := range tgs {
			if p.SkipEmptyTags && t == "" {
				continue
			}
			tags[k] = t
		}

		tags = p.getExtraMetricTags(gid, tags, pm)

		if math.IsNaN(value) || math.IsInf(value, 0) {
			bs, _ := json.Marshal(tags)
			p.Log.Debugf("[%d] %s skipped NaN/Inf value for: %v[%v]", gid, p.Name, pm.Name, string(bs))
			return
		}
		p.acc.AddFields(p.Prefix, p.addFields(pm.Name, v), tags, stamp)
	}

	if ds == nil {
		switch p.Version {
		case "v1":
			ds = NewPrometheusHttpV1(p.Name, p.Log, context.Background(), p.URL, p.User, p.Password, int(timeout), step, params)
		}
	}

	if ds != nil {
		period := p.getMetricPeriod(pm)
		err := ds.GetData(pm.Query, period, push)
		if err != nil {
			p.Log.Error(err)
		}
		callback(err)
	}
}

func (p *PrometheusHttp) gatherMetrics(gid uint64, ds PrometheusHttpDatasource) error {

	var wg sync.WaitGroup

	tags := make(map[string]string)
	tags[fmt.Sprintf("%s_name", pluginName)] = p.Name
	tags[fmt.Sprintf("%s_url", pluginName)] = p.URL

	for _, m := range p.Metrics {

		if m.Name == "" {
			err := fmt.Errorf("[%d] %s no metric name found", gid, p.Name)
			p.Log.Error(err)
			return err
		}

		wg.Add(1)

		go p.setMetrics(&wg, m, ds, func(err error) {

			p.requests.Incr(1)
			if err != nil {
				p.errors.Incr(1)
			}
		})
	}
	wg.Wait()

	// availability = (requests - errors) / requests * 100
	// availability = (100 - 0) / 100 * 100 = 100%
	// availability = (100 - 1) / 100 * 100 = 99%
	// availability = (100 - 10) / 100 * 100 = 90%
	// availability = (100 - 100) / 100 * 100 = 0%

	fields := p.addFields("requests", p.requests.counter.Value())
	fields["errors"] = p.errors.counter.Value()

	r1 := float64(p.requests.counter.Value())
	r2 := float64(p.errors.counter.Value())
	if r1 > 0 {
		fields["availability"] = (r1 - r2) / r1 * 100
	}
	p.acc.AddFields(pluginName, fields, tags, time.Now())

	return nil
}

func (ptt *PrometheusHttpTextTemplate) FCacheRegexMatchObjectNameByField(obj map[string]interface{}, field, value string) string {

	if obj == nil || utils.IsEmpty(field) || utils.IsEmpty(value) {
		return ""
	}
	name := fmt.Sprintf("%s.%s", ptt.name, field)
	if ptt.input.cache != nil {
		v := ptt.input.cache[ptt.hash][name]
		if v != nil {
			return fmt.Sprintf("%v", v)
		}
	}
	r := ""
	v := ptt.template.RegexMatchFindKeys(obj, field, value)
	//v := ptt.template.RegexMatchObjectNameByField(obj, field, value)
	if v != nil && ptt.input.cache != nil {
		m := ptt.input.cache[ptt.hash]
		if m == nil {
			m = make(map[string]interface{})
		}
		m[name] = v
		ptt.input.cache[ptt.hash] = m
		r = fmt.Sprintf("%v", v)
	}
	return r
}

func (p *PrometheusHttp) getDefaultTemplate(m *PrometheusHttpMetric, name, value string) *toolsRender.TextTemplate {

	if value == "" {
		return nil
	}

	ptt := &PrometheusHttpTextTemplate{}

	funcs := make(map[string]any)
	funcs["renderMetricTag"] = p.fRenderMetricTag
	funcs["regexMatchObjectNameByField"] = ptt.FCacheRegexMatchObjectNameByField

	tpl, err := toolsRender.NewTextTemplate(toolsRender.TemplateOptions{
		Name:        fmt.Sprintf("%s_template", name),
		Content:     value,
		Funcs:       funcs,
		FilterFuncs: true,
	}, p)

	if err != nil {
		p.Log.Error(err)
		return nil
	}
	ptt.template = tpl
	ptt.input = p
	ptt.name = name
	ptt.hash = byteHash64(byteSha512([]byte(m.Query)))
	return tpl
}

func (p *PrometheusHttp) ifTemplate(s string) (bool, string) {

	only := ""
	if strings.TrimSpace(s) == "" {
		return false, only
	}
	// find {{ }} to pass templates
	l := len("{{")
	idx1 := strings.Index(s, "{{")
	if idx1 == -1 {
		return false, only
	}
	s1 := s[idx1+l:]
	idx2 := strings.LastIndex(s1, "}}")
	if idx2 == -1 {
		return false, only
	}
	s2 := strings.TrimSpace(s1[0:idx2])
	arr := strings.Split(s2, " ")
	if len(arr) == 1 {
		if idx1 == 0 && strings.HasPrefix(arr[0], ".") {
			only = arr[0]
		}
	}
	return true, only
}

func (p *PrometheusHttp) findTagsOnVars(ident, name, value string, tags map[string]string, stack []string) []string {

	var r []string
	if len(tags) == 0 {
		return r
	}
	for k, v := range tags {
		pattern := fmt.Sprintf(".vars.%s", k)
		if strings.Contains(value, pattern) && !utils.Contains(r, k) {
			if utils.Contains(stack, k) {
				return append(r, k)
			}
			r = append(r, k)
			d := p.findTagsOnVars(ident, k, v, tags, append(stack, r...))
			if len(d) > 0 {
				for _, k1 := range d {
					if !utils.Contains(r, k1) {
						r = append(r, k1)
					}
				}
			}
		}
	}
	return r
}

func (p *PrometheusHttp) setDefaultMetric(gid uint64, m *PrometheusHttpMetric) {

	if m.Name == "" {
		return
	}
	if m.Transform != "" {
		m.template = p.getDefaultTemplate(m, m.Name, m.Transform)
	}
	if len(m.Tags) > 0 {
		m.templates = make(map[string]*toolsRender.TextTemplate)
	}
	m.dependecies = make(map[string][]string)
	m.only = make(map[string]string)
	for k, v := range m.Tags {

		b, only := p.ifTemplate(v)
		if b {

			n := fmt.Sprintf("%s_%s", m.Name, k)
			d := p.findTagsOnVars(m.Name, k, v, m.Tags, []string{k})
			if utils.Contains(d, k) {
				p.Log.Errorf("[%d] %s metric %s: %s dependency contains in %s", gid, p.Name, m.Name, k, d)
				continue
			}
			if len(d) > 0 {
				p.Log.Debugf("[%d] %s metric %s %s dependencies are %s", gid, p.Name, m.Name, k, d)
			}
			m.dependecies[k] = d
			if only == "" {
				m.templates[k] = p.getDefaultTemplate(m, n, v)
			} else {
				m.only[k] = only
			}
		}
	}
	if m.uniques == nil {
		m.uniques = make(map[uint64]bool)
	}
}

func (p *PrometheusHttp) readJson(bytes []byte) (interface{}, error) {

	var v interface{}
	err := json.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (p *PrometheusHttp) readToml(bytes []byte) (interface{}, error) {

	return nil, fmt.Errorf("toml is not implemented")
}

func (p *PrometheusHttp) readYaml(bytes []byte) (interface{}, error) {

	var v interface{}
	err := yaml.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (p *PrometheusHttp) readFiles(gid uint64, files *sync.Map) {

	for _, v := range p.Files {

		_, ok := files.Load(v.Name)
		if ok {
			p.Log.Debugf("[%d] %s cache file: %s", gid, p.Name, v.Path)
			continue
		}

		if _, err := os.Stat(v.Path); err == nil {

			p.Log.Debugf("[%d] %s read file: %s", gid, p.Name, v.Path)

			bytes, err := os.ReadFile(v.Path)
			if err != nil {
				p.Log.Error(err)
				continue
			}

			tp := strings.Replace(filepath.Ext(v.Path), ".", "", 1)
			if v.Type != "" {
				tp = v.Type
			}

			var obj interface{}
			switch {
			case tp == "json":
				obj, err = p.readJson(bytes)
			case tp == "toml":
				obj, err = p.readToml(bytes)
			case (tp == "yaml") || (tp == "yml"):
				obj, err = p.readYaml(bytes)
			default:
				obj, err = p.readJson(bytes)
			}
			if err != nil {
				p.Log.Error(err)
				continue
			}
			files.Store(v.Name, obj)
		}
	}
}

// Gather is called by telegraf when the plugin is executed on its interval.
func (p *PrometheusHttp) Gather(acc telegraf.Accumulator) error {

	p.acc = acc

	var ds PrometheusHttpDatasource = nil
	gid := utils.GoRoutineID()
	// Gather data
	err := p.gatherMetrics(gid, ds)
	return err
}

func (p *PrometheusHttp) Printf(format string, v ...interface{}) {
	p.Log.Debugf(format, v)
}

func (p *PrometheusHttp) Init() error {

	gid := utils.GoRoutineID()

	if p.Name == "" {
		p.Name = "unknown"
	}
	if p.Timeout == 0 {
		p.Timeout = config.Duration(time.Second) * 5
	}
	if p.Version == "" {
		p.Version = "v1"
	}
	if p.Step == "" {
		p.Step = "60"
	}
	if p.Prefix == "" {
		p.Prefix = pluginName
	}

	if len(p.Metrics) == 0 {
		err := fmt.Errorf("[%d] %s no metrics found", gid, p.Name)
		p.Log.Error(err)
		return err
	}

	p.Log.Debugf("[%d] %s metrics amount: %d", gid, p.Name, len(p.Metrics))

	p.cache = make(map[uint64]map[string]interface{})
	for _, m := range p.Metrics {
		p.setDefaultMetric(gid, m)
	}

	if len(p.Files) > 0 {
		p.readFiles(gid, &globalFiles)
	}

	p.requests = NewRateCounter(time.Duration(p.Interval))
	p.errors = NewRateCounter(time.Duration(p.Interval))

	return nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		return &PrometheusHttp{}
	})
}
