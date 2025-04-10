# Icinga2 Input Plugin

This plugin gather services and hosts status information using the
[Icinga2 remote API][remote_api].

⭐ Telegraf v1.8.0
🏷️ network, server, system
💻 all

[remote_api]: https://docs.icinga.com/icinga2/latest/doc/module/icinga2/chapter/icinga2-api

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather Icinga2 status
[[inputs.icinga2]]
  ## Required Icinga2 server address
  # server = "https://localhost:5665"

  ## Collected Icinga2 objects ("services", "hosts")
  ## Specify at least one object to collect from /v1/objects endpoint.
  # objects = ["services"]

  ## Collect metrics from /v1/status endpoint
  ## Choose from:
  ##     "ApiListener", "CIB", "IdoMysqlConnection", "IdoPgsqlConnection"
  # status = []

  ## Credentials for basic HTTP authentication
  # username = "admin"
  # password = "admin"

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
```

## Metrics

- `icinga2_hosts`
  - tags
    - `check_command` - The short name of the check command
    - `display_name` - The name of the host
    - `state` - The state: UP/DOWN
    - `source` - The icinga2 host
    - `port` - The icinga2 port
    - `scheme` - The icinga2 protocol (http/https)
    - `server` - The server the check_command is running for
  - fields
    - `name` (string)
    - `state_code` (int)
- `icinga2_services`
  - tags
    - `check_command` - The short name of the check command
    - `display_name` - The name of the service
    - `state` - The state: OK/WARNING/CRITICAL/UNKNOWN for services
    - `source` - The icinga2 host
    - `port` - The icinga2 port
    - `scheme` - The icinga2 protocol (http/https)
    - `server` - The server the check_command is running for
  - fields
    - `name` (string)
    - `state_code` (int)
- `icinga2_status`
  - component:
    - `ApiListener`
      - tags
        - `component` name
      - fields
        - `api_num_conn_endpoints`
        - `api_num_endpoint`
        - `api_num_http_clients`
        - `api_num_json_rpc_anonymous_clients`
        - `api_num_json_rpc_relay_queue_item_rate`
        - `api_num_json_rpc_relay_queue_items`
        - `api_num_json_rpc_sync_queue_item_rate`
        - `api_num_json_rpc_sync_queue_items`
        - `api_num_json_rpc_work_queue_item_rate`
        - `api_num_not_conn_endpoints`
    - `CIB`
      - tags
        - `component` name
      - fields
        - `active_host_checks`
        - `active_host_checks_15min`
        - `active_host_checks_1min`
        - `active_host_checks_5min`
        - `active_service_checks`
        - `active_service_checks_15min`
        - `active_service_checks_1min`
        - `active_service_checks_5min`
        - `avg_execution_time`
        - `avg_latency`
        - `current_concurrent_checks`
        - `current_pending_callbacks`
        - `max_execution_time`
        - `max_latency`
        - `min_execution_time`
        - `min_latency`
        - `num_hosts_acknowledged`
        - `num_hosts_down`
        - `num_hosts_flapping`
        - `num_hosts_handled`
        - `num_hosts_in_downtime`
        - `num_hosts_pending`
        - `num_hosts_problem`
        - `num_hosts_unreachable`
        - `num_hosts_up`
        - `num_services_acknowledged`
        - `num_services_critical`
        - `num_services_flapping`
        - `num_services_handled`
        - `num_services_in_downtime`
        - `num_services_ok`
        - `num_services_pending`
        - `num_services_problem`
        - `num_services_unknown`
        - `num_services_unreachable`
        - `num_services_warning`
        - `passive_host_checks`
        - `passive_host_checks_15min`
        - `passive_host_checks_1min`
        - `passive_host_checks_5min`
        - `passive_service_checks`
        - `passive_service_checks_15min`
        - `passive_service_checks_1min`
        - `passive_service_checks_5min`
        - `remote_check_queue`
        - `uptime`
    - `IdoMysqlConnection`
      - tags
        - `component` name
      - fields
        - `mysql_queries_1min`
        - `mysql_queries_5mins`
        - `mysql_queries_15mins`
        - `mysql_queries_rate`
        - `mysql_query_queue_item_rate`
        - `mysql_query_queue_items`
    - `IdoPgsqlConnection`
      - tags
        - `component` name
      - fields
        - `pgsql_queries_1min`
        - `pgsql_queries_5mins`
        - `pgsql_queries_15mins`
        - `pgsql_queries_rate`
        - `pgsql_query_queue_item_rate`
        - `pgsql_query_queue_items`

## Sample Queries

```sql
SELECT * FROM "icinga2_services" WHERE state_code = 0 AND time > now() - 24h // Service with OK status
SELECT * FROM "icinga2_services" WHERE state_code = 1 AND time > now() - 24h // Service with WARNING status
SELECT * FROM "icinga2_services" WHERE state_code = 2 AND time > now() - 24h // Service with CRITICAL status
SELECT * FROM "icinga2_services" WHERE state_code = 3 AND time > now() - 24h // Service with UNKNOWN status
```

## Example Output

```text
icinga2_hosts,display_name=router-fr.eqx.fr,check_command=hostalive-custom,host=test-vm,source=localhost,port=5665,scheme=https,state=ok name="router-fr.eqx.fr",state=0 1492021603000000000
```
