package chrony

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"testing"
	"time"

	fbchrony "github.com/facebook/time/ntp/chrony"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	// Setup a mock server
	server := Server{
		TrackingInfo: &fbchrony.Tracking{
			RefID:              0xA29FC87B,
			IPAddr:             net.ParseIP("192.168.1.22"),
			Stratum:            3,
			LeapStatus:         3,
			RefTime:            time.Now(),
			CurrentCorrection:  0.000020390,
			LastOffset:         0.000012651,
			RMSOffset:          0.000025577,
			FreqPPM:            -16.001,
			ResidFreqPPM:       0.0,
			SkewPPM:            0.006,
			RootDelay:          0.001655,
			RootDispersion:     0.003307,
			LastUpdateInterval: 507.2,
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin
	plugin := &Chrony{
		Server: "udp://" + addr,
		Log:    testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Start the plugin, do a gather and stop everything
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()
	require.NoError(t, plugin.Gather(&acc))
	plugin.Stop()
	server.Shutdown()

	// Do the comparison
	expected := []telegraf.Metric{
		metric.New(
			"chrony",
			map[string]string{
				"reference_id": "A29FC87B",
				"leap_status":  "not synchronized",
				"stratum":      "3",
			},
			map[string]interface{}{
				"system_time":     0.000020390,
				"last_offset":     0.000012651,
				"rms_offset":      0.000025577,
				"frequency":       -16.001,
				"residual_freq":   0.0,
				"skew":            0.006,
				"root_delay":      0.001655,
				"root_dispersion": 0.003307,
				"update_interval": 507.2,
			},
			time.Unix(0, 0),
		),
	}

	options := []cmp.Option{
		// tests on linux with go1.20 will add a warning about code coverage, ignore that tag
		testutil.IgnoreTags("warning"),
		testutil.IgnoreTime(),
		cmpopts.EquateApprox(0.001, 0),
	}

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, options...)
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container
	container := testutil.Container{
		Image:        "dockurr/chrony",
		ExposedPorts: []string{"323/udp"},
		Files: map[string]string{
			"/etc/telegraf-chrony.conf": "testdata/chrony.conf",
			"/start.sh":                 "testdata/start.sh",
		},
		Entrypoint: []string{"/start.sh"},
		WaitingFor: wait.ForLog("Selected source"),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	plugin := &Chrony{
		Server: "udp://" + container.Address + ":" + container.Ports["323"],
		Log:    testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()
	require.NoError(t, plugin.Gather(&acc))

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"chrony",
			map[string]string{
				"leap_status":  "normal",
				"reference_id": "A29FC87B",
				"stratum":      "4",
			},
			map[string]interface{}{
				"frequency":       float64(0),
				"last_offset":     float64(0),
				"residual_freq":   float64(0),
				"rms_offset":      float64(0),
				"root_delay":      float64(0),
				"root_dispersion": float64(0),
				"skew":            float64(0),
				"system_time":     float64(0),
				"update_interval": float64(0),
			},
			time.Unix(0, 0),
		),
	}

	options := []cmp.Option{
		testutil.IgnoreTags("leap_status", "reference_id", "stratum"),
		testutil.IgnoreTime(),
	}

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, options...)
}

type Server struct {
	TrackingInfo *fbchrony.Tracking

	conn net.PacketConn
}

func (s *Server) Shutdown() {
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *Server) Listen(t *testing.T) (string, error) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	s.conn = conn
	addr := s.conn.LocalAddr().String()

	go s.serve(t)

	return addr, nil
}

func (s *Server) serve(t *testing.T) {
	defer s.conn.Close()

	for {
		buf := make([]byte, 4096)
		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			return
		}
		t.Logf("mock server: received %d bytes from %q\n", n, addr.String())

		var header fbchrony.RequestHead
		data := bytes.NewBuffer(buf)
		if err := binary.Read(data, binary.BigEndian, &header); err != nil {
			t.Logf("mock server: reading request header failed: %v", err)
			return
		}
		seqno := header.Sequence + 1

		switch header.Command {
		case 33: // tracking
			_, err := s.conn.WriteTo(s.encodeTrackingReply(seqno), addr)
			if err != nil {
				t.Logf("mock server [tracking]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [tracking]: successfully wrote reply")
			}
		default:
			t.Logf("mock server: unhandled command %v", header.Command)
		}
	}
}

