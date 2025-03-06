# Dovecot Input Plugin

This plugin uses the Dovecot [v2.1 stats protocol][stats] to gather
metrics on configured domains of [Dovecot][dovecot] servers. You can use this
plugin on Dovecot up to and including version v2.3.x.

> [!IMPORTANT]
> Dovecot v2.4+ has the old protocol removed and this plugin will not work.
> Please use Dovecot's [Openmetrics exporter][openmetrics] in combination with
> the [http input plugin][http_plugin] and `openmetrics` data format for newer
> versions of Dovecot.

⭐ Telegraf v0.10.3
🏷️ server
💻 all

[dovecot]: https://www.dovecot.org/
[stats]: https://doc.dovecot.org/configuration_manual/stats/old_statistics/#old-statistics
[http_plugin]: /plugins/inputs/http/README.md
[openmetrics]: https://doc.dovecot.org/latest/core/config/statistics.html#openmetrics

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about dovecot servers
[[inputs.dovecot]]
  ## specify dovecot servers via an address:port list
  ##  e.g.
  ##    localhost:24242
  ## or as an UDS socket
  ##  e.g.
  ##    /var/run/dovecot/old-stats
  ##
  ## If no servers are specified, then localhost is used as the host.
  servers = ["localhost:24242"]

  ## Type is one of "user", "domain", "ip", or "global"
  type = "global"

  ## Wildcard matches like "*.com". An empty string "" is same as "*"
  ## If type = "ip" filters should be <IP/network>
  filters = [""]
```

## Metrics

- dovecot
  - tags:
    - server (hostname)
    - type (query type)
    - ip (ip addr)
    - user (username)
    - domain (domain name)
  - fields:
    - reset_timestamp (string)
    - last_update (string)
    - num_logins (integer)
    - num_cmds (integer)
    - num_connected_sessions (integer)
    - user_cpu (float)
    - sys_cpu (float)
    - clock_time (float)
    - min_faults (integer)
    - maj_faults (integer)
    - vol_cs (integer)
    - invol_cs (integer)
    - disk_input (integer)
    - disk_output (integer)
    - read_count (integer)
    - read_bytes (integer)
    - write_count (integer)
    - write_bytes (integer)
    - mail_lookup_path (integer)
    - mail_lookup_attr (integer)
    - mail_read_count (integer)
    - mail_read_bytes (integer)
    - mail_cache_hits (integer)

## Example Output

```text
dovecot,server=dovecot-1.domain.test,type=global clock_time=101196971074203.94,disk_input=6493168218112i,disk_output=17978638815232i,invol_cs=1198855447i,last_update="2016-04-08 11:04:13.000379245 +0200 CEST",mail_cache_hits=68192209i,mail_lookup_attr=0i,mail_lookup_path=653861i,mail_read_bytes=86705151847i,mail_read_count=566125i,maj_faults=17208i,min_faults=1286179702i,num_cmds=917469i,num_connected_sessions=8896i,num_logins=174827i,read_bytes=30327690466186i,read_count=1772396430i,reset_timestamp="2016-04-08 10:28:45 +0200 CEST",sys_cpu=157965.692,user_cpu=219337.48,vol_cs=2827615787i,write_bytes=17150837661940i,write_count=992653220i 1460106266642153907
```
