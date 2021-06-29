package suricata

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var ex2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"capture":{"kernel_packets":905344474,"kernel_drops":78355440,"kernel_packets_delta":2376742,"kernel_drops_delta":82049}}}`
var ex3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W#05-wlp4s0": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`

func TestSuricataLarge(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	data, err := ioutil.ReadFile("testdata/test1.json")
	require.NoError(t, err)

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write(data)
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.Wait(1)
}

func TestSuricata(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(ex2))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.Wait(1)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"suricata",
			map[string]string{
				"thread": "total",
			},
			map[string]interface{}{
				"capture.kernel_packets":       float64(905344474),
				"capture.kernel_drops":         float64(78355440),
				"capture.kernel_packets_delta": float64(2376742),
				"capture.kernel_drops_delta":   float64(82049),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestThreadStats(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(""))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	_, err = c.Write([]byte("foobard}\n"))
	require.NoError(t, err)
	_, err = c.Write([]byte(ex3))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())
	acc.Wait(2)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"suricata",
			map[string]string{
				"thread": "W#05-wlp4s0",
			},
			map[string]interface{}{
				"capture.kernel_packets": float64(905344474),
				"capture.kernel_drops":   float64(78355440),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestSuricataInvalid(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	acc.SetDebug(true)

	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte("sfjiowef"))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.WaitError(1)
}

func TestSuricataInvalidPath(t *testing.T) {
	tmpfn := fmt.Sprintf("/t%d/X", rand.Int63())
	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}

	acc := testutil.Accumulator{}
	require.Error(t, s.Start(&acc))
}

func TestSuricataTooLongLine(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}

	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(strings.Repeat("X", 20000000)))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.WaitError(1)
}

func TestSuricataEmptyJSON(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.WaitError(1)
}

func TestSuricataDisconnectSocket(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}

	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(ex2))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	c, err = net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(ex3))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.Wait(2)
}

func TestSuricataStartStop(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	s.Stop()
}

func TestSuricataParse(t *testing.T) {
	tests := []struct {
		filename string
		expected []telegraf.Metric
	}{{
		filename: "test2.json",
		expected: []telegraf.Metric{
			testutil.MustMetric(
				"suricata",
				map[string]string{
					"thread": "W#01-ens2f1",
				},
				map[string]interface{}{
					"detect_alert":                float64(0),
					"detect_engines_id":           float64(0),
					"detect_engines_last_reload":  "2021-06-08T06:33:05.084872+0000",
					"detect_engines_rules_failed": float64(0),
					"detect_engines_rules_loaded": float64(22712),
				},
				time.Unix(0, 0),
			),
		},
	},
	}

	for _, tc := range tests {
		data, err := ioutil.ReadFile("testdata/" + tc.filename)
		require.NoError(t, err)
		s := Suricata{
			Delimiter: "_",
		}
		acc := testutil.Accumulator{}
		s.parse(&acc, data)

		testutil.RequireMetricsEqual(t, tc.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
	}
}
