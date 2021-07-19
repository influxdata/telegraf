# Varnish Input Plugin

This plugin gathers stats from [Varnish HTTP Cache](https://varnish-cache.org/)

### Configuration:

```toml
[[inputs.varnish_cache]]
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the varnishstat binary can be overridden with:
  binary = "/usr/bin/varnishstat"
  
  ## Optional command line arguments	
  # args = ["-j"]

  ## Optional name for the varnish instance (or working directory) to query
  ## Usually append after -n in varnish cli
  # instance_name = instanceName

  ## Timeout for varnishstat command
  # timeout = "1s"
```

### Measurements & Fields:

The plugin runs `varnishstat -j` command and parses the JSON output into metrics. 
Varnish stats will be grouped into measurements by their lowercase prefix. For better organization
`varnish_` prefix will be appended.

Suffix after the last "." is a field name. The middle part splits into tags.

Examples:
Varnish counter:
```
  "MAIN.cache_hit": {
    "description": "Cache hits",
    "flag": "c", "format": "i",
    "value": 51
  },
```
Influx metric:
measurement: "varnish_main"
field name: "cache_hit"
value: 51

```
"MEMPOOL.req1.live": {
  "description": "In use",
  "flag": "g", "format": "i",
  "value": 39
},
```
Influx metric:
measurement: "varnish_mempool"
field name: "live"
tag "id": "req1"
value: 39


#### Varnish reload
This VarnishCache Telegraf Plugin supports Varnish reloads.  

Note that removed backends after ``systemctrl reload varnish`` will be reported by `varnishstat` until Varnish 
will be restarted. The VarnishCache Plugin uses the most recent version of `reload_`.

Example of metric after reload:
```
  "VBE.reload_20210622_153544_23757.default.happy": {
    "description": "Happy health probes",
    "flag": "b", "format": "b",
    "value": 0
  },
```
The `reload_20210622_153544_23757` part will be stripped from metrics.


### Permissions:

It's important to note that this plugin references varnishstat, which may require additional permissions to execute successfully.
Depending on the user/group permissions of the telegraf user executing this plugin, you may need to alter the group membership, set facls, or use sudo.

**Group membership (Recommended)**:
```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G varnish telegraf

$ groups telegraf
telegraf : telegraf varnish
```

**Extended filesystem ACL's**:
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

###Sudo privileges
If you use this method, you will need the following in your telegraf config:
```toml
[[inputs.varnish]]
  use_sudo = true
```

###Custom arguments
Custom arguments can be used for use cases like running `varnishstat` on remote machine  
```toml
[[inputs.varnish]]
  binary = "/usr/bin/ssh"
  args = ["root@10.100.0.108", "varnishstat", "-j"]
```
or in local docker container
```toml
  binary = "/usr/local/bin/docker"
  args = ["exec", "-t", "varnish_varnish_1", "varnishstat",  "-j"]
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

### Example Output:

```
varnish_main,host=kozel.local sc_req_http10=0 1624553280000000000
varnish_lck,host=kozel.local,id=cli creat=1 1624553280000000000
varnish_lck,host=kozel.local,id=hcb destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcl dbg_try_fail=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 allocs=242 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 toosmall=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local helddown=0 1624553280000000000
varnish_main,host=kozel.local backend_unhealthy=0 1624553280000000000
varnish_main,host=kozel.local backend_fail=0 1624553280000000000
varnish_main,host=kozel.local sess_closed_err=9 1624553280000000000
varnish_main,host=kozel.local esi_errors=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcapace creat=1 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 allocs=17 1624553280000000000
varnish_vbe,backend=default,host=kozel.local conn=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local fail_other=0 1624553280000000000
varnish_main,host=kozel.local backend_busy=0 1624553280000000000
varnish_main,host=kozel.local ws_session_overflow=0 1624553280000000000
varnish_main,host=kozel.local bans_deleted=0 1624553280000000000
varnish_main,host=kozel.local exp_mailed=7 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj recycle=7 1624553280000000000
varnish_lck,host=kozel.local,id=sma dbg_busy=0 1624553280000000000
varnish_main,host=kozel.local sess_fail_ebadf=0 1624553280000000000
varnish_main,host=kozel.local threads_limited=0 1624553280000000000
varnish_sma,host=kozel.local,id=s0 c_fail=0 1624553280000000000
varnish_sma,host=kozel.local,id=transient g_space=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 frees=242 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local fail_enetunreach=0 1624553280000000000
varnish_main,host=kozel.local cache_hit_grace=0 1624553280000000000
varnish_main,host=kozel.local req_dropped=0 1624553280000000000
varnish_main,host=kozel.local backend_req=7 1624553280000000000
varnish_lck,host=kozel.local,id=cli destroy=0 1624553280000000000
varnish_mgt,host=kozel.local child_panic=0 1624553280000000000
varnish_main,host=kozel.local cache_hitmiss=0 1624553280000000000
varnish_main,host=kozel.local n_objecthead=3 1624553280000000000
varnish_main,host=kozel.local vmods=2 1624553280000000000
varnish_main,host=kozel.local n_test_gunzip=6 1624553280000000000
varnish_sma,host=kozel.local,id=s0 g_space=268435456 1624553280000000000
varnish_sma,host=kozel.local,id=transient c_freed=0 1624553280000000000
varnish_main,host=kozel.local sess_dropped=0 1624553280000000000
varnish_lck,host=kozel.local,id=vbe destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=vxid creat=1 1624553280000000000
varnish_main,host=kozel.local sess_herd=243 1624553280000000000
varnish_main,host=kozel.local ws_client_overflow=0 1624553280000000000
varnish_lck,host=kozel.local,id=pipestat destroy=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local fail=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local fail_etimedout=0 1624553280000000000
varnish_main,host=kozel.local client_req_417=0 1624553280000000000
varnish_main,host=kozel.local bans_lurker_tested=0 1624553280000000000
varnish_lck,host=kozel.local,id=tcp_pool dbg_busy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 timeout=2 1624553280000000000
varnish_sma,host=kozel.local,id=transient g_alloc=0 1624553280000000000
varnish_main,host=kozel.local client_req_400=0 1624553280000000000
varnish_main,host=kozel.local threads_created=200 1624553280000000000
varnish_main,host=kozel.local bans=1 1624553280000000000
varnish_lck,host=kozel.local,id=lru dbg_busy=0 1624553280000000000
varnish_lck,host=kozel.local,id=wstat dbg_busy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj sz_actual=65504 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj surplus=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 timeout=7 1624553280000000000
varnish_main,host=kozel.local sc_rx_overflow=0 1624553280000000000
varnish_main,host=kozel.local n_vcl_avail=3 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 recycle=242 1624553280000000000
varnish_main,host=kozel.local busy_sleep=0 1624553280000000000
varnish_main,host=kozel.local bans_obj_killed=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcapace dbg_busy=0 1624553280000000000
varnish_main,host=kozel.local cache_hit=259 1624553280000000000
varnish_main,host=kozel.local sc_req_http20=0 1624553280000000000
varnish_lck,host=kozel.local,id=lru creat=2 1624553280000000000
varnish_lck,host=kozel.local,id=tcp_pool destroy=0 1624553280000000000
varnish_sma,host=kozel.local,id=s0 c_req=20 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local helddown=0 1624553280000000000
varnish_mgt,host=kozel.local uptime=27851 1624553280000000000
varnish_main,host=kozel.local sess_drop=0 1624553280000000000
varnish_main,host=kozel.local client_req=266 1624553280000000000
varnish_main,host=kozel.local thread_queue_len=0 1624553280000000000
varnish_main,host=kozel.local sc_rx_junk=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local fail=0 1624553280000000000
varnish_main,host=kozel.local bans_persisted_bytes=16 1624553280000000000
varnish_lck,host=kozel.local,id=busyobj creat=10 1624553280000000000
varnish_lck,host=kozel.local,id=cli dbg_busy=0 1624553280000000000
varnish_lck,host=kozel.local,id=vbe creat=1 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 surplus=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 recycle=17 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local pipe_in=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local fail_econnrefused=0 1624553280000000000
varnish_main,host=kozel.local bans_completed=1 1624553280000000000
varnish_main,host=kozel.local n_purges=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcapace dbg_try_fail=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local bereq_hdrbytes=0 1624553280000000000
varnish_main,host=kozel.local sess_fail_eintr=0 1624553280000000000
varnish_main,host=kozel.local n_lru_nuked=0 1624553280000000000
varnish_main,host=kozel.local s_pass=0 1624553280000000000
varnish_main,host=kozel.local sc_range_short=0 1624553280000000000
varnish_main,host=kozel.local bans_lurker_tests_tested=0 1624553280000000000
varnish_main,host=kozel.local n_gunzip=0 1624553280000000000
varnish_lck,host=kozel.local,id=wq destroy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 pool=10 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local req=0 1624553280000000000
varnish_mgt,host=kozel.local child_died=0 1624553280000000000
varnish_main,host=kozel.local backend_recycle=6 1624553280000000000
varnish_main,host=kozel.local fetch_bad=0 1624553280000000000
varnish_main,host=kozel.local bans_tested=0 1624553280000000000
varnish_main,host=kozel.local bans_lurker_obj_killed_cutoff=0 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj randry=0 1624553280000000000
varnish_sma,host=kozel.local,id=transient c_bytes=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local bereq_hdrbytes=470 1624553280000000000
varnish_main,host=kozel.local shm_records=50961 1624553280000000000
varnish_lck,host=kozel.local,id=hcb locks=169 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 recycle=8 1624553280000000000
varnish_main,host=kozel.local sc_tx_pipe=0 1624553280000000000
varnish_lck,host=kozel.local,id=exp dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=vbe dbg_busy=0 1624553280000000000
varnish_main,host=kozel.local sc_overload=0 1624553280000000000
varnish_main,host=kozel.local client_resp_500=0 1624553280000000000
varnish_lck,host=kozel.local,id=backend dbg_busy=0 1624553280000000000
varnish_lck,host=kozel.local,id=exp locks=8743 1624553280000000000
varnish_lck,host=kozel.local,id=tcp_pool locks=5482 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 sz_actual=65504 1624553280000000000
varnish_vbe,backend=default,host=kozel.local fail_enetunreach=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local fail_eaddrnotavail=0 1624553280000000000
varnish_main,host=kozel.local fetch_no_thread=0 1624553280000000000
varnish_main,host=kozel.local threads=200 1624553280000000000
varnish_main,host=kozel.local s_pipe_in=0 1624553280000000000
varnish_main,host=kozel.local sc_rem_close=16 1624553280000000000
varnish_lck,host=kozel.local,id=sma dbg_try_fail=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local fail_eacces=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local fail_other=0 1624553280000000000
varnish_main,host=kozel.local fetch_head=0 1624553280000000000
varnish_lck,host=kozel.local,id=ban destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=hcb dbg_try_fail=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 frees=17 1624553280000000000
varnish_sma,host=kozel.local,id=s0 g_alloc=0 1624553280000000000
varnish_main,host=kozel.local n_object=0 1624553280000000000
varnish_main,host=kozel.local s_resp_hdrbytes=86255 1624553280000000000
varnish_main,host=kozel.local hcb_insert=7 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local conn=0 1624553280000000000
varnish_main,host=kozel.local summs=35838 1624553280000000000
varnish_main,host=kozel.local n_vampireobject=0 1624553280000000000
varnish_main,host=kozel.local bans_lurker_obj_killed=0 1624553280000000000
varnish_lck,host=kozel.local,id=lru dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcl dbg_busy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj sz_wanted=65536 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local unhealthy=0 1624553280000000000
varnish_lck,host=kozel.local,id=ban locks=1144 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local req=0 1624553280000000000
varnish_main,host=kozel.local sess_fail_enomem=0 1624553280000000000
varnish_main,host=kozel.local fetch_chunked=6 1624553280000000000
varnish_main,host=kozel.local s_req_bodybytes=0 1624553280000000000
varnish_main,host=kozel.local sc_pipe_overflow=0 1624553280000000000
varnish_lck,host=kozel.local,id=pipestat dbg_busy=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcapace destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=wstat dbg_try_fail=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local pipe_out=0 1624553280000000000
varnish_main,host=kozel.local backend_conn=5 1624553280000000000
varnish_main,host=kozel.local losthdr=0 1624553280000000000
varnish_sma,host=kozel.local,id=transient c_req=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local fail=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local fail_econnrefused=0 1624553280000000000
varnish_main,host=kozel.local n_gzip=0 1624553280000000000
varnish_lck,host=kozel.local,id=backend destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=lru destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=sess creat=25 1624553280000000000
varnish_lck,host=kozel.local,id=sess locks=39 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 live=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 randry=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 pool=10 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 sz_actual=480 1624553280000000000
varnish_vbe,backend=default,host=kozel.local busy=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local fail_enetunreach=0 1624553280000000000
varnish_main,host=kozel.local shm_cont=0 1624553280000000000
varnish_lck,host=kozel.local,id=objhdr dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=pipestat dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=waiter dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=wstat destroy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 live=0 1624553280000000000
varnish_lck,host=kozel.local,id=sma locks=40 1624553280000000000
varnish_main,host=kozel.local fetch_none=0 1624553280000000000
varnish_main,host=kozel.local fetch_failed=0 1624553280000000000
varnish_main,host=kozel.local ws_thread_overflow=0 1624553280000000000
varnish_main,host=kozel.local bans_lurker_contention=0 1624553280000000000
varnish_lck,host=kozel.local,id=busyobj locks=55 1624553280000000000
varnish_lck,host=kozel.local,id=pipestat locks=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcl destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcl locks=48 1624553280000000000
varnish_vbe,backend=default,host=kozel.local helddown=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local bereq_bodybytes=0 1624553280000000000
varnish_lck,host=kozel.local,id=ban creat=1 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 surplus=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local pipe_hdrbytes=0 1624553280000000000
varnish_main,host=kozel.local n_vcl=3 1624553280000000000
varnish_lck,host=kozel.local,id=busyobj dbg_busy=0 1624553280000000000
varnish_lck,host=kozel.local,id=exp destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=sess destroy=25 1624553280000000000
varnish_lck,host=kozel.local,id=tcp_pool dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=vbe dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcl creat=1 1624553280000000000
varnish_lck,host=kozel.local,id=waiter creat=2 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj live=0 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj allocs=7 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 recycle=17 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local pipe_hdrbytes=0 1624553280000000000
varnish_main,host=kozel.local sess_fail_other=0 1624553280000000000
varnish_main,host=kozel.local s_fetch=7 1624553280000000000
varnish_main,host=kozel.local sess_closed=0 1624553280000000000
varnish_lck,host=kozel.local,id=backend locks=53022 1624553280000000000
varnish_lck,host=kozel.local,id=cli locks=9306 1624553280000000000
varnish_lck,host=kozel.local,id=exp creat=1 1624553280000000000
varnish_lck,host=kozel.local,id=mempool locks=123979 1624553280000000000
varnish_lck,host=kozel.local,id=objhdr dbg_busy=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local fail_eacces=0 1624553280000000000
varnish_main,host=kozel.local fetch_204=0 1624553280000000000
varnish_main,host=kozel.local threads_failed=0 1624553280000000000
varnish_main,host=kozel.local s_req_hdrbytes=147137 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj timeout=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 allocs=17 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local happy=18446744073709552000 1624553280000000000
varnish_mgt,host=kozel.local child_dump=0 1624553280000000000
varnish_main,host=kozel.local sess_fail_emfile=0 1624553280000000000
varnish_main,host=kozel.local s_pipe_out=0 1624553280000000000
varnish_lck,host=kozel.local,id=ban dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=cli dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=objhdr destroy=7 1624553280000000000
varnish_sma,host=kozel.local,id=transient g_bytes=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local bereq_bodybytes=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local pipe_in=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local busy=0 1624553280000000000
varnish_main,host=kozel.local shm_cycles=0 1624553280000000000
varnish_main,host=kozel.local bans_tests_tested=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 timeout=8 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local unhealthy=0 1624553280000000000
varnish_main,host=kozel.local sc_resp_close=0 1624553280000000000
varnish_main,host=kozel.local bans_obj=0 1624553280000000000
varnish_main,host=kozel.local bans_persisted_fragmentation=0 1624553280000000000
varnish_main,host=kozel.local hcb_lock=7 1624553280000000000
varnish_lck,host=kozel.local,id=objhdr locks=595 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj frees=7 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 frees=17 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local beresp_bodybytes=0 1624553280000000000
varnish_main,host=kozel.local fetch_1xx=0 1624553280000000000
varnish_main,host=kozel.local fetch_304=0 1624553280000000000
varnish_main,host=kozel.local pools=2 1624553280000000000
varnish_lck,host=kozel.local,id=hcb creat=1 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 pool=10 1624553280000000000
varnish_main,host=kozel.local fetch_eof=1 1624553280000000000
varnish_lck,host=kozel.local,id=wq dbg_busy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 sz_wanted=512 1624553280000000000
varnish_sma,host=kozel.local,id=s0 c_bytes=118925 1624553280000000000
varnish_sma,host=kozel.local,id=s0 c_freed=118925 1624553280000000000
varnish_sma,host=kozel.local,id=s0 g_bytes=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 pool=10 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 sz_actual=65504 1624553280000000000
varnish_vbe,backend=default,host=kozel.local pipe_out=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local fail_eacces=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local beresp_hdrbytes=0 1624553280000000000
varnish_lck,host=kozel.local,id=lru locks=29 1624553280000000000
varnish_lck,host=kozel.local,id=vxid dbg_try_fail=0 1624553280000000000
varnish_main,host=kozel.local backend_reuse=2 1624553280000000000
varnish_main,host=kozel.local sess_readahead=0 1624553280000000000
varnish_main,host=kozel.local hcb_nolock=266 1624553280000000000
varnish_lck,host=kozel.local,id=vbe locks=26982 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 toosmall=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local beresp_hdrbytes=258 1624553280000000000
varnish_lck,host=kozel.local,id=sess dbg_busy=0 1624553280000000000
varnish_lck,host=kozel.local,id=sess dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=vxid locks=3 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 sz_actual=480 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 allocs=8 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 surplus=0 1624553280000000000
varnish_main,host=kozel.local sess_fail_econnaborted=0 1624553280000000000
varnish_main,host=kozel.local busy_killed=0 1624553280000000000
varnish_main,host=kozel.local sc_rx_bad=0 1624553280000000000
varnish_main,host=kozel.local ws_backend_overflow=0 1624553280000000000
varnish_main,host=kozel.local shm_writes=36922 1624553280000000000
varnish_lck,host=kozel.local,id=sma creat=2 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local busy=0 1624553280000000000
varnish_main,host=kozel.local fetch_length=0 1624553280000000000
varnish_main,host=kozel.local sc_vcl_failure=0 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj pool=10 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 sz_wanted=65536 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local pipe_hdrbytes=0 1624553280000000000
varnish_main,host=kozel.local cache_hitpass=0 1624553280000000000
varnish_lck,host=kozel.local,id=ban dbg_busy=0 1624553280000000000
varnish_sma,host=kozel.local,id=transient c_fail=0 1624553280000000000
varnish_main,host=kozel.local n_expired=7 1624553280000000000
varnish_lck,host=kozel.local,id=tcp_pool creat=4 1624553280000000000
varnish_lck,host=kozel.local,id=waiter locks=1539 1624553280000000000
varnish_mempool,host=kozel.local,id=busyobj toosmall=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 randry=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local req=1 1624553280000000000
varnish_vbe,backend=default,host=kozel.local fail_econnrefused=0 1624553280000000000
varnish_main,host=kozel.local cache_miss=7 1624553280000000000
varnish_main,host=kozel.local n_objectcore=3 1624553280000000000
varnish_main,host=kozel.local s_pipe=0 1624553280000000000
varnish_main,host=kozel.local s_synth=0 1624553280000000000
varnish_main,host=kozel.local sc_rx_timeout=9 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 sz_wanted=65536 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local beresp_bodybytes=0 1624553280000000000
varnish_main,host=kozel.local s_pipe_hdrbytes=0 1624553280000000000
varnish_main,host=kozel.local bans_dups=0 1624553280000000000
varnish_lck,host=kozel.local,id=backend dbg_try_fail=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 live=0 1624553280000000000
varnish_lck,host=kozel.local,id=sma destroy=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local happy=18446744073709552000 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local pipe_out=0 1624553280000000000
varnish_mgt,host=kozel.local child_stop=0 1624553280000000000
varnish_main,host=kozel.local n_obj_purged=0 1624553280000000000
varnish_lck,host=kozel.local,id=busyobj dbg_try_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=vcapace locks=0 1624553280000000000
varnish_lck,host=kozel.local,id=wq creat=3 1624553280000000000
varnish_lck,host=kozel.local,id=wstat creat=1 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 frees=8 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 sz_wanted=512 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 randry=0 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local fail_etimedout=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local happy=18446744073709552000 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local beresp_hdrbytes=0 1624553280000000000
varnish_mgt,host=kozel.local child_exit=0 1624553280000000000
varnish_main,host=kozel.local sess_fail=0 1624553280000000000
varnish_main,host=kozel.local n_lru_limited=0 1624553280000000000
varnish_lck,host=kozel.local,id=exp dbg_busy=0 1624553280000000000
varnish_main,host=kozel.local bans_req=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local fail_eaddrnotavail=0 1624553280000000000
varnish_main,host=kozel.local sc_req_close=0 1624553280000000000
varnish_lck,host=kozel.local,id=objhdr creat=11 1624553280000000000
varnish_main,host=kozel.local threads_destroyed=0 1624553280000000000
varnish_lck,host=kozel.local,id=wstat locks=27343 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local fail_etimedout=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local fail_other=0 1624553280000000000
varnish_main,host=kozel.local n_backend=9 1624553280000000000
varnish_main,host=kozel.local shm_flushes=0 1624553280000000000
varnish_main,host=kozel.local n_vcl_discard=0 1624553280000000000
varnish_main,host=kozel.local bans_added=1 1624553280000000000
varnish_main,host=kozel.local esi_warnings=0 1624553280000000000
varnish_lck,host=kozel.local,id=waiter destroy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req1 toosmall=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local pipe_in=0 1624553280000000000
varnish_main,host=kozel.local vcl_fail=0 1624553280000000000
varnish_lck,host=kozel.local,id=hcb dbg_busy=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local beresp_bodybytes=384 1624553280000000000
varnish_vbe,backend=server1,host=kozel.local bereq_hdrbytes=0 1624553280000000000
varnish_main,host=kozel.local sess_queued=0 1624553280000000000
varnish_main,host=kozel.local sc_rx_body=0 1624553280000000000
varnish_main,host=kozel.local sc_tx_eof=0 1624553280000000000
varnish_lck,host=kozel.local,id=mempool dbg_try_fail=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 surplus=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 randry=0 1624553280000000000
varnish_vbe,backend=default,host=kozel.local unhealthy=0 1624553280000000000
varnish_main,host=kozel.local sc_tx_error=0 1624553280000000000
varnish_main,host=kozel.local exp_received=7 1624553280000000000
varnish_lck,host=kozel.local,id=mempool creat=5 1624553280000000000
varnish_lck,host=kozel.local,id=vxid dbg_busy=0 1624553280000000000
varnish_lck,host=kozel.local,id=wq dbg_try_fail=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local conn=0 1624553280000000000
varnish_main,host=kozel.local n_lru_moved=15 1624553280000000000
varnish_lck,host=kozel.local,id=backend creat=10 1624553280000000000
varnish_lck,host=kozel.local,id=vxid destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=wq locks=99839 1624553280000000000
varnish_lck,host=kozel.local,id=busyobj destroy=7 1624553280000000000
varnish_vbe,backend=default,host=kozel.local fail_eaddrnotavail=0 1624553280000000000
varnish_vbe,backend=server2,host=kozel.local bereq_bodybytes=0 1624553280000000000
varnish_main,host=kozel.local uptime=27851 1624553280000000000
varnish_main,host=kozel.local backend_retry=0 1624553280000000000
varnish_main,host=kozel.local busy_wakeup=0 1624553280000000000
varnish_main,host=kozel.local s_resp_bodybytes=9501 1624553280000000000
varnish_lck,host=kozel.local,id=mempool destroy=0 1624553280000000000
varnish_lck,host=kozel.local,id=pipestat creat=1 1624553280000000000
varnish_lck,host=kozel.local,id=waiter dbg_busy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=req0 timeout=1 1624553280000000000
varnish_mgt,host=kozel.local child_start=1 1624553280000000000
varnish_main,host=kozel.local sess_conn=25 1624553280000000000
varnish_main,host=kozel.local s_sess=25 1624553280000000000
varnish_lck,host=kozel.local,id=mempool dbg_busy=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess0 toosmall=0 1624553280000000000
varnish_mempool,host=kozel.local,id=sess1 live=0 1624553280000000000

```

You can merge metrics together into a metric with multiple fields into the most 
memory and network transfer efficient form using `aggregators.merge`
```toml
[[aggregators.merge]]
  drop_original = true
```

The output will be:
```
varnish_mgt,host=kozel.local child_dump=0,child_stop=0,uptime=28330,child_panic=0,child_died=0,child_exit=0,child_start=1 1624553760000000000
varnish_main,host=kozel.local shm_records=51569,backend_retry=0,bans_tests_tested=0,losthdr=0,uptime=28331,fetch_1xx=0,threads=200,s_resp_hdrbytes=86255,exp_mailed=7,fetch_none=0,sc_tx_pipe=0,shm_cycles=0,backend_req=7,vcl_fail=0,backend_unhealthy=0,n_vcl=3,bans_obj=0,bans_lurker_contention=0,sess_fail_eintr=0,fetch_head=0,sc_resp_close=0,n_test_gunzip=6,threads_destroyed=0,thread_queue_len=0,client_req=266,backend_fail=0,sc_req_http10=0,client_resp_500=0,shm_writes=37530,sc_rx_timeout=9,sc_tx_eof=0,bans=1,sess_conn=25,backend_reuse=2,bans_added=1,n_purges=0,hcb_lock=7,sess_fail_econnaborted=0,fetch_eof=1,summs=36414,sess_fail=0,fetch_failed=0,cache_hit_grace=0,ws_thread_overflow=0,s_pass=0,sess_fail_emfile=0,cache_hitmiss=0,cache_miss=7,n_backend=9,sess_dropped=0,n_vcl_discard=0,esi_errors=0,backend_recycle=6,sc_req_http20=0,sc_rx_body=0,sc_rx_overflow=0,backend_conn=5,n_vampireobject=0,s_pipe_hdrbytes=0,s_pipe=0,s_pipe_in=0,sc_range_short=0,backend_busy=0,sess_herd=243,bans_lurker_tested=0,busy_wakeup=0,sc_rx_bad=0,ws_backend_overflow=0,sess_drop=0,s_fetch=7,s_pipe_out=0,n_gunzip=0,n_objectcore=3,sc_rem_close=16,bans_persisted_fragmentation=0,fetch_204=0,sc_pipe_overflow=0,hcb_insert=7,n_gzip=0,sc_vcl_failure=0,bans_deleted=0,bans_obj_killed=0,n_obj_purged=0,sess_closed=0,sess_fail_other=0,busy_sleep=0,shm_flushes=0,vmods=2,n_lru_moved=15,n_lru_limited=0,fetch_length=0,pools=2,threads_created=200,threads_failed=0,bans_persisted_bytes=16,bans_dups=0,sess_fail_ebadf=0,bans_req=0,bans_tested=0,ws_client_overflow=0,s_resp_bodybytes=9501,fetch_bad=0,sess_queued=0,n_objecthead=3,n_vcl_avail=3,esi_warnings=0,sc_req_close=0,bans_completed=1,hcb_nolock=266,cache_hit=259,s_req_hdrbytes=147137,bans_lurker_obj_killed=0,bans_lurker_obj_killed_cutoff=0,exp_received=7,cache_hitpass=0,n_expired=7,n_lru_nuked=0,s_req_bodybytes=0,sc_overload=0,busy_killed=0,req_dropped=0,sess_closed_err=9,sess_fail_enomem=0,ws_session_overflow=0,bans_lurker_tests_tested=0,s_sess=25,sc_tx_error=0,shm_cont=0,client_req_400=0,client_req_417=0,fetch_no_thread=0,n_object=0,fetch_chunked=6,threads_limited=0,sess_readahead=0,s_synth=0,fetch_304=0,sc_rx_junk=0 1624553760000000000
varnish_lck,host=kozel.local,id=hcb destroy=0,creat=1,locks=172,dbg_try_fail=0,dbg_busy=0 1624553760000000000
varnish_vbe,backend=server1,host=kozel.local fail_other=0,pipe_out=0,beresp_hdrbytes=0,conn=0,busy=0,bereq_hdrbytes=0,beresp_bodybytes=0,pipe_hdrbytes=0,req=0,fail_eacces=0,fail=0,happy=18446744073709552000,fail_enetunreach=0,pipe_in=0,fail_eaddrnotavail=0,bereq_bodybytes=0,unhealthy=0,fail_etimedout=0,helddown=0,fail_econnrefused=0 1624553760000000000
varnish_vbe,backend=default,host=kozel.local beresp_hdrbytes=258,pipe_out=0,fail_eaddrnotavail=0,busy=0,unhealthy=0,fail_enetunreach=0,pipe_hdrbytes=0,fail_eacces=0,fail_etimedout=0,conn=0,helddown=0,happy=18446744073709552000,fail_other=0,bereq_bodybytes=0,beresp_bodybytes=384,fail_econnrefused=0,bereq_hdrbytes=470,fail=0,pipe_in=0,req=1 1624553760000000000
varnish_lck,host=kozel.local,id=vcapace destroy=0,dbg_busy=0,locks=0,dbg_try_fail=0,creat=1 1624553760000000000
varnish_mempool,host=kozel.local,id=sess0 live=0,toosmall=0,randry=0,allocs=8,frees=8,sz_actual=480,surplus=0,sz_wanted=512,timeout=2,recycle=8,pool=10 1624553760000000000
varnish_mempool,host=kozel.local,id=req1 pool=10,randry=0,sz_actual=65504,allocs=242,sz_wanted=65536,toosmall=0,surplus=0,timeout=8,frees=242,live=0,recycle=242 1624553760000000000
varnish_mempool,host=kozel.local,id=req0 toosmall=0,live=0,sz_wanted=65536,allocs=17,frees=17,sz_actual=65504,surplus=0,recycle=17,pool=10,timeout=1,randry=0 1624553760000000000
varnish_lck,host=kozel.local,id=sma locks=40,dbg_try_fail=0,creat=2,destroy=0,dbg_busy=0 1624553760000000000
varnish_vbe,backend=server2,host=kozel.local helddown=0,req=0,conn=0,beresp_bodybytes=0,busy=0,fail_eaddrnotavail=0,fail_etimedout=0,fail_other=0,pipe_in=0,pipe_hdrbytes=0,fail_econnrefused=0,beresp_hdrbytes=0,happy=18446744073709552000,fail_eacces=0,unhealthy=0,fail=0,bereq_hdrbytes=0,bereq_bodybytes=0,pipe_out=0,fail_enetunreach=0 1624553760000000000
varnish_lck,host=kozel.local,id=exp dbg_try_fail=0,creat=1,dbg_busy=0,locks=8896,destroy=0 1624553760000000000
varnish_lck,host=kozel.local,id=mempool creat=5,locks=126104,dbg_busy=0,destroy=0,dbg_try_fail=0 1624553760000000000
varnish_lck,host=kozel.local,id=tcp_pool destroy=0,creat=4,dbg_busy=0,dbg_try_fail=0,locks=5482 1624553760000000000
varnish_mempool,host=kozel.local,id=busyobj sz_actual=65504,frees=7,randry=0,pool=10,toosmall=0,recycle=7,surplus=0,allocs=7,sz_wanted=65536,timeout=0,live=0 1624553760000000000
varnish_lck,host=kozel.local,id=wq dbg_busy=0,dbg_try_fail=0,creat=3,destroy=0,locks=101475 1624553760000000000
varnish_mempool,host=kozel.local,id=sess1 timeout=7,toosmall=0,live=0,recycle=17,frees=17,randry=0,pool=10,allocs=17,sz_wanted=512,surplus=0,sz_actual=480 1624553760000000000
varnish_lck,host=kozel.local,id=lru creat=2,destroy=0,locks=29,dbg_try_fail=0,dbg_busy=0 1624553760000000000
varnish_lck,host=kozel.local,id=busyobj destroy=7,locks=55,dbg_try_fail=0,dbg_busy=0,creat=10 1624553760000000000
varnish_sma,host=kozel.local,id=transient c_bytes=0,c_fail=0,c_req=0,g_alloc=0,c_freed=0,g_space=0,g_bytes=0 1624553760000000000
varnish_lck,host=kozel.local,id=wstat creat=1,destroy=0,locks=27796,dbg_busy=0,dbg_try_fail=0 1624553760000000000
varnish_sma,host=kozel.local,id=s0 g_bytes=0,c_req=20,c_freed=118925,c_fail=0,g_alloc=0,c_bytes=118925,g_space=268435456 1624553760000000000
varnish_lck,host=kozel.local,id=sess dbg_try_fail=0,creat=25,destroy=25,locks=39,dbg_busy=0 1624553760000000000
varnish_lck,host=kozel.local,id=vbe destroy=0,creat=1,dbg_busy=0,dbg_try_fail=0,locks=27430 1624553760000000000
varnish_lck,host=kozel.local,id=vcl locks=48,creat=1,dbg_busy=0,destroy=0,dbg_try_fail=0 1624553760000000000
varnish_lck,host=kozel.local,id=pipestat creat=1,locks=0,dbg_try_fail=0,dbg_busy=0,destroy=0 1624553760000000000
varnish_lck,host=kozel.local,id=waiter creat=2,destroy=0,dbg_try_fail=0,dbg_busy=0,locks=1549 1624553760000000000
varnish_lck,host=kozel.local,id=vxid locks=3,creat=1,destroy=0,dbg_try_fail=0,dbg_busy=0 1624553760000000000
varnish_lck,host=kozel.local,id=ban dbg_try_fail=0,creat=1,dbg_busy=0,locks=1162,destroy=0 1624553760000000000
varnish_lck,host=kozel.local,id=cli destroy=0,locks=9466,creat=1,dbg_try_fail=0,dbg_busy=0 1624553760000000000
varnish_lck,host=kozel.local,id=backend dbg_try_fail=0,creat=10,destroy=0,locks=53886,dbg_busy=0 1624553760000000000
varnish_lck,host=kozel.local,id=objhdr dbg_try_fail=0,destroy=7,creat=11,locks=595,dbg_busy=0 1624553760000000000
```