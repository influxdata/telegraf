# HTTP Output Plugin

This plugin sends metrics in a HTTP message encoded using one of the output
data formats. For data_formats that support batching, metrics are sent in batch format.

### Configuration:

```toml
# A plugin that can transmit metrics over HTTP
[[outputs.http]]
  ## URL is the address to send metrics to
  url = "http://127.0.0.1:8080/telegraf"

  ## Timeout for HTTP message
  # timeout = "5s"

  ## HTTP method, one of: "POST" or "PUT"
  # method = "POST"

  ## HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## OAuth2 Client Credentials Grant
  # client_id = "clientid"
  # client_secret = "secret"
  # token_url = "https://indentityprovider/oauth2/v1/token"
  # scopes = ["urn:opc:idm:__myscopes__"]

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
  # cookie_auth_body = '{"username": "user", "password": "pa$$word", "authenticate": "me"}'
  ## cookie_auth_renewal not set or set to "0" will auth once and never renew the cookie
  # cookie_auth_renewal = "5m"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"

  ## HTTP Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## Additional HTTP headers
  # [outputs.http.headers]
  #   # Should be set manually to "application/json" for json data_format
  #   Content-Type = "text/plain; charset=utf-8"

  ## Idle (keep-alive) connection timeout.
  ## Maximum amount of time before idle connection is closed.
  ## Zero means no limit.
  # idle_conn_timeout = 0
```

### Optional Cookie Authentication Settings:

The optional Cookie Authentication Settings will retrieve a cookie from the given authorization endpoint, and use it in subsequent API requests.  This is useful for services that do not provide OAuth or Basic Auth authentication, e.g. the [Tesla Powerwall API](https://www.tesla.com/support/energy/powerwall/own/monitoring-from-home-network), which uses a Cookie Auth Body to retrieve an authorization cookie.  The Cookie Auth Renewal interval will renew the authorization by retrieving a new cookie at the given interval.
