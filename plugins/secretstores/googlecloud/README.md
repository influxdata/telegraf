# GoogleCloud Secrets Secret-Store Plugin

The `googlecloud` plugin allows to fetch token from google cloud auth library

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret-store support` section, it will detail which options support secret
store usage.

This plugin currently supports parameters required for getting GDCH credentials.
More parameters can be added based on [GoogleCloud Auth DetectOptions](https://github.com/googleapis/google-cloud-go/blob/main/auth/credentials/detect.go#L154)

## Configuration

```toml @sample.conf
# Secret-store to retrieve secrets from Google Cloud Authenticator
[[secretstores.googlecloud]]
  id = "googlecloud_secret"

  ## Path to the service account JSON key file
  service_account_file = "./testdata/gdch.json"
  sts_audience = "https://{AUDIENCE_URL}"
```

### Referencing Secret within a Plugin

Referencing the secret within a plugin occurs by:

```toml
[[inputs.http]]
  token = "@{googlecloud_secret:token}"
```

## Additional Information

[How to generate GDCH service account file][1]

[Learn about Google Cloud Detect Options][2]

[1]: https://docs.cloud.google.com/distributed-cloud/hosted/docs/latest/gdch/application/ao-user/iam/service-identities

[2]: https://github.com/googleapis/google-cloud-go/blob/main/auth/credentials/detect.go
