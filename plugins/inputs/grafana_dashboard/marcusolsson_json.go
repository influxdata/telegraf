package grafana_dashboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/blues/jsonata-go"
	"github.com/grafana-tools/sdk"
	"github.com/influxdata/telegraf"
)

type MarcusolssonJson struct {
	log     telegraf.Logger
	grafana *Grafana
}

func (mj *MarcusolssonJson) GetData(t *sdk.Target, ds *sdk.Datasource, period *GrafanaDashboardPeriod, push GrafanaDatasourcePushFunc) error {

	if t.URLPath == nil {
		return nil
	}
	query, ok := t.URLPath.(string)
	if !ok {
		return nil
	}

	method := "GET"
	if t.Method != nil {
		m, ok := t.Method.(string)
		if ok {
			method = m
		}
	}

	var body []byte
	if t.Body != nil {
		b, ok := t.Body.(string)
		if ok {
			body = []byte(b)
		}
	}

	params := make(url.Values)
	vars := make(map[string]string)

	t1, t2 := period.StartEnd()
	start := int(t1.UTC().UnixMilli())
	end := int(t2.UTC().UnixMilli())

	vars["__from"] = strconv.Itoa(start)
	vars["__to"] = strconv.Itoa(end)

	if t.Params != nil {

		ps, ok := t.Params.([]interface{})
		if ok {
			for _, v := range ps {

				mm, ok := v.([]interface{})
				if !ok {
					continue
				}
				if len(mm) != 2 {
					continue
				}
				pn, ok := mm[0].(string)
				if !ok {
					continue
				}
				pv, ok := mm[1].(string)
				if !ok {
					continue
				}
				params.Add(pn, mj.grafana.setVariables(vars, pv))
			}
		}
	}

	//idb.log.Debugf("MarcusolssonJson request params is %s", string(params))
	mj.log.Debugf("MarcusolssonJson request body => %s", body)

	when := time.Now()

	URL := fmt.Sprintf("/api/datasources/proxy/%d%s", ds.ID, query)
	raw, code, err := mj.grafana.httpDoRequest(method, URL, params, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("MarcusolssonJson HTTP error %d: returns %s", code, raw)
	}
	var res map[string]interface{}
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return err
	}
	if t.Fields == nil {
		return fmt.Errorf("MarcusolssonJson has no fields")
	}

	var times []float64
	var series = make(map[string][]float64)

	fs, ok := t.Fields.([]interface{})
	if !ok {
		return nil
	}

	for _, v := range fs {
		arr, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		jsonPath, ok := arr["jsonPath"].(string)
		if !ok {
			continue
		}
		language, ok := arr["language"].(string)
		if !ok {
			continue
		}
		name, ok := arr["name"].(string)
		if !ok {
			continue
		}
		ftype, ok := arr["type"].(string)
		if !ok {
			continue
		}

		if language == "jsonata" {
			expr := jsonata.MustCompile(jsonPath)
			data, err := expr.Eval(res)
			if err != nil {
				return err
			}
			d, ok := data.([]interface{})
			if !ok {
				continue
			}

			if ftype == "time" {
				for _, v := range d {
					ts, ok := v.(float64)
					if ok {
						times = append(times, ts)
						continue
					}
					s, ok := v.(string)
					if !ok {
						continue
					}
					t, err := time.Parse(time.RFC3339, s)
					if err == nil {
						ts = float64(t.UTC().UnixMilli())
						times = append(times, ts)
					}
				}
			} else {
				for _, v := range d {
					n, ok := v.(float64)
					if ok {
						series[name] = append(series[name], n)
					}
				}
			}
		}
	}

	if len(times) == 0 {
		mj.log.Debug("MarcusolssonJson has no data")
		return nil
	}

	for i, t := range times {

		tags := make(map[string]string)

		for k, v := range series {

			tags["alias"] = k

			if len(v) > i {
				push(when, tags, time.UnixMilli(int64(t)), v[i])
			}
		}
	}
	return nil
}

func NewMarcusolssonJson(log telegraf.Logger, grafana *Grafana) *MarcusolssonJson {

	return &MarcusolssonJson{
		log:     log,
		grafana: grafana,
	}
}
