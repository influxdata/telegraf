package spectrum_scale

import (
	"bufio"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsField(t *testing.T) {
	assert.Equal(t, isField("_event_"), true)
	assert.Equal(t, isField("_response_"), true)
	assert.Equal(t, isField("_nn_"), true)
	assert.Equal(t, isField("_t_"), true)
	assert.Equal(t, isField("_t node_"), true)
	assert.Equal(t, isField("_mmpmon::test::field_"), true)
	assert.Equal(t, isField("event_"), false)
	assert.Equal(t, isField(""), false)
}

func TestParseConfiguration(t *testing.T) {
	config := []byte(`
[[inputs.spectrum_scale]]
   sensors = [
     "nsd_ds",
     "gfis",
     "fis",
   ]
   socketLocation = "/var/mmfs/mmpmon/mmpmonSocket"
   `)

	table, err := toml.Parse([]byte(config))
	require.NoError(t, err)

	inputs, ok := table.Fields["inputs"]
	require.True(t, ok)

	spectrum_scale, ok := inputs.(*ast.Table).Fields["spectrum_scale"]
	require.True(t, ok)

	var plugin Spectrum_Scale

	require.NoError(t, toml.UnmarshalTable(spectrum_scale.([]*ast.Table)[0], &plugin))

	assert.Equal(t, Spectrum_Scale{
		Sensors: []string{
			"nsd_ds",
			"gfis",
			"fis",
		},
		SocketLocation: "/var/mmfs/mmpmon/mmpmonSocket",
	}, plugin)
}

const gfisExample = `_response_ begin mmpmon gfis
_mmpmon::gfis_ _n_ 10.10.10.1 _nn_ node1.example.com _rc_ 0 _t_ 1427920124 _tu_ 292780 _cl_ cluster.example.com _fs_ example _d_ 2 _br_ 0 _bw_ 0 _c_ 0 _r_ 0 _w_ 0 _oc_ 6427 _cc_ 6427 _rdc_ 0 _wc_ 0 _dir_ 6426 _iu_ 1456 _irc_ 1820 _idc_ 0 _icc_ 0 _bc_ 0 _sch_ 142442 _scm_ 15045
_response_ end
`

const nsddsExample = `_response_ begin mmpmon nsd_ds
_mmpmon::nsd_ds_ _n_ 10.10.10.1 _nn_ node1.example.com _rc_ 0 _t_ 1427920124 _tu_ 292780 _dev_ rg_node1_rg1 _d_ rg_node1_rg1
_r_ _ops_ 138448 _b_ 563132928 _tw_ 2219.600760 _qt_ 2219.714397 _stw_ 0.000002 _sqt_ 0.000002 _ltw_ 0.230000 _lqt_ 0.230000 _stwpb_ 0.000017 _sqtpb_ 0.000017 _ltwpb_ 0.000021 _lqtpb_ 0.000021
_w_ _ops_ 9406 _b_ 202620928 _tw_ 3.652243 _qt_ 3.652243 _stw_ 0.000145 _sqt_ 0.000145 _ltw_ 0.141563 _lqt_ 0.141563 _stwpb_ 0.000211 _sqtpb_ 0.000211 _ltwpb_ 0.000609 _lqtpb_ 0.000609
_response_ end
`

const fisExample = `_response_ begin mmpmon fis
_mmpmon::fis_ _n_ 10.10.10.1 _nn_ node1.example.com _rc_ 0 _t_ 1427920124 _tu_ 292780 _cl_ cluster.example.com _fs_ example _d_ 2 _br_ 0 _bw_ 0 _oc_ 6631 _cc_ 6631 _rdc_ 0 _wc_ 0 _dir_ 6630 _iu_ 1493
_response_ end
`

func mmpmonServer(t *testing.T, conn net.Conn) {

	buf := make([]byte, 1024)
	reader := bufio.NewReader(conn)

	for {
		amount, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		message := string(buf[:amount])

		if strings.Contains(message, "gfis") {
			conn.Write([]byte(gfisExample))
		} else if strings.Contains(message, "nsd_ds") {
			conn.Write([]byte(nsddsExample))
		} else if strings.Contains(message, "fis") {
			conn.Write([]byte(fisExample))
		}
	}
	conn.Close()

}

func socketHelper(t *testing.T, socketPath string) {

	// spin up socket
	err := os.RemoveAll(socketPath)
	require.NoError(t, err)

	l, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	defer l.Close()

	// Accept connections and respond until all tests complete
	for {
		conn, err := l.Accept()
		require.NoError(t, err)
		mmpmonServer(t, conn)
	}
}

func verifySensorGfis(t *testing.T, socketPath string) {

	gfis := &Spectrum_Scale{
		Sensors:        []string{"gfis"},
		SocketLocation: socketPath,
	}

	var acc testutil.Accumulator

	err := gfis.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"cluster":          "cluster.example.com",
		"daemon_node_name": "node1.example.com",
		"filesystem":       "example",
		"ipaddr":           "10.10.10.1",
		"disk":             "2",
		"sensor":           "gfis",
	}
	fields := map[string]interface{}{
		"close":             int(6427),
		"inode_create":      int(0),
		"inode_del":         int(0),
		"inode_read":        int(1820),
		"inode_update":      int(1456),
		"open":              int(6427),
		"rc":                int(0),
		"read_bytes_cache":  int(0),
		"read_bytes_disk":   int(0),
		"read_calls_cache":  int(0),
		"read_calls_disk":   int(0),
		"read_req":          int(0),
		"readdir":           int(6426),
		"stat_cache_hit":    int(142442),
		"stat_cache_miss":   int(15045),
		"time":              int(1427920124),
		"time_microseconds": int(292780),
		"write_bytes":       int(0),
		"write_calls":       int(0),
		"write_req":         int(0),
	}

	acc.AssertContainsTaggedFields(t, "spectrum_scale", fields, tags)
}

