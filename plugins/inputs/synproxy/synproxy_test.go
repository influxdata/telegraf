// +build linux

package synproxy

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestSynproxyFileNormal(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(SynproxyFile_Normal))
	defer os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.NoError(t, err)

	fields := map[string]interface{}{
		"entries":        uint32(0x00000000),
		"syn_received":   uint32(0x0003e27b),
		"cookie_invalid": uint32(0x0001493e),
		"cookie_valid":   uint32(0x0003d7cf),
		"cookie_retrans": uint32(0x00000000),
		"conn_reopened":  uint32(0x00000000),
	}
	acc.AssertContainsFields(t, "synproxy", fields)
}

func TestSynproxyFileOverflow(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(SynproxyFile_Overflow))
	defer os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.NoError(t, err)

	fields := map[string]interface{}{
		"entries":        uint32(0x00000000),
		"syn_received":   uint32(0x00000004),
		"cookie_invalid": uint32(0xd0000009),
		"cookie_valid":   uint32(0x00000004),
		"cookie_retrans": uint32(0x00000000),
		"conn_reopened":  uint32(0x00000000),
	}
	acc.AssertContainsFields(t, "synproxy", fields)
}

func TestSynproxyFileHeaderMismatch(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(SynproxyFile_HeaderMismatch))
	defer os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid number of columns in data")
}

func TestSynproxyFileInvalidHex(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(SynproxyFile_InvalidHex))
	defer os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value")
}

func TestNoSynproxyFile(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(SynproxyFile_Normal))
	// Remove file to generate "no such file" error
	os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
}

// Valid Synproxy file
const SynproxyFile_Normal = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans  conn_reopened
00000000        00007a88        00002af7        00007995        00000000        00000000
00000000        0000892c        000015e3        00008852        00000000        00000000
00000000        00007a80        00002ccc        0000796a        00000000        00000000
00000000        000079f7        00002bf5        0000790a        00000000        00000000
00000000        00007a08        00002c9a        00007901        00000000        00000000
00000000        00007cfc        00002b36        000078fd        00000000        00000000
00000000        000079c2        00002c2b        000078d6        00000000        00000000
00000000        0000798a        00002ba8        000078a0        00000000        00000000`

const SynproxyFile_Overflow = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans  conn_reopened
00000000        80000001        e0000000        80000001        00000000        00000000
00000000        80000003        f0000009        80000003        00000000        00000000`

const SynproxyFile_HeaderMismatch = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans
00000000        00000002        00000000        00000002        00000000        00000000
00000000        00000004        00000015        00000004        00000000        00000000
00000000        00000003        00000000        00000003        00000000        00000000
00000000        00000002        00000000        00000002        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000
00000000        00000001        00000000        00000001        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000`

const SynproxyFile_InvalidHex = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans  conn_reopened
entries        00000002        00000000        00000002        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000`

func makeFakeSynproxyFile(content []byte) string {
	tmpfile, err := ioutil.TempFile("", "synproxy_test")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile.Name()
}
