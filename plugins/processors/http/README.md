# HTTP Processor Plugin

This plugin sends each incoming metric to an HTTP endpoint, parses the
response using the configured parser, and emits transformed metrics. It is
useful for enrichment and transformation workflows that require external HTTP
services or to call HTTP APIs where the request data is derived from the incoming
metrics.

⭐ Telegraf v0.0.0 <!-- introduction version TBD -->
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Enrich or replace metrics by sending them to an HTTP endpoint
[[processors.http]]
  ## Required HTTP endpoint to call for each metric
  url = "https://api.example.com/enrich"

  ## HTTP method, one of: "POST", "PUT", or "PATCH"
  # method = "POST"

  ## If true, incoming metrics are not emitted.
  # drop_original = false

  ## Merge Behavior
  ## Possible options are:
  ##  - none: keep the newly parsed metrics as-is
  ##  - override: emit a single metric with all tags and fields of newly parsed
  ##    merged but retaining the first timestamp. If drop_original is
  ##    false, all metrics are merged into the original metric
  ##    NOTE: Existing field or tag values will be overridden.
  ##  - override-with-timestamp: same as "override", but the timestamp is set
  ##    based on the new metrics if present
  ##  - parent: emit one metric per newly parsed metric with each newly parsed
  ##    metric is merged individually into the parent metric keeping the parent
  ##    timestamp
  ##  - parent-with-timestamp: same as "parent", but the timestamp is set
  ##    based on the new metric if present
  # merge = "none"

  ## What to do when the HTTP request or response parsing fails.
  ##   keep - pass the original metric through unchanged
  ##   drop - discard the metric
  # on_error = "keep"

  ## List of acceptable HTTP response status codes
  # success_status_codes = [200]

  ## HTTP Content-Encoding for the request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## Google API Auth
  # google_application_credentials = "/etc/telegraf/example_secret.json"

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## HTTP connection settings
  # idle_conn_timeout = "0s"
  # max_idle_conn = 0
  # max_idle_conn_per_host = 0
  # response_timeout = "0s"

  ## Use the local address for connecting, assigned by the OS by default
  # local_address = ""

  ## Optional proxy settings
  # use_system_proxy = false
  # http_proxy_url = ""

  ## Optional TLS settings
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

  ## OAuth2 Client Credentials. The options 'client_id', 'client_secret', and 'token_url' are required to use OAuth2.
  # client_id = "clientid"
  # client_secret = "secret"
  # token_url = "https://indentityprovider/oauth2/v1/token"
  # audience = ""
  # scopes = ["urn:opc:idm:__myscopes__"]

  ## Optional Cookie authentication
  # cookie_auth_url = "https://localhost/authMe"
  # cookie_auth_method = "POST"
  # cookie_auth_username = "username"
  # cookie_auth_password = "pa$$word"
  # cookie_auth_headers = { Content-Type = "application/json", X-MY-HEADER = "hello" }
  # cookie_auth_body = '{"username": "user", "password": "pa$$word", "authenticate": "me"}'
  ## cookie_auth_renewal not set or set to "0" will auth once and never renew the cookie
  # cookie_auth_renewal = "0s"

  ## Amazon Region
  #region = "us-east-1"

  ## Amazon Credentials
  ## Amazon Credentials are not built unless the following aws_service
  ## setting is set to a non-empty string.
  #aws_service = "execute-api"

  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #web_identity_token_file = ""
  #role_session_name = ""
  #profile = ""
  #shared_credential_file = ""

  ## Data format used for both request serialization and response parsing.
  ## The format must support both serialization and parsing (e.g. "influx", "json").
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## NOTE: Due to the way TOML is parsed, tables must be at the END of the
  ## plugin definition, otherwise additional config options are read as part of
  ## the table

  ## Additional HTTP headers
  # [processors.http.headers]
  #   Content-Type = "application/json"
  #   Authorization = "Bearer ..."
```

### Data format requirements

The configured `data_format` must exist as both a Telegraf serializer and
parser because the processor uses the same format for the request body and
response body.

Examples:

- Supported: `influx`, `json`, and other formats available in both parser and
  serializer registries.
- Rejected: serializer-only or parser-only formats.

## HTTP metadata

The processor annotates metrics with request status information where applicable:

- tags:
  - `status_code` (response status code)
  - `result` ([see below](#result))
- fields:
  - `http_error` (string, set only on failure when `on_error = "keep"`)

On success, emitted metrics include `status_code` and `result=success`.

On failure with `on_error = "keep"`, the original metric is passed through with
`http_error` and `result`. The `status_code` tag is set when an HTTP response
was received.

### `result`

Upon finishing the HTTP request for each metric, the plugin sets the `result`
tag to describe the outcome:

- `success`: acceptable status code and response parsed successfully
- `body_read_error`: response body could not be parsed, or the parser returned
  no metrics
- `connection_failed`: network error not otherwise classified below
- `timeout`: request timed out
- `dns_error`: DNS lookup failed
- `response_status_code_mismatch`: status code not listed in
  `success_status_codes`

## Example

```toml
[[processors.http]]
  namepass = ["ip_lookup"]
  url = "http://ip-api.com/batch"
  method = "POST"
  data_format = "json"
  merge = "parent"
  drop_original = true
  on_error = "keep"
  timeout = "15s"

  # Reshape Telegraf JSON message to API request body...
  # ip-api batch format: [{"query": "<ip_address>"}]
  json_transformation = '[{"query": fields.ip_address}]'

  # List the expected string fields in the API response.
  json_string_fields = [
    "query", "status", "country", "countryCode", "region",
    "regionName", "city", "zip", "isp", "org", "as",
  ]

  [processors.http.headers]
    Content-Type = "application/json"
```
