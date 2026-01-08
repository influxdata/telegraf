# SIP Input Plugin

The SIP (Session Initiation Protocol) input plugin monitors the health and
availability of SIP servers such as PBX systems, SIP proxies, registrars, and
VoIP service providers. It sends SIP requests (typically OPTIONS) and measures
response times and status codes.

⭐ Telegraf v1.38.0
🏷️ network
💻 all

This plugin is particularly useful for:

- Monitoring VoIP infrastructure availability
- Measuring SIP service response times
- Verifying SIP server connectivity
- Alerting on SIP service degradation

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# SIP (Session Initiation Protocol) health check plugin
[[inputs.sip]]
  ## List of SIP servers to monitor (RFC 3261 compliant SIP URIs)
  ## Formats:
  ##   sip://host:port (uses UDP transport, default port 5060)
  ##   sips://host:port (uses TLS transport, default port 5061)
  ##   sip://host:port;transport=tcp (explicit transport parameter)
  ##   sip://host:port;transport=udp
  ##   sip://host:port;transport=ws (WebSocket)
  ##   sips://host:port;transport=wss (Secure WebSocket)
  ##
  ## Examples:
  servers = [
    "sip://sip.example.com:5060",              # UDP (default)
    "sips://secure.example.com:5061",          # TLS
    "sip://192.168.1.100:5060;transport=tcp"   # TCP with explicit parameter
  ]

  ## SIP method to use for health checks
  ## Valid values: OPTIONS, INVITE, MESSAGE
  ## Default: OPTIONS (recommended for health checks)
  # method = "OPTIONS"

  ## Request timeout
  ## Default: 5s
  # timeout = "5s"

  ## From user (appears in SIP From header)
  ## Default: telegraf
  # from_user = "telegraf"

  ## From domain (domain part of From header)
  ## If not specified, uses the server hostname
  # from_domain = "monitoring.example.com"

  ## To user (appears in SIP To/Request URI)
  ## If not specified, uses the same value as from_user
  # to_user = ""

  ## User-Agent string
  ## Default: Telegraf SIP Monitor
  # user_agent = "Telegraf SIP Monitor"

  ## Expected SIP response code
  ## The response code is compared to this value. If they match,
  ## the field "response_code_match" will be 1, otherwise it will be 0.
  ## If set to 0 (default), any 2xx response is considered success.
  ## Common values: 200 (OK), 404 (Not Found), 407 (Proxy Auth Required)
  # expect_code = 200

  ## Local address to use for outgoing requests
  ## Leave empty to use system default
  # local_address = ""

  ## Optional SIP digest authentication credentials
  ## If the SIP server responds with 401 Unauthorized or 407 Proxy Authentication
  ## Required, the plugin will automatically attempt digest authentication using
  ## these credentials. The credentials are never included in the sip_uri tag.
  ## Note: Leave empty if the server does not require authentication.
  # username = "user"
  # password = "pa$$word"

  ## Optional TLS Config (only used for sips:// URLs or ;transport=tls/wss)
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

## Metrics

- sip
  - tags:
    - server (the SIP server address)
    - method (the SIP method used: OPTIONS, INVITE, MESSAGE)
    - transport (the transport protocol: udp, tcp, tls, ws, wss)
    - sip_uri (the complete SIP URI used in the request)
    - status_code (the SIP response status code, e.g., "200", "404")
    - result (result type: success, timeout, connection_failed, etc.)
    - reason (SIP response reason phrase, e.g., "OK", "Not Found")
    - server_agent (optional: the Server header from the response)
  - fields:
    - response_time (float, seconds) - Time taken to receive response
    - response_code_match (integer, 0 or 1) - Whether response code matched
      expected value
    - result_type (string) - Type of result (matches result tag)
    - result_code (integer) - Numeric result code for categorization

### Result Types and Codes

