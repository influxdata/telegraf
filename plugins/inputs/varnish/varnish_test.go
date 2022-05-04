//go:build !windows
// +build !windows

package varnish

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func fakeVarnishRunner(output string) func(string, bool, []string, config.Duration) (*bytes.Buffer, error) {
	return func(string, bool, []string, config.Duration) (*bytes.Buffer, error) {
		return bytes.NewBuffer([]byte(output)), nil
	}
}

func TestGather(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Varnish{
		run:   fakeVarnishRunner(smOutput),
		Stats: []string{"*"},
	}
	require.NoError(t, v.Gather(acc))

	acc.HasMeasurement("varnish")
	for tag, fields := range parsedSmOutput {
		acc.AssertContainsTaggedFields(t, "varnish", fields, map[string]string{
			"section": tag,
		})
	}
}

func TestParseFullOutput(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Varnish{
		run:   fakeVarnishRunner(fullOutput),
		Stats: []string{"*"},
	}
	require.NoError(t, v.Gather(acc))

	acc.HasMeasurement("varnish")
	flat := flatten(acc.Metrics)
	require.Len(t, acc.Metrics, 6)
	require.Equal(t, 293, len(flat))
}

func TestFilterSomeStats(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Varnish{
		run:   fakeVarnishRunner(fullOutput),
		Stats: []string{"MGT.*", "VBE.*"},
	}
	require.NoError(t, v.Gather(acc))

	acc.HasMeasurement("varnish")
	flat := flatten(acc.Metrics)
	require.Len(t, acc.Metrics, 2)
	require.Equal(t, 16, len(flat))
}

func TestFieldConfig(t *testing.T) {
	expect := map[string]int{
		"*":                                     293,
		"":                                      0, // default
		"MAIN.uptime":                           1,
		"MEMPOOL.req0.sz_needed,MAIN.fetch_bad": 2,
	}

	for fieldCfg, expected := range expect {
		acc := &testutil.Accumulator{}
		v := &Varnish{
			run:   fakeVarnishRunner(fullOutput),
			Stats: strings.Split(fieldCfg, ","),
		}
		require.NoError(t, v.Gather(acc))

		acc.HasMeasurement("varnish")
		flat := flatten(acc.Metrics)
		require.Equal(t, expected, len(flat))
	}
}

func flatten(metrics []*testutil.Metric) map[string]interface{} {
	flat := map[string]interface{}{}
	for _, m := range metrics {
		buf := &bytes.Buffer{}
		for k, v := range m.Tags {
			_, err := buf.WriteString(fmt.Sprintf("%s=%s", k, v))
			if err != nil {
				return nil
			}
		}
		for k, v := range m.Fields {
			flat[fmt.Sprintf("%s %s", buf.String(), k)] = v
		}
	}
	return flat
}

var smOutput = `
MAIN.uptime                895         1.00 Child process uptime
MAIN.cache_hit                     95         0.00 Cache hits
MAIN.cache_miss                    5          0.00 Cache misses
MGT.uptime                         896         1.00 Management process uptime
MGT.child_start                    1         0.00 Child process started
MEMPOOL.vbc.live                   0          .   In use
MEMPOOL.vbc.pool                   10          .   In Pool
MEMPOOL.vbc.sz_wanted              88          .   Size requested
`

var parsedSmOutput = map[string]map[string]interface{}{
	"MAIN": {
		"uptime":     uint64(895),
		"cache_hit":  uint64(95),
		"cache_miss": uint64(5),
	},
	"MGT": {
		"uptime":      uint64(896),
		"child_start": uint64(1),
	},
	"MEMPOOL": {
		"vbc.live":      uint64(0),
		"vbc.pool":      uint64(10),
		"vbc.sz_wanted": uint64(88),
	},
}

