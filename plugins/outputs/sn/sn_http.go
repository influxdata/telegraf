package sn

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)
/*
Example for the JSON generated and pushed to the MID

{
	"metric_type":"telegraf.cpu_usage_system",
	"timestamp":1487365430,
	"value":0.8991008991008991,
	"ci_identifier":{"cpu":"cpu5","host":"MTLVIT1012136"},
	"source":"telegraf",
	"node":"MTLVIT1012136",
	"resource":
}

*/
type HttpMetric struct {
	Metric    string            `json:"metric_type"`
	Timestamp int64             `json:"timestamp"`
	Value     interface{}       `json:"value"`
	Tags      map[string]string `json:"ci_identifier"`
	Source    string 			`json:"source"`
	Node    string 				`json:"node"`
	Resource    string 			`json:"resource"`
}

type SNHttp struct {
	Url      string	
	Host      string	
	Scheme    string
	Username      string
	Password      string
	BatchSize int
	Debug     bool
	metricCounter int
	body          requestBody
}

type requestBody struct {
	b bytes.Buffer
	g *gzip.Writer

	dbgB bytes.Buffer

	w   io.Writer
	enc *json.Encoder

	empty bool
}

func (r *requestBody) reset(debug bool) {
	r.b.Reset()
	r.dbgB.Reset()

	if r.g == nil {
		r.g = gzip.NewWriter(&r.b)
	} else {
		r.g.Reset(&r.b)
	}

	r.w = io.MultiWriter(r.g, &r.dbgB)

	r.enc = json.NewEncoder(r.w)

	io.WriteString(r.w, "[")

	r.empty = true
}

func (r *requestBody) addMetric(metric *HttpMetric) error {
	if !r.empty {
		io.WriteString(r.w, ",")
	}

	//metricsBytes, err := json.Marshal(metric)
	//io.WriteString(r.w, string(metricsBytes))

	if err := r.enc.Encode(metric); err != nil {
		return fmt.Errorf("Metric serialization error %s", err.Error())
	}

	r.empty = false

	return nil
}

func (r *requestBody) close() error {
	io.WriteString(r.w, "]")

	if err := r.g.Close(); err != nil {
		return fmt.Errorf("Error when closing gzip writer: %s", err.Error())
	}

	return nil
}

func (o *SNHttp) sendDataPoint(metric *HttpMetric) error {
	if o.metricCounter == 0 {
		o.body.reset(o.Debug)
	}

	if err := o.body.addMetric(metric); err != nil {
		return err
	}

	o.metricCounter++
	if o.metricCounter == o.BatchSize {
		if err := o.flush(); err != nil {
			return err
		}

		o.metricCounter = 0
	}

	return nil
}

func (o *SNHttp) flush() error {
	if o.metricCounter == 0 {
		return nil
	}

	o.body.close()

	//var jsonStr = []byte(`{"metric_type":"telegraf.mysql_tc_log_page_waits","timestamp":1487454360000,"value":0,"ci_identifier":{"host":"MTLVIT1012136","server":"127.0.0.1_3306"},"source":"telegraf","node":"MTLVIT1012136","resource":""}`)
	req, err := http.NewRequest("POST", o.Url , bytes.NewBuffer([]byte (o.body.dbgB.String())))
	//req, err := http.NewRequest("POST", o.Url , &o.body.b)
	//req, err := http.NewRequest("POST", o.Url , bytes.NewBuffer(jsonStr))
	if err != nil {
		return fmt.Errorf("Error when building request: %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.SetBasicAuth(o.Username,o.Password) 

	if o.Debug {
		dump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			return fmt.Errorf("Error when dumping request: %s", err.Error())
		}

		fmt.Printf("Sending metrics:\n%s", dump)
		fmt.Printf("Body:\n%s\n\n", o.body.dbgB.String())
	}

	client := &http.Client{}
    resp, err := client.Do(req)

	//resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error when sending metrics: %s", err.Error())
	}
	defer resp.Body.Close()

	if o.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("Error when dumping response: %s", err.Error())
		}

		fmt.Printf("Received response\n%s\n\n", dump)
	} else {
		// Important so http client reuse connection for next request if need be.
		io.Copy(ioutil.Discard, resp.Body)
	}

	//fmt.Printf("Got status back:"+ resp.StatusCode)
	if resp.StatusCode/100 != 2 {
		if resp.StatusCode/100 == 4 {
			log.Printf("E! Received %d status code. Dropping metrics to avoid overflowing buffer.",
				resp.StatusCode)
		} else {
			return fmt.Errorf("Error when sending metrics. Received status %d",
				resp.StatusCode)
		}
	}

	return nil
}
