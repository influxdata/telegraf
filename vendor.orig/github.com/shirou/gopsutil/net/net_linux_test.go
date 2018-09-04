package net

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/shirou/gopsutil/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestIOCountersByFileParsing(t *testing.T) {
	// Prpare a temporary file, which will be read during the test
	tmpfile, err := ioutil.TempFile("", "proc_dev_net")
	defer os.Remove(tmpfile.Name()) // clean up

	assert.Nil(t, err, "Temporary file creation failed: ", err)

	cases := [4][2]string{
		[2]string{"eth0:   ", "eth1:   "},
		[2]string{"eth0:0:   ", "eth1:0:   "},
		[2]string{"eth0:", "eth1:"},
		[2]string{"eth0:0:", "eth1:0:"},
	}
	for _, testCase := range cases {
		err = tmpfile.Truncate(0)
		assert.Nil(t, err, "Temporary file truncating problem: ", err)

		// Parse interface name for assertion
		interface0 := strings.TrimSpace(testCase[0])
		interface0 = interface0[:len(interface0)-1]

		interface1 := strings.TrimSpace(testCase[1])
		interface1 = interface1[:len(interface1)-1]

		// Replace the interfaces from the test case
		proc := []byte(fmt.Sprintf("Inter-|   Receive                                                |  Transmit\n face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n  %s1       2    3    4    5     6          7         8        9       10    11    12    13     14       15          16\n    %s100 200    300   400    500     600          700         800 900 1000    1100    1200    1300    1400       1500          1600\n", testCase[0], testCase[1]))

		// Write /proc/net/dev sample output
		_, err = tmpfile.Write(proc)
		assert.Nil(t, err, "Temporary file writing failed: ", err)

		counters, err := IOCountersByFile(true, tmpfile.Name())

		assert.Nil(t, err)
		assert.NotEmpty(t, counters)
		assert.Equal(t, 2, len(counters))
		assert.Equal(t, interface0, counters[0].Name)
		assert.Equal(t, 1, int(counters[0].BytesRecv))
		assert.Equal(t, 2, int(counters[0].PacketsRecv))
		assert.Equal(t, 3, int(counters[0].Errin))
		assert.Equal(t, 4, int(counters[0].Dropin))
		assert.Equal(t, 5, int(counters[0].Fifoin))
		assert.Equal(t, 9, int(counters[0].BytesSent))
		assert.Equal(t, 10, int(counters[0].PacketsSent))
		assert.Equal(t, 11, int(counters[0].Errout))
		assert.Equal(t, 12, int(counters[0].Dropout))
		assert.Equal(t, 13, int(counters[0].Fifoout))
		assert.Equal(t, interface1, counters[1].Name)
		assert.Equal(t, 100, int(counters[1].BytesRecv))
		assert.Equal(t, 200, int(counters[1].PacketsRecv))
		assert.Equal(t, 300, int(counters[1].Errin))
		assert.Equal(t, 400, int(counters[1].Dropin))
		assert.Equal(t, 500, int(counters[1].Fifoin))
		assert.Equal(t, 900, int(counters[1].BytesSent))
		assert.Equal(t, 1000, int(counters[1].PacketsSent))
		assert.Equal(t, 1100, int(counters[1].Errout))
		assert.Equal(t, 1200, int(counters[1].Dropout))
		assert.Equal(t, 1300, int(counters[1].Fifoout))
	}

	err = tmpfile.Close()
	assert.Nil(t, err, "Temporary file closing failed: ", err)
}

func TestGetProcInodesAll(t *testing.T) {
	if os.Getenv("CIRCLECI") == "true" {
		t.Skip("Skip CI")
	}

	root := common.HostProc("")
	v, err := getProcInodesAll(root, 0)
	assert.Nil(t, err)
	assert.NotEmpty(t, v)
}

func TestConnectionsMax(t *testing.T) {
	if os.Getenv("CIRCLECI") == "true" {
		t.Skip("Skip CI")
	}

	max := 10
	v, err := ConnectionsMax("tcp", max)
	assert.Nil(t, err)
	assert.NotEmpty(t, v)

	cxByPid := map[int32]int{}
	for _, cx := range v {
		if cx.Pid > 0 {
			cxByPid[cx.Pid]++
		}
	}
	for _, c := range cxByPid {
		assert.True(t, c <= max)
	}
}

type AddrTest struct {
	IP    string
	Port  int
	Error bool
}

func TestDecodeAddress(t *testing.T) {
	assert := assert.New(t)

	addr := map[string]AddrTest{
		"0500000A:0016": {
			IP:   "10.0.0.5",
			Port: 22,
		},
		"0100007F:D1C2": {
			IP:   "127.0.0.1",
			Port: 53698,
		},
		"11111:0035": {
			Error: true,
		},
		"0100007F:BLAH": {
			Error: true,
		},
		"0085002452100113070057A13F025401:0035": {
			IP:   "2400:8500:1301:1052:a157:7:154:23f",
			Port: 53,
		},
		"00855210011307F025401:0035": {
			Error: true,
		},
	}

	for src, dst := range addr {
		family := syscall.AF_INET
		if len(src) > 13 {
			family = syscall.AF_INET6
		}
		addr, err := decodeAddress(uint32(family), src)
		if dst.Error {
			assert.NotNil(err, src)
		} else {
			assert.Nil(err, src)
			assert.Equal(dst.IP, addr.IP, src)
			assert.Equal(dst.Port, int(addr.Port), src)
		}
	}
}

func TestReverse(t *testing.T) {
	src := []byte{0x01, 0x02, 0x03}
	assert.Equal(t, []byte{0x03, 0x02, 0x01}, Reverse(src))
}
