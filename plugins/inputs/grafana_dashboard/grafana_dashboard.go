package grafana_dashboard

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/araddon/dateparse"
	"github.com/grafana-tools/sdk"
	"github.com/jinzhu/copier"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// GrafanaDashboardMetric struct
type GrafanaDashboardMetric struct {
	Name      string
	Transform string
	template  *template.Template
	Panels    []string
	Duration  config.Duration
	From      string
	Timeout   config.Duration
	Interval  config.Duration
	Tags      map[string]string
	templates map[string]*template.Template
}

// GrafanaDatasourcePushFunc func
type GrafanaDatasourcePushFunc = func(when time.Time, tags map[string]string, stamp time.Time, value float64)

// GrafanaDashboardPeriod struct
type GrafanaDashboardPeriod struct {
	duration config.Duration
	from     string
}

// GrafanaDatasource interface
type GrafanaDatasource interface {
	GetData(t *sdk.Target, ds *sdk.Datasource, period *GrafanaDashboardPeriod, push GrafanaDatasourcePushFunc) error
}

// GrafanaDashboard struct
type GrafanaDashboard struct {
	URL        string
	APIKey     string
	Dashboards []string
	Rows       []string
	Metrics    []*GrafanaDashboardMetric `toml:"metric"`
	Duration   config.Duration
	From       string
	Timeout    config.Duration

	Log telegraf.Logger `toml:"-"`

	acc    telegraf.Accumulator
	client *http.Client
	ctx    context.Context
}

var description = "Collect Grafana dashboard data"

// Description will return a short string to explain what the plugin does.
func (*GrafanaDashboard) Description() string {
	return description
}

var sampleConfig = `
#
`

func (p *GrafanaDashboardPeriod) Duration() config.Duration {
	return p.duration
}

func (p *GrafanaDashboardPeriod) DurationHuman() string {
	return time.Duration(p.duration).String()
}

func (p *GrafanaDashboardPeriod) From() time.Time {
	t, err := dateparse.ParseAny(p.from)
	if err == nil {
		return t
	}
	return time.Now()
}

func (p *GrafanaDashboardPeriod) StartEnd() (time.Time, time.Time) {

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
func (*GrafanaDashboard) SampleConfig() string {
	return sampleConfig
}

func (g *GrafanaDashboard) makeHttpClient(timeout time.Duration) *http.Client {

	var transport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: timeout}).Dial,
		TLSHandshakeTimeout: timeout,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	var client = &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return client
}

func (g *GrafanaDashboard) findDashboard(c *sdk.Client, title string) (*sdk.Board, error) {

	//var tags []string
	//boards, err := c.SearchDashboards(g.ctx, title, false, tags...)

	board, _, err := c.GetDashboardByUID(g.ctx, title)
	if err == nil {
		return &board, nil
	}

	boards, err := c.Search(g.ctx, sdk.SearchType(sdk.SearchTypeDashboard), sdk.SearchQuery(title))
	if err != nil {
		return nil, err
	}

	if len(boards) > 0 {

		var b *sdk.FoundBoard

		for _, v := range boards {
			if v.Title == title {
				b = &v
				break
			}
		}
		if b == nil {
			return nil, errors.New("board not found")
		}

		board, _, err := c.GetDashboardByUID(g.ctx, b.UID)
		if err != nil {
			return nil, err
		}
		return &board, nil
	}
	return nil, errors.New("dashboard not found")
}

func (g *GrafanaDashboard) findDatasource(name string, dss []sdk.Datasource) *sdk.Datasource {

	for _, ds := range dss {
		if (ds.UID == name) || (ds.Name == name) {
			return &ds
		}
	}
	return nil
}

func (g *GrafanaDashboard) findDefaultDatasource(dss []sdk.Datasource) *sdk.Datasource {

	for _, ds := range dss {
		if ds.IsDefault {
			return &ds
		}
	}
	return nil
}

func (g *GrafanaDashboard) findMetrics(name string) []*GrafanaDashboardMetric {

	var r []*GrafanaDashboardMetric
	for _, m := range g.Metrics {
		for _, s := range m.Panels {
			if b, _ := regexp.MatchString(name, s); b {
				r = append(r, m)
			}
		}
	}
	return r
}

