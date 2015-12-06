package expvar

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdb/telegraf/testutil"
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
		if r.URL.Path == "/expvar" {
			_, _ = w.Write([]byte(js))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	expvar := &Expvar{
		Services: []Service{
			{
				Name: "test",
				URLs: []string{fakeServer.URL + "/expvar"},
			},
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, expvar.Gather(&acc))

	require.Len(t, acc.Points, 2)
	require.NoError(t, acc.ValidateTaggedFieldsValue(
		"test_foo",
		map[string]interface{}{
			// JSON will truncate floats to integer representations.
			// Since there's no distinction in JSON, we can't assume it's an int.
			"i": -1.0,
			"f": 0.5,
			"b": true,
			"s": "string",
		},
		map[string]string{
			"id":         "ex1",
			"expvar_url": fakeServer.URL + "/expvar",
		},
	))
	require.NoError(t, acc.ValidateTaggedFieldsValue(
		"test_bar",
		map[string]interface{}{
			"x": "x",
		},
		map[string]string{
			"id":         "ex2",
			"expvar_url": fakeServer.URL + "/expvar",
		},
	))
}
