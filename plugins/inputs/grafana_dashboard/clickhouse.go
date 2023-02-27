package grafana_dashboard

import (
	"encoding/json"
	"fmt"
	"github.com/grafana-tools/sdk"
	"github.com/influxdata/telegraf"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ClickhouseResponseField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ClickhouseResponse struct {
	Meta       []ClickhouseResponseField `json:"meta"`
	Data       []json.RawMessage         `json:"data"`
	Rows       int                       `json:"rows"`
	Statistics struct {
		Elapsed   float64 `json:"elapsed"`
		RowsRead  int     `json:"rows_read"`
		BytesRead int     `json:"bytes_read"`
	} `json:"statistics"`
}

type Clickhouse struct {
	log     telegraf.Logger
	grafana *Grafana
}

type ResRow struct {
	t     time.Time
	tags  map[string]string
	value float64
}

func mapToString(m map[string]string) string {
	var ss []string
	for key, value := range m {
		ss = append(ss, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(ss, ",")
}

func makeMeta(m []ClickhouseResponseField) map[string]string {
	res := make(map[string]string)
	for _, field := range m {
		res[field.Name] = field.Type
	}
	return res
}

func (c *Clickhouse) makeData(rawData []json.RawMessage, meta map[string]string) (res []ResRow) {
	for _, datum := range rawData {
		var i interface{}
		var ts time.Time
		var value float64
		tags := make(map[string]string)

		err := json.Unmarshal(datum, &i)

		if err != nil {
			c.log.Error(err)
			continue
		}

		d := i.(map[string]interface{})
		for key, v := range d {
			if m, found := meta[key]; found {
				if key == "t" && m == "Int64" {
					st, err := strconv.Atoi(v.(string))
					if err != nil {
						c.log.Error(err)
						continue
					}
					ts = time.UnixMilli(int64(st))
				} else if m == "String" {
					tags[key] = v.(string)
				} else {
					float, err := strconv.ParseFloat(v.(string), 64)
					if err != nil {
						c.log.Error(err)
						continue
					}
					value = float
				}
			}
		}
		res = append(res, ResRow{ts, tags, value})
	}
	return res
}

func (c *Clickhouse) GetData(t *sdk.Target, ds *sdk.Datasource, period *GrafanaDashboardPeriod, push GrafanaDatasourcePushFunc) error {
	//TODO implement me
	// https://grafana.exness.io/api/datasources/proxy/83/?query=SELECT%0A%20%20%20%20(intDiv((request_time%20%2F%201000)%2C%2060)%20*%2060)%20*%201000%20as%20t%2C%0A%20%20%20%20count(1)%20as%20overall%0AFROM%20research.mms_response_events%0AWHERE%0A%20%20%20%20(date%20%3E%3D%20toDate(1643544992)%20and%20date%20%3C%3D%20toDate(1643630983))%0A%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20AND%20transaction_type%20IN%20(1%2C%202)%0A%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20AND%20response_result%20IN%20(2%2C11%2C%2012%2C%2013%2C%2014%2C%2015%2C16)%0A%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20and%20date%20%3E%3D%20toDate(1643544992)%20AND%20date%20%3C%3D%20toDate(1643630983)%20AND%20updated%20%3E%3D%20toDateTime(1643544992)%20AND%20updated%20%3C%3D%20toDateTime(1643630983)%0AGROUP%20BY%20t%0AORDER%20BY%20t%20FORMAT%20JSON

	t1, t2 := period.StartEnd()
	start := int(t1.UTC().Unix())
	end := int(t2.UTC().Unix())

	vars := make(map[string]string)
	vars["timeFilter"] = fmt.Sprintf("(date >= toDate(%d) and date <= toDate(%d))", start, end)
	vars["from"] = fmt.Sprintf("%d", start)
	vars["to"] = fmt.Sprintf("%d", end)
	vars["table"] = fmt.Sprintf("%s.%s", t.Database, t.Table)

	params := make(url.Values)
	params.Add("query", c.grafana.setVariables(vars, t.Query)+" FORMAT JSON")

	chURL := fmt.Sprintf("/api/datasources/proxy/%d/", ds.ID)

	raw, code, err := c.grafana.getData(ds, chURL, params, nil)
	if err != nil {
		c.log.Error(err)
		return err
	}
	if code != 200 {
		return fmt.Errorf("clickhouse HTTP error %d: returns %s", code, raw)
	}

	var res ClickhouseResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		c.log.Error(err)
		return err
	}

	when := time.Now()

	meta := makeMeta(res.Meta)
	data := c.makeData(res.Data, meta)
	for _, datum := range data {
		push(when, datum.tags, datum.t, datum.value)
	}

	return nil
}

//func (c *Clickhouse) makeQuery(t *sdk.Target, start, end int) string {
//	res := strings.ReplaceAll(t.Query, "$table", fmt.Sprintf("%s.%s", t.Database, t.Table))
//	res = strings.ReplaceAll(res, "$from", fmt.Sprintf("%d", start))
//	res = strings.ReplaceAll(res, "$to", fmt.Sprintf("%d", end))
//	res = strings.ReplaceAll(res, "$timeFilter", fmt.Sprintf("(date >= toDate(%d) and date <= toDate(%d))", start, end))
//	res += " FORMAT JSON"
//	return res
//}

func NewClickhouse(log telegraf.Logger, grafana *Grafana) *Clickhouse {
	return &Clickhouse{
		log:     log,
		grafana: grafana,
	}
}
