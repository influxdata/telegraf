# HTTP Output Plugin

This plugin sends metrics in a HTTP message encoded using one of the output data
formats. For data_formats that support batching, metrics are sent in batch
format by default.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# A plugin that can transmit metrics over HTTP
[[outputs.http]]
  ## URL is the address to send metrics to
  url = "http://127.0.0.1:8080/telegraf"

  ## Timeout for HTTP message
  # timeout = "5s"

  ## HTTP method, one of: "POST" or "PUT" or "PATCH"
  # method = "POST"

  ## HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## OAuth2 Client Credentials Grant
  # client_id = "clientid"
  # client_secret = "secret"
  # token_url = "https://indentityprovider/oauth2/v1/token"
  # audience = ""
  # scopes = ["urn:opc:idm:__myscopes__"]

  ## Goole API Auth
  # google_application_credentials = "/etc/telegraf/example_secret.json"

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

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"

  ## Use batch serialization format (default) instead of line based format.
  ## Batch format is more efficient and should be used unless line based
  ## format is really needed.
  # use_batch_format = true

  ## HTTP Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## Additional HTTP headers
  # [outputs.http.headers]
  #   # Should be set manually to "application/json" for json data_format
  #   Content-Type = "text/plain; charset=utf-8"

  ## MaxIdleConns controls the maximum number of idle (keep-alive)
  ## connections across all hosts. Zero means no limit.
  # max_idle_conn = 0

  ## MaxIdleConnsPerHost, if non-zero, controls the maximum idle
  ## (keep-alive) connections to keep per-host. If zero,
  ## DefaultMaxIdleConnsPerHost is used(2).
  # max_idle_conn_per_host = 2

  ## Idle (keep-alive) connection timeout.
  ## Maximum amount of time before idle connection is closed.
  ## Zero means no limit.
  # idle_conn_timeout = 0

  ## Amazon Region
  #region = "us-east-1"

  ## Amazon Credentials
  ## Amazon Credentials are not built unless the following aws_service
  ## setting is set to a non-empty string. It may need to match the name of
  ## the service output to as well
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

  ## Optional list of statuscodes (<200 or >300) upon which requests should not be retried
  # non_retryable_statuscodes = [409, 413]
```

### Google API Auth

The `google_application_credentials` setting is used with Google Cloud APIs.
It specifies the json key file. To learn about creating Google service accounts,
consult Google's [oauth2 service account documentation][create_service_account].
An example use case is a metrics proxy deployed to Cloud Run. In this example,
the service account must have the "run.routes.invoke" permission.

[create_service_account]: https://cloud.google.com/docs/authentication/production#create_service_account

### Optional Cookie Authentication Settings

The optional Cookie Authentication Settings will retrieve a cookie from the
given authorization endpoint, and use it in subsequent API requests.  This is
useful for services that do not provide OAuth or Basic Auth authentication,
e.g. the [Tesla Powerwall API][powerwall], which uses a Cookie Auth Body to
retrieve an authorization cookie.  The Cookie Auth Renewal interval will renew
the authorization by retrieving a new cookie at the given interval.

[powerwall]: https://www.tesla.com/support/energy/powerwall/own/monitoring-from-home-network
