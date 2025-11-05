# Docker Secrets Secret-Store Plugin

The `gdchauth` plugin allows to fetch token for a service account file and STS audience.

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret-store support` section, it will detail which options support secret
store usage.

## Configuration

```toml @sample.conf
# Secret-store to retrieve secrets from Google Cloud Authenticator
[secretstores.gdchauth]
  id = "gdchauth_secret"
  ## Path to the GDCH service account JSON key file
  service_account_file = "/etc/telegraf/service-account.json"
  audience = "https://{SERVICE_URL}"
  
  ##Time before token expiry to fetch a new one.
  # token_expiry_buffer = "5m"
```

### Referencing Secret within a Plugin

Referencing the secret within a plugin occurs by:

```toml
[[inputs.http]]
  token = "@{gdchauth_secret:token}"
```

## Additional Information