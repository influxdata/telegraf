package enterprise_test

import (
	"encoding/json"
	"fmt"
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

	shutdown := make(chan struct{})
	defer close(shutdown)
	e := enterprise.NewEnterprise(c, expected, shutdown)
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

func Test_StartsAdminInterface(t *testing.T) {
	hostname := "localhost"
	adminPort := 2300

	success, srv := mockEnterprise(func(c *client.Product, err error) {})
	defer srv.Close()

	c := enterprise.Config{
		Hosts: []*client.Host{
			&client.Host{URL: srv.URL},
		},
		AdminPort: 2300,
	}

	shutdown := make(chan struct{})
	defer close(shutdown)
	e := enterprise.NewEnterprise(c, hostname, shutdown)
	e.Open()

	timeout := time.After(1 * time.Millisecond)
	for {
		select {
		case <-success:
			_, err := http.Get(fmt.Sprintf("http://%s:%d", hostname, adminPort))
			if err != nil {
				t.Errorf("Unable to connect to admin interface: err: %s", err)
			}
			return
		case <-timeout:
			t.Fatal("Expected to receive call to Enterprise API, but received none")
		}
	}
}

func Test_ClosesAdminInterface(t *testing.T) {
	hostname := "localhost"
	adminPort := 2300

	success, srv := mockEnterprise(func(c *client.Product, err error) {})
	defer srv.Close()

	c := enterprise.Config{
		Hosts: []*client.Host{
			&client.Host{URL: srv.URL},
		},
		AdminPort: 2300,
	}

	shutdown := make(chan struct{})
	e := enterprise.NewEnterprise(c, hostname, shutdown)
	e.Open()

	timeout := time.After(1 * time.Millisecond)
	for {
		select {
		case <-success:
			// Ensure that the admin interface is running
			_, err := http.Get(fmt.Sprintf("http://%s:%d", hostname, adminPort))
			if err != nil {
				t.Errorf("Unable to connect to admin interface: err: %s", err)
			}
			close(shutdown)

			// ...and that it's not running after we shut it down
			_, err = http.Get(fmt.Sprintf("http://%s:%d", hostname, adminPort))
			if err == nil {
				t.Errorf("Admin interface continued running after shutdown")
			}
			return
		case <-timeout:
			t.Fatal("Expected to receive call to Enterprise API, but received none")
		}
	}
}
