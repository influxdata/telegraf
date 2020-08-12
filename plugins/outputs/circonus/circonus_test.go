// Package circonus contains the output plugin used to write metric data to a
// Circonus broker.
package circonus

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	itls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

func TestCirconus(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}

			exp := `{"test1.value|ST[tag1:value1]":{"_type":"L","_value":1}}`
			if string(body) != exp {
				t.Errorf("Expected: %v, got: %v", exp, string(body))
			}

			wg.Done()
		}))

	defer ts.Close()

	cli := &Circonus{
		Checks:       map[string]string{".*": ts.URL},
		APIURL:       "http://test.com",
		APIToken:     "11223344-5566-7788-9900-aabbccddeeff",
		APIApp:       "telegraf",
		ClientConfig: itls.ClientConfig{InsecureSkipVerify: true},
	}

	err := cli.Init()
	if err != nil {
		t.Fatal(err)
	}

	if cli.SampleConfig() != sampleConfig {
		t.Errorf("Expected config: %v, got: %v", sampleConfig,
			cli.SampleConfig())
	}

	if cli.Description() != description {
		t.Errorf("Expected description: %v, got: %v", description,
			cli.Description())
	}

	err = cli.Connect()
	if err != nil {
		t.Fatal(err)
	}

	err = cli.Write(testutil.MockMetrics())
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	err = cli.Close()
	if err != nil {
		t.Fatal(err)
	}
}
