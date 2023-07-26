package prometheus_http

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
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
)

// PrometheusHttpMetric struct
type PrometheusHttpMetric struct {
	Name      string `toml:"name"`
	Query     string `toml:"query"`
	Transform string `toml:"transform"`
	template  *toolsRender.TextTemplate
	Duration  config.Duration   `toml:"duration"`
	From      string            `toml:"from"`
	Step      string            `toml:"step"`
	Params    string            `toml:"params"`
	Timeout   config.Duration   `toml:"timeout"`
	Interval  config.Duration   `toml:"interval"`
	Tags      map[string]string `toml:"tags"`
	UniqueBy  []string          `toml:"unique_by"`
	templates map[string]*toolsRender.TextTemplate
	uniques   map[uint64]bool
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
	Metrics       []*PrometheusHttpMetric `toml:"metric"`
	Duration      config.Duration         `toml:"duration"`
	From          string                  `toml:"from"`
	Timeout       config.Duration         `toml:"timeout"`
	Version       string                  `toml:"version"`
	Step          string                  `toml:"step"`
	Params        string                  `toml:"params"`
	Prefix        string                  `toml:"prefix"`
	SkipEmptyTags bool                    `toml:"skip_empty_tags"`
	Files         []*PrometheusHttpFile   `toml:"file"`

	Log   telegraf.Logger `toml:"-"`
	acc   telegraf.Accumulator
	files map[string]interface{}
}

type PrometheusHttpPushFunc = func(when time.Time, tags map[string]string, stamp time.Time, value float64)

type PrometheusHttpDatasource interface {
	GetData(query string, period *PrometheusHttpPeriod, push PrometheusHttpPushFunc) error
}

var description = "Collect data from Prometheus http api"

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

/*func (p *PrometheusHttp) mergeMaps(maps ...map[string]interface{}) map[string]interface{} {

	r := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			r[k] = v
		}
	}
	return r
}*/

func (p *PrometheusHttp) setExtraMetricTag(t *toolsRender.TextTemplate, valueTags, metricTags map[string]string) string {

	tgs := make(map[string]interface{})
	for k, v := range valueTags {
		tgs[k] = v
	}

	m := tgs
	m["values"] = valueTags
	m["tags"] = metricTags
	m["files"] = p.files

	b, err := t.RenderObject(&m)
	if err != nil {
		p.Log.Errorf("%s failed to execute template: %v", p.Name, err)
		return err.Error()
	}
	r := strings.TrimSpace(string(b))
	// simplify <no value> => empty string
	return strings.ReplaceAll(r, "<no value>", "")
}

func (p *PrometheusHttp) getExtraMetricTags(tags map[string]string, m *PrometheusHttpMetric) map[string]string {

	if m.templates == nil {
		return tags
	}
	tgs := make(map[string]string)
	for v, t := range m.Tags {

		tpl := m.templates[v]
		if tpl != nil {
			tgs[v] = p.setExtraMetricTag(m.templates[v], tags, m.Tags)
			if !p.SkipEmptyTags && tgs[v] == "" {
				tgs[v] = t
			}
		} else {
			tgs[v] = t
		}
	}
	return tgs
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

func (p *PrometheusHttp) setMetrics(w *sync.WaitGroup, pm *PrometheusHttpMetric, ds PrometheusHttpDatasource) {

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

		fields := make(map[string]interface{})
		fields[pm.Name] = v

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

		tags = p.getExtraMetricTags(tags, pm)

		if math.IsNaN(value) || math.IsInf(value, 0) {
			bs, _ := json.Marshal(tags)
			p.Log.Debugf("%s skipped NaN/Inf value for: %v[%v]", p.Name, pm.Name, string(bs))
			return
		}
		p.acc.AddFields(p.Prefix, fields, tags, stamp)
	}

	if ds == nil {
		switch p.Version {
		case "v1":
			ds = NewPrometheusHttpV1(p.Name, p.Log, context.Background(), p.URL, int(timeout), step, params)
		}
	}

	if ds != nil {
		period := p.getMetricPeriod(pm)
		err := ds.GetData(pm.Query, period, push)
		if err != nil {
			p.Log.Error(err)
		}
	}
}

func (p *PrometheusHttp) gatherMetrics(ds PrometheusHttpDatasource) error {

	var wg sync.WaitGroup

	for _, m := range p.Metrics {

		if m.Name == "" {
			err := fmt.Errorf("%s no metric name found", p.Name)
			p.Log.Error(err)
			return err
		}

		wg.Add(1)
		go p.setMetrics(&wg, m, ds)
	}
	wg.Wait()
	return nil
}

func (p *PrometheusHttp) fRenderMetricTag(template string, obj interface{}) interface{} {

	t, err := toolsRender.NewTextTemplate(toolsRender.TemplateOptions{
		Content: template,
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

func (p *PrometheusHttp) getDefaultTemplate(name, value string) *toolsRender.TextTemplate {

	if value == "" {
		return nil
	}

	funcs := make(map[string]any)
	funcs["renderMetricTag"] = p.fRenderMetricTag

	tpl, err := toolsRender.NewTextTemplate(toolsRender.TemplateOptions{
		Name:    fmt.Sprintf("%s_template", name),
		Content: value,
		Funcs:   funcs,
	}, p)

	if err != nil {
		p.Log.Error(err)
		return nil
	}
	return tpl
}

func (p *PrometheusHttp) ifTemplate(s string) bool {

	if strings.TrimSpace(s) == "" {
		return false
	}
	// find {{ }} to pass templates
	l := len("{{")
	idx := strings.Index(s, "{{")
	if idx == -1 {
		return false
	}
	s1 := s[idx+l+1:]
	return strings.Contains(s1, "}}")
}

func (p *PrometheusHttp) setDefaultMetric(m *PrometheusHttpMetric) {

	if m.Name == "" {
		return
	}
	if m.Transform != "" {
		m.template = p.getDefaultTemplate(m.Name, m.Transform)
	}
	if len(m.Tags) > 0 {
		m.templates = make(map[string]*toolsRender.TextTemplate)
	}
	for k, v := range m.Tags {
		if p.ifTemplate(v) {
			m.templates[k] = p.getDefaultTemplate(fmt.Sprintf("%s_%s", m.Name, k), v)
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

func (p *PrometheusHttp) readFiles() map[string]interface{} {

	r := make(map[string]interface{})
	for _, v := range p.Files {

		if _, err := os.Stat(v.Path); err == nil {

			p.Log.Debugf("read file: %s", v.Path)

			bytes, err := ioutil.ReadFile(v.Path)
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
			r[v.Name] = obj
		}
	}
	return r
}

// Gather is called by telegraf when the plugin is executed on its interval.
func (p *PrometheusHttp) Gather(acc telegraf.Accumulator) error {

	p.acc = acc

	var ds PrometheusHttpDatasource = nil
	// Gather data
	err := p.gatherMetrics(ds)
	if err != nil {
		return err
	}

	return nil
}

func (p *PrometheusHttp) Init() error {

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
		p.Prefix = "prometheus_http"
	}

	if len(p.Metrics) == 0 {
		err := fmt.Errorf("%s no metrics found", p.Name)
		p.Log.Error(err)
		return err
	}

	p.Log.Debugf("%s metrics: %d", p.Name, len(p.Metrics))

	for _, m := range p.Metrics {
		p.setDefaultMetric(m)
	}

	if len(p.Files) > 0 {
		p.files = p.readFiles()
	}

	return nil
}

func init() {
	inputs.Add("prometheus_http", func() telegraf.Input {
		return &PrometheusHttp{}
	})
}
