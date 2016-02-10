package enterprise_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/enterprise-client/v2"
	"github.com/influxdata/telegraf/services/enterprise"
)

func mockEnterprise(srvFunc func(*client.Product, error)) (chan struct{}, *httptest.Server) {
	success := make(chan struct{})
	me := http.NewServeMux()
	me.HandleFunc("/api/v2/products", func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			c := &client.Product{}
			d := json.NewDecoder(r.Body)
			err := d.Decode(c)
			srvFunc(c, err)
			close(success)
		}
	})
	srv := httptest.NewServer(me)
	return success, srv
}

func Test_RegistersWithEnterprise(t *testing.T) {
	expected := "www.example.com"
	var actualHostname string
	success, srv := mockEnterprise(func(c *client.Product, err error) {
		if err != nil {
			t.Error(err.Error())
		}
		actualHostname = c.Host
	})
	defer srv.Close()

	c := enterprise.Config{
		Hosts: []*client.Host{
			&client.Host{URL: srv.URL},
		},
	}
	e := enterprise.NewEnterprise(c, expected)
	e.Open()

	timeout := time.After(1 * time.Millisecond)
	for {
		select {
		case <-success:
			if actualHostname != expected {
				t.Errorf("Expected hostname to be %s but was %s", expected, actualHostname)
			}
			return
		case <-timeout:
			t.Fatal("Expected to receive call to Enterprise API, but received none")
		}
	}
}
