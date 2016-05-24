// +build !windows

package varnish

import (
	"bytes"
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func fakeVarnishStat(output string) func(string) (*bytes.Buffer, error) {
	return func(string) (*bytes.Buffer, error) {
		return bytes.NewBuffer([]byte(output)), nil
	}
}

func TestConfigsUsed(t *testing.T) {
	saved := varnishStat
	defer func() {
		varnishStat = saved
	}()

	expecations := map[string]string{
		"":             defaultBinary,
		"/foo/bar/baz": "/foo/bar/baz",
	}

	for in, expected := range expecations {
		varnishStat = func(actual string) (*bytes.Buffer, error) {
			assert.Equal(t, expected, actual)
			return &bytes.Buffer{}, nil
		}

		acc := &testutil.Accumulator{}
		v := &Varnish{Binary: in}
		v.Gather(acc)
	}
}

func TestGather(t *testing.T) {
	saved := varnishStat
	defer func() {
		varnishStat = saved
	}()
	varnishStat = fakeVarnishStat(smOutput)

	acc := &testutil.Accumulator{}
	v := &Varnish{Stats: []string{"all"}}
	v.Gather(acc)

	acc.HasMeasurement("varnish")
	for tag, fields := range parsedSmOutput {
		acc.AssertContainsTaggedFields(t, "varnish", fields, map[string]string{
			"section": tag,
		})
	}
}

func TestParseFullOutput(t *testing.T) {
	saved := varnishStat
	defer func() {
		varnishStat = saved
	}()
	varnishStat = fakeVarnishStat(fullOutput)

	acc := &testutil.Accumulator{}
	v := &Varnish{Stats: []string{"all"}}
	err := v.Gather(acc)

	assert.NoError(t, err)
	acc.HasMeasurement("varnish")
	flat := flatten(acc.Metrics)
	assert.Len(t, acc.Metrics, 6)
	assert.Equal(t, 293, len(flat))
}

func TestFieldConfig(t *testing.T) {
	saved := varnishStat
	defer func() {
		varnishStat = saved
	}()
	varnishStat = fakeVarnishStat(fullOutput)

	expect := map[string]int{
		"all":                                   293,
		"":                                      0, // default
		"MAIN.uptime":                           1,
		"MEMPOOL.req0.sz_needed,MAIN.fetch_bad": 2,
	}

	for fieldCfg, expected := range expect {
		acc := &testutil.Accumulator{}
		v := &Varnish{Stats: strings.Split(fieldCfg, ",")}
		err := v.Gather(acc)

		assert.NoError(t, err)
		acc.HasMeasurement("varnish")
		flat := flatten(acc.Metrics)
		assert.Equal(t, expected, len(flat))
	}
}

func flatten(metrics []*testutil.Metric) map[string]interface{} {
	flat := map[string]interface{}{}
	for _, m := range metrics {
		buf := &bytes.Buffer{}
		for k, v := range m.Tags {
			buf.WriteString(fmt.Sprintf("%s=%s", k, v))
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
	"MAIN": map[string]interface{}{
		"uptime":     895,
		"cache_hit":  95,
		"cache_miss": 5,
	},
	"MGT": map[string]interface{}{
		"uptime":      896,
		"child_start": 1,
	},
	"MEMPOOL": map[string]interface{}{
		"vbc.live":      0,
		"vbc.pool":      10,
		"vbc.sz_wanted": 88,
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
MAIN.s_synth                       0         0.00 Total synthethic responses made
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
