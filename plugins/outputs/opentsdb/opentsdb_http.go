package opentsdb

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type HTTPMetric struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     interface{}       `json:"value"`
	Tags      map[string]string `json:"tags"`
}

type openTSDBHttp struct {
	Host      string
	Port      int
	Scheme    string
	User      *url.Userinfo
	BatchSize int
	Path      string
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

	if debug {
		r.w = io.MultiWriter(r.g, &r.dbgB)
	} else {
		r.w = r.g
	}

	r.enc = json.NewEncoder(r.w)

	io.WriteString(r.w, "[")

	r.empty = true
}

func (r *requestBody) addMetric(metric *HTTPMetric) error {
	if !r.empty {
		io.WriteString(r.w, ",")
	}

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

func (o *openTSDBHttp) sendDataPoint(metric *HTTPMetric) error {
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

func (o *openTSDBHttp) flush() error {
	if o.metricCounter == 0 {
		return nil
	}

	o.body.close()

	u := url.URL{
		Scheme: o.Scheme,
		User:   o.User,
		Host:   fmt.Sprintf("%s:%d", o.Host, o.Port),
		Path:   o.Path,
	}

	if o.Debug {
		u.RawQuery = "details"
	}

	req, err := http.NewRequest("POST", u.String(), &o.body.b)
	if err != nil {
		return fmt.Errorf("Error when building request: %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if o.Debug {
		dump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			return fmt.Errorf("Error when dumping request: %s", err.Error())
		}

		fmt.Printf("Sending metrics:\n%s", dump)
		fmt.Printf("Body:\n%s\n\n", o.body.dbgB.String())
	}

	resp, err := http.DefaultClient.Do(req)
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
		_, _ = io.Copy(io.Discard, resp.Body)
	}

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