var fullOutput = `
MAIN.uptime               2872         1.00 Child process uptime
MAIN.sess_conn               0         0.00 Sessions accepted
MAIN.sess_drop               0         0.00 Sessions dropped
MAIN.sess_fail               0         0.00 Session accept failures
MAIN.sess_pipe_overflow            0         0.00 Session pipe overflow
MAIN.client_req_400                0         0.00 Client requests received, subject to 400 errors
MAIN.client_req_411                0         0.00 Client requests received, subject to 411 errors
MAIN.client_req_413                0         0.00 Client requests received, subject to 413 errors
MAIN.client_req_417                0         0.00 Client requests received, subject to 417 errors
MAIN.client_req                    0         0.00 Good client requests received
MAIN.cache_hit                     0         0.00 Cache hits
MAIN.cache_hitpass                 0         0.00 Cache hits for pass
MAIN.cache_miss                    0         0.00 Cache misses
MAIN.backend_conn                  0         0.00 Backend conn. success
MAIN.backend_unhealthy             0         0.00 Backend conn. not attempted
MAIN.backend_busy                  0         0.00 Backend conn. too many
MAIN.backend_fail                  0         0.00 Backend conn. failures
MAIN.backend_reuse                 0         0.00 Backend conn. reuses
MAIN.backend_toolate               0         0.00 Backend conn. was closed
MAIN.backend_recycle               0         0.00 Backend conn. recycles
MAIN.backend_retry                 0         0.00 Backend conn. retry
MAIN.fetch_head                    0         0.00 Fetch no body (HEAD)
MAIN.fetch_length                  0         0.00 Fetch with Length
MAIN.fetch_chunked                 0         0.00 Fetch chunked
MAIN.fetch_eof                     0         0.00 Fetch EOF
MAIN.fetch_bad                     0         0.00 Fetch bad T-E
MAIN.fetch_close                   0         0.00 Fetch wanted close
MAIN.fetch_oldhttp                 0         0.00 Fetch pre HTTP/1.1 closed
MAIN.fetch_zero                    0         0.00 Fetch zero len body
MAIN.fetch_1xx                     0         0.00 Fetch no body (1xx)
MAIN.fetch_204                     0         0.00 Fetch no body (204)
MAIN.fetch_304                     0         0.00 Fetch no body (304)
MAIN.fetch_failed                  0         0.00 Fetch failed (all causes)
MAIN.fetch_no_thread               0         0.00 Fetch failed (no thread)
MAIN.pools                         2          .   Number of thread pools
MAIN.threads                     200          .   Total number of threads
MAIN.threads_limited               0         0.00 Threads hit max
MAIN.threads_created             200         0.07 Threads created
MAIN.threads_destroyed             0         0.00 Threads destroyed
MAIN.threads_failed                0         0.00 Thread creation failed
MAIN.thread_queue_len              0          .   Length of session queue
MAIN.busy_sleep                    0         0.00 Number of requests sent to sleep on busy objhdr
MAIN.busy_wakeup                   0         0.00 Number of requests woken after sleep on busy objhdr
MAIN.sess_queued                   0         0.00 Sessions queued for thread
MAIN.sess_dropped                  0         0.00 Sessions dropped for thread
MAIN.n_object                      0          .   object structs made
MAIN.n_vampireobject               0          .   unresurrected objects
MAIN.n_objectcore                  0          .   objectcore structs made
MAIN.n_objecthead                  0          .   objecthead structs made
MAIN.n_waitinglist                 0          .   waitinglist structs made
MAIN.n_backend                     1          .   Number of backends
MAIN.n_expired                     0          .   Number of expired objects
MAIN.n_lru_nuked                   0          .   Number of LRU nuked objects
MAIN.n_lru_moved                   0          .   Number of LRU moved objects
MAIN.losthdr                       0         0.00 HTTP header overflows
MAIN.s_sess                        0         0.00 Total sessions seen
MAIN.s_req                         0         0.00 Total requests seen
MAIN.s_pipe                        0         0.00 Total pipe sessions seen
MAIN.s_pass                        0         0.00 Total pass-ed requests seen
MAIN.s_fetch                       0         0.00 Total backend fetches initiated
MAIN.s_synth                       0         0.00 Total synthetic responses made
MAIN.s_req_hdrbytes                0         0.00 Request header bytes
MAIN.s_req_bodybytes               0         0.00 Request body bytes
MAIN.s_resp_hdrbytes               0         0.00 Response header bytes
MAIN.s_resp_bodybytes              0         0.00 Response body bytes
MAIN.s_pipe_hdrbytes               0         0.00 Pipe request header bytes
MAIN.s_pipe_in                     0         0.00 Piped bytes from client
MAIN.s_pipe_out                    0         0.00 Piped bytes to client
MAIN.sess_closed                   0         0.00 Session Closed
MAIN.sess_pipeline                 0         0.00 Session Pipeline
MAIN.sess_readahead                0         0.00 Session Read Ahead
MAIN.sess_herd                     0         0.00 Session herd
MAIN.shm_records                1918         0.67 SHM records
MAIN.shm_writes                 1918         0.67 SHM writes
MAIN.shm_flushes                   0         0.00 SHM flushes due to overflow
MAIN.shm_cont                      0         0.00 SHM MTX contention
MAIN.shm_cycles                    0         0.00 SHM cycles through buffer
MAIN.sms_nreq                      0         0.00 SMS allocator requests
MAIN.sms_nobj                      0          .   SMS outstanding allocations
MAIN.sms_nbytes                    0          .   SMS outstanding bytes
MAIN.sms_balloc                    0          .   SMS bytes allocated
MAIN.sms_bfree                     0          .   SMS bytes freed
MAIN.backend_req                   0         0.00 Backend requests made
MAIN.n_vcl                         1         0.00 Number of loaded VCLs in total
MAIN.n_vcl_avail                   1         0.00 Number of VCLs available
MAIN.n_vcl_discard                 0         0.00 Number of discarded VCLs
MAIN.bans                          1          .   Count of bans
MAIN.bans_completed                1          .   Number of bans marked 'completed'
MAIN.bans_obj                      0          .   Number of bans using obj.*
MAIN.bans_req                      0          .   Number of bans using req.*
MAIN.bans_added                    1         0.00 Bans added
MAIN.bans_deleted                  0         0.00 Bans deleted
MAIN.bans_tested                   0         0.00 Bans tested against objects (lookup)
MAIN.bans_obj_killed               0         0.00 Objects killed by bans (lookup)
MAIN.bans_lurker_tested            0         0.00 Bans tested against objects (lurker)
MAIN.bans_tests_tested             0         0.00 Ban tests tested against objects (lookup)
MAIN.bans_lurker_tests_tested            0         0.00 Ban tests tested against objects (lurker)
MAIN.bans_lurker_obj_killed              0         0.00 Objects killed by bans (lurker)
MAIN.bans_dups                           0         0.00 Bans superseded by other bans
MAIN.bans_lurker_contention              0         0.00 Lurker gave way for lookup
MAIN.bans_persisted_bytes               13          .   Bytes used by the persisted ban lists
MAIN.bans_persisted_fragmentation            0          .   Extra bytes in persisted ban lists due to fragmentation
MAIN.n_purges                                0          .   Number of purge operations executed
MAIN.n_obj_purged                            0          .   Number of purged objects
MAIN.exp_mailed                              0         0.00 Number of objects mailed to expiry thread
MAIN.exp_received                            0         0.00 Number of objects received by expiry thread
MAIN.hcb_nolock                              0         0.00 HCB Lookups without lock
MAIN.hcb_lock                                0         0.00 HCB Lookups with lock
MAIN.hcb_insert                              0         0.00 HCB Inserts
MAIN.esi_errors                              0         0.00 ESI parse errors (unlock)
MAIN.esi_warnings                            0         0.00 ESI parse warnings (unlock)
MAIN.vmods                                   0          .   Loaded VMODs
MAIN.n_gzip                                  0         0.00 Gzip operations
MAIN.n_gunzip                                0         0.00 Gunzip operations
MAIN.vsm_free                           972528          .   Free VSM space
MAIN.vsm_used                         83962080          .   Used VSM space
MAIN.vsm_cooling                             0          .   Cooling VSM space
MAIN.vsm_overflow                            0          .   Overflow VSM space
MAIN.vsm_overflowed                          0         0.00 Overflowed VSM space
MGT.uptime                                2871         1.00 Management process uptime
MGT.child_start                              1         0.00 Child process started
MGT.child_exit                               0         0.00 Child process normal exit
MGT.child_stop                               0         0.00 Child process unexpected exit
MGT.child_died                               0         0.00 Child process died (signal)
MGT.child_dump                               0         0.00 Child process core dumped
MGT.child_panic                              0         0.00 Child process panic
MEMPOOL.vbc.live                             0          .   In use
MEMPOOL.vbc.pool                            10          .   In Pool
MEMPOOL.vbc.sz_wanted                       88          .   Size requested
MEMPOOL.vbc.sz_needed                      120          .   Size allocated
MEMPOOL.vbc.allocs                           0         0.00 Allocations
MEMPOOL.vbc.frees                            0         0.00 Frees
MEMPOOL.vbc.recycle                          0         0.00 Recycled from pool
MEMPOOL.vbc.timeout                          0         0.00 Timed out from pool
MEMPOOL.vbc.toosmall                         0         0.00 Too small to recycle
MEMPOOL.vbc.surplus                          0         0.00 Too many for pool
MEMPOOL.vbc.randry                           0         0.00 Pool ran dry
MEMPOOL.busyobj.live                         0          .   In use
MEMPOOL.busyobj.pool                        10          .   In Pool
MEMPOOL.busyobj.sz_wanted                65536          .   Size requested
MEMPOOL.busyobj.sz_needed                65568          .   Size allocated
MEMPOOL.busyobj.allocs                       0         0.00 Allocations
MEMPOOL.busyobj.frees                        0         0.00 Frees
MEMPOOL.busyobj.recycle                      0         0.00 Recycled from pool
MEMPOOL.busyobj.timeout                      0         0.00 Timed out from pool
MEMPOOL.busyobj.toosmall                     0         0.00 Too small to recycle
MEMPOOL.busyobj.surplus                      0         0.00 Too many for pool
MEMPOOL.busyobj.randry                       0         0.00 Pool ran dry
MEMPOOL.req0.live                            0          .   In use
MEMPOOL.req0.pool                           10          .   In Pool
MEMPOOL.req0.sz_wanted                   65536          .   Size requested
MEMPOOL.req0.sz_needed                   65568          .   Size allocated
MEMPOOL.req0.allocs                          0         0.00 Allocations
MEMPOOL.req0.frees                           0         0.00 Frees
MEMPOOL.req0.recycle                         0         0.00 Recycled from pool
MEMPOOL.req0.timeout                         0         0.00 Timed out from pool
MEMPOOL.req0.toosmall                        0         0.00 Too small to recycle
MEMPOOL.req0.surplus                         0         0.00 Too many for pool
MEMPOOL.req0.randry                          0         0.00 Pool ran dry
MEMPOOL.sess0.live                           0          .   In use
MEMPOOL.sess0.pool                          10          .   In Pool
MEMPOOL.sess0.sz_wanted                    384          .   Size requested
MEMPOOL.sess0.sz_needed                    416          .   Size allocated
MEMPOOL.sess0.allocs                         0         0.00 Allocations
MEMPOOL.sess0.frees                          0         0.00 Frees
MEMPOOL.sess0.recycle                        0         0.00 Recycled from pool
MEMPOOL.sess0.timeout                        0         0.00 Timed out from pool
MEMPOOL.sess0.toosmall                       0         0.00 Too small to recycle
MEMPOOL.sess0.surplus                        0         0.00 Too many for pool
MEMPOOL.sess0.randry                         0         0.00 Pool ran dry
MEMPOOL.req1.live                            0          .   In use
MEMPOOL.req1.pool                           10          .   In Pool
MEMPOOL.req1.sz_wanted                   65536          .   Size requested
MEMPOOL.req1.sz_needed                   65568          .   Size allocated
MEMPOOL.req1.allocs                          0         0.00 Allocations
MEMPOOL.req1.frees                           0         0.00 Frees
MEMPOOL.req1.recycle                         0         0.00 Recycled from pool
MEMPOOL.req1.timeout                         0         0.00 Timed out from pool
MEMPOOL.req1.toosmall                        0         0.00 Too small to recycle
MEMPOOL.req1.surplus                         0         0.00 Too many for pool
MEMPOOL.req1.randry                          0         0.00 Pool ran dry
MEMPOOL.sess1.live                           0          .   In use
MEMPOOL.sess1.pool                          10          .   In Pool
MEMPOOL.sess1.sz_wanted                    384          .   Size requested
MEMPOOL.sess1.sz_needed                    416          .   Size allocated
MEMPOOL.sess1.allocs                         0         0.00 Allocations
MEMPOOL.sess1.frees                          0         0.00 Frees
MEMPOOL.sess1.recycle                        0         0.00 Recycled from pool
MEMPOOL.sess1.timeout                        0         0.00 Timed out from pool
MEMPOOL.sess1.toosmall                       0         0.00 Too small to recycle
MEMPOOL.sess1.surplus                        0         0.00 Too many for pool
MEMPOOL.sess1.randry                         0         0.00 Pool ran dry
SMA.s0.c_req                                 0         0.00 Allocator requests
SMA.s0.c_fail                                0         0.00 Allocator failures
SMA.s0.c_bytes                               0         0.00 Bytes allocated
SMA.s0.c_freed                               0         0.00 Bytes freed
SMA.s0.g_alloc                               0          .   Allocations outstanding
SMA.s0.g_bytes                               0          .   Bytes outstanding
SMA.s0.g_space                       268435456          .   Bytes available
SMA.Transient.c_req                          0         0.00 Allocator requests
SMA.Transient.c_fail                         0         0.00 Allocator failures
SMA.Transient.c_bytes                        0         0.00 Bytes allocated
SMA.Transient.c_freed                        0         0.00 Bytes freed
SMA.Transient.g_alloc                        0          .   Allocations outstanding
SMA.Transient.g_bytes                        0          .   Bytes outstanding
SMA.Transient.g_space                        0          .   Bytes available
VBE.default(127.0.0.1,,8080).vcls            1          .   VCL references
VBE.default(127.0.0.1,,8080).happy            0          .   Happy health probes
VBE.default(127.0.0.1,,8080).bereq_hdrbytes            0         0.00 Request header bytes
VBE.default(127.0.0.1,,8080).bereq_bodybytes            0         0.00 Request body bytes
VBE.default(127.0.0.1,,8080).beresp_hdrbytes            0         0.00 Response header bytes
VBE.default(127.0.0.1,,8080).beresp_bodybytes            0         0.00 Response body bytes
VBE.default(127.0.0.1,,8080).pipe_hdrbytes               0         0.00 Pipe request header bytes
VBE.default(127.0.0.1,,8080).pipe_out                    0         0.00 Piped bytes to backend
VBE.default(127.0.0.1,,8080).pipe_in                     0         0.00 Piped bytes from backend
LCK.sms.creat                                            0         0.00 Created locks
LCK.sms.destroy                                          0         0.00 Destroyed locks
LCK.sms.locks                                            0         0.00 Lock Operations
LCK.smp.creat                                            0         0.00 Created locks
LCK.smp.destroy                                          0         0.00 Destroyed locks
LCK.smp.locks                                            0         0.00 Lock Operations
LCK.sma.creat                                            2         0.00 Created locks
LCK.sma.destroy                                          0         0.00 Destroyed locks
LCK.sma.locks                                            0         0.00 Lock Operations
LCK.smf.creat                                            0         0.00 Created locks
LCK.smf.destroy                                          0         0.00 Destroyed locks
LCK.smf.locks                                            0         0.00 Lock Operations
LCK.hsl.creat                                            0         0.00 Created locks
LCK.hsl.destroy                                          0         0.00 Destroyed locks
LCK.hsl.locks                                            0         0.00 Lock Operations
LCK.hcb.creat                                            1         0.00 Created locks
LCK.hcb.destroy                                          0         0.00 Destroyed locks
LCK.hcb.locks                                           16         0.01 Lock Operations
LCK.hcl.creat                                            0         0.00 Created locks
LCK.hcl.destroy                                          0         0.00 Destroyed locks
LCK.hcl.locks                                            0         0.00 Lock Operations
LCK.vcl.creat                                            1         0.00 Created locks
LCK.vcl.destroy                                          0         0.00 Destroyed locks
LCK.vcl.locks                                            2         0.00 Lock Operations
LCK.sessmem.creat                                        0         0.00 Created locks
LCK.sessmem.destroy                                      0         0.00 Destroyed locks
LCK.sessmem.locks                                        0         0.00 Lock Operations
LCK.sess.creat                                           0         0.00 Created locks
LCK.sess.destroy                                         0         0.00 Destroyed locks
LCK.sess.locks                                           0         0.00 Lock Operations
LCK.wstat.creat                                          1         0.00 Created locks
LCK.wstat.destroy                                        0         0.00 Destroyed locks
LCK.wstat.locks                                        930         0.32 Lock Operations
LCK.herder.creat                                         0         0.00 Created locks
LCK.herder.destroy                                       0         0.00 Destroyed locks
LCK.herder.locks                                         0         0.00 Lock Operations
LCK.wq.creat                                             3         0.00 Created locks
LCK.wq.destroy                                           0         0.00 Destroyed locks
LCK.wq.locks                                          1554         0.54 Lock Operations
LCK.objhdr.creat                                         1         0.00 Created locks
LCK.objhdr.destroy                                       0         0.00 Destroyed locks
LCK.objhdr.locks                                         0         0.00 Lock Operations
LCK.exp.creat                                            1         0.00 Created locks
LCK.exp.destroy                                          0         0.00 Destroyed locks
LCK.exp.locks                                          915         0.32 Lock Operations
LCK.lru.creat                                            2         0.00 Created locks
LCK.lru.destroy                                          0         0.00 Destroyed locks
LCK.lru.locks                                            0         0.00 Lock Operations
LCK.cli.creat                                            1         0.00 Created locks
LCK.cli.destroy                                          0         0.00 Destroyed locks
LCK.cli.locks                                          970         0.34 Lock Operations
LCK.ban.creat                                            1         0.00 Created locks
LCK.ban.destroy                                          0         0.00 Destroyed locks
LCK.ban.locks                                         9413         3.28 Lock Operations
LCK.vbp.creat                                            1         0.00 Created locks
LCK.vbp.destroy                                          0         0.00 Destroyed locks
LCK.vbp.locks                                            0         0.00 Lock Operations
LCK.backend.creat                                        1         0.00 Created locks
LCK.backend.destroy                                      0         0.00 Destroyed locks
LCK.backend.locks                                        0         0.00 Lock Operations
LCK.vcapace.creat                                        1         0.00 Created locks
LCK.vcapace.destroy                                      0         0.00 Destroyed locks
LCK.vcapace.locks                                        0         0.00 Lock Operations
LCK.nbusyobj.creat                                       0         0.00 Created locks
LCK.nbusyobj.destroy                                     0         0.00 Destroyed locks
LCK.nbusyobj.locks                                       0         0.00 Lock Operations
LCK.busyobj.creat                                        0         0.00 Created locks
LCK.busyobj.destroy                                      0         0.00 Destroyed locks
LCK.busyobj.locks                                        0         0.00 Lock Operations
LCK.mempool.creat                                        6         0.00 Created locks
LCK.mempool.destroy                                      0         0.00 Destroyed locks
LCK.mempool.locks                                    15306         5.33 Lock Operations
LCK.vxid.creat                                           1         0.00 Created locks
LCK.vxid.destroy                                         0         0.00 Destroyed locks
LCK.vxid.locks                                           0         0.00 Lock Operations
LCK.pipestat.creat                                       1         0.00 Created locks
LCK.pipestat.destroy                                     0         0.00 Destroyed locks
LCK.pipestat.locks                                       0         0.00 Lock Operations
`

