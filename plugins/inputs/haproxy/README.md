# HAProxy Input Plugin

This plugin gathers statistics of [HAProxy][haproxy] servers using sockets or
the HTTP protocol.

⭐ Telegraf v0.1.5
🏷️ network, server
💻 all

[haproxy]: http://www.haproxy.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

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

## Metrics

For more details about collected metrics reference the [HAProxy CSV format
documentation][6].

- haproxy
  - tags:
    - `server` - address of the server data was gathered from
    - `proxy` - proxy name
    - `sv` - service name
    - `type` - proxy session type
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
haproxy,server=/run/haproxy/admin.sock,proxy=public,sv=FRONTEND,type=frontend http_response.other=0i,req_rate_max=1i,comp_byp=0i,status="OPEN",rate_lim=0i,dses=0i,req_rate=0i,comp_rsp=0i,bout=9287i,comp_in=0i,mode="http",smax=1i,slim=2000i,http_response.1xx=0i,conn_rate=0i,dreq=0i,ereq=0i,iid=2i,rate_max=1i,http_response.2xx=1i,comp_out=0i,intercepted=1i,stot=2i,pid=1i,http_response.5xx=1i,http_response.3xx=0i,http_response.4xx=0i,conn_rate_max=1i,conn_tot=2i,dcon=0i,bin=294i,rate=0i,sid=0i,req_tot=2i,scur=0i,dresp=0i 1513293519000000000
```
