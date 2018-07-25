package apcupsd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

func TestApcupsdGather(t *testing.T) {
	input, ok := inputs.Inputs["apcupsd"]
	if !ok {
		t.Fatal("Input not defined")
	}
	apc := input().(*ApcUpsd)

	apc.Description()
	apc.SampleConfig()

	one := boolToInt(true)
	if one != 1 {
		t.Errorf("boolToInt failed")
	}

	if testing.Short() {
		t.Skip("Skipping network dependent tests in short mode")
	}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		if err != nil {
			return
		}

		in := make([]byte, 128)
		n, err := conn.Read(in)
		if err != nil {
			t.Fatal(fmt.Sprintf("failed to read from connection - %s", err.Error()))
		}

		status := []byte{0, 6, 's', 't', 'a', 't', 'u', 's'}
		if want, got := status, in[:n]; !bytes.Equal(want, got) {
			t.Fatal(fmt.Sprintf("unexpected request from Client:\n- want: %v\n - got: %v",
				want, got))
		}

		// Run against test function and append EOF to end of output bytes
		out := genOutput()
		out = append(out, []byte{0, 0})

		for _, o := range out {
			if _, err := conn.Write(o); err != nil {
				t.Fatal(fmt.Sprintf("failed to write to connection - %s", err.Error()))
			}
		}
	}()

	var acc testutil.Accumulator

	apc.Servers = []string{ln.Addr().String()}

	err = apc.Gather(&acc)
	if err == nil {
		t.Fatal("Should have failed but didn't")
	}

	apc.Servers = []string{"127.0.0.3"}

	err = apc.Gather(&acc)
	if err == nil {
		t.Fatal("Should have failed but didn't")
	}

	apc.Servers = []string{"tcp://" + ln.Addr().String()}

	err = apc.Gather(&acc)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed gathering when shouldn't - %s", err.Error()))
	}
	wg.Wait()
}

// The following functionality is straight from apcupsd tests.

// kvBytes is a helper to generate length and key/value byte buffers.
func kvBytes(kv string) ([]byte, []byte) {
	lenb := make([]byte, 2)
	binary.BigEndian.PutUint16(lenb, uint16(len(kv)))

	return lenb, []byte(kv)
}

func genOutput() [][]byte {
	kvs := []string{
		"DATE     : 2016-09-06 22:13:28 -0400",
		"HOSTNAME : example",
		"LOADPCT  :  13.0 Percent Load Capacity",
		"BATTDATE : 2016-09-06",
		"TIMELEFT :  46.5 Minutes",
		"TONBATT  : 0 seconds",
		"NUMXFERS : 0",
		"SELFTEST : NO",
		"NOMPOWER : 865 Watts",
	}

	var out [][]byte
	for _, kv := range kvs {
		lenb, kvb := kvBytes(kv)
		out = append(out, lenb)
		out = append(out, kvb)
	}

	return out
}