type testConfig struct {
	vName         string
	tags          map[string]string
	field         string
	activeVcl     string
	customRegexps []string
}

func TestV2ParseVarnishNames(t *testing.T) {
	for _, c := range []testConfig{
		{
			vName: "MGT.uptime",
			tags:  map[string]string{"section": "MGT"},
			field: "uptime",
		},
		{
			vName:     "VBE.boot.default.fail",
			tags:      map[string]string{"backend": "default", "section": "VBE"},
			field:     "fail",
			activeVcl: "boot",
		},
		{
			vName: "MEMPOOL.req1.allocs",
			tags:  map[string]string{"section": "MEMPOOL"},
			field: "req1.allocs",
		},
		{
			vName: "SMF.s0.c_bytes",
			tags:  map[string]string{"section": "SMF"},
			field: "s0.c_bytes",
		},
		{
			vName:     "VBE.reload_20210622_153544_23757.server1.happy",
			tags:      map[string]string{"backend": "server1", "section": "VBE"},
			field:     "happy",
			activeVcl: "reload_20210622_153544_23757",
		},
		{
			vName: "XXX.YYY.AAA",
			tags:  map[string]string{"section": "XXX"},
			field: "YYY.AAA",
		},
		{
			vName:     "VBE.vcl_20211502_214503.goto.000007d4.(10.100.0.1).(https://example.com:443).(ttl:10.000000).beresp_bodybytes",
			tags:      map[string]string{"backend": "10.100.0.1", "server": "https://example.com:443", "section": "VBE"},
			activeVcl: "vcl_20211502_214503",
			field:     "beresp_bodybytes",
		},
		{
			vName:     "VBE.VCL_xxxx_xxx_VOD_SHIELD_Vxxxxxxxxxxxxx_xxxxxxxxxxxxx.default.bereq_hdrbytes",
			tags:      map[string]string{"backend": "default", "section": "VBE"},
			activeVcl: "VCL_xxxx_xxx_VOD_SHIELD_Vxxxxxxxxxxxxx_xxxxxxxxxxxxx",
			field:     "bereq_hdrbytes",
		},
		{
			vName:     "VBE.VCL_ROUTER_V123_123.default.happy",
			tags:      map[string]string{"backend": "default", "section": "VBE"},
			field:     "happy",
			activeVcl: "VCL_ROUTER_V123_123",
		},
		{
			vName:     "KVSTORE.ds_stats.VCL_xxxx_xxx_A_B_C.shield",
			tags:      map[string]string{"id": "ds_stats", "section": "KVSTORE"},
			field:     "shield",
			activeVcl: "VCL_xxxx_xxx_A_B_C",
		},
		{
			vName:     "LCK.goto.director.destroy",
			tags:      map[string]string{"section": "LCK"},
			field:     "goto.director.destroy",
			activeVcl: "",
		},
		{
			vName:     "XCNT.1111.XXX+_LINE.cr.deliver_stub_restart.val",
			tags:      map[string]string{"group": "XXX+_LINE.cr", "section": "XCNT"},
			field:     "deliver_stub_restart",
			activeVcl: "1111",
		},
		{
			vName:     "VBE.VCL_1023_DIS_VOD_SHIELD_V1629295401194_1629295437531.goto.00000000.(111.112.113.114).(http://abc-ede.xyz.yyy.com:80).(ttl:3600.000000).is_healthy",
			tags:      map[string]string{"section": "VBE", "serial_1": "0", "backend_1": "111.112.113.114", "server_1": "http://abc-ede.xyz.yyy.com:80", "ttl": "3600.000000"},
			field:     "is_healthy",
			activeVcl: "VCL_1023_DIS_VOD_SHIELD_V1629295401194_1629295437531",
			customRegexps: []string{
				`^VBE\.(?P<_vcl>[\w\-]*)\.goto\.(?P<serial_1>[[:alnum:]])+\.\((?P<backend_1>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\)\.\((?P<server_1>.*)\)\.\(ttl:(?P<ttl>\d*\.\d*.)*\)`,
				`^VBE\.(?P<_vcl>[\w\-]*)\.goto\.(?P<serial_2>[[:alnum:]])+\.\((?P<backend_2>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\)\.\((?P<server_2>.*)\)\.\(ttl:(?P<ttl>\d*\.\d*.)*\)`,
			},
		},
	} {
		v := &Varnish{regexpsCompiled: defaultRegexps, Regexps: c.customRegexps}
		require.NoError(t, v.Init())
		vMetric := v.parseMetricV2(c.vName)
		require.Equal(t, c.activeVcl, vMetric.vclName)
		require.Equal(t, "varnish", vMetric.measurement, c.vName)
		require.Equal(t, c.field, vMetric.fieldName)
		require.Equal(t, c.tags, vMetric.tags)
	}
}

