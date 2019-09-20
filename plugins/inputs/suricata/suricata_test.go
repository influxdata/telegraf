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

	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var ex2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"capture":{"kernel_packets":905344474,"kernel_drops":78355440,"kernel_packets_delta":2376742,"kernel_drops_delta":82049}}}`
var ex3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "foo": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`
var ex4 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W1#en..bar1": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`
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
	if err != nil {
		t.Fatal(err)
	}
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
	acc.SetDebug(true)
	assert.NoError(t, s.Start(&acc))

	data, err := ioutil.ReadFile("testdata/test1.json")
	if err != nil {
		t.Fatal(err)
	}

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		t.Fatal(err)
	}
	c.Write([]byte(data))
	c.Write([]byte("\n"))
	c.Close()

	acc.Wait(1)

	s.Stop()
}

func TestSuricata(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
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
	acc.SetDebug(true)
	assert.NoError(t, s.Start(&acc))

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		t.Fatalf("failed: %s", err.Error())
	}
	c.Write([]byte(ex2))
	c.Write([]byte("\n"))
	c.Close()

	acc.Wait(1)

	s.Stop()
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

	acc = testutil.Accumulator{}
	acc.SetDebug(true)
	assert.NoError(t, s.Start(&acc))

	c, err = net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte(""))
	c.Write([]byte("\n"))
	c.Write([]byte("foobard}\n"))
	c.Write([]byte(ex3))
	c.Write([]byte("\n"))
	c.Close()
	acc.Wait(1)

	s.Stop()

	acc.AssertContainsTaggedFields(t, "suricata",
		map[string]interface{}{
			"capture.kernel_packets": float64(905344474),
			"capture.kernel_drops":   float64(78355440),
		},
		map[string]string{"thread": "foo"})
}

func TestSuricataInvalid(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
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

	assert.NoError(t, s.Start(&acc))

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte("sfjiowef"))
	c.Write([]byte("\n"))
	c.Close()

	acc.WaitError(1)
	s.Stop()
}

func splitAtSingleDot(in string) []string {
	res := singleDotRegexp.FindAllStringIndex(in, -1)
	if res == nil {
		return []string{in}
	}
	ret := make([]string, 0)
	startpos := 0
	for _, v := range res {
		ret = append(ret, in[startpos:v[0]+1])
		startpos = v[1] - 1
	}
	return append(ret, in[startpos:])
}

func TestSuricataSplitDots(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	out := splitAtSingleDot("foo")
	if len(out) != 1 {
		t.Fatalf("splitting 'foo' should yield one result")
	}
	if out[0] != "foo" {
		t.Fatalf("splitting 'foo' should yield one result, 'foo'")
	}

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	acc.SetDebug(true)

	assert.NoError(t, s.Start(&acc))

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte(ex4))
	c.Write([]byte("\n"))
	c.Close()

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "suricata",
		map[string]interface{}{
			"capture.kernel_packets": float64(905344474),
			"capture.kernel_drops":   float64(78355440),
		},
		map[string]string{"thread": "W1#en..bar1"})

	s.Stop()
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
	acc.SetDebug(true)

	assert.Error(t, s.Start(&acc))
}

func TestSuricataTooLongLine(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
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

	assert.NoError(t, s.Start(&acc))

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte(strings.Repeat("X", 20000000)))
	c.Write([]byte("\n"))
	c.Close()

	acc.WaitError(1)

	s.Stop()
}

func TestSuricataEmptyJSON(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
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

	assert.NoError(t, s.Start(&acc))

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte("\n"))
	c.Close()

	acc.WaitError(1)

	s.Stop()
}

func TestSuricataInvalidInputs(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
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

	assert.NoError(t, s.Start(&acc))

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte(ex2))
	c.Write([]byte("\n"))
	c.Close()

	c, err = net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte(ex3))
	c.Write([]byte("\n"))
	c.Close()

	acc.Wait(2)

	s.Stop()
}

func TestSuricataPluginDesc(t *testing.T) {
	v, ok := inputs.Inputs["suricata"]
	if !ok {
		t.Fatal("suricata plugin not registered")
	}
	desc := v().Description()
	if desc != "Suricata stats plugin" {
		t.Fatal("invalid description ", desc)
	}
}

func TestSuricataStartStop(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
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
	assert.NoError(t, s.Start(&acc))
	s.Stop()
}

func TestSuricataGather(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
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
	assert.NoError(t, s.Gather(&acc))
}

func TestSuricataSampleConfig(t *testing.T) {
	v, ok := inputs.Inputs["suricata"]
	if !ok {
		t.Fatal("suricata plugin not registered")
	}
	if v().SampleConfig() != sampleConfig {
		t.Fatal("wrong sampleconfig")
	}
}
