# Sematext Output Plugin

The Sematext output plugin writes metrics to [Sematext](https://sematext.com/spm/).
Check the [docs](https://sematext.com/docs/monitoring) for more info.

### Configuration:

```toml
# Sematext output config
[[outputs.sematext]]
  ## Docs at https://sematext.com/docs/monitoring provide info about getting
  ## started with Sematext monitoring.

  ## URL of your Sematext metrics receiver. US-region metrics receiver is used
  ## in this example (it is also the default when receiver_url value is empty),
  ## but address of e.g. Sematext EU-region metrics receiver can be used
  ## instead.
  receiver_url = "https://spm-receiver.sematext.com"

  ## Token of the App to which the data is sent. Create an App of appropriate
  ## type in Sematext UI, instructions will show its token which can be used
  ## here.
  token = ""

  ## Optional TLS Config.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Optional flag for ignoring tls certificate check.
  # insecure_skip_verify = false
```
