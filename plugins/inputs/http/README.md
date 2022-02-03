# HTTP Input Plugin

The HTTP input plugin collects metrics from one or more HTTP(S) endpoints.  The endpoint should have metrics formatted in one of the supported [input data formats](../../../docs/DATA_FORMATS_INPUT.md).  Each data format has its own unique set of configuration options which can be added to the input configuration.

## Configuration

```toml
# Read formatted metrics from one or more HTTP endpoints
[[inputs.http]]
  ## One or more URLs from which to read formatted metrics
  urls = [
    "http://localhost/metrics"
  ]

  ## HTTP method
  # method = "GET"

  ## Optional HTTP headers
  # headers = {"X-Special-Header" = "Special-Value"}

  ## HTTP entity-body to send with POST/PUT requests.
  # body = ""

  ## HTTP Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## Optional file with Bearer token
  ## file content is added as an Authorization header
  # bearer_token = "/path/to/file"

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## OAuth2 Client Credentials. The options 'client_id', 'client_secret', and 'token_url' are required to use OAuth2.
  # client_id = "clientid"
  # client_secret = "secret"
  # token_url = "https://indentityprovider/oauth2/v1/token"
  # scopes = ["urn:opc:idm:__myscopes__"]

  ## HTTP Proxy support
  # http_proxy_url = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional Cookie authentication
  # cookie_auth_url = "https://localhost/authMe"
  # cookie_auth_method = "POST"
  # cookie_auth_username = "username"
  # cookie_auth_password = "pa$$word"
  # cookie_auth_headers = '{"Content-Type": "application/json", "X-MY-HEADER":"hello"}'
  # cookie_auth_body = '{"username": "user", "password": "pa$$word", "authenticate": "me"}'
  ## cookie_auth_renewal not set or set to "0" will auth once and never renew the cookie
  # cookie_auth_renewal = "5m"

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## List of success status codes
  # success_status_codes = [200]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"

```

## Metrics

The metrics collected by this input plugin will depend on the configured `data_format` and the payload returned by the HTTP endpoint(s).

The default values below are added if the input format does not specify a value:

- http
  - tags:
    - url

## Optional Cookie Authentication Settings

The optional Cookie Authentication Settings will retrieve a cookie from the given authorization endpoint, and use it in subsequent API requests.  This is useful for services that do not provide OAuth or Basic Auth authentication, e.g. the [Tesla Powerwall API](https://www.tesla.com/support/energy/powerwall/own/monitoring-from-home-network), which uses a Cookie Auth Body to retrieve an authorization cookie.  The Cookie Auth Renewal interval will renew the authorization by retrieving a new cookie at the given interval.
