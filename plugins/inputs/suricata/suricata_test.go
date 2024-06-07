package suricata

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var ex2 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":` +
	`{"capture":{"kernel_packets":905344474,"kernel_drops":78355440,"kernel_packets_delta":2376742,"kernel_drops_delta":82049}}}`
var ex3 = `{"timestamp":"2017-03-06T07:43:39.000397+0000","event_type":"stats","stats":` +
	`{"threads": { "W#05-wlp4s0": { "capture":{"kernel_packets":905344474,"kernel_drops":78355440}}}}}`

func TestSuricataLarge(t *testing.T) {
	dir := t.TempDir()
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Alerts:    true,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	data, err := os.ReadFile("testdata/test1.json")
	require.NoError(t, err)

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write(data)
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)

	//test suricata alerts
	data2, err := os.ReadFile("testdata/test2.json")
	require.NoError(t, err)
	_, err = c.Write(data2)
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.Wait(1)
}

func TestSuricataAlerts(t *testing.T) {
	dir := t.TempDir()
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Alerts:    true,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	data, err := os.ReadFile("testdata/test3.json")
	require.NoError(t, err)

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write(data)
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.Wait(1)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"suricata_alert",
			map[string]string{},
			map[string]interface{}{
				"action":       "allowed",
				"category":     "Misc activity",
				"gid":          float64(1),
				"rev":          float64(0),
				"signature":    "Corrupted HTTP body",
				"signature_id": float64(6),
				"severity":     float64(3),
				"source.ip":    "10.0.0.5",
				"target.ip":    "179.60.192.3",
				"source.port":  float64(18715),
				"target.port":  float64(80),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestSuricata(t *testing.T) {
	dir := t.TempDir()
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(ex2))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.Wait(1)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"suricata",
			map[string]string{
				"thread": "total",
			},
			map[string]interface{}{
				"capture.kernel_packets":       float64(905344474),
				"capture.kernel_drops":         float64(78355440),
				"capture.kernel_packets_delta": float64(2376742),
				"capture.kernel_drops_delta":   float64(82049),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestThreadStats(t *testing.T) {
	dir := t.TempDir()
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source:    tmpfn,
		Delimiter: ".",
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(""))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	_, err = c.Write([]byte("foobard}\n"))
	require.NoError(t, err)
	_, err = c.Write([]byte(ex3))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())
	acc.Wait(2)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"suricata",
			map[string]string{
				"thread": "W#05-wlp4s0",
			},
			map[string]interface{}{
				"capture.kernel_packets": float64(905344474),
				"capture.kernel_drops":   float64(78355440),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestSuricataInvalid(t *testing.T) {
	dir := t.TempDir()
	tmpfn := filepath.Join(dir, fmt.Sprintf("t%d", rand.Int63()))

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	acc.SetDebug(true)

	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte("sfjiowef"))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.WaitError(1)
}

func TestSuricataInvalidPath(t *testing.T) {
	tmpfn := fmt.Sprintf("/t%d/X", rand.Int63())
	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}

	acc := testutil.Accumulator{}
	require.Error(t, s.Start(&acc))
}

func TestSuricataTooLongLine(t *testing.T) {
	tmpfn := testutil.TempSocket(t)

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}

	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(strings.Repeat("X", 20000000)))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.WaitError(1)
}

func TestSuricataEmptyJSON(t *testing.T) {
	tmpfn := testutil.TempSocket(t)

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	if err != nil {
		log.Println(err)
	}
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.WaitError(1)
}

func TestSuricataDisconnectSocket(t *testing.T) {
	tmpfn := testutil.TempSocket(t)

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}

	require.NoError(t, s.Start(&acc))
	defer s.Stop()

	c, err := net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(ex2))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	c, err = net.Dial("unix", tmpfn)
	require.NoError(t, err)
	_, err = c.Write([]byte(ex3))
	require.NoError(t, err)
	_, err = c.Write([]byte("\n"))
	require.NoError(t, err)
	require.NoError(t, c.Close())

	acc.Wait(2)
}

func TestSuricataStartStop(t *testing.T) {
	tmpfn := testutil.TempSocket(t)

	s := Suricata{
		Source: tmpfn,
		Log: testutil.Logger{
			Name: "inputs.suricata",
		},
	}
	acc := testutil.Accumulator{}
	require.NoError(t, s.Start(&acc))
	s.Stop()
}

