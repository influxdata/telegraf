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
	c.Write([]byte(data))
	c.Write([]byte("\n"))
	c.Close()

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
	c.Write([]byte(ex2))
	c.Write([]byte("\n"))
	c.Close()

	acc.Wait(1)

	s = Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}

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
	c.Write([]byte(""))
	c.Write([]byte("\n"))
	c.Write([]byte("foobard}\n"))
	c.Write([]byte(ex3))
	c.Write([]byte("\n"))
	c.Close()
	acc.Wait(1)

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
	c.Write([]byte("sfjiowef"))
	c.Write([]byte("\n"))
	c.Close()

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
	c.Write([]byte(strings.Repeat("X", 20000000)))
	c.Write([]byte("\n"))
	c.Close()

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
	c.Write([]byte("\n"))
	c.Close()

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
	c.Write([]byte(ex2))
	c.Write([]byte("\n"))
	c.Close()

	c, err = net.Dial("unix", tmpfn)
	require.NoError(t, err)
	c.Write([]byte(ex3))
	c.Write([]byte("\n"))
	c.Close()

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
