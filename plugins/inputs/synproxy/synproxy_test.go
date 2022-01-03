//go:build linux
// +build linux

package synproxy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestSynproxyFileNormal(t *testing.T) {
	testSynproxyFileData(t, synproxyFileNormal, synproxyResultNormal)
}

func TestSynproxyFileOverflow(t *testing.T) {
	testSynproxyFileData(t, synproxyFileOverflow, synproxyResultOverflow)
}

func TestSynproxyFileExtended(t *testing.T) {
	testSynproxyFileData(t, synproxyFileExtended, synproxyResultNormal)
}

func TestSynproxyFileAltered(t *testing.T) {
	testSynproxyFileData(t, synproxyFileAltered, synproxyResultNormal)
}

func TestSynproxyFileHeaderMismatch(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(synproxyFileHeaderMismatch))
	defer os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid number of columns in data")
}

func TestSynproxyFileInvalidHex(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(synproxyFileInvalidHex))
	defer os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid value")
}

func TestNoSynproxyFile(t *testing.T) {
	tmpfile := makeFakeSynproxyFile([]byte(synproxyFileNormal))
	// Remove file to generate "no such file" error
	// Ignore errors if file does not yet exist
	//nolint:errcheck,revive
	os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.Error(t, err)
}

// Valid Synproxy file
const synproxyFileNormal = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans  conn_reopened
00000000        00007a88        00002af7        00007995        00000000        00000000
00000000        0000892c        000015e3        00008852        00000000        00000000
00000000        00007a80        00002ccc        0000796a        00000000        00000000
00000000        000079f7        00002bf5        0000790a        00000000        00000000
00000000        00007a08        00002c9a        00007901        00000000        00000000
00000000        00007cfc        00002b36        000078fd        00000000        00000000
00000000        000079c2        00002c2b        000078d6        00000000        00000000
00000000        0000798a        00002ba8        000078a0        00000000        00000000`

const synproxyFileOverflow = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans  conn_reopened
00000000        80000001        e0000000        80000001        00000000        00000000
00000000        80000003        f0000009        80000003        00000000        00000000`

const synproxyFileHeaderMismatch = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans
00000000        00000002        00000000        00000002        00000000        00000000
00000000        00000004        00000015        00000004        00000000        00000000
00000000        00000003        00000000        00000003        00000000        00000000
00000000        00000002        00000000        00000002        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000
00000000        00000001        00000000        00000001        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000`

const synproxyFileInvalidHex = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans  conn_reopened
entries        00000002        00000000        00000002        00000000        00000000
00000000        00000003        00000009        00000003        00000000        00000000`

const synproxyFileExtended = `entries         syn_received    cookie_invalid  cookie_valid    cookie_retrans  conn_reopened   new_counter
00000000        00007a88        00002af7        00007995        00000000        00000000        00000000
00000000        0000892c        000015e3        00008852        00000000        00000000        00000000
00000000        00007a80        00002ccc        0000796a        00000000        00000000        00000000
00000000        000079f7        00002bf5        0000790a        00000000        00000000        00000000
00000000        00007a08        00002c9a        00007901        00000000        00000000        00000000
00000000        00007cfc        00002b36        000078fd        00000000        00000000        00000000
00000000        000079c2        00002c2b        000078d6        00000000        00000000        00000000
00000000        0000798a        00002ba8        000078a0        00000000        00000000        00000000`

const synproxyFileAltered = `entries         cookie_invalid  cookie_valid    syn_received    conn_reopened
00000000        00002af7        00007995        00007a88        00000000
00000000        000015e3        00008852        0000892c        00000000
00000000        00002ccc        0000796a        00007a80        00000000
00000000        00002bf5        0000790a        000079f7        00000000
00000000        00002c9a        00007901        00007a08        00000000
00000000        00002b36        000078fd        00007cfc        00000000
00000000        00002c2b        000078d6        000079c2        00000000
00000000        00002ba8        000078a0        0000798a        00000000`

var synproxyResultNormal = map[string]interface{}{
	"entries":        uint32(0x00000000),
	"syn_received":   uint32(0x0003e27b),
	"cookie_invalid": uint32(0x0001493e),
	"cookie_valid":   uint32(0x0003d7cf),
	"cookie_retrans": uint32(0x00000000),
	"conn_reopened":  uint32(0x00000000),
}

var synproxyResultOverflow = map[string]interface{}{
	"entries":        uint32(0x00000000),
	"syn_received":   uint32(0x00000004),
	"cookie_invalid": uint32(0xd0000009),
	"cookie_valid":   uint32(0x00000004),
	"cookie_retrans": uint32(0x00000000),
	"conn_reopened":  uint32(0x00000000),
}

func testSynproxyFileData(t *testing.T, fileData string, telegrafData map[string]interface{}) {
	tmpfile := makeFakeSynproxyFile([]byte(fileData))
	defer os.Remove(tmpfile)

	k := Synproxy{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsFields(t, "synproxy", telegrafData)
}

func makeFakeSynproxyFile(content []byte) string {
	tmpfile, err := os.CreateTemp("", "synproxy_test")
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