func (s *Server) encodeTrackingReply(sequence uint32) []byte {
	t := s.TrackingInfo

	// Encode the header
	buf := []byte{
		0x06,       // version 6
		0x02,       // packet type 2: tracking
		0x00,       // res1
		0x00,       // res2
		0x00, 0x21, // command 33: tracking request
		0x00, 0x05, // reply 5: tracking reply
		0x00, 0x00, // status 0: success
		0x00, 0x00, // pad1
		0x00, 0x00, // pad2
		0x00, 0x00, // pad3
	}
	buf = binary.BigEndian.AppendUint32(buf, sequence)                // sequence number
	buf = append(buf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00) // pad 4 & 5

	// Encode data
	buf = binary.BigEndian.AppendUint32(buf, t.RefID)
	buf = append(buf, t.IPAddr.To16()...)
	if len(t.IPAddr) == 4 {
		buf = append(buf, 0x00, 0x01) // IPv4 address family
	} else {
		buf = append(buf, 0x00, 0x02) // IPv6 address family
	}
	buf = append(buf, 0x00, 0x00) // padding
	buf = binary.BigEndian.AppendUint16(buf, t.Stratum)
	buf = binary.BigEndian.AppendUint16(buf, t.LeapStatus)
	sec := uint64(t.RefTime.Unix())
	nsec := uint32(t.RefTime.UnixNano() % t.RefTime.Unix() * int64(time.Second))
	buf = binary.BigEndian.AppendUint32(buf, uint32(sec>>32))        // seconds high part
	buf = binary.BigEndian.AppendUint32(buf, uint32(sec&0xffffffff)) // seconds low part
	buf = binary.BigEndian.AppendUint32(buf, nsec)                   // nanoseconds
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.CurrentCorrection))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.LastOffset))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.RMSOffset))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.FreqPPM))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.ResidFreqPPM))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.SkewPPM))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.RootDelay))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.RootDispersion))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(t.LastUpdateInterval))

	return buf
}

// Modified based on https://github.com/mlichvar/chrony/blob/master/util.c
const (
	floatExpBits   = int32(7)
	floatCoeffBits = int32(25) // 32 - floatExpBits
	floatExpMin    = int32(-(1 << (floatExpBits - 1)))
	floatExpMax    = -floatExpMin - 1
	floatCoefMin   = int32(-(1 << (floatCoeffBits - 1)))
	floatCoefMax   = -floatCoefMin - 1
)

func encodeFloat(x float64) uint32 {
	var neg int32

	if math.IsNaN(x) {
		/* Save NaN as zero */
		x = 0.0
	} else if x < 0.0 {
		x = -x
		neg = 1
	}

	var exp, coef int32
	if x > 1.0e100 {
		exp = floatExpMax
		coef = floatCoefMax + neg
	} else if x > 1.0e-100 {
		exp = int32(math.Log2(x)) + 1
		coef = int32(x*math.Pow(2.0, float64(-exp+floatCoeffBits)) + 0.5)

		if coef <= 0 {
			panic(fmt.Errorf("invalid coefficient %v for value %f", coef, x))
		}

		/* we may need to shift up to two bits down */
		for coef > floatCoefMax+neg {
			coef >>= 1
			exp++
		}

		if exp > floatExpMax {
			/* overflow */
			exp = floatExpMax
			coef = floatCoefMax + neg
		} else if exp < floatExpMin {
			/* underflow */
			if exp+floatCoeffBits >= floatExpMin {
				coef >>= floatExpMin - exp
				exp = floatExpMin
			} else {
				exp = 0
				coef = 0
			}
		}
	}

	/* negate back */
	if neg != 0 {
		coef = int32(uint32(-coef) % (1 << floatCoeffBits))
	}

	return uint32(exp<<floatCoeffBits) | uint32(coef)
}