func (g *GrafanaDashboard) getMetricPeriod(m *GrafanaDashboardMetric) *GrafanaDashboardPeriod {

	duration := g.Duration
	if m != nil && m.Duration > 0 {
		duration = m.Duration
	}

	from := g.From
	if m != nil && m.From != "" {
		from = m.From
	}

	return &GrafanaDashboardPeriod{

		duration: duration,
		from:     from,
	}
}

func (g *GrafanaDashboard) getTemplateValue(t *template.Template, value float64) (float64, error) {

	if t == nil {
		return value, nil
	}

	var b strings.Builder
	err := t.Execute(&b, value)
	if err != nil {
		return value, err
	}
	v := b.String()
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return value, err
	}
	return f, nil
}

func (g *GrafanaDashboard) setExtraMetricTag(t *template.Template, tag string, tags map[string]string) {

	if t == nil || tag == "" {
		return
	}

	var b strings.Builder
	err := t.Execute(&b, &tags)
	if err != nil {
		g.Log.Errorf("failed to execute template: %v", err)
		return
	}
	tags[tag] = b.String()
}

func (g *GrafanaDashboard) setExtraMetricTags(tags map[string]string, m *GrafanaDashboardMetric) {

	if m.templates == nil {
		return
	}
	for v, t := range m.templates {
		g.setExtraMetricTag(t, v, tags)
	}
}

func (g *GrafanaDashboard) setData(b *sdk.Board, r string, p *sdk.Panel, ds *sdk.Datasource, dss []sdk.Datasource) {

	if p.GraphPanel != nil {

		title := p.CommonPanel.Title
		var metric *GrafanaDashboardMetric
		metrics := g.findMetrics(title)
		if len(metrics) == 0 {
			return
		}
		metric = metrics[0]
		if metric.Name == "" {
			return
		}

		var wg sync.WaitGroup

		for _, t := range p.GraphPanel.Targets {

			if t.Hide {
				continue
			}

			d := ds
			dsUid := sdk.GetDatasourceUID(t.Datasource)
			if dsUid != "" {
				td := g.findDatasource(dsUid, dss)
				if td != nil {
					d = td
				}
			}

			if d == nil {
				continue
			}
			if d.Access != "proxy" {
				continue
			}

			if t.Interval == "" {
				t.Interval = time.Duration(metric.Interval).String()
			}

			wg.Add(1)

			tnew := sdk.Target{}
			dsnew := sdk.Datasource{}

			copier.Copy(&dsnew, &d)
			copier.Copy(&tnew, &t)

			go func(w *sync.WaitGroup, wtt, wr, wdt string, wds *sdk.Datasource, wt *sdk.Target, wm *GrafanaDashboardMetric) {

				defer w.Done()

				var datasource GrafanaDatasource = nil

				var push = func(when time.Time, tgs map[string]string, stamp time.Time, value float64) {

					v, err := g.getTemplateValue(wm.template, value)
					if err != nil {
						g.Log.Error(err)
						return
					}

					fields := make(map[string]interface{})
					fields[wm.Name] = v

					millis := when.UTC().UnixMilli()
					tags := make(map[string]string)
					tags["timestamp"] = strconv.Itoa(int(millis))
					tags["duration_ms"] = strconv.Itoa(int(time.Now().UTC().UnixMilli()) - int(millis))
					tags["title"] = wtt
					if wr != "" {
						tags["row"] = wr
					}
					tags["datasource_type"] = wdt
					tags["datasource_name"] = wds.Name

					for k, t := range tgs {
						tags[k] = t
					}

					g.setExtraMetricTags(tags, wm)

					if math.IsNaN(value) || math.IsInf(value, 0) {
						bs, _ := json.Marshal(tags)
						g.Log.Debugf("Skipped NaN/Inf value for: %v[%v]", wm.Name, string(bs))
						return
					}

					g.acc.AddFields("grafana_dashboard", fields, tags, stamp)
				}

				client := g.makeHttpClient(time.Duration(wm.Timeout))
				grafana := NewGrafana(g.Log, g.URL, g.APIKey, client, context.Background())

				switch wdt {
				case "prometheus":
					datasource = NewPrometheus(g.Log, grafana)
				case "influxdb":
					datasource = NewInfluxDB(g.Log, grafana)
				case "alexanderzobnin-zabbix-datasource":
					datasource = NewAlexanderzobninZabbix(g.Log, grafana)
				case "marcusolsson-json-datasource":
					datasource = NewMarcusolssonJson(g.Log, grafana)
				case "elasticsearch":
					datasource = NewElasticsearch(g.Log, grafana)
				case "vertamedia-clickhouse-datasource":
					datasource = NewClickhouse(g.Log, grafana)
				default:
					g.Log.Debugf("%s is not implemented yet", wdt)
				}

				if datasource != nil {
					period := g.getMetricPeriod(wm)
					err := datasource.GetData(wt, wds, period, push)
					if err != nil {
						g.Log.Error(err)
					}
				}
			}(&wg, title, r, d.Type, &dsnew, &tnew, metric)
		}
		wg.Wait()
	}
}

