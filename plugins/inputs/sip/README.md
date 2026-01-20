# SIP Input Plugin

This plugin gathers metrics about the health and availability of
[SIP (Session Initiation Protocol)][sip] servers such as PBX systems, SIP
proxies, registrars, and VoIP service providers. It sends SIP requests
(typically OPTIONS) and measures response times and status codes.

⭐ Telegraf v1.38.0
🏷️ network
💻 all

[sip]: https://datatracker.ietf.org/doc/html/rfc3261

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# SIP (Session Initiation Protocol) health check plugin
[[inputs.sip]]
  ## SIP server address to monitor
  ## Format: sip://host[:port] or sips://host[:port]
  ##   sip://  - Standard SIP (default port 5060)
  ##   sips:// - Secure SIP with TLS (default port 5061)
  server = "sip://sip.example.com:5060"

  ## Transport protocol
  ## Valid values: udp, tcp, ws, wss
  ## Note: For TLS, use sips:// scheme instead of transport=tls (per RFC 3261)
  # transport = "udp"

  ## SIP method to use for health checks
  ## Valid values: OPTIONS, INVITE, MESSAGE
  # method = "OPTIONS"

  ## Request timeout
  # timeout = "5s"

  ## From user as it appears in SIP header
  # from_user = "telegraf"

  ## From domain (domain part of From header)
  ## If not specified, uses the server hostname
  # from_domain = ""

  ## To user as it appears in SIP header
  ## If not specified, uses the same value as from_user
  # to_user = ""

  ## Local address to use for outgoing requests
  # local_address = ""

  ## SIP digest authentication credentials
  ## Leave empty to use no authentication
  # username = ""
  # password = ""

  ## Optional TLS Config (only used for sips:// URLs or transport=tls/wss)
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  ## Use the given name as the SNI server name
  # tls_server_name = "sip.example.com"
  ## Minimum TLS version to accept, defaults to TLS 1.2
  # tls_min_version = "TLS12"
  ## Maximum TLS version to accept, defaults to TLS 1.3
  # tls_max_version = "TLS13"
  ## TLS renegotiation method, choose from "never", "once", "freely"
  # tls_renegotiation_method = "never"
```

> [!NOTE]
> Per RFC 3261, the use of `;transport=tls` is deprecated.
> Use the `sips://` URI scheme instead to indicate TLS transport.

This plugin is particularly useful for:

- Monitoring VoIP infrastructure availability
- Measuring SIP service response times
- Verifying SIP server connectivity
- Alerting on SIP service degradation

## Metrics

- sip
  - tags:
    - server (the SIP server address)
    - method (the SIP method used: OPTIONS, INVITE, MESSAGE)
    - transport (the transport protocol: udp, tcp, tls, ws, wss)
    - status_code (the SIP response status code, e.g., "200", "404")
    - server_agent (optional: the Server header from the response)
  - fields:
    - up (int) - Server availability indicator (1 = success, 0 = failure/timeout/error)
    - response_time (float, seconds, optional) - Time taken to receive
      response (not set for timeouts/connection failures)
    - reason (string, optional) - SIP response reason phrase, e.g., "OK", "Not Found"

### SIP Methods

The plugin supports the following SIP methods:

- **OPTIONS** (recommended): Standard SIP method for health checks. Queries
  server capabilities without establishing a session.
- **INVITE**: Initiates a session. Use with caution as it may create call
  records.
- **MESSAGE**: Sends an instant message. Useful for testing messaging
  infrastructure.

## Example Output

```text
sip,host=telegraf-host,method=OPTIONS,server=sip://sip.example.com:5060,status_code=200,transport=udp response_time=0.023,reason="OK",up=1i 1640000000000000000
sip,host=telegraf-host,method=OPTIONS,server=sip://unreachable.example.com:5060,transport=udp reason="Timeout",up=0i 1640000000000000000
sip,host=telegraf-host,method=OPTIONS,server=sip://sip.provider.com:5060,status_code=404,transport=udp response_time=0.045,reason="Not Found",up=0i 1640000000000000000
sip,host=telegraf-host,method=OPTIONS,server=sips://secure.voip.example.com:5061,status_code=200,transport=tcp response_time=0.067,reason="OK",up=1i 1640000000000000000
```

### Troubleshooting

#### Permission Issues

Some SIP implementations may require specific network permissions. If you
encounter permission errors, ensure Telegraf has appropriate network access.

#### Firewall Configuration

Ensure that:

- Outbound connections to SIP ports (typically 5060/5061) are allowed
- If using UDP, firewall allows UDP packets
- Return traffic is permitted for the transaction

#### Timeout Issues

If experiencing frequent timeouts:

- Increase the `timeout` value
- Verify network connectivity to the SIP server
- Check if the SIP server is configured to respond to the chosen method
- Ensure the correct transport protocol is selected

#### Response Codes

Different SIP servers may respond with different status codes to OPTIONS requests:

- `200 OK` - Server is operational and responding
- `404 Not Found` - User or resource doesn't exist (may still indicate healthy server)
- `401 Unauthorized` / `407 Proxy Authentication Required` - Authentication required
