# ProxySQL Input Plugin

This plugin gathers the statistic data from ProxySQL server.

* Connection Pool
* Commands Counters
* Global Stats

### Configuration

```toml
[[inputs.proxysql]]
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    servers = ["user:passwd@tcp(127.0.0.1:6032)/"]
  ##    servers = ["user@tcp(127.0.0.1:6032)/"]
  #
  ## If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:6032)/"]

  ## Selects the metric output format.
  ##
  ## This option exists to maintain backwards compatibility, if you have
  ## existing metrics do not set or change this value until you are ready to
  ## migrate to the new format.
  ##
  ## If you do not have existing metrics from this plugin set to the latest
  ## version.
  ##
  ## Telegraf >=1.6: metric_version = 2
  ##           <1.6: metric_version = 1 (or unset)
  metric_version = 2

  ## the limits for metrics form perf_events_statements
  perf_events_statements_digest_text_limit  = 120
  perf_events_statements_limit              = 250
  perf_events_statements_time_limit         = 86400
  #
  ## gather metrics from stats.stats_mysql_connection_pool
  gather_connection_pool                    = true
  #
  ## gather metrics from stats.stats_mysql_commands_counters
  gather_commands_counters                  = true
  #
  ## gather metrics from stats.stats_mysql_global
  gather_global_stats                       = true
```

ProxySQl use the same MySQL Protocol to get all metrics but in other port number,
in this case use `6032`.
