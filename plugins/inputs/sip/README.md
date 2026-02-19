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
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Password for the key file if it is encrypted
  # tls_key_pwd = ""
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## List of ciphers to accept, by default all secure ciphers will be accepted
  ## See https://pkg.go.dev/crypto/tls#pkg-constants for supported values.
  ## Use "all", "secure" and "insecure" to add all support ciphers, secure
  ## suites or insecure suites respectively.
  # tls_cipher_suites = ["secure"]
  ## Renegotiation method, "never", "once" or "freely"
  # tls_renegotiation_method = "never"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### SIP Methods

The plugin supports the following SIP methods:

- **OPTIONS** (recommended): Standard SIP method for health checks. Queries
  server capabilities without establishing a session.
- **INVITE**: Initiates a session. Use with caution as it may create call
  records.
- **MESSAGE**: Sends an instant message. Useful for testing messaging
  infrastructure.

## Troubleshooting

### Permission Issues

Some SIP implementations may require specific network permissions. If you
encounter permission errors, ensure Telegraf has appropriate network access.

### Firewall Configuration

Ensure that:

- Outbound connections to SIP ports (typically 5060/5061) are allowed
- If using UDP, firewall allows UDP packets
- Return traffic is permitted for the transaction

### Timeout Issues

If experiencing frequent timeouts:

- Increase the `timeout` value
- Verify network connectivity to the SIP server
- Check if the SIP server is configured to respond to the chosen method
- Ensure the correct transport protocol is selected

### Response Codes

Different SIP servers may respond with different status codes to OPTIONS requests:

- `200 OK` - Server is operational and responding
- `404 Not Found` - User or resource doesn't exist (may still indicate healthy server)
- `401 Unauthorized` / `407 Proxy Authentication Required` - Authentication required

## Metrics

- sip
  - tags:
    - source (the SIP server address)
    - method (the SIP method used, lowercase: options, invite, message)
    - transport (the transport protocol: udp, tcp, ws, wss)
    - status_code (the SIP response status code, e.g., "200", "404"; not always present, e.g. on timeout)
  - fields:
    - response_time_s (float, seconds) - Time taken to receive response
      (for timeouts, this equals the configured timeout value)
    - result (string) - The outcome of the request: the SIP reason phrase when
      a response is received (e.g. "OK", "Not Found", "Unauthorized"), or a
      sentinel value when no valid response is received (`Timeout`, `Error`,
      `No Response`)
    - server_agent (string, optional) - Value of the `Server` header from the
      SIP response, identifying the remote server software

## Example Output

```text
sip,host=telegraf-host,method=options,source=sip://sip.example.com:5060,status_code=200,transport=udp response_time_s=0.023,result="OK" 1640000000000000000
sip,host=telegraf-host,method=options,source=sip://unreachable.example.com:5060,transport=udp response_time_s=5.0,result="Timeout" 1640000000000000000
sip,host=telegraf-host,method=options,source=sip://sip.provider.com:5060,status_code=404,transport=udp response_time_s=0.045,result="Not Found" 1640000000000000000
sip,host=telegraf-host,method=options,source=sips://secure.voip.example.com:5061,status_code=200,transport=tcp response_time_s=0.067,result="OK",server_agent="Asterisk PBX 18.15.0" 1640000000000000000
```
