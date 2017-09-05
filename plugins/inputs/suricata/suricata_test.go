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

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var ex2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"capture":{"kernel_packets":905344474,"kernel_drops":78355440,"kernel_packets_delta":2376742,"kernel_drops_delta":82049}}}`
var ex3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "foo": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`
var ex4 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "W1#en..bar1": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`

func TestSuricata(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
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

	acc.Wait(1)

	s.Stop()

	acc.AssertContainsTaggedFields(t, "suricata",
		map[string]interface{}{
			"capture.kernel_packets":               float64(905344474),
			"capture.kernel_drops":                 float64(78355440),
			"capture.kernel_packets_delta":         float64(2376742),
			"capture.kernel_drops_delta":           float64(82049),
			"capture.kernel_drop_percentage":       float64(0.07965380385303154),
			"capture.kernel_drop_delta_percentage": float64(0.033369651995635255),
			"event_type":                           "stats",
			"timestamp":                            "2017-03-06T07:43:39.000397+0000",
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

	acc.AssertContainsTaggedFields(t, "suricata",
		map[string]interface{}{
			"event_type": "stats",
			"timestamp":  "2017-03-06T07:43:39.000397+0000",
		},
		map[string]string{"thread": "total"})
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
		Source: tmpfn,
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

func TestSuricataDisconnectSocket(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
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
	v, ok := inputs.Inputs["suricata"]
	if !ok {
		t.Fatal("suricata plugin not registered")
	}
	s := v().(telegraf.ServiceInput)
	acc := testutil.Accumulator{}
	acc.SetDebug(true)
	assert.NoError(t, s.Start(&acc))
	s.Stop()
}

func TestSuricataDoubleStart(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
	}
	acc := testutil.Accumulator{}
	acc.SetDebug(true)
	assert.NoError(t, s.Start(&acc))
	assert.NoError(t, s.Start(&acc))

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	c.Write([]byte(ex2))
	c.Write([]byte("\n"))
	c.Close()

	acc.Wait(1)
	s.Stop()
}

func TestSuricataDoubleStop(t *testing.T) {
	v, ok := inputs.Inputs["suricata"]
	if !ok {
		t.Fatal("suricata plugin not registered")
	}
	s := v().(telegraf.ServiceInput)
	acc := testutil.Accumulator{}
	acc.SetDebug(true)
	assert.NoError(t, s.Start(&acc))
	s.Stop()
	s.Stop()
}

func TestSuricataGather(t *testing.T) {
	v, ok := inputs.Inputs["suricata"]
	if !ok {
		t.Fatal("suricata plugin not registered")
	}
	s := v().(telegraf.ServiceInput)
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
