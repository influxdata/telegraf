package v2

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertGlobalStatus(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       sql.RawBytes
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "default",
			key:         "ssl_ctx_verify_depth",
			value:       []byte("0"),
			expected:    uint64(0),
			expectedErr: nil,
		},
		{
			name:        "overflow int64",
			key:         "ssl_ctx_verify_depth",
			value:       []byte("18446744073709551615"),
			expected:    uint64(18446744073709551615),
			expectedErr: nil,
		},
		{
			name:        "defined variable but unset",
			key:         "ssl_ctx_verify_depth",
			value:       []byte(""),
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:  "multiple values in one metric converted to a map",
			key:   "wsrep_evs_repl_latency",
			value: []byte("0.000160108/0.000386178/0.00964884/0.000488261/816"),
			expected: map[string]interface{}{
				"min":         0.000160108,
				"avg":         0.000386178,
				"max":         0.00964884,
				"stdev":       0.000488261,
				"sample_size": 816.0,
			},
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ConvertGlobalStatus(tt.key, tt.value)
			require.Equal(t, tt.expectedErr, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestConvertGlobalVariables(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       sql.RawBytes
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "boolean type mysql<=5.6",
			key:         "gtid_mode",
			value:       []byte("ON"),
			expected:    int64(1),
			expectedErr: nil,
		},
		{
			name:        "enum type mysql>=5.7",
			key:         "gtid_mode",
			value:       []byte("ON_PERMISSIVE"),
			expected:    int64(1),
			expectedErr: nil,
		},
		{
			name:        "defined variable but unset",
			key:         "ssl_ctx_verify_depth",
			value:       []byte(""),
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "multiple values in one metric converted to a map",
			key:  "wsrep_provider_options",
			value: []byte(
				"allocator.disk_pages_encryption = no; allocator.encryption_cache_page_size = 32K; allocator.encryption_cache_size = 16777216; " +
					"base_dir = /var/lib/mysql; base_host = 192.168.1.1; base_port = 4567; cert.log_conflicts = no; cert.optimistic_pa = NO; debug = no; " +
					"evs.auto_evict = 0; evs.causal_keepalive_period = PT1S; evs.debug_log_mask = 0x1; evs.delay_margin = PT1S; " +
					"evs.delayed_keep_period = PT30S; evs.inactive_check_period = PT0.5S; evs.inactive_timeout = PT15S; evs.info_log_mask = 0; " +
					"evs.install_timeout = PT7.5S; evs.join_retrans_period = PT1S; evs.keepalive_period = PT1S; evs.max_install_timeouts = 3; " +
					"evs.send_window = 10; evs.stats_report_period = PT1M; evs.suspect_timeout = PT5S; evs.use_aggregate = true; evs.user_send_window = 4; " +
					"evs.version = 1; evs.view_forget_timeout = P1D; gcache.dir = /var/lib/mysql; gcache.encryption = no; " +
					"gcache.encryption_cache_page_size = 32K; gcache.encryption_cache_size = 16777216; gcache.freeze_purge_at_seqno = -1; " +
					"gcache.keep_pages_count = 0; gcache.keep_pages_size = 0; gcache.mem_size = 0; gcache.name = galera.cache; gcache.page_size = 128M; " +
					"gcache.recover = yes; gcache.size = 128M; gcomm.thread_prio = ; gcs.fc_auto_evict_threshold = 0.75; gcs.fc_auto_evict_window = 0; " +
					"gcs.fc_debug = 0; gcs.fc_factor = 1.0; gcs.fc_limit = 100; gcs.fc_master_slave = no; gcs.fc_single_primary = no; " +
					"gcs.max_packet_size = 64500; gcs.max_throttle = 0.25; gcs.recv_q_hard_limit = 9223372036854775807; gcs.recv_q_soft_limit = 0.25; " +
					"gcs.sync_donor = no; gmcast.listen_addr = tcp://0.0.0.0:4567; gmcast.mcast_addr = ; gmcast.mcast_ttl = 1; gmcast.peer_timeout = PT3S; " +
					"gmcast.segment = 0; gmcast.time_wait = PT5S; gmcast.version = 0; ist.recv_addr = 192.168.1.1; pc.announce_timeout = PT3S; " +
					"pc.checksum = false; pc.ignore_quorum = false; pc.ignore_sb = false; pc.linger = PT20S; pc.npvo = false; pc.recovery = true; " +
					"pc.version = 0; pc.wait_prim = true; pc.wait_prim_timeout = PT30S; pc.wait_restored_prim_timeout = PT0S; pc.weight = 1; " +
					"protonet.backend = asio; protonet.version = 0; repl.causal_read_timeout = PT30S; repl.commit_order = 3; repl.key_format = FLAT8; " +
					"repl.max_ws_size = 2147483647; repl.proto_max = 11; socket.checksum = 2; socket.recv_buf_size = auto; socket.send_buf_size = auto; ",
			),
			expected: map[string]interface{}{
				"gcache_size": uint64(134217728),
			},
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ConvertGlobalVariables(tt.key, tt.value)
			require.Equal(t, tt.expectedErr, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestParseValue(t *testing.T) {
	testCases := []struct {
		rawByte sql.RawBytes
		output  interface{}
		err     string
	}{
		{sql.RawBytes("123"), int64(123), ""},
		{sql.RawBytes("abc"), "abc", ""},
		{sql.RawBytes("10.1"), 10.1, ""},
		{sql.RawBytes("ON"), int64(1), ""},
		{sql.RawBytes("OFF"), int64(0), ""},
		{sql.RawBytes("NO"), int64(0), ""},
		{sql.RawBytes("YES"), int64(1), ""},
		{sql.RawBytes("No"), int64(0), ""},
		{sql.RawBytes("Yes"), int64(1), ""},
		{sql.RawBytes("-794"), int64(-794), ""},
		{sql.RawBytes("2147483647"), int64(2147483647), ""},                       // max int32
		{sql.RawBytes("2147483648"), int64(2147483648), ""},                       // too big for int32
		{sql.RawBytes("9223372036854775807"), int64(9223372036854775807), ""},     // max int64
		{sql.RawBytes("9223372036854775808"), uint64(9223372036854775808), ""},    // too big for int64
		{sql.RawBytes("18446744073709551615"), uint64(18446744073709551615), ""},  // max uint64
		{sql.RawBytes("18446744073709551616"), float64(18446744073709552000), ""}, // too big for uint64
		{sql.RawBytes("18446744073709552333"), float64(18446744073709552000), ""}, // too big for uint64
		{sql.RawBytes(""), nil, "unconvertible value"},
	}
	for _, cases := range testCases {
		got, err := ParseValue(cases.rawByte)

		if err != nil && cases.err == "" {
			t.Errorf("for %q got unexpected error: %q", string(cases.rawByte), err.Error())
		} else if err != nil && !strings.HasPrefix(err.Error(), cases.err) {
			t.Errorf("for %q wanted error %q, got %q", string(cases.rawByte), cases.err, err.Error())
		} else if err == nil && cases.err != "" {
			t.Errorf("for %q did not get expected error: %s", string(cases.rawByte), cases.err)
		} else if got != cases.output {
			t.Errorf("for %q wanted %#v (%T), got %#v (%T)", string(cases.rawByte), cases.output, cases.output, got, got)
		}
	}
}