func TestVersions(t *testing.T) {
	server := &Varnish{regexpsCompiled: defaultRegexps}
	require.NoError(t, server.Init())
	acc := &testutil.Accumulator{}

	require.Equal(t, 0, len(acc.Metrics))

	type testConfig struct {
		jsonFile           string
		activeReloadPrefix string
		size               int
	}

	for _, c := range []testConfig{
		{jsonFile: "varnish_types.json", activeReloadPrefix: "", size: 3},
		{jsonFile: "varnish6.2.1_reload.json", activeReloadPrefix: "reload_20210623_170621_31083", size: 374},
		{jsonFile: "varnish6.2.1_reload.json", activeReloadPrefix: "", size: 434},
		{jsonFile: "varnish6.6.json", activeReloadPrefix: "boot", size: 358},
		{jsonFile: "varnish4_4.json", activeReloadPrefix: "boot", size: 295},
	} {
		output, _ := ioutil.ReadFile("test_data/" + c.jsonFile)
		err := server.processMetricsV2(c.activeReloadPrefix, acc, bytes.NewBuffer(output))
		require.NoError(t, err)
		require.Equal(t, c.size, len(acc.Metrics))
		for _, m := range acc.Metrics {
			require.NotEmpty(t, m.Fields)
			require.Equal(t, m.Measurement, "varnish")
			for field := range m.Fields {
				require.NotContains(t, field, "reload_")
			}
			for tag := range m.Tags {
				require.NotContains(t, tag, "reload_")
			}
		}
		acc.ClearMetrics()
	}
}

