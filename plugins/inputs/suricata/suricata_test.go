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

var ex2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}`
var ex3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":{"threads": { "foo": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`

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
			"capture.kernel_packets": float64(905344474),
			"capture.kernel_drops":   float64(78355440),
			"event_type":             "stats",
			"timestamp":              "2017-03-06T07:43:39.000397+0000",
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
