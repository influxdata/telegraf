package bind

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestBindXmlStatsV2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "bindstats-v2.xml")
	}))
	defer ts.Close()

	b := Bind{
		Urls: []string{ts.URL},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)
}

func TestBindXmlStatsV3(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "bindstats-v3.xml")
	}))
	defer ts.Close()

	b := Bind{
		Urls: []string{ts.URL},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)
}