func TestJsonTypes(t *testing.T) {
	json := `{
		"timestamp": "2021-06-23T17:06:37",
			"counters": {
			"XXX.floatTest": {
				"description": "floatTest",
					"flag": "c",
					"format": "d",
					"value": 123.45
			},
			"XXX.stringTest": {
				"description": "stringTest",
					"flag": "c",
					"format": "d",
					"value": "abc_def"
			},
			"XXX.intTest": {
				"description": "intTest",
					"flag": "c",
					"format": "d",
					"value": 12345
			},
			"XXX.uintTest": {
				"description": "intTest",
					"flag": "b",
					"format": "b",
					"value": 18446744073709551615
			}
		}}`
	exp := map[string]interface{}{
		"floatTest":  123.45,
		"stringTest": "abc_def",
		"intTest":    int64(12345),
		"uintTest":   uint64(18446744073709551615),
	}
	acc := &testutil.Accumulator{}
	v := &Varnish{
		run:             fakeVarnishRunner(json),
		regexpsCompiled: defaultRegexps,
		Stats:           []string{"*"},
		MetricVersion:   2,
	}
	require.NoError(t, v.Gather(acc))
	require.Equal(t, len(exp), len(acc.Metrics))
	for _, metric := range acc.Metrics {
		require.Equal(t, "varnish", metric.Measurement)
		for fieldName, value := range metric.Fields {
			require.Equal(t, exp[fieldName], value)
		}
	}
}

func TestVarnishAdmJson(t *testing.T) {
	admJSON, _ := ioutil.ReadFile("test_data/" + "varnishadm-200.json")
	activeVcl, err := getActiveVCLJson(bytes.NewBuffer(admJSON))
	require.NoError(t, err)
	require.Equal(t, activeVcl, "boot-123")

	admJSON, _ = ioutil.ReadFile("test_data/" + "varnishadm-reload.json")
	activeVcl, err = getActiveVCLJson(bytes.NewBuffer(admJSON))
	require.NoError(t, err)
	require.Equal(t, activeVcl, "reload_20210723_091821_2056185")
}
