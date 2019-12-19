package suricata

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ex2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"capture":{"kernel_packets":905344474,"kernel_drops":78355440,"kernel_packets_delta":2376742,"kernel_drops_delta":82049}}}`
var ex3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W#05-wlp4s0": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`
var brokenType1 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W1#en..bar1": { "capture":{"kernel_packets":905344474,"kernel_drops": true}}}}}`
var brokenType2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W1#en..bar1": { "capture":{"kernel_packets":905344474,"kernel_drops": ["foo"]}}}}}`
var brokenType3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W1#en..bar1": { "capture":{"kernel_packets":905344474,"kernel_drops":"none this time"}}}}}`
var brokenType4 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W1#en..bar1": { "capture":{"kernel_packets":905344474,"kernel_drops":null}}}}}`
var brokenType5 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"foo": null}}`
var brokenStruct1 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": ["foo"]}}`
var brokenStruct2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats"}`
var brokenStruct3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats": "foobar"}`
var brokenStruct4 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats": null}`
var singleDotRegexp = regexp.MustCompilePOSIX(`[^.]\.[^.]`)

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

	acc.AssertContainsTaggedFields(t, "suricata",
		map[string]interface{}{
			"capture.kernel_packets":       float64(905344474),
			"capture.kernel_drops":         float64(78355440),
			"capture.kernel_packets_delta": float64(2376742),
			"capture.kernel_drops_delta":   float64(82049),
		},
		map[string]string{"thread": "total"})
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

	acc.AssertContainsTaggedFields(t, "suricata",
		map[string]interface{}{
			"capture.kernel_packets": float64(905344474),
			"capture.kernel_drops":   float64(78355440),
		},
		map[string]string{"thread": "W#05-wlp4s0"})
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

func TestSuricataInvalidInputs(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	for input, errmsg := range map[string]string{
		brokenType1:   `Unsupported type bool encountered`,
		brokenType2:   `Unsupported type []interface {} encountered`,
		brokenType3:   `Unsupported type string encountered`,
		brokenType4:   `Unsupported type <nil> encountered`,
		brokenType5:   `Unsupported type <nil> encountered`,
		brokenStruct1: `The 'threads' sub-object does not have required structure`,
		brokenStruct2: `Input does not contain necessary 'stats' sub-object`,
		brokenStruct3: `The 'stats' sub-object does not have required structure`,
		brokenStruct4: `The 'stats' sub-object does not have required structure`,
	} {
		var logBuf buffer
		logBuf.Reset()
		log.SetOutput(&logBuf)

		acc := testutil.Accumulator{}
		acc.SetDebug(true)

		s := Suricata{
			Source:    tmpfn,
			Delimiter: ".",
			Log: testutil.Logger{
				Name: "inputs.suricata",
			},
		}
		assert.NoError(t, s.Start(&acc))

		c, err := net.Dial("unix", tmpfn)
		if err != nil {
			t.Fatal(err)
		}
		c.Write([]byte(input))
		c.Write([]byte("\n"))
		c.Close()

		for {
			if bytes.Count(logBuf.Bytes(), []byte{'\n'}) > 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		assert.Contains(t, logBuf.String(), errmsg)
		s.Stop()
	}
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
