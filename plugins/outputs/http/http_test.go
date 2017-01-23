package http

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/influxdata/telegraf/metric"
	"time"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf"
	"io/ioutil"
	"fmt"
	"net/http"
	"sync"
)

var (
	tags = map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}

	fields = map[string]interface{}{
		"usage_idle": float64(91.5),
	}
)

func TestHttpWriteWithoutURL(t *testing.T) {
	m, _ := metric.New("cpu", tags, fields, time.Now())
	metrics := []telegraf.Metric{m}

	http := &Http{
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix: "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()

	if err := http.Write(metrics); err != nil {
		assert.Equal(t, "Http Output URL Option is empty! It is necessary.", err.Error())
	}
}

func TestHttpWriteNormalCase(t *testing.T) {
	now := time.Now()
	HTTPServer(t, now, 9880)

	m, _ := metric.New("cpu", tags, fields, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:"http://127.0.0.1:9880/metric1",
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix: "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	http.Write(metrics)
}

func TestHttpWriteWithIncorrectURLForRetry(t *testing.T) {
	now := time.Now()

	m, _ := metric.New("cpu", tags, fields, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:"http://127.0.0.1:9880/incorrect/url",
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix: "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	if err := http.Write(metrics); err != nil {
		assert.Equal(t, fmt.Sprintf("E! Since the retry limit %d has been reached, this request is discarded.", http.Retry), err.Error())
	}
}

func HTTPServer(t *testing.T, now time.Time, port int) {
	http.HandleFunc("/metric", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)

		assert.Equal(t, fmt.Sprintf("telegraf.cpu0.us-west-2.localhost.cpu.usage_idle 91.5 %d\n", now.Unix()), string(body))

		fmt.Fprintf(w, "ok")
	})

	var wg sync.WaitGroup
	wg.Add(1)

	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}