package flume

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//Flume config
type Flume struct {
	Server string
}

func (f *Flume) Description() string {
	return "Read metrics from one server"
}

func (f *Flume) SampleConfig() string {
	return `
  ## specify servers via a url matching:
  ##
  server = "http://localhost:6666/metrics"
`
}

func (f *Flume) Gather(acc telegraf.Accumulator) error {

	url := f.Server

	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var metrics map[string]json.RawMessage

	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return err
	}

	for c, mm := range metrics {

		tags := map[string]string{"component": c}

		metric := map[string]interface{}{}

		err := json.Unmarshal([]byte(mm), &metric)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{}
		for n, value := range metric {
			fields[n] = value
		}
		acc.AddFields("flume", fields, tags)

	}

	return nil
}
func init() {
	inputs.Add("flume", func() telegraf.Input { return &Flume{} })
}
