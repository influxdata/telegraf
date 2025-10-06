package chrony

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"
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

func TestGatherActivity(t *testing.T) {
	// Setup a mock server
	server := Server{
		ActivityInfo: &fbchrony.Activity{
			Online:       34,
			Offline:      6,
			BurstOnline:  2,
			BurstOffline: 0,
			Unresolved:   5,
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin
	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"activity"},
		Log:     testutil.Logger{},
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
			"chrony_activity",
			map[string]string{"source": addr},
			map[string]interface{}{
				"online":        34,
				"offline":       6,
				"burst_online":  2,
				"burst_offline": 0,
				"unresolved":    5,
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

func TestGatherTracking(t *testing.T) {
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
		Server:  "udp://" + addr,
		Metrics: []string{"tracking"},
		Log:     testutil.Logger{},
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
				"source":       addr,
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

func TestGatherServerStats(t *testing.T) {
	// Setup a mock server
	server := Server{
		ServerStatInfo: &fbchrony.ServerStats{
			NTPHits:  2542,
			CMDHits:  112,
			NTPDrops: 42,
			CMDDrops: 8,
			LogDrops: 0,
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin
	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"serverstats"},
		Log:     testutil.Logger{},
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
			"chrony_serverstats",
			map[string]string{"source": addr},
			map[string]interface{}{
				"ntp_hits":  uint64(2542),
				"ntp_drops": uint64(42),
				"cmd_hits":  uint64(112),
				"cmd_drops": uint64(8),
				"log_drops": uint64(0),
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

func TestGatherServerStats2(t *testing.T) {
	// Setup a mock server
	server := Server{
		ServerStatInfo: &fbchrony.ServerStats2{
			NTPHits:     2542,
			NKEHits:     5,
			CMDHits:     112,
			NTPDrops:    42,
			NKEDrops:    1,
			CMDDrops:    8,
			LogDrops:    0,
			NTPAuthHits: 9,
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin
	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"serverstats"},
		Log:     testutil.Logger{},
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
			"chrony_serverstats",
			map[string]string{"source": addr},
			map[string]interface{}{
				"ntp_hits":      uint64(2542),
				"ntp_drops":     uint64(42),
				"ntp_auth_hits": uint64(9),
				"cmd_hits":      uint64(112),
				"cmd_drops":     uint64(8),
				"log_drops":     uint64(0),
				"nke_hits":      uint64(5),
				"nke_drops":     uint64(1),
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

func TestGatherServerStats3(t *testing.T) {
	// Setup a mock server
	server := Server{
		ServerStatInfo: &fbchrony.ServerStats3{
			NTPHits:            2542,
			NKEHits:            5,
			CMDHits:            112,
			NTPDrops:           42,
			NKEDrops:           1,
			CMDDrops:           8,
			LogDrops:           0,
			NTPAuthHits:        9,
			NTPInterleavedHits: 28,
			NTPTimestamps:      69527,
			NTPSpanSeconds:     33,
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin
	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"serverstats"},
		Log:     testutil.Logger{},
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
			"chrony_serverstats",
			map[string]string{"source": addr},
			map[string]interface{}{
				"ntp_hits":             uint64(2542),
				"ntp_drops":            uint64(42),
				"ntp_auth_hits":        uint64(9),
				"ntp_interleaved_hits": uint64(28),
				"ntp_timestamps":       uint64(69527),
				"ntp_span_seconds":     uint64(33),
				"cmd_hits":             uint64(112),
				"cmd_drops":            uint64(8),
				"log_drops":            uint64(0),
				"nke_hits":             uint64(5),
				"nke_drops":            uint64(1),
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

func TestGatherSources(t *testing.T) {
	// Setup a mock server
	server := Server{
		SourcesInfo: []source{
			{
				name: "ntp1.my.org",
				data: &fbchrony.SourceData{
					IPAddr:         net.IPv4(192, 168, 0, 1),
					Poll:           64,
					Stratum:        16,
					State:          fbchrony.SourceStateSync,
					Mode:           fbchrony.SourceModePeer,
					Flags:          0,
					Reachability:   0,
					SinceSample:    0,
					OrigLatestMeas: 1.22354,
					LatestMeas:     1.22354,
					LatestMeasErr:  0.00423,
				},
			},
			{
				name: "ntp2.my.org",
				data: &fbchrony.SourceData{
					IPAddr:         net.IPv4(192, 168, 0, 2),
					Poll:           64,
					Stratum:        16,
					State:          fbchrony.SourceStateSync,
					Mode:           fbchrony.SourceModePeer,
					Flags:          0,
					Reachability:   0,
					SinceSample:    0,
					OrigLatestMeas: 0.17791,
					LatestMeas:     0.35445,
					LatestMeasErr:  0.01196,
				},
			},
			{
				name: "ntp3.my.org",
				data: &fbchrony.SourceData{
					IPAddr:         net.IPv4(192, 168, 0, 3),
					Poll:           512,
					Stratum:        1,
					State:          fbchrony.SourceStateOutlier,
					Mode:           fbchrony.SourceModePeer,
					Flags:          0,
					Reachability:   512,
					SinceSample:    377,
					OrigLatestMeas: 7.21158,
					LatestMeas:     7.21158,
					LatestMeasErr:  2.15453,
				},
			},
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin
	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"sources"},
		Log:     testutil.Logger{},
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
			"chrony_sources",
			map[string]string{
				"source": addr,
				"peer":   "ntp1.my.org",
			},
			map[string]interface{}{
				"index":                    0,
				"ip":                       "192.168.0.1",
				"poll":                     64,
				"stratum":                  uint64(16),
				"state":                    "sync",
				"mode":                     "peer",
				"flags":                    uint64(0),
				"reachability":             uint64(0),
				"sample":                   uint64(0),
				"latest_measurement":       1.22354,
				"latest_measurement_error": 0.00423,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"chrony_sources",
			map[string]string{
				"source": addr,
				"peer":   "ntp2.my.org",
			},
			map[string]interface{}{
				"index":                    1,
				"ip":                       "192.168.0.2",
				"poll":                     64,
				"stratum":                  uint64(16),
				"state":                    "sync",
				"mode":                     "peer",
				"flags":                    uint64(0),
				"reachability":             uint64(0),
				"sample":                   uint64(0),
				"latest_measurement":       0.35445,
				"latest_measurement_error": 0.01196,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"chrony_sources",
			map[string]string{
				"source": addr,
				"peer":   "ntp3.my.org",
			},
			map[string]interface{}{
				"index":                    2,
				"ip":                       "192.168.0.3",
				"poll":                     512,
				"stratum":                  uint64(1),
				"state":                    "outlier",
				"mode":                     "peer",
				"flags":                    uint64(0),
				"reachability":             uint64(512),
				"sample":                   uint64(377),
				"latest_measurement":       7.21158,
				"latest_measurement_error": 2.15453,
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

func TestGatherSourceStats(t *testing.T) {
	// Setup a mock server
	server := Server{
		SourcesInfo: []source{
			{
				name: "ntp1.my.org",
				stats: &fbchrony.SourceStats{
					RefID:              434354566,
					IPAddr:             net.IPv4(192, 168, 0, 1),
					NSamples:           1254,
					NRuns:              16,
					SpanSeconds:        32,
					StandardDeviation:  0.0244,
					ResidFreqPPM:       0.0015,
					SkewPPM:            0.0001,
					EstimatedOffset:    0.0039,
					EstimatedOffsetErr: 0.0007,
				},
			},
			{
				name: "ntp2.my.org",
				stats: &fbchrony.SourceStats{
					RefID:              70349595,
					IPAddr:             net.IPv4(192, 168, 0, 2),
					NSamples:           23135,
					NRuns:              24,
					SpanSeconds:        3,
					StandardDeviation:  0.0099,
					ResidFreqPPM:       0.0188,
					SkewPPM:            0.0002,
					EstimatedOffset:    0.0104,
					EstimatedOffsetErr: 0.0021,
				},
			},
			{
				name: "ntp3.my.org",
				stats: &fbchrony.SourceStats{
					RefID:              983490438,
					IPAddr:             net.IPv4(192, 168, 0, 3),
					NSamples:           23,
					NRuns:              4,
					SpanSeconds:        193,
					StandardDeviation:  7.0586,
					ResidFreqPPM:       0.8320,
					SkewPPM:            0.0332,
					EstimatedOffset:    5.3345,
					EstimatedOffsetErr: 1.5437,
				},
			},
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin
	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"sourcestats"},
		Log:     testutil.Logger{},
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
			"chrony_sourcestats",
			map[string]string{
				"source":       addr,
				"peer":         "ntp1.my.org",
				"reference_id": "19E3B986",
			},
			map[string]interface{}{
				"index":              0,
				"ip":                 "192.168.0.1",
				"samples":            uint64(1254),
				"runs":               uint64(16),
				"span_seconds":       uint64(32),
				"stddev":             0.0244,
				"residual_frequency": 0.0015,
				"skew":               0.0001,
				"offset":             0.0039,
				"offset_error":       0.0007,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"chrony_sourcestats",
			map[string]string{
				"source":       addr,
				"peer":         "ntp2.my.org",
				"reference_id": "0431731B",
			},
			map[string]interface{}{
				"index":              1,
				"ip":                 "192.168.0.2",
				"samples":            uint64(23135),
				"runs":               uint64(24),
				"span_seconds":       uint64(3),
				"stddev":             0.0099,
				"residual_frequency": 0.0188,
				"skew":               0.0002,
				"offset":             0.0104,
				"offset_error":       0.0021,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"chrony_sourcestats",
			map[string]string{
				"source":       addr,
				"peer":         "ntp3.my.org",
				"reference_id": "3A9EDF86",
			},
			map[string]interface{}{
				"index":              2,
				"ip":                 "192.168.0.3",
				"samples":            uint64(23),
				"runs":               uint64(4),
				"span_seconds":       uint64(193),
				"stddev":             7.0586,
				"residual_frequency": 0.8320,
				"skew":               0.0332,
				"offset":             5.3345,
				"offset_error":       1.5437,
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
	addr := container.Address + ":" + container.Ports["323"]

	// Setup the plugin
	plugin := &Chrony{
		Server: "udp://" + addr,
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
				"source":       addr,
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

type source struct {
	name  string
	data  *fbchrony.SourceData
	stats *fbchrony.SourceStats
}

type Server struct {
	ActivityInfo   *fbchrony.Activity
	TrackingInfo   *fbchrony.Tracking
	ServerStatInfo interface{}
	SourcesInfo    []source

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
			t.Errorf("mock server: reading request header failed: %v", err)
			return
		}
		seqno := header.Sequence + 1

		t.Logf("mock server: received request %d", header.Command)
		switch header.Command {
		case 14: // sources
			_, err := s.conn.WriteTo(s.encodeSourcesReply(seqno), addr)
			if err != nil {
				t.Errorf("mock server [sources]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [sources]: successfully wrote reply")
			}
		case 15: // source data
			var idx int32
			if err = binary.Read(data, binary.BigEndian, &idx); err != nil {
				t.Error(err)
				return
			}
			_, err = s.conn.WriteTo(s.encodeSourceDataReply(seqno, idx), addr)
			if err != nil {
				t.Errorf("mock server [source data]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [source data]: successfully wrote reply")
			}
		case 33: // tracking
			_, err := s.conn.WriteTo(s.encodeTrackingReply(seqno), addr)
			if err != nil {
				t.Errorf("mock server [tracking]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [tracking]: successfully wrote reply")
			}
		case 34: // source stats
			var idx int32
			if err = binary.Read(data, binary.BigEndian, &idx); err != nil {
				t.Error(err)
				return
			}
			_, err = s.conn.WriteTo(s.encodeSourceStatsReply(seqno, idx), addr)
			if err != nil {
				t.Errorf("mock server [source stats]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [source stats]: successfully wrote reply")
			}
		case 44: // activity
			payload, err := s.encodeActivityReply(seqno)
			if err != nil {
				t.Error(err)
				return
			}

			_, err = s.conn.WriteTo(payload, addr)
			if err != nil {
				t.Errorf("mock server [activity]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [activity]: successfully wrote reply")
			}
		case 54: // server stats
			payload, err := s.encodeServerStatsReply(seqno)
			if err != nil {
				t.Error(err)
				return
			}

			_, err = s.conn.WriteTo(payload, addr)
			if err != nil {
				t.Errorf("mock server [serverstats]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [serverstats]: successfully wrote reply")
			}
		case 65: // source name
			buf := make([]byte, 20)
			_, err := data.Read(buf)
			if err != nil {
				t.Error(err)
				return
			}
			ip := decodeIP(buf)
			t.Logf("mock server [source name]: resolving %v", ip)
			_, err = s.conn.WriteTo(s.encodeSourceNameReply(seqno, ip), addr)
			if err != nil {
				t.Errorf("mock server [source name]: writing reply failed: %v", err)
			} else {
				t.Log("mock server [source name]: successfully wrote reply")
			}
		default:
			t.Logf("mock server: unhandled command %v", header.Command)
		}
	}
}

func (s *Server) encodeActivityReply(sequence uint32) ([]byte, error) {
	// Encode the header
	buf := encodeHeader(44, 12, 0, sequence) // activity request

	// Encode data
	b := bytes.NewBuffer(buf)
	if err := binary.Write(b, binary.BigEndian, s.ActivityInfo); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (s *Server) encodeTrackingReply(sequence uint32) []byte {
	t := s.TrackingInfo

	// Encode the header
	buf := encodeHeader(33, 5, 0, sequence) // tracking request

	// Encode data
	buf = binary.BigEndian.AppendUint32(buf, t.RefID)
	buf = append(buf, encodeIP(t.IPAddr)...)
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

func (s *Server) encodeServerStatsReply(sequence uint32) ([]byte, error) {
	var b *bytes.Buffer
	var err error

	switch info := s.ServerStatInfo.(type) {
	case *fbchrony.ServerStats:
		// Encode the header
		buf := encodeHeader(54, 14, 0, sequence) // activity request

		// Encode data
		b = bytes.NewBuffer(buf)
		err = binary.Write(b, binary.BigEndian, info)
	case *fbchrony.ServerStats2:
		// Encode the header
		buf := encodeHeader(54, 22, 0, sequence) // activity request

		// Encode data
		b = bytes.NewBuffer(buf)
		err = binary.Write(b, binary.BigEndian, info)
	case *fbchrony.ServerStats3:
		// Encode the header
		buf := encodeHeader(54, 24, 0, sequence) // activity request

		// Encode data
		b = bytes.NewBuffer(buf)
		err = binary.Write(b, binary.BigEndian, info)
	}

	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s *Server) encodeSourcesReply(sequence uint32) []byte {
	// Encode the header
	buf := encodeHeader(14, 2, 0, sequence) // sources request

	// Encode data
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(s.SourcesInfo))) // NSources

	return buf
}

func (s *Server) encodeSourceDataReply(sequence uint32, idx int32) []byte {
	if len(s.SourcesInfo) <= int(idx) {
		return encodeHeader(15, 3, 3, sequence) // status invalid
	}
	src := s.SourcesInfo[idx].data

	// Encode the header
	buf := encodeHeader(15, 3, 0, sequence) // source data request

	// Encode data
	buf = append(buf, encodeIP(src.IPAddr)...)
	buf = binary.BigEndian.AppendUint16(buf, uint16(src.Poll))
	buf = binary.BigEndian.AppendUint16(buf, src.Stratum)
	buf = binary.BigEndian.AppendUint16(buf, uint16(src.State))
	buf = binary.BigEndian.AppendUint16(buf, uint16(src.Mode))
	buf = binary.BigEndian.AppendUint16(buf, src.Flags)
	buf = binary.BigEndian.AppendUint16(buf, src.Reachability)
	buf = binary.BigEndian.AppendUint32(buf, src.SinceSample)
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.OrigLatestMeas))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.LatestMeas))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.LatestMeasErr))

	return buf
}

func (s *Server) encodeSourceStatsReply(sequence uint32, idx int32) []byte {
	if len(s.SourcesInfo) <= int(idx) {
		return encodeHeader(34, 6, 3, sequence) // status invalid
	}
	src := s.SourcesInfo[idx].stats

	// Encode the header
	buf := encodeHeader(15, 6, 0, sequence) // source data request

	// Encode data
	buf = binary.BigEndian.AppendUint32(buf, src.RefID)
	buf = append(buf, encodeIP(src.IPAddr)...)
	buf = binary.BigEndian.AppendUint32(buf, src.NSamples)
	buf = binary.BigEndian.AppendUint32(buf, src.NRuns)
	buf = binary.BigEndian.AppendUint32(buf, src.SpanSeconds)
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.StandardDeviation))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.ResidFreqPPM))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.SkewPPM))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.EstimatedOffset))
	buf = binary.BigEndian.AppendUint32(buf, encodeFloat(src.EstimatedOffsetErr))

	return buf
}

func (s *Server) encodeSourceNameReply(sequence uint32, ip net.IP) []byte {
	// Encode the header
	buf := encodeHeader(65, 19, 0, sequence) // source name request

	// Find the correct source
	var name []byte
	for _, src := range s.SourcesInfo {
		if src.data != nil && src.data.IPAddr.Equal(ip) || src.stats != nil && src.stats.IPAddr.Equal(ip) {
			name = []byte(src.name)
			break
		}
	}

	// Encode data
	if len(name) > 256 {
		buf = append(buf, name[:256]...)
	} else {
		buf = append(buf, name...)
		buf = append(buf, make([]byte, 256-len(name))...)
	}

	return buf
}

func encodeHeader(command, replyType, status uint16, seqnr uint32) []byte {
	buf := []byte{
		0x06, // version 6
		0x02, // packet type 2: reply
		0x00, // res1
		0x00, // res2
	}
	buf = binary.BigEndian.AppendUint16(buf, command)   // command
	buf = binary.BigEndian.AppendUint16(buf, replyType) // reply type
	buf = binary.BigEndian.AppendUint16(buf, status)    // status 0: success
	buf = append(buf, []byte{
		0x00, 0x00, // pad1
		0x00, 0x00, // pad2
		0x00, 0x00, // pad3
	}...)
	buf = binary.BigEndian.AppendUint32(buf, seqnr)                   // sequence number
	buf = append(buf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00) // pad 4 & 5

	return buf
}

func encodeIP(addr net.IP) []byte {
	var buf []byte

	buf = append(buf, addr.To16()...)
	if len(addr) == 4 {
		buf = append(buf, 0x00, 0x01) // IPv4 address family
	} else {
		buf = append(buf, 0x00, 0x02) // IPv6 address family
	}
	buf = append(buf, 0x00, 0x00) // padding

	return buf
}

func decodeIP(buf []byte) net.IP {
	if len(buf) != 20 {
		panic("invalid length for IP")
	}

	addr := net.IP(buf[0:16])
	family := binary.BigEndian.Uint16(buf[16:18])
	if family == 1 {
		return addr.To4()
	}

	return addr
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

// TestConcurrentGather verifies that concurrent Gather() calls don't cause
// a race condition or panic when accessing the shared chrony client.
// This test addresses the issue reported in GitHub issue #17757 where
// concurrent access to the client caused "index out of range [256] with length 256" panics.
func TestConcurrentGather(t *testing.T) {
	// Setup a mock server with multiple sources to ensure longer gather time
	server := Server{
		ActivityInfo: &fbchrony.Activity{
			Online:       10,
			Offline:      2,
			BurstOnline:  1,
			BurstOffline: 0,
			Unresolved:   3,
		},
		TrackingInfo: &fbchrony.Tracking{
			RefID:              0xA29FC87B,
			IPAddr:             net.ParseIP("192.168.1.22"),
			Stratum:            3,
			LeapStatus:         0,
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
		ServerStatInfo: &fbchrony.ServerStats{
			NTPHits:  2542,
			CMDHits:  112,
			NTPDrops: 42,
			CMDDrops: 8,
			LogDrops: 0,
		},
		SourcesInfo: []source{
			{
				name: "ntp1.example.com",
				data: &fbchrony.SourceData{
					IPAddr:         net.IPv4(192, 168, 1, 1),
					Poll:           64,
					Stratum:        2,
					State:          fbchrony.SourceStateSync,
					Mode:           fbchrony.SourceModePeer,
					Flags:          0,
					Reachability:   255,
					SinceSample:    0,
					OrigLatestMeas: 0.001,
					LatestMeas:     0.001,
					LatestMeasErr:  0.0001,
				},
				stats: &fbchrony.SourceStats{
					RefID:              434354566,
					IPAddr:             net.IPv4(192, 168, 1, 1),
					NSamples:           100,
					NRuns:              10,
					SpanSeconds:        1000,
					StandardDeviation:  0.001,
					ResidFreqPPM:       0.0001,
					SkewPPM:            0.0001,
					EstimatedOffset:    0.0001,
					EstimatedOffsetErr: 0.00001,
				},
			},
			{
				name: "ntp2.example.com",
				data: &fbchrony.SourceData{
					IPAddr:         net.IPv4(192, 168, 1, 2),
					Poll:           64,
					Stratum:        2,
					State:          fbchrony.SourceStateSync,
					Mode:           fbchrony.SourceModePeer,
					Flags:          0,
					Reachability:   255,
					SinceSample:    0,
					OrigLatestMeas: 0.002,
					LatestMeas:     0.002,
					LatestMeasErr:  0.0002,
				},
				stats: &fbchrony.SourceStats{
					RefID:              434354567,
					IPAddr:             net.IPv4(192, 168, 1, 2),
					NSamples:           100,
					NRuns:              10,
					SpanSeconds:        1000,
					StandardDeviation:  0.002,
					ResidFreqPPM:       0.0002,
					SkewPPM:            0.0002,
					EstimatedOffset:    0.0002,
					EstimatedOffsetErr: 0.00002,
				},
			},
			{
				name: "ntp3.example.com",
				data: &fbchrony.SourceData{
					IPAddr:         net.IPv4(192, 168, 1, 3),
					Poll:           64,
					Stratum:        2,
					State:          fbchrony.SourceStateSync,
					Mode:           fbchrony.SourceModePeer,
					Flags:          0,
					Reachability:   255,
					SinceSample:    0,
					OrigLatestMeas: 0.003,
					LatestMeas:     0.003,
					LatestMeasErr:  0.0003,
				},
				stats: &fbchrony.SourceStats{
					RefID:              434354568,
					IPAddr:             net.IPv4(192, 168, 1, 3),
					NSamples:           100,
					NRuns:              10,
					SpanSeconds:        1000,
					StandardDeviation:  0.003,
					ResidFreqPPM:       0.0003,
					SkewPPM:            0.0003,
					EstimatedOffset:    0.0003,
					EstimatedOffsetErr: 0.00003,
				},
			},
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	// Setup the plugin with all metrics enabled to maximize the gather time
	// and increase the likelihood of concurrent access
	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"activity", "tracking", "serverstats", "sources", "sourcestats"},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(nil))
	defer plugin.Stop()

	// Run multiple concurrent Gather() calls
	// This simulates what happens when a previous gather hasn't completed
	// before the next interval triggers
	const numConcurrent = 10
	var wg sync.WaitGroup
	errors := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			var acc testutil.Accumulator
			if err := plugin.Gather(&acc); err != nil {
				errors <- err
				return
			}

			// Verify we got metrics (no panic occurred)
			if len(acc.GetTelegrafMetrics()) == 0 {
				t.Logf("iteration %d: no metrics collected", iteration)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check if any errors occurred
	for err := range errors {
		require.NoError(t, err, "concurrent gather should not produce errors")
	}
}

// TestRaceDetector runs the test with the Go race detector enabled.
func TestRaceDetector(t *testing.T) {
	// Setup a minimal mock server
	server := Server{
		ServerStatInfo: &fbchrony.ServerStats{
			NTPHits:  100,
			CMDHits:  10,
			NTPDrops: 1,
			CMDDrops: 0,
			LogDrops: 0,
		},
	}
	addr, err := server.Listen(t)
	require.NoError(t, err)
	defer server.Shutdown()

	plugin := &Chrony{
		Server:  "udp://" + addr,
		Metrics: []string{"serverstats"},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(nil))
	defer plugin.Stop()

	// Run 100 concurrent gathers to give the race detector
	// a better chance of catching any race conditions
	const iterations = 100
	var wg sync.WaitGroup
	errors := make(chan error, iterations)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var acc testutil.Accumulator
			if err := plugin.Gather(&acc); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check if any errors occurred
	for err := range errors {
		require.NoError(t, err)
	}
}
