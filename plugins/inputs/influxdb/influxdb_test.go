package influxdb_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/influxdb"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	js := `
{
  "_1": {
    "name": "foo",
    "tags": {
      "id": "ex1"
    },
    "values": {
      "i": -1,
      "f": 0.5,
      "b": true,
      "s": "string"
    }
  },
  "ignored": {
    "willBeRecorded": false
  },
  "ignoredAndNested": {
    "hash": {
      "is": "nested"
    }
  },
  "array": [
   "makes parsing more difficult than necessary"
  ],
  "string": "makes parsing more difficult than necessary",
  "_2": {
    "name": "bar",
    "tags": {
      "id": "ex2"
    },
    "values": {
      "x": "x"
    }
  },
  "pointWithoutFields_willNotBeIncluded": {
    "name": "asdf",
    "tags": {
      "id": "ex3"
    },
    "values": {}
  }
}
`
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(js))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &influxdb.InfluxDB{
		URLs: []string{fakeServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Metrics, 2)
	fields := map[string]interface{}{
		// JSON will truncate floats to integer representations.
		// Since there's no distinction in JSON, we can't assume it's an int.
		"i": -1.0,
		"f": 0.5,
		"b": true,
		"s": "string",
	}
	tags := map[string]string{
		"id":  "ex1",
		"url": fakeServer.URL + "/endpoint",
	}
	acc.AssertContainsTaggedFields(t, "influxdb_foo", fields, tags)

	fields = map[string]interface{}{
		"x": "x",
	}
	tags = map[string]string{
		"id":  "ex2",
		"url": fakeServer.URL + "/endpoint",
	}
	acc.AssertContainsTaggedFields(t, "influxdb_bar", fields, tags)
}
