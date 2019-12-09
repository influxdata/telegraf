package sflow

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"testing"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/wlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEmptyLog is a helper function to ensure no data is written to log.
// Should be called at the start of the test, and returns a function which should run at the end.
func testEmptyLog(t *testing.T) func() {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(wlog.NewWriter(buf))

	level := wlog.WARN
	wlog.SetLevel(level)

	return func() {
		log.SetOutput(os.Stderr)

		for {
			line, err := buf.ReadBytes('\n')
			if err != nil {
				assert.Equal(t, io.EOF, err)
				break
			}
			assert.Empty(t, string(line), "log not empty")
		}
	}
}

func TestNetflowDescription(t *testing.T) {
	sl := newListener()
	assert.NotEmpty(t, sl.Description())
}

func TestNetflowSampleConfig(t *testing.T) {
	sl := newListener()
	assert.NotEmpty(t, sl.SampleConfig())
}

func TestNetflowGather(t *testing.T) {
	sl := newListener()
	assert.Nil(t, sl.Gather(nil))
}

func TestNetflowToMetrics(t *testing.T) {
	defer testEmptyLog(t)()

	sl := newListener()
	sl.ServiceAddress = "udp://127.0.0.1:0"
	sl.ReadBufferSize = internal.Size{Size: 1024}
	sl.DNSFQDNResolve = false

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	client, err := net.Dial("udp", sl.Closer.(net.PacketConn).LocalAddr().String())
	require.NoError(t, err)

	template257And258 := []byte("00090004000071d45dc583690000000000000041000000840101000f00010004000200040004000100050001000600010007000200080004000a0004000b0002000c0004000e0004001000040011000400150004001600040102000f000100040002000400040001000500010006000100070002000a0004000b0002000e000400100004001100040015000400160004001b0010001c00100001001801030004000800010004002a000400290004000001030010000000000000000100000000")
	dataAgainst257And258 := []byte("00090004000071d45dc583690000000100000041010100340000004800000001110000e115ac10ec0100000000e115ac10ecff000000000000000000000000000000000000000004")
	expected := "[netflow map[agentAddress:127.0.0.1 bgpDestinationAsNumber:0 bgpSourceAsNumber:0 destinationIPv4Address:172.16.236.255 destinationTransportPort:57621 destinationTransportSvc:57621 egressInterface:0 ingressInterface:0 ipClassOfService:0 protocolIdentifier:17 sourceID:65 sourceIPv4Address:172.16.236.1 sourceTransportPort:57621 sourceTransportSvc:57621 tcpControlBits:0] map[flowEndSysUpTime:0 flowStartSysUpTime:0 octetDeltaCount:72 packetDeltaCount:1]]"

	packetBytes := make([]byte, hex.DecodedLen(len(template257And258)))
	_, err = hex.Decode(packetBytes, template257And258)
	client.Write(packetBytes)

	packetBytes = make([]byte, hex.DecodedLen(len(dataAgainst257And258)))
	_, err = hex.Decode(packetBytes, dataAgainst257And258)
	client.Write(packetBytes)

	acc.Wait(1)
	acc.Lock()
	actual := fmt.Sprintf(("%s"), acc.Metrics)
	acc.Unlock()

	assert.Equal(t, expected, actual)
}
