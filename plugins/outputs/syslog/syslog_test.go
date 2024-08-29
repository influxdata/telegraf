package syslog

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
	"github.com/leodido/go-syslog/v4/nontransparent"
)

func TestGetSyslogMessageWithFramingOctetCounting(t *testing.T) {
	// Init plugin
	s := newSyslog()
	require.NoError(t, s.Init())
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
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

	require.Equal(t, "59 <13>1 2010-11-10T23:00:00Z testhost Telegraf - testmetric -", string(messageBytesWithFraming), "Incorrect Octet counting framing")
}

func TestGetSyslogMessageWithFramingNonTransparent(t *testing.T) {
	// Init plugin
	s := newSyslog()
	require.NoError(t, s.Init())
	s.initializeSyslogMapper()
	s.Framing = "non-transparent"

	// Init metrics
	m1 := metric.New(
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

	require.Equal(t, "<13>1 2010-11-10T23:00:00Z testhost Telegraf - testmetric -\n", string(messageBytesWithFraming), "Incorrect Octet counting framing")
}

func TestGetSyslogMessageWithFramingNonTransparentNul(t *testing.T) {
	// Init plugin
	s := newSyslog()
	require.NoError(t, s.Init())
	s.initializeSyslogMapper()
	s.Framing = "non-transparent"
	s.Trailer = nontransparent.NUL

	// Init metrics
	m1 := metric.New(
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

	require.Equal(t, "<13>1 2010-11-10T23:00:00Z testhost Telegraf - testmetric -\x00", string(messageBytesWithFraming), "Incorrect Octet counting framing")
}

func TestSyslogWriteWithTcp(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := newSyslog()
	require.NoError(t, s.Init())
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
	require.NoError(t, s.Init())
	s.Address = "udp://" + listener.LocalAddr().String()

	err = s.Connect()
	require.NoError(t, err)

	testSyslogWriteWithPacket(t, s, listener)
}

func testSyslogWriteWithStream(t *testing.T, s *Syslog, lconn net.Conn) {
	metrics := []telegraf.Metric{}
	m1 := metric.New(
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
	require.Equal(t, string(messageBytesWithFraming), string(buf[:n]))
}

func testSyslogWriteWithPacket(t *testing.T, s *Syslog, lconn net.PacketConn) {
	s.Framing = "non-transparent"
	metrics := []telegraf.Metric{}
	m1 := metric.New(
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
	require.Equal(t, string(messageBytesWithFraming), string(buf[:n]))
}

func TestSyslogWriteErr(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := newSyslog()
	require.NoError(t, s.Init())
	s.Address = "tcp://" + listener.Addr().String()

	err = s.Connect()
	require.NoError(t, err)
	err = s.Conn.(*net.TCPConn).SetReadBuffer(256)
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	err = lconn.(*net.TCPConn).SetWriteBuffer(256)
	require.NoError(t, err)

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}

	// close the socket to generate an error
	err = lconn.Close()
	require.NoError(t, err)

	err = s.Conn.Close()
	require.NoError(t, err)

	err = s.Write(metrics)
	require.Error(t, err)
	require.Nil(t, s.Conn)
}

func TestSyslogWriteReconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := newSyslog()
	require.NoError(t, s.Init())
	s.Address = "tcp://" + listener.Addr().String()

	err = s.Connect()
	require.NoError(t, err)
	err = s.Conn.(*net.TCPConn).SetReadBuffer(256)
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	err = lconn.(*net.TCPConn).SetWriteBuffer(256)
	require.NoError(t, err)
	err = lconn.Close()
	require.NoError(t, err)
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
	require.NoError(t, lerr)

	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(metrics[0])
	require.NoError(t, err)
	messageBytesWithFraming, err := s.getSyslogMessageBytesWithFraming(syslogMessage)
	require.NoError(t, err)
	buf := make([]byte, 256)
	n, err := lconn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, string(messageBytesWithFraming), string(buf[:n]))
}

func TestStartupErrorBehaviorDefault(t *testing.T) {
	// Setup a dummy listener but do not accept connections
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	listener.Close()

	// Setup the plugin and the model to be able to use the startup retry strategy
	plugin := &Syslog{
		Address:             "tcp://" + address,
		Trailer:             nontransparent.LF,
		Separator:           "_",
		DefaultSeverityCode: uint8(5), // notice
		DefaultFacilityCode: uint8(1), // user-level
		DefaultAppname:      "Telegraf",
	}

	model := models.NewRunningOutput(
		plugin,
		&models.OutputConfig{
			Name: "syslog",
		},
		10, 100,
	)
	require.NoError(t, model.Init())

	// Starting the plugin will fail with an error because the server does not listen
	err = model.Connect()
	require.Error(t, err, "connection should be refused")
	var serr *internal.StartupError
	require.ErrorAs(t, err, &serr)
}

func TestStartupErrorBehaviorError(t *testing.T) {
	// Setup a dummy listener but do not accept connections
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	listener.Close()

	// Setup the plugin and the model to be able to use the startup retry strategy
	plugin := &Syslog{
		Address:             "tcp://" + address,
		Trailer:             nontransparent.LF,
		Separator:           "_",
		DefaultSeverityCode: uint8(5), // notice
		DefaultFacilityCode: uint8(1), // user-level
		DefaultAppname:      "Telegraf",
	}

	model := models.NewRunningOutput(
		plugin,
		&models.OutputConfig{
			Name:                 "syslog",
			StartupErrorBehavior: "error",
		},
		10, 100,
	)
	require.NoError(t, model.Init())

	// Starting the plugin will fail with an error because the server does not listen
	err = model.Connect()
	require.Error(t, err, "connection should be refused")
	var serr *internal.StartupError
	require.ErrorAs(t, err, &serr)
}

func TestStartupErrorBehaviorIgnore(t *testing.T) {
	// Setup a dummy listener but do not accept connections
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	listener.Close()

	// Setup the plugin and the model to be able to use the startup retry strategy
	plugin := &Syslog{
		Address:             "tcp://" + address,
		Trailer:             nontransparent.LF,
		Separator:           "_",
		DefaultSeverityCode: uint8(5), // notice
		DefaultFacilityCode: uint8(1), // user-level
		DefaultAppname:      "Telegraf",
	}

	model := models.NewRunningOutput(
		plugin,
		&models.OutputConfig{
			Name:                 "syslog",
			StartupErrorBehavior: "ignore",
		},
		10, 100,
	)
	require.NoError(t, model.Init())

	// Starting the plugin will fail because the server does not accept connections.
	// The model code should convert it to a fatal error for the agent to remove
	// the plugin.
	err = model.Connect()
	require.Error(t, err, "connection should be refused")
	var fatalErr *internal.FatalError
	require.ErrorAs(t, err, &fatalErr)
}

func TestStartupErrorBehaviorRetry(t *testing.T) {
	// Setup a dummy listener but do not accept connections
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	listener.Close()

	// Setup the plugin and the model to be able to use the startup retry strategy
	plugin := &Syslog{
		Address:             "tcp://" + address,
		Trailer:             nontransparent.LF,
		Separator:           "_",
		DefaultSeverityCode: uint8(5), // notice
		DefaultFacilityCode: uint8(1), // user-level
		DefaultAppname:      "Telegraf",
	}

	model := models.NewRunningOutput(
		plugin,
		&models.OutputConfig{
			Name:                 "syslog",
			StartupErrorBehavior: "retry",
		},
		10, 100,
	)
	require.NoError(t, model.Init())

	// Starting the plugin will return no error because the plugin will
	// retry to connect in every write cycle.
	require.NoError(t, model.Connect())
	defer model.Close()

	// Writing metrics in this state should fail because we are not fully
	// started up
	metrics := testutil.MockMetrics()
	for _, m := range metrics {
		model.AddMetric(m)
	}
	require.ErrorIs(t, model.WriteBatch(), internal.ErrNotConnected)

	// Startup an actually working listener we can connect and write to
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	var wg sync.WaitGroup
	buf := make([]byte, 256)

	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, err := listener.Accept()
		if err != nil {
			t.Logf("accepting connection failed: %v", err)
			t.Fail()
			return
		}

		if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			t.Logf("setting read deadline failed: %v", err)
			t.Fail()
			return
		}

		if _, err := conn.Read(buf); err != nil {
			t.Logf("reading failed: %v", err)
			t.Fail()
		}
	}()

	// Update the plugin's address and write again. This time the write should
	// succeed.
	plugin.Address = "tcp://" + listener.Addr().String()
	require.NoError(t, model.WriteBatch())
	wg.Wait()
	require.NotEmpty(t, string(buf))
}
