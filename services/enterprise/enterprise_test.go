package enterprise_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/enterprise-client/v2"
	"github.com/influxdata/telegraf/services/enterprise"
)

func Test_RegistersWithEnterprise(t *testing.T) {
	success := make(chan struct{})
	mockEnterprise := http.NewServeMux()
	mockEnterprise.HandleFunc("/api/v2/products", func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			close(success)
		}
	})
	srv := httptest.NewServer(mockEnterprise)
	defer srv.Close()

	c := enterprise.Config{
		Hosts: []*client.Host{
			&client.Host{URL: srv.URL},
		},
	}
	e := enterprise.NewEnterprise(c)
	e.Open()

	timeout := time.After(50 * time.Millisecond)
	for {
		select {
		case <-success:
			return
		case <-timeout:
			t.Fatal("Expected to receive call to Enterprise API, but received none")
		}
	}
}
