# OpenSMTPD Input Plugin

This plugin gathers stats from [OpenSMTPD - a FREE implementation of the server-side SMTP protocol](https://www.opensmtpd.org/)

### Configuration:

```toml
 # A plugin to collect stats from OpenSMTPD - a FREE implementation of the server-side SMTP protocol
 [[inputs.smtpctl]]
   ## If running as a restricted user you can prepend sudo for additional access:
   #use_sudo = false

   ## The default location of the smtpctl binary can be overridden with:
   binary = "/usr/sbin/smtpctl"

   # The default timeout of 1000ms can be overriden with (in milliseconds):
   #timeout = 1000

  ## By default, telegraf gather stats for 4 metric points.
  ## Setting stats will override the defaults shown below.
  ## Glob matching can be used, ie, stats = ["total.*"]
  ## stats may also be set to ["*"], which will collect all stats
  ## except histogram.* statistics that will never be collected.
  stats = ["total.*", "num.*","time.up", "mem.*"]
```

### Measurements & Fields:

This is the full list of stats provided by smtpctl and potentially collected by telegram
depending of your smtpctl configuration.

- smtpctl
    control_session=2
    mda_envelope=0
    mda_pending=0
    mda_running=0
    mda_user=0
    queue_evpcache_load_hit=2
    queue_evpcache_size=1
    scheduler_delivery_ok=1
    scheduler_envelope=0
    scheduler_envelope_incoming=1
    scheduler_envelope_inflight=0
    scheduler_ramqueue_envelope=1
    scheduler_ramqueue_message=1
    scheduler_ramqueue_update=1
    smtp_session=1
    smtp_session_local=2
    uptime=21

### Permissions:

It's important to note that this plugin references smtpctl, which may require additional permissions to execute successfully.
Depending on the user/group permissions of the telegraf user executing this plugin, you may need to alter the group membership, set facls, or use sudo.

**Group membership (Recommended)**:
```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G opensmtpd telegraf

$ groups telegraf
telegraf : telegraf opensmtpd
```

**Sudo privileges**:
If you use this method, you will need the following in your telegraf config:
```toml
[[inputs.opensmtpd]]
  use_sudo = true
```

You will also need to update your sudoers file:
```bash
$ visudo
# Add the following line:
telegraf ALL=(ALL) NOPASSWD: /usr/sbin/smtpctl
```

Please use the solution you see as most appropriate.

### Example Output:

```
 telegraf --config etc/telegraf.conf --input-filter smtpctl --test
* Plugin: inputs.smtpctl, Collection 1
> smtpctl,host=localhost total_num_cachehits=0,total_num_prefetch=0,total_requestlist_avg=0,total_requestlist_max=0,total_recursion_time_median=0,total_num_queries=0,total_requestlist_overwritten=0,total_requestlist_current_all=0,time_up=159185.583967,total_num_recursivereplies=0,total_requestlist_exceeded=0,total_requestlist_current_user=0,total_recursion_time_avg=0,total_tcpusage=0,total_num_cachemiss=0 1510130793000000000

```
