# Tacacs Input Plugin

The Tacacs plugin collects successful tacacs authentication response times
from tacacs servers such as Aruba ClearPass, FreeRADIUS or tac_plus (TACACS+).
It is primarily meant to monitor how long it takes for the server to fully
handle an auth request, including all potential dependent calls (for example
to AD servers, or other sources of truth for auth the tacacs server uses).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Tacacs plugin collects successful tacacs authentication response times.
[[inputs.tacacs]]
  ## An array of Server IPs (or hostnames) and ports to gather from. If none specified, defaults to localhost.
  # servers = ["127.0.0.1:49"]

  ## Request source server IP, normally the server running telegraf.
  # request_ip = "127.0.0.1"

  ## Credentials for tacacs authentication.
  username = "myuser"
  password = "mypassword"
  secret = "mysecret"

  ## Maximum time to receive response.
  # response_timeout = "5s"
```

## Metrics

- tacacs
  - tags:
    - source
  - fields:
    - response_status (string, [see below](#field-response_status)))
    - responsetime_ms (int64 [see below](#field-responsetime_ms)))

### field `response_status`

The field "response_status" is either a translated raw code returned
by the tacacs server, or filled by telegraf in case of a timeout.

| Field Value          | Raw Code     | From          | responsetime_ms
| -------------------- | ------------ | ------------- | ---------------
| AuthenStatusPass     | 1 (0x1)      | tacacs server | real value
| AuthenStatusFail     | 2 (0x2)      | tacacs server | real value
| AuthenStatusGetData  | 3 (0x3)      | tacacs server | real value
| AuthenStatusGetUser  | 4 (0x4)      | tacacs server | real value
| AuthenStatusGetPass  | 5 (0x5)      | tacacs server | real value
| AuthenStatusRestart  | 6 (0x6)      | tacacs server | real value
| AuthenStatusError    | 7 (0x7)      | tacacs server | real value
| AuthenStatusFollow   | 33 (0x21)    | tacacs server | real value
| Timeout              | Timeout      | telegraf      | eq. to response_timeout

### field `responsetime_ms`

The field responsetime_ms is response time of the tacacs server
in miliseconds of the furthest achieved stage of auth.
In case of timeout, its filled by telegraf to be the value of
the configured response_timeout.

## Example Output

```text
tacacs,source=127.0.0.1:49 responsetime_ms=311i,response_status="AuthenStatusPass" 1677526200000000000
```
