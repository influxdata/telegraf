# Varnish Input Plugin

This plugin gathers stats from [Varnish HTTP Cache](https://varnish-cache.org/)

## Configuration

```toml
# A plugin to collect stats from Varnish HTTP Cache
[[inputs.varnish]]
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the varnishstat binary can be overridden with:
  binary = "/usr/bin/varnishstat"

  ## Additional custom arguments for the varnishstat command
  # binary_args = ["-f", "MAIN.*"]

  ## The default location of the varnishadm binary can be overridden with:
  adm_binary = "/usr/bin/varnishadm"

  ## Custom arguments for the varnishadm command
  # adm_binary_args = [""]

  ## Metric version defaults to metric_version=1, use metric_version=2 for removal of nonactive vcls
  ## Varnish 6.0.2 and newer is required for metric_version=2.
  metric_version = 1

  ## Additional regexps to override builtin conversion of varnish metrics into telegraf metrics.
  ## Regexp group "_vcl" is used for extracting the VCL name. Metrics that contain nonactive VCL's are skipped.
  ## Regexp group "_field" overrides the field name. Other named regexp groups are used as tags.
  # regexps = ['^XCNT\.(?P<_vcl>[\w\-]*)(\.)*(?P<group>[\w\-.+]*)\.(?P<_field>[\w\-.+]*)\.val']

  ## By default, telegraf gather stats for 3 metric points.
  ## Setting stats will override the defaults shown below.
  ## Glob matching can be used, ie, stats = ["MAIN.*"]
  ## stats may also be set to ["*"], which will collect all stats
  stats = ["MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"]

  ## Optional name for the varnish instance (or working directory) to query
  ## Usually append after -n in varnish cli
  # instance_name = instanceName

  ## Timeout for varnishstat command
  # timeout = "1s"
```

### Measurements & Fields (metric_version=1)

This is the full list of stats provided by varnish. Stats will be grouped by their capitalized prefix (eg MAIN, MEMPOOL,
etc). In the output, the prefix will be used as a tag, and removed from field names.

- varnish
  - MAIN.uptime                                    (uint64, count, Child process uptime)
  - MAIN.sess_conn                                 (uint64, count, Sessions accepted)
  - MAIN.sess_drop                                 (uint64, count, Sessions dropped)
  - MAIN.sess_fail                                 (uint64, count, Session accept failures)
  - MAIN.sess_pipe_overflow                        (uint64, count, Session pipe overflow)
  - MAIN.client_req_400                            (uint64, count, Client requests received,)
  - MAIN.client_req_411                            (uint64, count, Client requests received,)
  - MAIN.client_req_413                            (uint64, count, Client requests received,)
  - MAIN.client_req_417                            (uint64, count, Client requests received,)
  - MAIN.client_req                                (uint64, count, Good client requests)
  - MAIN.cache_hit                                 (uint64, count, Cache hits)
  - MAIN.cache_hitpass                             (uint64, count, Cache hits for)
  - MAIN.cache_miss                                (uint64, count, Cache misses)
  - MAIN.backend_conn                              (uint64, count, Backend conn. success)
  - MAIN.backend_unhealthy                         (uint64, count, Backend conn. not)
  - MAIN.backend_busy                              (uint64, count, Backend conn. too)
  - MAIN.backend_fail                              (uint64, count, Backend conn. failures)
  - MAIN.backend_reuse                             (uint64, count, Backend conn. reuses)
  - MAIN.backend_toolate                           (uint64, count, Backend conn. was)
  - MAIN.backend_recycle                           (uint64, count, Backend conn. recycles)
  - MAIN.backend_retry                             (uint64, count, Backend conn. retry)
  - MAIN.fetch_head                                (uint64, count, Fetch no body)
  - MAIN.fetch_length                              (uint64, count, Fetch with Length)
  - MAIN.fetch_chunked                             (uint64, count, Fetch chunked)
  - MAIN.fetch_eof                                 (uint64, count, Fetch EOF)
  - MAIN.fetch_bad                                 (uint64, count, Fetch bad T- E)
  - MAIN.fetch_close                               (uint64, count, Fetch wanted close)
  - MAIN.fetch_oldhttp                             (uint64, count, Fetch pre HTTP/1.1)
  - MAIN.fetch_zero                                (uint64, count, Fetch zero len)
  - MAIN.fetch_1xx                                 (uint64, count, Fetch no body)
  - MAIN.fetch_204                                 (uint64, count, Fetch no body)
  - MAIN.fetch_304                                 (uint64, count, Fetch no body)
  - MAIN.fetch_failed                              (uint64, count, Fetch failed (all)
  - MAIN.fetch_no_thread                           (uint64, count, Fetch failed (no)
  - MAIN.pools                                     (uint64, count, Number of thread)
  - MAIN.threads                                   (uint64, count, Total number of)
  - MAIN.threads_limited                           (uint64, count, Threads hit max)
  - MAIN.threads_created                           (uint64, count, Threads created)
  - MAIN.threads_destroyed                         (uint64, count, Threads destroyed)
  - MAIN.threads_failed                            (uint64, count, Thread creation failed)
  - MAIN.thread_queue_len                          (uint64, count, Length of session)
  - MAIN.busy_sleep                                (uint64, count, Number of requests)
  - MAIN.busy_wakeup                               (uint64, count, Number of requests)
  - MAIN.sess_queued                               (uint64, count, Sessions queued for)
  - MAIN.sess_dropped                              (uint64, count, Sessions dropped for)
  - MAIN.n_object                                  (uint64, count, object structs made)
  - MAIN.n_vampireobject                           (uint64, count, unresurrected objects)
  - MAIN.n_objectcore                              (uint64, count, objectcore structs made)
  - MAIN.n_objecthead                              (uint64, count, objecthead structs made)
  - MAIN.n_waitinglist                             (uint64, count, waitinglist structs made)
  - MAIN.n_backend                                 (uint64, count, Number of backends)
  - MAIN.n_expired                                 (uint64, count, Number of expired)
  - MAIN.n_lru_nuked                               (uint64, count, Number of LRU)
  - MAIN.n_lru_moved                               (uint64, count, Number of LRU)
  - MAIN.losthdr                                   (uint64, count, HTTP header overflows)
  - MAIN.s_sess                                    (uint64, count, Total sessions seen)
  - MAIN.s_req                                     (uint64, count, Total requests seen)
  - MAIN.s_pipe                                    (uint64, count, Total pipe sessions)
  - MAIN.s_pass                                    (uint64, count, Total pass- ed requests)
  - MAIN.s_fetch                                   (uint64, count, Total backend fetches)
  - MAIN.s_synth                                   (uint64, count, Total synthetic responses)
  - MAIN.s_req_hdrbytes                            (uint64, count, Request header bytes)
  - MAIN.s_req_bodybytes                           (uint64, count, Request body bytes)
  - MAIN.s_resp_hdrbytes                           (uint64, count, Response header bytes)
  - MAIN.s_resp_bodybytes                          (uint64, count, Response body bytes)
  - MAIN.s_pipe_hdrbytes                           (uint64, count, Pipe request header)
  - MAIN.s_pipe_in                                 (uint64, count, Piped bytes from)
  - MAIN.s_pipe_out                                (uint64, count, Piped bytes to)
  - MAIN.sess_closed                               (uint64, count, Session Closed)
  - MAIN.sess_pipeline                             (uint64, count, Session Pipeline)
  - MAIN.sess_readahead                            (uint64, count, Session Read Ahead)
  - MAIN.sess_herd                                 (uint64, count, Session herd)
  - MAIN.shm_records                               (uint64, count, SHM records)
  - MAIN.shm_writes                                (uint64, count, SHM writes)
  - MAIN.shm_flushes                               (uint64, count, SHM flushes due)
  - MAIN.shm_cont                                  (uint64, count, SHM MTX contention)
  - MAIN.shm_cycles                                (uint64, count, SHM cycles through)
  - MAIN.sms_nreq                                  (uint64, count, SMS allocator requests)
  - MAIN.sms_nobj                                  (uint64, count, SMS outstanding allocations)
  - MAIN.sms_nbytes                                (uint64, count, SMS outstanding bytes)
  - MAIN.sms_balloc                                (uint64, count, SMS bytes allocated)
  - MAIN.sms_bfree                                 (uint64, count, SMS bytes freed)
  - MAIN.backend_req                               (uint64, count, Backend requests made)
  - MAIN.n_vcl                                     (uint64, count, Number of loaded)
  - MAIN.n_vcl_avail                               (uint64, count, Number of VCLs)
  - MAIN.n_vcl_discard                             (uint64, count, Number of discarded)
  - MAIN.bans                                      (uint64, count, Count of bans)
  - MAIN.bans_completed                            (uint64, count, Number of bans)
  - MAIN.bans_obj                                  (uint64, count, Number of bans)
  - MAIN.bans_req                                  (uint64, count, Number of bans)
  - MAIN.bans_added                                (uint64, count, Bans added)
  - MAIN.bans_deleted                              (uint64, count, Bans deleted)
  - MAIN.bans_tested                               (uint64, count, Bans tested against)
  - MAIN.bans_obj_killed                           (uint64, count, Objects killed by)
  - MAIN.bans_lurker_tested                        (uint64, count, Bans tested against)
  - MAIN.bans_tests_tested                         (uint64, count, Ban tests tested)
  - MAIN.bans_lurker_tests_tested                  (uint64, count, Ban tests tested)
  - MAIN.bans_lurker_obj_killed                    (uint64, count, Objects killed by)
  - MAIN.bans_dups                                 (uint64, count, Bans superseded by)
  - MAIN.bans_lurker_contention                    (uint64, count, Lurker gave way)
  - MAIN.bans_persisted_bytes                      (uint64, count, Bytes used by)
  - MAIN.bans_persisted_fragmentation              (uint64, count, Extra bytes in)
  - MAIN.n_purges                                  (uint64, count, Number of purge)
  - MAIN.n_obj_purged                              (uint64, count, Number of purged)
  - MAIN.exp_mailed                                (uint64, count, Number of objects)
  - MAIN.exp_received                              (uint64, count, Number of objects)
  - MAIN.hcb_nolock                                (uint64, count, HCB Lookups without)
  - MAIN.hcb_lock                                  (uint64, count, HCB Lookups with)
  - MAIN.hcb_insert                                (uint64, count, HCB Inserts)
  - MAIN.esi_errors                                (uint64, count, ESI parse errors)
  - MAIN.esi_warnings                              (uint64, count, ESI parse warnings)
  - MAIN.vmods                                     (uint64, count, Loaded VMODs)
  - MAIN.n_gzip                                    (uint64, count, Gzip operations)
  - MAIN.n_gunzip                                  (uint64, count, Gunzip operations)
  - MAIN.vsm_free                                  (uint64, count, Free VSM space)
  - MAIN.vsm_used                                  (uint64, count, Used VSM space)
  - MAIN.vsm_cooling                               (uint64, count, Cooling VSM space)
  - MAIN.vsm_overflow                              (uint64, count, Overflow VSM space)
  - MAIN.vsm_overflowed                            (uint64, count, Overflowed VSM space)
  - MGT.uptime                                     (uint64, count, Management process uptime)
  - MGT.child_start                                (uint64, count, Child process started)
  - MGT.child_exit                                 (uint64, count, Child process normal)
  - MGT.child_stop                                 (uint64, count, Child process unexpected)
  - MGT.child_died                                 (uint64, count, Child process died)
  - MGT.child_dump                                 (uint64, count, Child process core)
  - MGT.child_panic                                (uint64, count, Child process panic)
  - MEMPOOL.vbc.live                               (uint64, count, In use)
  - MEMPOOL.vbc.pool                               (uint64, count, In Pool)
  - MEMPOOL.vbc.sz_wanted                          (uint64, count, Size requested)
  - MEMPOOL.vbc.sz_needed                          (uint64, count, Size allocated)
  - MEMPOOL.vbc.allocs                             (uint64, count, Allocations )
  - MEMPOOL.vbc.frees                              (uint64, count, Frees )
  - MEMPOOL.vbc.recycle                            (uint64, count, Recycled from pool)
  - MEMPOOL.vbc.timeout                            (uint64, count, Timed out from)
  - MEMPOOL.vbc.toosmall                           (uint64, count, Too small to)
  - MEMPOOL.vbc.surplus                            (uint64, count, Too many for)
  - MEMPOOL.vbc.randry                             (uint64, count, Pool ran dry)
  - MEMPOOL.busyobj.live                           (uint64, count, In use)
  - MEMPOOL.busyobj.pool                           (uint64, count, In Pool)
  - MEMPOOL.busyobj.sz_wanted                      (uint64, count, Size requested)
  - MEMPOOL.busyobj.sz_needed                      (uint64, count, Size allocated)
  - MEMPOOL.busyobj.allocs                         (uint64, count, Allocations )
  - MEMPOOL.busyobj.frees                          (uint64, count, Frees )
  - MEMPOOL.busyobj.recycle                        (uint64, count, Recycled from pool)
  - MEMPOOL.busyobj.timeout                        (uint64, count, Timed out from)
  - MEMPOOL.busyobj.toosmall                       (uint64, count, Too small to)
  - MEMPOOL.busyobj.surplus                        (uint64, count, Too many for)
  - MEMPOOL.busyobj.randry                         (uint64, count, Pool ran dry)
  - MEMPOOL.req0.live                              (uint64, count, In use)
  - MEMPOOL.req0.pool                              (uint64, count, In Pool)
  - MEMPOOL.req0.sz_wanted                         (uint64, count, Size requested)
  - MEMPOOL.req0.sz_needed                         (uint64, count, Size allocated)
  - MEMPOOL.req0.allocs                            (uint64, count, Allocations )
  - MEMPOOL.req0.frees                             (uint64, count, Frees )
  - MEMPOOL.req0.recycle                           (uint64, count, Recycled from pool)
  - MEMPOOL.req0.timeout                           (uint64, count, Timed out from)
  - MEMPOOL.req0.toosmall                          (uint64, count, Too small to)
  - MEMPOOL.req0.surplus                           (uint64, count, Too many for)
  - MEMPOOL.req0.randry                            (uint64, count, Pool ran dry)
  - MEMPOOL.sess0.live                             (uint64, count, In use)
  - MEMPOOL.sess0.pool                             (uint64, count, In Pool)
  - MEMPOOL.sess0.sz_wanted                        (uint64, count, Size requested)
  - MEMPOOL.sess0.sz_needed                        (uint64, count, Size allocated)
  - MEMPOOL.sess0.allocs                           (uint64, count, Allocations )
  - MEMPOOL.sess0.frees                            (uint64, count, Frees )
  - MEMPOOL.sess0.recycle                          (uint64, count, Recycled from pool)
  - MEMPOOL.sess0.timeout                          (uint64, count, Timed out from)
  - MEMPOOL.sess0.toosmall                         (uint64, count, Too small to)
  - MEMPOOL.sess0.surplus                          (uint64, count, Too many for)
  - MEMPOOL.sess0.randry                           (uint64, count, Pool ran dry)
  - MEMPOOL.req1.live                              (uint64, count, In use)
  - MEMPOOL.req1.pool                              (uint64, count, In Pool)
  - MEMPOOL.req1.sz_wanted                         (uint64, count, Size requested)
  - MEMPOOL.req1.sz_needed                         (uint64, count, Size allocated)
  - MEMPOOL.req1.allocs                            (uint64, count, Allocations )
  - MEMPOOL.req1.frees                             (uint64, count, Frees )
  - MEMPOOL.req1.recycle                           (uint64, count, Recycled from pool)
  - MEMPOOL.req1.timeout                           (uint64, count, Timed out from)
  - MEMPOOL.req1.toosmall                          (uint64, count, Too small to)
  - MEMPOOL.req1.surplus                           (uint64, count, Too many for)
  - MEMPOOL.req1.randry                            (uint64, count, Pool ran dry)
  - MEMPOOL.sess1.live                             (uint64, count, In use)
  - MEMPOOL.sess1.pool                             (uint64, count, In Pool)
  - MEMPOOL.sess1.sz_wanted                        (uint64, count, Size requested)
  - MEMPOOL.sess1.sz_needed                        (uint64, count, Size allocated)
  - MEMPOOL.sess1.allocs                           (uint64, count, Allocations )
  - MEMPOOL.sess1.frees                            (uint64, count, Frees )
  - MEMPOOL.sess1.recycle                          (uint64, count, Recycled from pool)
  - MEMPOOL.sess1.timeout                          (uint64, count, Timed out from)
  - MEMPOOL.sess1.toosmall                         (uint64, count, Too small to)
  - MEMPOOL.sess1.surplus                          (uint64, count, Too many for)
  - MEMPOOL.sess1.randry                           (uint64, count, Pool ran dry)
  - SMA.s0.c_req                                   (uint64, count, Allocator requests)
  - SMA.s0.c_fail                                  (uint64, count, Allocator failures)
  - SMA.s0.c_bytes                                 (uint64, count, Bytes allocated)
  - SMA.s0.c_freed                                 (uint64, count, Bytes freed)
  - SMA.s0.g_alloc                                 (uint64, count, Allocations outstanding)
  - SMA.s0.g_bytes                                 (uint64, count, Bytes outstanding)
  - SMA.s0.g_space                                 (uint64, count, Bytes available)
  - SMA.Transient.c_req                            (uint64, count, Allocator requests)
  - SMA.Transient.c_fail                           (uint64, count, Allocator failures)
  - SMA.Transient.c_bytes                          (uint64, count, Bytes allocated)
  - SMA.Transient.c_freed                          (uint64, count, Bytes freed)
  - SMA.Transient.g_alloc                          (uint64, count, Allocations outstanding)
  - SMA.Transient.g_bytes                          (uint64, count, Bytes outstanding)
  - SMA.Transient.g_space                          (uint64, count, Bytes available)
  - VBE.default(127.0.0.1,,8080).vcls              (uint64, count, VCL references)
  - VBE.default(127.0.0.1,,8080).happy             (uint64, count, Happy health probes)
  - VBE.default(127.0.0.1,,8080).bereq_hdrbytes    (uint64, count, Request header bytes)
  - VBE.default(127.0.0.1,,8080).bereq_bodybytes   (uint64, count, Request body bytes)
  - VBE.default(127.0.0.1,,8080).beresp_hdrbytes   (uint64, count, Response header bytes)
  - VBE.default(127.0.0.1,,8080).beresp_bodybytes  (uint64, count, Response body bytes)
  - VBE.default(127.0.0.1,,8080).pipe_hdrbytes     (uint64, count, Pipe request header)
  - VBE.default(127.0.0.1,,8080).pipe_out          (uint64, count, Piped bytes to)
  - VBE.default(127.0.0.1,,8080).pipe_in           (uint64, count, Piped bytes from)
  - LCK.sms.creat                                  (uint64, count, Created locks)
  - LCK.sms.destroy                                (uint64, count, Destroyed locks)
  - LCK.sms.locks                                  (uint64, count, Lock Operations)
  - LCK.smp.creat                                  (uint64, count, Created locks)
  - LCK.smp.destroy                                (uint64, count, Destroyed locks)
  - LCK.smp.locks                                  (uint64, count, Lock Operations)
  - LCK.sma.creat                                  (uint64, count, Created locks)
  - LCK.sma.destroy                                (uint64, count, Destroyed locks)
  - LCK.sma.locks                                  (uint64, count, Lock Operations)
  - LCK.smf.creat                                  (uint64, count, Created locks)
  - LCK.smf.destroy                                (uint64, count, Destroyed locks)
  - LCK.smf.locks                                  (uint64, count, Lock Operations)
  - LCK.hsl.creat                                  (uint64, count, Created locks)
  - LCK.hsl.destroy                                (uint64, count, Destroyed locks)
  - LCK.hsl.locks                                  (uint64, count, Lock Operations)
  - LCK.hcb.creat                                  (uint64, count, Created locks)
  - LCK.hcb.destroy                                (uint64, count, Destroyed locks)
  - LCK.hcb.locks                                  (uint64, count, Lock Operations)
  - LCK.hcl.creat                                  (uint64, count, Created locks)
  - LCK.hcl.destroy                                (uint64, count, Destroyed locks)
  - LCK.hcl.locks                                  (uint64, count, Lock Operations)
  - LCK.vcl.creat                                  (uint64, count, Created locks)
  - LCK.vcl.destroy                                (uint64, count, Destroyed locks)
  - LCK.vcl.locks                                  (uint64, count, Lock Operations)
  - LCK.sessmem.creat                              (uint64, count, Created locks)
  - LCK.sessmem.destroy                            (uint64, count, Destroyed locks)
  - LCK.sessmem.locks                              (uint64, count, Lock Operations)
  - LCK.sess.creat                                 (uint64, count, Created locks)
  - LCK.sess.destroy                               (uint64, count, Destroyed locks)
  - LCK.sess.locks                                 (uint64, count, Lock Operations)
  - LCK.wstat.creat                                (uint64, count, Created locks)
  - LCK.wstat.destroy                              (uint64, count, Destroyed locks)
  - LCK.wstat.locks                                (uint64, count, Lock Operations)
  - LCK.herder.creat                               (uint64, count, Created locks)
  - LCK.herder.destroy                             (uint64, count, Destroyed locks)
  - LCK.herder.locks                               (uint64, count, Lock Operations)
  - LCK.wq.creat                                   (uint64, count, Created locks)
  - LCK.wq.destroy                                 (uint64, count, Destroyed locks)
  - LCK.wq.locks                                   (uint64, count, Lock Operations)
  - LCK.objhdr.creat                               (uint64, count, Created locks)
  - LCK.objhdr.destroy                             (uint64, count, Destroyed locks)
  - LCK.objhdr.locks                               (uint64, count, Lock Operations)
  - LCK.exp.creat                                  (uint64, count, Created locks)
  - LCK.exp.destroy                                (uint64, count, Destroyed locks)
  - LCK.exp.locks                                  (uint64, count, Lock Operations)
  - LCK.lru.creat                                  (uint64, count, Created locks)
  - LCK.lru.destroy                                (uint64, count, Destroyed locks)
  - LCK.lru.locks                                  (uint64, count, Lock Operations)
  - LCK.cli.creat                                  (uint64, count, Created locks)
  - LCK.cli.destroy                                (uint64, count, Destroyed locks)
  - LCK.cli.locks                                  (uint64, count, Lock Operations)
  - LCK.ban.creat                                  (uint64, count, Created locks)
  - LCK.ban.destroy                                (uint64, count, Destroyed locks)
  - LCK.ban.locks                                  (uint64, count, Lock Operations)
  - LCK.vbp.creat                                  (uint64, count, Created locks)
  - LCK.vbp.destroy                                (uint64, count, Destroyed locks)
  - LCK.vbp.locks                                  (uint64, count, Lock Operations)
  - LCK.backend.creat                              (uint64, count, Created locks)
  - LCK.backend.destroy                            (uint64, count, Destroyed locks)
  - LCK.backend.locks                              (uint64, count, Lock Operations)
  - LCK.vcapace.creat                              (uint64, count, Created locks)
  - LCK.vcapace.destroy                            (uint64, count, Destroyed locks)
  - LCK.vcapace.locks                              (uint64, count, Lock Operations)
  - LCK.nbusyobj.creat                             (uint64, count, Created locks)
  - LCK.nbusyobj.destroy                           (uint64, count, Destroyed locks)
  - LCK.nbusyobj.locks                             (uint64, count, Lock Operations)
  - LCK.busyobj.creat                              (uint64, count, Created locks)
  - LCK.busyobj.destroy                            (uint64, count, Destroyed locks)
  - LCK.busyobj.locks                              (uint64, count, Lock Operations)
  - LCK.mempool.creat                              (uint64, count, Created locks)
  - LCK.mempool.destroy                            (uint64, count, Destroyed locks)
  - LCK.mempool.locks                              (uint64, count, Lock Operations)
  - LCK.vxid.creat                                 (uint64, count, Created locks)
  - LCK.vxid.destroy                               (uint64, count, Destroyed locks)
  - LCK.vxid.locks                                 (uint64, count, Lock Operations)
  - LCK.pipestat.creat                             (uint64, count, Created locks)
  - LCK.pipestat.destroy                           (uint64, count, Destroyed locks)
  - LCK.pipestat.locks                             (uint64, count, Lock Operations)

### Tags

As indicated above, the prefix of a varnish stat will be used as it's 'section' tag. So section tag may have one of the
following values:

- section:
  - MAIN
  - MGT
  - MEMPOOL
  - SMA
  - VBE
  - LCK

## Measurements & Fields (metric_version=2)

When `metric_version=2` is enabled, the plugin runs `varnishstat -j` command and parses the JSON output into metrics.

Plugin uses `varnishadm vcl.list -j` commandline to find the active VCL. Metrics that are related to the nonactive VCL
are excluded from monitoring.

### Requirements

- Varnish 6.0.2+ is required (older versions do not support JSON output from CLI tools)

#### Examples

Varnish counter:

```json
{
  "MAIN.cache_hit": {
    "description": "Cache hits",
    "flag": "c",
    "format": "i",
    "value": 51
  }
}
```

Influx metric:
`varnish,section=MAIN cache_hit=51i 1462765437090957980`

### Advanced customizations using regexps

Finding the VCL in a varnish measurement and parsing into tags can be adjusted by using GO regular expressions.

Regexps use a special named group `(?P<_vcl>[\w\-]*)(\.)` to extract VCL name. `(?P<_field>[\w\-.+]*)\.val` regexp group
extracts the field name. All other named regexp groups like `(?P<my_tag>[\w\-.+]*)` are tags.

_Tip: It is useful to verify regexps using online tools like <https://regoio.herokuapp.com/>._

By default, the plugin has a builtin list of regexps for following VMODs:

- Dynamic Backends (goto)
  - regexp: `^VBE\.(?P<_vcl>[\w\-]*)\.goto\.[[:alnum:]]+\.\((?P<backend>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\)\.\((?P<server>.*)\)\.\(ttl:\d*\.\d*.*\)`
    - `VBE.VCL12323.goto.000007c8.(123.123.123.123).(http://aaa.xxcc:80).(ttl:3600.000000).cache_hit` -> `varnish,section=VBE,backend="123.123.123.123",server="http://aaa.xxcc:80" cache_hit=51i 1462765437090957980`

- Key value storage (kvstore)
  - regexp `^KVSTORE\.(?P<id>[\w\-]*)\.(?P<_vcl>[\w\-]*)\.([\w\-]*)`
    - `KVSTORE.object_name.vcl_name.key` -> `varnish,section=KVSTORE,id=object_name key=5i`
- XCNT (libvmod-xcounter)
  - regexp `^XCNT\.(?P<_vcl>[\w\-]*)(\.)*(?P<group>[\w\-.+]*)\.(?P<_field>[\w\-.+]*)\.val`
    - `XCNT.abc1234.XXX+_YYYY.cr.pass.val` -> `varnish,section=XCNT,group="XXX+_YYYY.cr" pass=5i`

- standard VBE metrics
  - regexp `^VBE\.(?P<_vcl>[\w\-]*)\.(?P<backend>[\w\-]*)\.([\w\-]*)`
    - `VBE.reload_20210622_153544_23757.default.unhealthy` -> `varnish,section=VBE,backend="default" unhealthy=51i 1462765437090957980`
- default generic metric
  - regexp `([\w\-]*)\.(?P<_field>[\w\-.]*)`
    - `MSE_STORE.store-1-1.g_aio_running_bytes_write` -> `varnish,section=MSE_STORE store-1-1.g_aio_running_bytes_write=5i`

The default regexps list can be extended in the telegraf config. The following example shows a config with a custom
regexp for parsing of `accounting` VMOD metrics in `ACCG.<namespace>.<key>.<stat_name>` format. The namespace value will
be used as a tag.

```toml
[[inputs.varnish]]
    regexps = ['^ACCG.(?P<namespace>[\w-]*).(?P<_field>[\w-.]*)']
```

### Custom arguments

You can change the default binary location and custom arguments for `varnishstat` and `varnishadm` command output. This
is useful when running varnish in docker or executing using varnish by SSH on a different machine.

It's important to note that `instance_name` parameter is not take into account when using custom `binary_args` or
`adm_binary_args`. You have to add `"-n", "/instance_name"` manually into configuration.

#### Example for SSH

```toml
[[inputs.varnish]]
  binary = "/usr/bin/ssh"
  binary_args = ["root@10.100.0.112", "varnishstat", "-n", "/var/lib/varnish/ubuntu", "-j"]
  adm_binary   =  "/usr/bin/ssh"
  adm_binary_args = ["root@10.100.0.112", "varnishadm", "-n", "/var/lib/varnish/ubuntu", "vcl.list", "-j"]
  metric_version = 2
  stats = ["*"]
```

#### Example for Docker

```toml
[[inputs.varnish]]
  binary = "/usr/local/bin/docker"
  binary_args = ["exec", "-t", "container_name", "varnishstat",  "-j"]
  adm_binary   =  "/usr/local/bin/docker"
  adm_binary_args =  ["exec", "-t", "container_name", "varnishadm", "vcl.list", "-j"]
  metric_version = 2
  stats = ["*"]
```

### Permissions

It's important to note that this plugin references `varnishstat` and `varnishadm`, which may require additional permissions to execute successfully.
Depending on the user/group permissions of the telegraf user executing this plugin, you may need to alter the group membership, set facls, or use sudo.

#### Group membership (Recommended)

```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G varnish telegraf

$ groups telegraf
telegraf : telegraf varnish
```

#### Extended filesystem ACL's

```bash
$ getfacl /var/lib/varnish/<hostname>/_.vsm
# file: var/lib/varnish/<hostname>/_.vsm
# owner: root
# group: root
user::rw-
group::r--
other::---

$ setfacl -m u:telegraf:r /var/lib/varnish/<hostname>/_.vsm

$ getfacl /var/lib/varnish/<hostname>/_.vsm
# file: var/lib/varnish/<hostname>/_.vsm
# owner: root
# group: root
user::rw-
user:telegraf:r--
group::r--
mask::r--
other::---
```

**Sudo privileges**:
If you use this method, you will need the following in your telegraf config:

```toml
[[inputs.varnish]]
  use_sudo = true
```

You will also need to update your sudoers file:

```bash
$ visudo
# Add the following line:
Cmnd_Alias VARNISHSTAT = /usr/bin/varnishstat
telegraf  ALL=(ALL) NOPASSWD: VARNISHSTAT
Defaults!VARNISHSTAT !logfile, !syslog, !pam_session
```

Please use the solution you see as most appropriate.

### Example Output

```bash
 telegraf --config etc/telegraf.conf --input-filter varnish --test
* Plugin: varnish, Collection 1
> varnish,host=rpercy-VirtualBox,section=MAIN cache_hit=0i,cache_miss=0i,uptime=8416i 1462765437090957980
```

### Output (when metric_version = 2)

```bash
telegraf --config etc/telegraf.conf --input-filter varnish --test
> varnish,host=kozel.local,section=MAIN n_vampireobject=0i 1631121567000000000
> varnish,backend=server_test1,host=kozel.local,section=VBE fail_eacces=0i 1631121567000000000
> varnish,backend=default,host=kozel.local,section=VBE req=0i 1631121567000000000
> varnish,host=kozel.local,section=MAIN client_req_400=0i 1631121567000000000
> varnish,host=kozel.local,section=MAIN shm_cycles=10i 1631121567000000000
> varnish,backend=default,host=kozel.local,section=VBE pipe_hdrbytes=0i 1631121567000000000
```

You can merge metrics together into a metric with multiple fields into the most
memory and network transfer efficient form using `aggregators.merge`

```toml
[[aggregators.merge]]
  drop_original = true
```

The output will be:

```bash
telegraf --config etc/telegraf.conf --input-filter varnish --test
> varnish,host=kozel.local,section=MAIN backend_busy=0i,backend_conn=19i,backend_fail=0i,backend_recycle=8i,backend_req=19i,backend_retry=0i,backend_reuse=0i,backend_unhealthy=0i,bans=1i,bans_added=1i,bans_completed=1i,bans_deleted=0i,bans_dups=0i,bans_lurker_contention=0i,bans_lurker_obj_killed=0i,bans_lurker_obj_killed_cutoff=0i,bans_lurker_tested=0i,bans_lurker_tests_tested=0i,bans_obj=0i,bans_obj_killed=0i,bans_persisted_bytes=16i,bans_persisted_fragmentation=0i,bans_req=0i,bans_tested=0i,bans_tests_tested=0i,busy_killed=0i,busy_sleep=0i,busy_wakeup=0i,cache_hit=643999i,cache_hit_grace=22i,cache_hitmiss=0i,cache_hitpass=0i,cache_miss=1i,client_req=644000i,client_req_400=0i,client_req_417=0i,client_resp_500=0i,esi_errors=0i,esi_warnings=0i,exp_mailed=37i,exp_received=37i,fetch_1xx=0i,fetch_204=0i,fetch_304=2i,fetch_bad=0i,fetch_chunked=6i,fetch_eof=0i,fetch_failed=0i,fetch_head=0i,fetch_length=11i,fetch_no_thread=0i,fetch_none=0i,hcb_insert=1i,hcb_lock=1i,hcb_nolock=644000i,losthdr=0i,n_backend=19i,n_expired=1i,n_gunzip=289204i,n_gzip=0i,n_lru_limited=0i,n_lru_moved=843i,n_lru_nuked=0i,n_obj_purged=0i,n_object=0i,n_objectcore=40i,n_objecthead=40i,n_purges=0i,n_test_gunzip=6i,n_vampireobject=0i,n_vcl=7i,n_vcl_avail=7i,n_vcl_discard=0i,pools=2i,req_dropped=0i,s_fetch=1i,s_pass=0i,s_pipe=0i,s_pipe_hdrbytes=0i,s_pipe_in=0i,s_pipe_out=0i,s_req_bodybytes=0i,s_req_hdrbytes=54740000i,s_resp_bodybytes=341618192i,s_resp_hdrbytes=190035576i,s_sess=651038i,s_synth=0i,sc_overload=0i,sc_pipe_overflow=0i,sc_range_short=0i,sc_rem_close=7038i,sc_req_close=0i,sc_req_http10=644000i,sc_req_http20=0i,sc_resp_close=0i,sc_rx_bad=0i,sc_rx_body=0i,sc_rx_junk=0i,sc_rx_overflow=0i,sc_rx_timeout=0i,sc_tx_eof=0i,sc_tx_error=0i,sc_tx_pipe=0i,sc_vcl_failure=0i,sess_closed=644000i,sess_closed_err=644000i,sess_conn=651038i,sess_drop=0i,sess_dropped=0i,sess_fail=0i,sess_fail_ebadf=0i,sess_fail_econnaborted=0i,sess_fail_eintr=0i,sess_fail_emfile=0i,sess_fail_enomem=0i,sess_fail_other=0i,sess_herd=11i,sess_queued=0i,sess_readahead=0i,shm_cont=3572i,shm_cycles=10i,shm_flushes=0i,shm_records=30727866i,shm_writes=4661979i,summs=2225754i,thread_queue_len=0i,threads=200i,threads_created=200i,threads_destroyed=0i,threads_failed=0i,threads_limited=0i,uptime=4416326i,vcl_fail=0i,vmods=2i,ws_backend_overflow=0i,ws_client_overflow=0i,ws_session_overflow=0i,ws_thread_overflow=0i 1631121675000000000
> varnish,backend=default,host=kozel.local,section=VBE bereq_bodybytes=0i,bereq_hdrbytes=0i,beresp_bodybytes=0i,beresp_hdrbytes=0i,busy=0i,conn=0i,fail=0i,fail_eacces=0i,fail_eaddrnotavail=0i,fail_econnrefused=0i,fail_enetunreach=0i,fail_etimedout=0i,fail_other=0i,happy=9223372036854775807i,helddown=0i,pipe_hdrbytes=0i,pipe_in=0i,pipe_out=0i,req=0i,unhealthy=0i 1631121675000000000
> varnish,backend=server1,host=kozel.local,section=VBE bereq_bodybytes=0i,bereq_hdrbytes=0i,beresp_bodybytes=0i,beresp_hdrbytes=0i,busy=0i,conn=0i,fail=0i,fail_eacces=0i,fail_eaddrnotavail=0i,fail_econnrefused=30609i,fail_enetunreach=0i,fail_etimedout=0i,fail_other=0i,happy=0i,helddown=3i,pipe_hdrbytes=0i,pipe_in=0i,pipe_out=0i,req=0i,unhealthy=0i 1631121675000000000
> varnish,backend=server2,host=kozel.local,section=VBE bereq_bodybytes=0i,bereq_hdrbytes=0i,beresp_bodybytes=0i,beresp_hdrbytes=0i,busy=0i,conn=0i,fail=0i,fail_eacces=0i,fail_eaddrnotavail=0i,fail_econnrefused=30609i,fail_enetunreach=0i,fail_etimedout=0i,fail_other=0i,happy=0i,helddown=3i,pipe_hdrbytes=0i,pipe_in=0i,pipe_out=0i,req=0i,unhealthy=0i 1631121675000000000
> varnish,backend=server_test1,host=kozel.local,section=VBE bereq_bodybytes=0i,bereq_hdrbytes=0i,beresp_bodybytes=0i,beresp_hdrbytes=0i,busy=0i,conn=0i,fail=0i,fail_eacces=0i,fail_eaddrnotavail=0i,fail_econnrefused=49345i,fail_enetunreach=0i,fail_etimedout=0i,fail_other=0i,happy=0i,helddown=2i,pipe_hdrbytes=0i,pipe_in=0i,pipe_out=0i,req=0i,unhealthy=0i 1631121675000000000
```
