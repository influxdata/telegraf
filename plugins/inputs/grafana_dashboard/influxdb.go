package grafana_dashboard

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/influxdata/telegraf"
)

type InfluxDBResponseResultSeria struct {
	Columns []string          `json:"columns"`
	Name    string            `json:"name"`
	Tags    map[string]string `json:"tags,omitempty"`
	Values  [][]interface{}   `json:"values"`
}

type InfluxDBResponseResult struct {
	Series []InfluxDBResponseResultSeria `json:"series,omitempty"`
}

type InfluxDBResponse struct {
	Results []InfluxDBResponseResult `json:"results,omitempty"`
}

type InfluxDB struct {
	log     telegraf.Logger
	grafana *Grafana
}

func (idb *InfluxDB) GetData(t *sdk.Target, ds *sdk.Datasource, period *GrafanaDashboardPeriod, push GrafanaDatasourcePushFunc) error {

	params := make(url.Values)
	params.Add("db", *ds.Database)

	vars := make(map[string]string)

	t1, t2 := period.StartEnd()
	start := t1.UTC().Format("2006-01-02 15:04:05")
	end := t2.UTC().Format("2006-01-02 15:04:05")

	vars["timeFilter"] = fmt.Sprintf("time >= '%s' and time < '%s'", start, end)

	//vars["timeFilter"] = fmt.Sprintf("time >= now() - %s", period.DurationHuman())
	params.Add("q", idb.grafana.setVariables(vars, t.Query))
	params.Add("epoch", "ms")

	//idb.log.Debugf("Influxdb params => %s", string(params))

	when := time.Now()

	URL := fmt.Sprintf("/api/datasources/proxy/%d/query", ds.ID)
	raw, code, err := idb.grafana.getData(ds, URL, params, nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("Influxdb HTTP error %d: returns %s", code, raw)
	}
	var res InfluxDBResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return err
	}
	if res.Results == nil {
		idb.log.Debug("InfluxDB has no data")
		return nil
	}

	for _, r := range res.Results {

		for _, s := range r.Series {

			tags := make(map[string]string)
			for k, t := range s.Tags {
				tags[k] = t
			}

			for _, v := range s.Values {
				if len(v) == 2 {

					vt, ok := v[0].(float64)
					if !ok {
						continue
					}

					vv, ok := v[1].(float64)
					if !ok {
						continue
					}

					push(when, tags, time.UnixMilli(int64(vt)), vv)
				}
			}
		}
	}
	return nil
}

func NewInfluxDB(log telegraf.Logger, grafana *Grafana) *InfluxDB {

	return &InfluxDB{
		log:     log,
		grafana: grafana,
	}
}
