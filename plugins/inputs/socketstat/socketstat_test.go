//go:build !windows
// +build !windows

package socketstat

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSocketstat_Gather(t *testing.T) {
	tests := []struct {
		name     string
		proto    []string
		filename string
		tags     []map[string]string
		fields   [][]map[string]interface{}
		err      error
	}{
		{
			name:     "tcp - no sockets => no results",
			proto:    []string{"tcp"},
			filename: "tcp_no_sockets.txt",
		},
		{
			name:     "udp - no sockets => no results",
			proto:    []string{"udp"},
			filename: "udp_no_sockets.txt",
		},
		{
			name:     "tcp sockets captured",
			proto:    []string{"tcp"},
			filename: "tcp_traffic.txt",
			tags: []map[string]string{
				{"proto": "tcp", "local_addr": "192.168.1.21", "local_port": "6514", "remote_addr": "192.168.1.21", "remote_port": "443"},
				{"proto": "tcp", "local_addr": "192.168.122.1", "local_port": "55194", "remote_addr": "192.168.122.1", "remote_port": "6514"},
				{"proto": "tcp", "local_addr": "127.0.0.1", "local_port": "7778", "remote_addr": "127.0.0.1", "remote_port": "50378"},
			},
			fields: [][]map[string]interface{}{
				{map[string]interface{}{"state": "ESTAB", "bytes_acked": uint64(1126), "bytes_received": uint64(532644751), "segs_out": uint64(211249), "segs_in": uint64(211254), "data_segs_out": uint64(2), "data_segs_in": uint64(211251), "recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"state": "ESTAB", "bytes_acked": uint64(790782896), "bytes_received": uint64(1126), "segs_out": uint64(333361), "segs_in": uint64(333361), "data_segs_out": uint64(333358), "data_segs_in": uint64(2), "recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"state": "ESTAB", "bytes_acked": uint64(19983121), "bytes_received": uint64(266383), "segs_out": uint64(15431), "segs_in": uint64(17633), "data_segs_out": uint64(15119), "data_segs_in": uint64(5098), "recv_q": uint64(0), "send_q": uint64(0)}},
			},
		},
		{
			name:     "udp packets captured",
			proto:    []string{"udp"},
			filename: "udp_traffic.txt",
			tags: []map[string]string{
				{"proto": "udp", "local_addr": "10.10.0.4", "local_port": "33149", "remote_addr": "10.10.0.5", "remote_port": "53"},
				{"proto": "udp", "local_addr": "10.10.0.4", "local_port": "54276", "remote_addr": "10.10.0.6", "remote_port": "53"},
				{"proto": "udp", "local_addr": "10.10.0.4", "local_port": "38312", "remote_addr": "10.10.0.7", "remote_port": "53"},
			},
			fields: [][]map[string]interface{}{
				{map[string]interface{}{"recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"recv_q": uint64(0), "send_q": uint64(0)}},
			},
		},
	}
	for i, tt := range tests {
		octets, err := os.ReadFile(filepath.Join("testdata", tt.filename))
		require.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			i++
			ss := &Socketstat{
				SocketProto: tt.proto,
			}
			acc := new(testutil.Accumulator)

			err := ss.Init()
			if err != nil {
				require.EqualError(t, err, "exec: \"ss\": executable file not found in $PATH")
			}
			ss.lister = func(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
				return bytes.NewBuffer(octets), nil
			}

			err = acc.GatherError(ss.Gather)
			require.ErrorIs(t, err, tt.err)
			if len(tt.proto) == 0 {
				n := acc.NFields()
				require.Equalf(t, 0, n, "%d: expected 0 values got %d", i, n)
				return
			}
			if len(tt.tags) == 0 {
				n := acc.NFields()
				require.Equalf(t, 0, n, "%d: expected 0 values got %d", i, n)
				return
			}
			n := 0
			for j, tags := range tt.tags {
				for k, fields := range tt.fields[j] {
					require.Greater(t, len(acc.Metrics), n)
					m := acc.Metrics[n]
					require.Equal(t, measurement, m.Measurement, "%d %d %d: expected measurement '%#v' got '%#v'\n", i, j, k, measurement, m.Measurement)
					require.Equal(t, tags, m.Tags, "%d %d %d: expected tags\n%#v got\n%#v\n", i, j, k, tags, m.Tags)
					require.Equal(t, fields, m.Fields, "%d %d %d: expected fields\n%#v got\n%#v\n", i, j, k, fields, m.Fields)
					n++
				}
			}
		})
	}
}

func TestSocketstat_Gather_listerError(t *testing.T) {
	errorMessage := "error foobar"
	errFoo := errors.New(errorMessage)
	ss := &Socketstat{
		SocketProto: []string{"foobar"},
	}
	ss.lister = func(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
		return new(bytes.Buffer), errFoo
	}
	acc := new(testutil.Accumulator)
	err := acc.GatherError(ss.Gather)
	require.EqualError(t, err, errorMessage)
}
