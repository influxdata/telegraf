package common

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestRootCAs(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "https://portal.azure.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	c := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: "portal.azure.com",
			RootCAs:    RootCAs(),
		},
	}}
	res, err := c.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
}
