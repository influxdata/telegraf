package fireboard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestFireboard(t *testing.T) {
	// Create a test server with the const response JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintln(w, response)
		require.NoError(t, err)
	}))
	defer ts.Close()

	// Parse the URL of the test server, used to verify the expected host
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)

	// Create a new fb instance with our given test server
	fireboard := NewFireboard()
	fireboard.AuthToken = "b4bb6e6a7b6231acb9f71b304edb2274693d8849"
	fireboard.URL = u.String()

	// Create a test accumulator
	acc := &testutil.Accumulator{}

	// Gather data from the test server
	err = fireboard.Gather(acc)
	require.NoError(t, err)

	// Expect the correct values for all known keys
	expectFields := map[string]interface{}{
		"temperature": float64(79.9),
	}
	// Expect the correct values for all tags
	expectTags := map[string]string{
		"title":   "telegraf-FireBoard",
		"uuid":    "b55e766c-b308-49b5-93a4-df89fe31efd0",
		"channel": strconv.FormatInt(1, 10),
		"scale":   "Fahrenheit",
	}

	acc.AssertContainsTaggedFields(t, "fireboard", expectFields, expectTags)
}

var response = `
[{
	"id": 99999,
	"title": "telegraf-FireBoard",
	"created": "2019-03-23T16:48:32.152010Z",
	"uuid": "b55e766c-b308-49b5-93a4-df89fe31efd0",
	"hardware_id": "XXXXXXXXX",
	"latest_temps": [
	  {
		"temp": 79.9,
		"channel": 1,
		"degreetype": 2,
		"created": "2019-06-25T06:07:10Z"
	  }
	],
	"last_templog": "2019-06-25T06:06:40Z",
	"model": "FBX11E",
	"channel_count": 6,
	"degreetype": 2
  }]
`
