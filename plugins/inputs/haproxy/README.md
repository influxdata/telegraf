# HAProxy Input Plugin

This plugin gathers statistics of [HAProxy][haproxy] servers using sockets or
the HTTP protocol.

⭐ Telegraf v0.1.5
🏷️ network, server
💻 all

[haproxy]: http://www.haproxy.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics of HAProxy, via stats socket or http endpoints
[[inputs.haproxy]]
  ## List of stats endpoints. Metrics can be collected from both http and socket
  ## endpoints. Examples of valid endpoints:
  ##   - http://myhaproxy.com:1936/haproxy?stats
  ##   - https://myhaproxy.com:8000/stats
  ##   - socket:/run/haproxy/admin.sock
  ##   - /run/haproxy/*.sock
  ##   - tcp://127.0.0.1:1936
  ##
  ## Server addresses not starting with 'http://', 'https://', 'tcp://' will be
  ## treated as possible sockets. When specifying local socket, glob patterns are
  ## supported.
  servers = ["http://myhaproxy.com:1936/haproxy?stats"]

  ## By default, some of the fields are renamed from what haproxy calls them.
  ## Setting this option to true results in the plugin keeping the original
  ## field names.
  # keep_field_names = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Master socket support (experimental)
  ## When enabled, the plugin will try to query the HAProxy master socket
  ## for worker PIDs and request statistics from each worker. This helps to
  ## collect metrics from old worker processes that continue handling
  ## connections after a reload.
  # use_master = false
  # master_socket = "/run/haproxy-master.sock" # optional - plugin will try common locations if empty
  # aggregate_workers = false # if true, aggregate (sum) stats across workers per proxy/service
  # add_source_tag = false # if true, adds a `source` tag like "master:pid=1234" or "master:aggregated"
  # concurrency = 4 # max concurrent requests to workers via master socket
```

### HAProxy Configuration

The following information may be useful when getting started, but please consult
the HAProxy documentation for complete and up to date instructions.

The [`stats enable`][4] option can be used to add unauthenticated access over
HTTP using the default settings.  To enable the unix socket begin by reading
about the [`stats socket`][5] option.

[4]: https://cbonte.github.io/haproxy-dconv/1.8/configuration.html#4-stats%20enable
[5]: https://cbonte.github.io/haproxy-dconv/1.8/configuration.html#3.1-stats%20socket

### servers

Server addresses must explicitly start with 'http' if you wish to use HAProxy
status page.  Otherwise, addresses will be assumed to be an UNIX socket and any
protocol (if present) will be discarded.

When using socket names, wildcard expansion is supported so plugin can gather
stats from multiple sockets at once.

To use HTTP Basic Auth add the username and password in the userinfo section of
the URL: `http://user:password@1.2.3.4/haproxy?stats`.  The credentials are sent
via the `Authorization` header and not using the request URL.

### keep_field_names

By default, some of the fields are renamed from what haproxy calls them.
Setting the `keep_field_names` parameter to `true` will result in the plugin
keeping the original field names.

The following renames are made:

- `pxname` -> `proxy`
- `svname` -> `sv`
- `act` -> `active_servers`
- `bck` -> `backup_servers`
- `cli_abrt` -> `cli_abort`
- `srv_abrt` -> `srv_abort`
- `hrsp_1xx` -> `http_response.1xx`
- `hrsp_2xx` -> `http_response.2xx`
- `hrsp_3xx` -> `http_response.3xx`
- `hrsp_4xx` -> `http_response.4xx`
- `hrsp_5xx` -> `http_response.5xx`
- `hrsp_other` -> `http_response.other`

## Master socket and worker aggregation

When HAProxy is reloaded it may spawn new worker processes while old workers
continue handling existing connections. Those old workers expose statistics via
the master socket using `@!<pid> show stat` which allows collecting metrics for
workers that otherwise would be invisible to the admin socket paths. The
plugin can optionally query the master socket and request stats for each
worker process.

- use_master (bool) - enable master-socket based collection. Default: false
- master_socket (string) - path to master socket. If empty the plugin will try
  common locations like `/run/haproxy-master.sock` and `/run/haproxy/master.sock`
- aggregate_workers (bool) - if true, sum stats across workers per proxy/service
- add_source_tag (bool) - if true, adds a `source` tag like `master:pid=1234`
- concurrency (int) - maximum concurrent requests to the master socket (default 4)

## Metrics

For more details about collected metrics reference the [HAProxy CSV format
documentation][6].

- haproxy
  - tags:
    - `server` - address of the server data was gathered from
    - `proxy` - proxy name
    - `sv` - service name
    - `type` - proxy session type
    - `source` - (optional) when gathering via master socket, identifies the
      worker pid or `master:aggregated`
  - fields:
    - `status` (string)
    - `check_status` (string)
    - `last_chk` (string)
    - `mode` (string)
    - `tracked` (string)
    - `agent_status` (string)
    - `last_agt` (string)
    - `addr` (string)
    - `cookie` (string)
    - `lastsess` (int)
    - **all other stats** (int)

[6]: https://cbonte.github.io/haproxy-dconv/1.8/management.html#9.1

## Example Output

```text
haproxy,server=/run/haproxy/admin.sock,proxy=public,sv=FRONTEND,type=frontend http_response.other=0i,req_rate_max=1i,comp_byp=0i,status="OPEN",rate_lim=0i,dses=0i,req_rate=0i,comp_rsp=0i,bout=9287i,co[...]
```

```