func (g *GrafanaDashboard) rowExists(p *sdk.Panel) bool {

	if p == nil {
		return false
	}
	if len(g.Rows) == 0 {
		return true
	}

	for _, v := range g.Rows {
		if b, _ := regexp.MatchString(p.CommonPanel.Title, v); b {
			return true
		}
	}
	return false
}

func (g *GrafanaDashboard) processDashboard(c *sdk.Client, b *sdk.Board, dss []sdk.Datasource) {

	var rowPanel *sdk.RowPanel = nil
	rowExists := false
	rowTitle := ""

	for _, p := range b.Panels {

		if p.RowPanel != nil {
			rowPanel = p.RowPanel
			rowExists = g.rowExists(p)
			if rowExists {
				rowTitle = p.CommonPanel.Title
			}
			continue
		}
		if p.GraphPanel == nil {
			continue
		}

		if rowPanel != nil && !rowExists {
			continue
		}

		var ds *sdk.Datasource

		if p.CommonPanel.Datasource != nil {
			dsUid := sdk.GetDatasourceUID(p.CommonPanel.Datasource)
			if dsUid != "-- Mixed --" {
				ds = g.findDatasource(dsUid, dss)
				if ds == nil {
					continue
				}
			}
		} else {
			ds = g.findDefaultDatasource(dss)
		}

		g.setData(b, rowTitle, p, ds, dss)
	}
}

func (g *GrafanaDashboard) GrafanaGather() error {

	client := g.makeHttpClient(time.Duration(g.Timeout))
	c, err := sdk.NewClient(g.URL, g.APIKey, client)
	if err != nil {
		g.Log.Error(err)
		return err
	}
	g.client = client

	ctx := context.Background()
	dss, err := c.GetAllDatasources(ctx)
	if err != nil {
		g.Log.Error(err)
		return err
	}
	g.ctx = ctx

	for _, d := range g.Dashboards {
		b, err := g.findDashboard(c, d)
		if err != nil {
			g.Log.Errorf("%s: %s", d, err.Error())
			continue
		}
		if b == nil {
			continue
		}
		g.processDashboard(c, b, dss)
	}
	return nil
}

func (g *GrafanaDashboard) getDefaultTemplate(name, value string) *template.Template {

	if value == "" {
		return nil
	}

	t, err := template.New(fmt.Sprintf("%s_template", name)).Funcs(sprig.TxtFuncMap()).Parse(value)
	if err != nil {
		g.Log.Error(err)
		return nil
	}
	return t
}

func (g *GrafanaDashboard) setDefaultMetric(m *GrafanaDashboardMetric) {

	if m.Name == "" {
		return
	}
	if m.Transform != "" {
		m.template = g.getDefaultTemplate(m.Name, m.Transform)
	}
	if len(m.Tags) > 0 {
		m.templates = make(map[string]*template.Template)
	}
	for k, v := range m.Tags {
		m.templates[k] = g.getDefaultTemplate(fmt.Sprintf("%s_%s", m.Name, k), v)
	}
}

// Gather is called by telegraf when the plugin is executed on its interval.
func (g *GrafanaDashboard) Gather(acc telegraf.Accumulator) error {

	// Set default values
	if g.Duration == 0 {
		g.Duration = config.Duration(time.Second) * 5
	}
	if g.Timeout == 0 {
		g.Timeout = config.Duration(time.Second) * 5
	}
	g.acc = acc

	if len(g.Metrics) == 0 {
		return errors.New("no metrics found")
	}

	for _, m := range g.Metrics {
		g.setDefaultMetric(m)
	}

	// Gather data
	err := g.GrafanaGather()
	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("grafana_dashboard", func() telegraf.Input {
		return &GrafanaDashboard{}
	})
}
