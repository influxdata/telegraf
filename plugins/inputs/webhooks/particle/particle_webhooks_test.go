package particle

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(rb *ParticleWebhook, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(eventBody))
	w := httptest.NewRecorder()
	w.Code = 500

	rb.eventHandler(w, req)

	return w
}

func TestNewItem(t *testing.T) {
	t.Parallel()
	var acc testutil.Accumulator
	rb := &ParticleWebhook{Path: "/particle", acc: &acc}
	resp := postWebhooks(rb, NewItemJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"temp_c":    26.680000,
		"temp_f":    80.024001,
		"infrared":  528.0,
		"lux":       0.0,
		"humidity":  44.937500,
		"pressure":  998.998901,
		"altitude":  119.331436,
		"broadband": 1266.0,
	}

	tags := map[string]string{
		"id":       "230035001147343438323536",
		"location": "TravelingWilbury",
	}

	acc.AssertContainsTaggedFields(t, "temperature", fields, tags)
}

func TestUnknowItem(t *testing.T) {
	t.Parallel()
	var acc testutil.Accumulator
	rb := &ParticleWebhook{Path: "/particle", acc: &acc}
	resp := postWebhooks(rb, UnknowJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST unknown returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}
}

func NewItemJSON() string {
	return `
	{
	  "event": "temperature",
	  "data": {
		  "tags": {
			  "id": "230035001147343438323536",
			  "location": "TravelingWilbury"
		  },
		  "values": {
			  "temp_c": 26.680000,
			  "temp_f": 80.024001,
			  "humidity": 44.937500,
			  "pressure": 998.998901,
			  "altitude": 119.331436,
			  "broadband": 1266.0,
			  "infrared": 528.0,
			  "lux": 0.0
		  }
	  },
	  "ttl": 60,
	  "published_at": "2017-09-28T21:54:10.897Z",
	  "coreid": "123456789938323536",
	  "userid": "1234ee123ac8e5ec1231a123d",
	  "version": 10,
	  "public": false,
	  "productID": 1234,
	  "name": "sensor",
	  "measurement": "mydata"
  }`
}

func UnknowJSON() string {
	return `
    {
      "event": "roger"
    }`
}