| Result Type            | Result Code | Description                                             |
| ---------------------- | ----------- | ------------------------------------------------------- |
| success                | 0           | Request completed successfully                          |
| response_code_mismatch | 1           | Response received but code doesn't match expected       |
| timeout                | 2           | Request timed out waiting for response                  |
| connection_refused     | 3           | Connection refused by server                            |
| connection_failed      | 4           | Connection failed (general network error)               |
| no_response            | 5           | Transaction completed but no response received          |
| parse_error            | 6           | Failed to parse server address                          |
| request_error          | 7           | Failed to create SIP request                            |
| transaction_error      | 8           | SIP transaction error                                   |
| no_route               | 9           | No route to host                                        |
| network_unreachable    | 10          | Network is unreachable                                  |
| error_response         | 11          | Received error response (4xx, 5xx, 6xx)                 |
| auth_required          | 12          | Authentication required but no credentials provided     |
| auth_failed            | 13          | Authentication attempt failed                           |
| auth_error             | 14          | Error retrieving authentication credentials             |

## SIP Methods

The plugin supports the following SIP methods:

- **OPTIONS** (recommended): Standard SIP method for health checks. Queries
  server capabilities without establishing a session.
- **INVITE**: Initiates a session. Use with caution as it may create call
  records.
- **MESSAGE**: Sends an instant message. Useful for testing messaging
  infrastructure.

## Server URL Format (RFC 3261 Compliant)

The plugin uses RFC 3261 compliant SIP URI format for server addresses:

### Basic Formats

- `sip://host:port` - Standard SIP with UDP transport (default port 5060)
- `sips://host:port` - Secure SIP with TLS transport (default port 5061)

### With Explicit Transport Parameters

- `sip://host:port;transport=udp` - UDP transport (explicit)
- `sip://host:port;transport=tcp` - TCP transport
- `sip://host:port;transport=ws` - WebSocket transport
- `sips://host:port;transport=wss` - Secure WebSocket transport

**Note:** Per RFC 3261, the use of `;transport=tls` is deprecated. Use the
`sips://` URI scheme instead to indicate TLS transport.

### Examples

```toml
[[inputs.sip]]
  servers = [
    "sip://sip.example.com:5060",              # UDP (default)
    "sips://secure.example.com:5061",          # TLS
    "sip://192.168.1.100:5060;transport=tcp",  # TCP with parameter
    "sip://ws.example.com:5060;transport=ws",  # WebSocket
    "sips://wss.example.com:5061;transport=wss" # Secure WebSocket
  ]
```

## TLS/SSL Configuration

When using `sips://` URLs or `;transport=wss`, the plugin supports all standard
Telegraf TLS configuration options:

- Certificate verification with custom CA
- Client certificate authentication
- Skip certificate verification (for testing)
- SNI (Server Name Indication) configuration
- TLS version constraints (1.2, 1.3)
- TLS renegotiation control

Example TLS configuration:

```toml
[[inputs.sip]]
  servers = ["sips://secure.example.com:5061"]
  tls_ca = "/etc/telegraf/ca.pem"
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"
  tls_server_name = "sip.example.com"
```

## Example Output

```text
sip,host=telegraf-host,method=OPTIONS,result=success,server=sip://sip.example.com:5060,sip_uri=sip:telegraf@sip.example.com:5060,status_code=200,transport=udp,reason=OK response_code_match=1i,response_time=0.023,result_code=0i,result_type="success" 1640000000000000000
sip,host=telegraf-host,method=OPTIONS,result=timeout,server=sip://unreachable.example.com:5060,transport=udp response_time=5.001,result_code=2i,result_type="timeout" 1640000000000000000
sip,host=telegraf-host,method=OPTIONS,result=response_code_mismatch,server=sip://sip.provider.com:5060,sip_uri=sip:telegraf@sip.provider.com:5060,status_code=404,transport=udp,reason="Not Found" response_code_match=0i,response_time=0.045,result_code=1i,result_type="response_code_mismatch" 1640000000000000000
sip,host=telegraf-host,method=OPTIONS,result=success,server=sips://secure.voip.example.com:5061,sip_uri=sips:telegraf@secure.voip.example.com:5061,status_code=200,transport=tls,reason=OK response_code_match=1i,response_time=0.067,result_code=0i,result_type="success" 1640000000000000000
```

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

### Response Code Mismatches

Different SIP servers may respond with different status codes:

- Some may respond with 200 OK to OPTIONS requests
- Others may respond with 404 Not Found if the user doesn't exist
- Some may require authentication (401 or 407)
- Adjust `expect_code` based on your server's behavior

## Implementation Details

This plugin uses the [sipgo](https://github.com/emiago/sipgo) library, a
high-performance SIP stack for Go that supports all standard SIP transports and
operations.
