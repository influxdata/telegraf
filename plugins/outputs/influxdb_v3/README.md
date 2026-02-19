# InfluxDB v3.x Output Plugin

This plugin writes metrics to a [InfluxDB v3.x][influxdb_v3] Core or Enterprise
instance via the HTTP API.

⭐ Telegraf v1.38.0
🏷️ datastore
💻 all

[influxdb_v3]: https://docs.influxdata.com

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Configuration for sending metrics to InfluxDB 3.x Core and Enterprise
[[outputs.influxdb_v3]]
  ## Multiple URLs can be specified but only ONE of them will be selected
  ## randomly in each interval for writing. If endpoints are unavailable another
  ## one will be used until all are exhausted or the write succeeds.
  urls = ["http://127.0.0.1:8181"]

  ## Token for authentication
  token = ""

  ## Destination database to write into
  database = ""

  ## The value of this tag will be used to determine the database. If this
  ## tag is not set the 'database' option is used as the default.
  # database_tag = ""

  ## If true, the database tag will not be added to the metric
  # exclude_database_tag = false

  ## Wait for WAL persistence to complete synchronization
  ## Setting this to false reduces latency but increases the risk of data loss.
  ## See https://docs.influxdata.com/influxdb3/enterprise/write-data/http-api/v3-write-lp/#use-no_sync-for-immediate-write-responses
  # sync = true

  ## Enable or disable conversion of unsigned integer fields to signed integers
  ## This is useful if existing data exist as signed integers e.g. from previous
  ## versions of InfluxDB.
  # convert_uint_to_int = false

  ## Omit the timestamp of the metrics when sending to allow InfluxDB to set the
  ## timestamp of the data during ingestion. You likely want this to be false
  ## to submit the metric timestamp
  # omit_timestamp = false

  ## HTTP User-Agent
  # user_agent = "telegraf"

  ## Content-Encoding for write request body, available values are "gzip",
  ## "none" and "identity"
  # content_encoding = "gzip"

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
```
