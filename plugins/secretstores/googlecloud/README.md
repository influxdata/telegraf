# GoogleCloud Secrets Secret-Store Plugin

This plugin allows to retrieve token-based [Google Cloud Credentials][gc_auth].

[gc_auth]: https://docs.cloud.google.com/docs/authentication

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
## Fetch tokens from Google Cloud Authentication
[[secretstores.googlecloud]]
  ## Unique identifier for the secret-store.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:token}(mandatory)
  id = "googlecloud_secret"

  ## Path to the service account credentials file
  credentials_file = "./testdata/gdch.json"

  ## Audience sent to when retrieving an STS token.
  ## Currently only used for GDCH auth flow
  sts_audience = "https://{AUDIENCE_URL}"
```

> [!IMPORTANT]
> This plugin only provides one secret with the key `token`,
> other keys lead to errors.

## Additional Information

To generate a Google-Distributed-Cloud-Hosted service account credentials file
check the [Manage service accounts][gdch_service_docs].

[gdch_service_docs]: https://docs.cloud.google.com/distributed-cloud/hosted/docs/latest/gdch/application/ao-user/iam/service-identities
