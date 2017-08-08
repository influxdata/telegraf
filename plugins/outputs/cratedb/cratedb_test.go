package cratedb

import (
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	return
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	c := &CrateDB{
		URL:         testURL(),
		Table:       "test",
		Timeout:     internal.Duration{Duration: time.Second * 5},
		TableCreate: true,
	}

	require.NoError(t, c.Connect())
	require.NoError(t, c.Write(testutil.MockMetrics()))
	require.NoError(t, c.Close())
}

func Test_insertSQL(t *testing.T) {
	return
	tests := []struct {
		Metrics []telegraf.Metric
		Want    string
	}{
		{
			Metrics: testutil.MockMetrics(),
			Want:    "INSERT ...",
		},
	}

	for _, test := range tests {
		if got, err := insertSQL("my_table", test.Metrics); err != nil {
			t.Error(err)
		} else if got != test.Want {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, test.Want)
		}
	}
}

func Test_escapeValue(t *testing.T) {
	tests := []struct {
		Val  interface{}
		Want string
	}{
		{`foo`, `'foo'`},
		{`foo'bar 'yeah`, `'foo''bar ''yeah'`},
		{time.Date(2017, 8, 7, 16, 44, 52, 123*1000*1000, time.FixedZone("Dreamland", 5400)), `'2017-08-07T16:44:52.123+0130'`},
		{map[string]string{"foo": "bar"}, `{"foo" = 'bar'}`},
	}

	for _, test := range tests {
		if got, err := escapeValue(test.Val); err != nil {
			t.Errorf("val: %#v: %s", test.Val, err)
		} else if got != test.Want {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, test.Want)
		}
	}
}

func testURL() string {
	url := os.Getenv("CRATE_URL")
	if url == "" {
		// @TODO use telegraf helper func for hostname
		return "postgres://localhost:6543/test?sslmode=disable"
	}
	return url
}
