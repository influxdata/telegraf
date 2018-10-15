# ProxySQL Input Plugin

This plugin gathers the statistic data from ProxySQL server

* Global statuses
* Global variables
* Commands Counters
* Connection pool
* Users
* Rules
* Queries
* Memory
* Process list

### Configuration

```toml
# Read metrics from one or many proxysql servers
[[inputs.proxysql]]
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    servers = ["user:passwd@tcp(127.0.0.1:6032)/?tls=false"]
  ##    servers = ["user@tcp(127.0.0.1:6032)/?tls=false"]
  #
  ## If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:6032)/"]
  #
  ## gather metrics from stats_mysql_global
  gather_global                             = true  
  ## gather metrics from stats_mysql_commands_counters
  gather_commands_counters                  = true
  ## gather metrics from stats_mysql_connection_pool
  gather_connection_pool                    = true
  ## gather metrics from stats_mysql_users
  gather_users                              = true
  ## gather metrics from stats_mysql_query_rules
  gather_rules                              = true
  ## gather metrics from stats_mysql_query_digest
  gather_queries                            = true
  ## gather metrics from stats_memory_metrics
  gather_memory_metrics                     = true
  ## gather thread state counts from stats_mysql_processlist
  gather_process_list                       = true
  #
  ## Some queries we may want to run less often (such as SHOW GLOBAL VARIABLES)
  interval_slow                             = "30m"

  ## Optional TLS Config (will be used if tls=custom parameter specified in server uri)
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Metrics:

* Global statuses - all numeric and boolean values of `stats_mysql_global`
* Global variables - all numeric and boolean values of `global_variables`
* Commands Counters - from `stats_mysql_commands_counters`
* Connection pool - from `stats_mysql_connection_pool`
* Users - from `stats_mysql_users`
* Rules - from `stats_mysql_query_rules`
* Queries - from `stats_mysql_query_digest`
* Memory - from `stats_memory_metrics`
* Process list - connection metrics from `stats_mysql_processlist` for each connection status
    * proxysql_process_list(number)

## Tags

* All measurements has following tags
    * server (the host name from which the metrics are gathered)
* Users stats
    * user (username for whom the metrics are gathered)
* Query stats
    * user (username for that did the query)
    * schema_name
    * hostgroup
* Connection pool
    * hostgroup
    * host
* Rules
    * rule_id