func TestSuricataParse(t *testing.T) {
	tests := []struct {
		filename string
		expected []telegraf.Metric
	}{{
		filename: "test2.json",
		expected: []telegraf.Metric{
			testutil.MustMetric(
				"suricata",
				map[string]string{
					"thread": "W#01-ens2f1",
				},
				map[string]interface{}{
					"detect_alert":                float64(0),
					"detect_engines_id":           float64(0),
					"detect_engines_last_reload":  "2021-06-08T06:33:05.084872+0000",
					"detect_engines_rules_failed": float64(0),
					"detect_engines_rules_loaded": float64(22712),
				},
				time.Unix(0, 0),
			),
		},
	},
	}

	for _, tc := range tests {
		data, err := os.ReadFile("testdata/" + tc.filename)
		require.NoError(t, err)

		s := Suricata{
			Delimiter: "_",
		}
		acc := testutil.Accumulator{}
		err = s.parse(&acc, data)
		require.NoError(t, err)

		testutil.RequireMetricsEqual(t, tc.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
	}
}

func TestSuricataParseVersion2(t *testing.T) {
	tests := []struct {
		filename string
		expected []telegraf.Metric
	}{
		{
			filename: "v2/alert.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"suricata",
					map[string]string{
						"event_type": "alert",
						"in_iface":   "s1-suricata",
						"proto":      "TCP",
					},
					map[string]interface{}{
						"action":       "allowed",
						"category":     "Misc activity",
						"dest_ip":      "179.60.192.3",
						"dest_port":    int64(80),
						"gid":          float64(1),
						"rev":          float64(0),
						"severity":     float64(3),
						"signature":    "Corrupted HTTP body",
						"signature_id": float64(6),
						"sourceip":     "10.0.0.5",
						"sourceport":   float64(18715),
						"src_ip":       "10.0.0.5",
						"src_port":     int64(18715),
						"targetip":     "179.60.192.3",
						"targetport":   float64(80),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			filename: "v2/dns.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"suricata",
					map[string]string{
						"event_type": "dns",
						"in_iface":   "eth1",
						"proto":      "UDP",
					},
					map[string]interface{}{
						"dest_ip":   "192.168.0.1",
						"dest_port": int64(53),
						"id":        float64(7145),
						"rrname":    "reddit.com",
						"rrtype":    "A",
						"src_ip":    "192.168.0.100",
						"type":      "query",
						"src_port":  int64(39262),
						"tx_id":     float64(10),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			filename: "v2/drop.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"suricata",
					map[string]string{
						"event_type": "drop",
						"in_iface":   "eth1",
						"proto":      "TCP",
					},
					map[string]interface{}{
						"dest_ip":   "54.192.18.125",
						"dest_port": int64(443),
						"ipid":      float64(62316),
						"len":       float64(76),
						"reason":    "stream error",
						"src_ip":    "192.168.0.110",
						"src_port":  int64(46016),
						"tcpack":    float64(2339873683),
						"tcpres":    float64(0),
						"tcpseq":    float64(3900248957),
						"tcpurgp":   float64(0),
						"tcpwin":    float64(501),
						"tos":       float64(0),
						"ttl":       float64(64),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			filename: "v2/flow.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"suricata",
					map[string]string{
						"event_type": "flow",
						"in_iface":   "eth1",
						"proto":      "TCP",
					},
					map[string]interface{}{
						"age":       float64(0),
						"dest_ip":   "142.251.130.3",
						"dest_port": int64(443),
						"src_ip":    "192.168.0.121",
						"src_port":  int64(50212),
						"state":     "new",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			filename: "v2/http.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"suricata",
					map[string]string{
						"event_type": "http",
						"in_iface":   "eth2",
						"proto":      "TCP",
					},
					map[string]interface{}{
						"dest_ip":           "203.205.239.179",
						"dest_port":         int64(80),
						"hostname":          "hkminorshort.weixin.qq.com",
						"http_content_type": "application/octet-stream",
						"http_method":       "POST",
						"http_user_agent":   "MicroMessenger Client",
						"length":            float64(245),
						"protocol":          "HTTP/1.1",
						"src_ip":            "192.168.0.120",
						"src_port":          int64(33950),
						"status":            float64(200),
						"url":               "/mmtls/2d6d45f1",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			filename: "v2/status.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"suricata",
					map[string]string{
						"event_type": "stats",
					},
					map[string]interface{}{
						"captureerrors":          float64(0),
						"capturekernel_drops":    float64(0),
						"capturekernel_packets":  float64(522),
						"flowemerg_mode_entered": float64(0),
						"flowemerg_mode_over":    float64(0),
						"flowmemcap":             float64(0),
						"flowmemuse":             float64(9965056),
						"flowmgrclosed_pruned":   float64(0),
						"flowmgrfull_hash_pass":  float64(1),
						"flowmgrnew_pruned":      float64(0),
						"flowspare":              float64(10100),
						"flowtcp":                float64(15),
						"flowudp":                float64(13),
						"flowwrkspare_sync":      float64(11),
						"flowwrkspare_sync_avg":  float64(100),
						"uptime":                 float64(160),
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tc := range tests {
		data, err := os.ReadFile("testdata/" + tc.filename)
		require.NoError(t, err)

		s := Suricata{
			Version: "2",
			Log:     testutil.Logger{},
		}
		acc := testutil.Accumulator{}
		require.NoError(t, s.parse(&acc, data))

		testutil.RequireMetricsEqual(t, tc.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
	}
}
