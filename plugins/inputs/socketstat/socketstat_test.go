package socketstat

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
        "github.com/stretchr/testify/require"
)

func TestSocketstat_Gather(t *testing.T) {
	tests := []struct {
		name   string
		proto  []string
		value  string
		tags   []map[string]string
		fields [][]map[string]interface{}
		err    error
	}{
		{
			name:  "tcp - no sockets => no results",
			proto: []string{"tcp"},
			value: `State      Recv-Q Send-Q       Local Address:Port                      Peer Address:Port`,
		},
		{
			name:  "udp - no sockets => no results",
			proto: []string{"udp"},
			value: `Recv-Q Send-Q            Local Address:Port                           Peer Address:Port`,
		},
		{
			name:  "tcp sockets captured",
			proto: []string{"tcp"},
			value: `State      Recv-Q Send-Q       Local Address:Port                      Peer Address:Port
ESTAB      0      0             192.168.1.21:6514                      192.168.1.21:443
	cubic wscale:7,7 rto:204 rtt:0.057/0.033 ato:40 mss:22976 cwnd:10 bytes_acked:1126 bytes_received:532644751 segs_out:211249 segs_in:211254 data_segs_out:2 data_segs_in:211251 send 32247.0Mbps lastsnd:299082764 lastrcv:5248 lastack:5252 rcv_rtt:3.532 rcv_space:186557 minrtt:0.047
ESTAB      0      0            192.168.122.1:55194                    192.168.122.1:6514
	cubic wscale:7,7 rto:204 rtt:0.034/0.01 ato:40 mss:65483 cwnd:10 bytes_acked:790782896 bytes_received:1126 segs_out:333361 segs_in:333361 data_segs_out:333358 data_segs_in:2 send 154077.6Mbps lastsnd:5248 lastrcv:443892492 lastack:5248 rcv_rtt:250 rcv_space:43690 minrtt:0.009
ESTAB      0      0                127.0.0.1:7778                         127.0.0.1:50378
	cubic wscale:7,7 rto:220 rtt:16.009/21.064 ato:44 mss:65483 cwnd:10 bytes_acked:19983121 bytes_received:266383 segs_out:15431 segs_in:17633 data_segs_out:15119 data_segs_in:5098 send 327.2Mbps lastsnd:9792 lastrcv:9840 lastack:9748 pacing_rate 654.4Mbps retrans:0/1 rcv_rtt:129800 rcv_space:44057 minrtt:0.043`,
			tags: []map[string]string{
				map[string]string{"proto": "tcp", "local_addr": "192.168.1.21", "local_port": "6514", "remote_addr": "192.168.1.21", "remote_port": "443"},
				map[string]string{"proto": "tcp", "local_addr": "192.168.122.1", "local_port": "55194", "remote_addr": "192.168.122.1", "remote_port": "6514"},
				map[string]string{"proto": "tcp", "local_addr": "127.0.0.1", "local_port": "7778", "remote_addr": "127.0.0.1", "remote_port": "50378"},
			},
			fields: [][]map[string]interface{}{
				{map[string]interface{}{"state": "ESTAB", "bytes_acked": uint64(1126), "bytes_received": uint64(532644751), "segs_out": uint64(211249), "segs_in": uint64(211254), "data_segs_out": uint64(2), "data_segs_in": uint64(211251), "recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"state": "ESTAB", "bytes_acked": uint64(790782896), "bytes_received": uint64(1126), "segs_out": uint64(333361), "segs_in": uint64(333361), "data_segs_out": uint64(333358), "data_segs_in": uint64(2), "recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"state": "ESTAB", "bytes_acked": uint64(19983121), "bytes_received": uint64(266383), "segs_out": uint64(15431), "segs_in": uint64(17633), "data_segs_out": uint64(15119), "data_segs_in": uint64(5098), "recv_q": uint64(0), "send_q": uint64(0)}},
			},
		},
		{
			name:  "udp packets captured",
			proto: []string{"udp"},
			value: `Recv-Q Send-Q                   Local Address:Port                                  Peer Address:Port
0      0                         10.10.0.4:33149                                 10.10.0.5:53

0      0                         10.10.0.4:54276                                 10.10.0.6:53

0      0                         10.10.0.4:38312                                 10.10.0.7:53`,
			tags: []map[string]string{
				map[string]string{"proto": "udp", "local_addr": "10.10.0.4", "local_port": "33149", "remote_addr": "10.10.0.5", "remote_port": "53"},
				map[string]string{"proto": "udp", "local_addr": "10.10.0.4", "local_port": "54276", "remote_addr": "10.10.0.6", "remote_port": "53"},
				map[string]string{"proto": "udp", "local_addr": "10.10.0.4", "local_port": "38312", "remote_addr": "10.10.0.7", "remote_port": "53"},
			},
			fields: [][]map[string]interface{}{
				{map[string]interface{}{"recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"recv_q": uint64(0), "send_q": uint64(0)}},
				{map[string]interface{}{"recv_q": uint64(0), "send_q": uint64(0)}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i++
			ss := &Socketstat{
				SocketProto: tt.proto,
				lister: func(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
					return bytes.NewBufferString(tt.value), nil
				},
			}
			acc := new(testutil.Accumulator)

                        err := ss.Init()
                        require.NoError(t, err)

			err = acc.GatherError(ss.Gather)
			if !reflect.DeepEqual(tt.err, err) {
				t.Errorf("%d: expected error '%#v' got '%#v'", i, tt.err, err)
			}
			if len(tt.proto) == 0 {
				n := acc.NFields()
				if n != 0 {
					t.Errorf("%d: expected 0 fields if no protocol specified got %d", i, n)
				}
				return
			}
			if len(tt.tags) == 0 {
				n := acc.NFields()
				if n != 0 {
					t.Errorf("%d: expected 0 values got %d", i, n)
				}
				return
			}
			n := 0
			for j, tags := range tt.tags {
				for k, fields := range tt.fields[j] {
					if len(acc.Metrics) < n+1 {
						t.Errorf("%d: expected at least %d values got %d", i, n+1, len(acc.Metrics))
						break
					}
					m := acc.Metrics[n]
					if !reflect.DeepEqual(m.Measurement, ss.measurement) {
						t.Errorf("%d %d %d: expected measurement '%#v' got '%#v'\n", i, j, k, ss.measurement, m.Measurement)
					}
					if !reflect.DeepEqual(m.Tags, tags) {
						t.Errorf("%d %d %d: expected tags\n%#v got\n%#v\n", i, j, k, tags, m.Tags)
					}
					if !reflect.DeepEqual(m.Fields, fields) {
						t.Errorf("%d %d %d: expected fields\n%#v got\n%#v\n", i, j, k, fields, m.Fields)
					}
					n++
				}
			}
		})
	}
}

func TestSocketstat_Gather_listerError(t *testing.T) {
	errFoo := errors.New("error foobar")
	ss := &Socketstat{
		SocketProto: []string{"foobar"},
		lister: func(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
			return new(bytes.Buffer), errFoo
		},
	}
	acc := new(testutil.Accumulator)
	err := acc.GatherError(ss.Gather)
	if !reflect.DeepEqual(err, errFoo) {
		t.Errorf("Expected error %#v got\n%#v\n", errFoo, err)
	}
}
