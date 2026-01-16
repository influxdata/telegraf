# HTTP Output Plugin

This plugin writes metrics to a HTTP endpoint using one of the supported
[data formats][data_formats]. For data formats supporting batching, metrics are
sent in batches by default.

⭐ Telegraf v1.7.0
🏷️ applications
💻 all

[data_formats]: /docs/DATA_FORMATS_OUTPUT.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username`, `password`
`headers`, and `cookie_auth_headers` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# A plugin that can transmit metrics over HTTP
[[outputs.http]]
  ## URL is the address to send metrics to
  url = "http://127.0.0.1:8080/telegraf"

  ## HTTP method, one of: "POST" or "PUT" or "PATCH"
  # method = "POST"

  ## HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## Goole API Auth
  # google_application_credentials = "/etc/telegraf/example_secret.json"

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## HTTP connection settings
  # idle_conn_timeout = "0s"
  # max_idle_conn = 0
  # max_idle_conn_per_host = 0
  # response_timeout = "0s"

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

  ## NOTE: Due to the way TOML is parsed, tables must be at the END of the
  ## plugin definition, otherwise additional config options are read as part of
  ## the table

  ## Additional HTTP headers
  # [outputs.http.headers]
  #   ## Should be set manually to "application/json" for json data_format
  #   Content-Type = "text/plain; charset=utf-8"
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
