package syslog

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	framing "github.com/influxdata/telegraf/internal/syslog"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSyslogMessageWithFramingOctectCounting(t *testing.T) {
	// Init plugin
	s := newSyslog()
	s.initializeSyslogMapper()

	// Init metrics
	m1, _ := metric.New(
		"testmetric",
		map[string]string{
			"hostname": "testhost",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	messageBytesWithFraming, err := s.getSyslogMessageBytesWithFraming(syslogMessage)
	require.NoError(t, err)

	assert.Equal(t, "59 <13>1 2010-11-10T23:00:00Z testhost Telegraf - testmetric -", string(messageBytesWithFraming), "Incorrect Octect counting framing")
}

func TestGetSyslogMessageWithFramingNonTransparent(t *testing.T) {
	// Init plugin
	s := newSyslog()
	s.initializeSyslogMapper()
	s.Framing = framing.NonTransparent

	// Init metrics
	m1, _ := metric.New(
		"testmetric",
		map[string]string{
			"hostname": "testhost",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	messageBytesWithFraming, err := s.getSyslogMessageBytesWithFraming(syslogMessage)
	require.NoError(t, err)

	assert.Equal(t, "<13>1 2010-11-10T23:00:00Z testhost Telegraf - testmetric -\x00", string(messageBytesWithFraming), "Incorrect Octect counting framing")
}

func TestSyslogWriteWithTcp(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := newSyslog()
	s.Address = "tcp://" + listener.Addr().String()

	err = s.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)

	testSyslogWriteWithStream(t, s, lconn)
}

func TestSyslogWriteWithUdp(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	s := newSyslog()
	s.Address = "udp://" + listener.LocalAddr().String()

	err = s.Connect()
	require.NoError(t, err)

	testSyslogWriteWithPacket(t, s, listener)
}

func testSyslogWriteWithStream(t *testing.T, s *Syslog, lconn net.Conn) {
	metrics := []telegraf.Metric{}
	m1, _ := metric.New(
		"testmetric",
		map[string]string{},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC))

	metrics = append(metrics, m1)
	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(metrics[0])
	require.NoError(t, err)
	messageBytesWithFraming, err := s.getSyslogMessageBytesWithFraming(syslogMessage)
	require.NoError(t, err)

	err = s.Write(metrics)
	require.NoError(t, err)

	buf := make([]byte, 256)
	n, err := lconn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, string(messageBytesWithFraming), string(buf[:n]))
}

func testSyslogWriteWithPacket(t *testing.T, s *Syslog, lconn net.PacketConn) {
	s.Framing = framing.NonTransparent
	metrics := []telegraf.Metric{}
	m1, _ := metric.New(
		"testmetric",
		map[string]string{},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC))

	metrics = append(metrics, m1)
	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(metrics[0])
	require.NoError(t, err)
	messageBytesWithFraming, err := s.getSyslogMessageBytesWithFraming(syslogMessage)
	require.NoError(t, err)

	err = s.Write(metrics)
	require.NoError(t, err)

	buf := make([]byte, 256)
	n, _, err := lconn.ReadFrom(buf)
	require.NoError(t, err)
	assert.Equal(t, string(messageBytesWithFraming), string(buf[:n]))
}

func TestSyslogWriteErr(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := newSyslog()
	s.Address = "tcp://" + listener.Addr().String()

	err = s.Connect()
	require.NoError(t, err)
	s.Conn.(*net.TCPConn).SetReadBuffer(256)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	lconn.(*net.TCPConn).SetWriteBuffer(256)

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}

	// close the socket to generate an error
	lconn.Close()
	s.Conn.Close()
	err = s.Write(metrics)
	require.Error(t, err)
	assert.Nil(t, s.Conn)
}

func TestSyslogWriteReconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := newSyslog()
	s.Address = "tcp://" + listener.Addr().String()

	err = s.Connect()
	require.NoError(t, err)
	s.Conn.(*net.TCPConn).SetReadBuffer(256)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	lconn.(*net.TCPConn).SetWriteBuffer(256)
	lconn.Close()
	s.Conn = nil

	wg := sync.WaitGroup{}
	wg.Add(1)
	var lerr error
	go func() {
		lconn, lerr = listener.Accept()
		wg.Done()
	}()

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}
	err = s.Write(metrics)
	require.NoError(t, err)

	wg.Wait()
	assert.NoError(t, lerr)

	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(metrics[0])
	require.NoError(t, err)
	messageBytesWithFraming, err := s.getSyslogMessageBytesWithFraming(syslogMessage)
	require.NoError(t, err)
	buf := make([]byte, 256)
	n, err := lconn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, string(messageBytesWithFraming), string(buf[:n]))
}