func verifySensorNsdds(t *testing.T, socketPath string) {

	gfis := &Spectrum_Scale{
		Sensors:        []string{"nsd_ds"},
		SocketLocation: socketPath,
	}

	var acc testutil.Accumulator

	err := gfis.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"daemon_node_name": "node1.example.com",
		"ipaddr":           "10.10.10.1",
		"device":           "rg_node1_rg1",
		"disk":             "rg_node1_rg1",
		"sensor":           "nsd_ds",
	}

	fields := map[string]interface{}{
		"longest_queued_time":  float64(0.141563),
		"longest_wait_time":    float64(0.141563),
		"lqtpb":                float64(0.000609),
		"ltwpb":                float64(0.000609),
		"rc":                   int(0),
		"read_bytes":           int(563132928),
		"read_calls":           int(138448),
		"shortest_queued_time": float64(0.000145),
		"shortest_wait_time":   float64(0.000145),
		"sqtpb":                float64(0.000211),
		"stwpb":                float64(0.000211),
		"time":                 int(1427920124),
		"time_microseconds":    int(292780),
		"total_queued_time":    float64(3.652243),
		"total_wait_time":      float64(3.652243),
		"write_bytes":          int(202620928),
		"write_calls":          int(9406),
	}

	acc.AssertContainsTaggedFields(t, "spectrum_scale", fields, tags)
}

func verifySensorFis(t *testing.T, socketPath string) {

	gfis := &Spectrum_Scale{
		Sensors:        []string{"fis"},
		SocketLocation: socketPath,
	}

	var acc testutil.Accumulator

	err := gfis.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"cluster":          "cluster.example.com",
		"daemon_node_name": "node1.example.com",
		"filesystem":       "example",
		"ipaddr":           "10.10.10.1",
		"disk":             "2",
		"sensor":           "fis",
	}
	fields := map[string]interface{}{
		"close":             int(6631),
		"inode_update":      int(1493),
		"open":              int(6631),
		"rc":                int(0),
		"read_bytes_disk":   int(0),
		"readdir":           int(6630),
		"time":              int(1427920124),
		"time_microseconds": int(292780),
		"write_bytes":       int(0),
		"write_req":         int(0),
		"read_req":          int(0),
	}

	acc.AssertContainsTaggedFields(t, "spectrum_scale", fields, tags)
}

func TestGather(t *testing.T) {

	// setup
	tempdir := os.TempDir() + "/telegraf/"
	socketName := "mmpmonSocket"

	err := os.MkdirAll(tempdir, 0755)
	require.NoError(t, err)

	socketPath := tempdir + socketName

	t.Logf("socketPath: %s", socketPath)

	go socketHelper(t, socketPath)

	// Initialize "server"
	time.Sleep(2 * time.Second)

	verifySensorGfis(t, socketPath)
	verifySensorNsdds(t, socketPath)
	verifySensorFis(t, socketPath)

	err = os.RemoveAll(os.TempDir() + "/telegraf")
	require.NoError(t, err)
}
