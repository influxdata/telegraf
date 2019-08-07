package apcupsd

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestApcupsdDocs(t *testing.T) {
	apc := &ApcUpsd{}
	apc.Description()
	apc.SampleConfig()
}

func TestApcupsdInit(t *testing.T) {
	input, ok := inputs.Inputs["apcupsd"]
	if !ok {
		t.Fatal("Input not defined")
	}

	_ = input().(*ApcUpsd)
}

func TestBoolToInt(t *testing.T) {
	one := boolToInt(true)
	if one != 1 {
		t.Errorf("boolToInt failed")
	}
}

func listen(t *testing.T) (string, error) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	go func() {
		for {
			defer ln.Close()

			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			defer conn.Close()

			in := make([]byte, 128)
			n, err := conn.Read(in)
			require.NoError(t, err, "failed to read from connection")

			status := []byte{0, 6, 's', 't', 'a', 't', 'u', 's'}
			want, got := status, in[:n]
			require.Equal(t, want, got)

			// Run against test function and append EOF to end of output bytes
			out := genOutput()
			out = append(out, []byte{0, 0})

			for _, o := range out {
				_, err := conn.Write(o)
				require.NoError(t, err, "failed to write to connection")
			}
		}
	}()

	return ln.Addr().String(), nil
}

func TestApcupsdGather(t *testing.T) {
	apc := &ApcUpsd{Timeout: defaultTimeout}

	lAddr, err := listen(t)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tests = []struct {
			servers []string
			err     bool
		}{
			{
				servers: []string{lAddr},
				err:     true,
			},
			{
				servers: []string{"127.0.0.3"},
				err:     true,
			},
			{
				servers: []string{"tcp://" + lAddr},
				err:     false,
			},
		}

		acc testutil.Accumulator
	)

	for _, test := range tests {
		fmt.Println("running test", test.servers)
		apc.Servers = test.servers

		err = apc.Gather(&acc)
		if test.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
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
