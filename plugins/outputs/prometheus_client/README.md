# Prometheus Client Service Output Plugin

This plugin starts a [Prometheus](https://prometheus.io/) Client, it exposes all metrics on `/metrics` (default) to be polled by a Prometheus server.

## Configuration

```toml
# Publish all metrics to /metrics for Prometheus to scrape
[[outputs.prometheus_client]]
  ## Address to listen on.
  listen = ":9273"

  ## Use HTTP Basic Authentication.
  # basic_username = "Foo"
  # basic_password = "Bar"

  ## If set, the IP Ranges which are allowed to access metrics.
  ##   ex: ip_range = ["192.168.0.0/24", "192.168.1.0/30"]
  # ip_range = []

  ## Path to publish the metrics on.
  # path = "/metrics"

  ## Expiration interval for each metric. 0 == no expiration
  # expiration_interval = "60s"

  ## Collectors to enable, valid entries are "gocollector" and "process".
  ## If unset, both are enabled.
  # collectors_exclude = ["gocollector", "process"]

  ## Send string metrics as Prometheus labels.
  ## Unless set to false all string metrics will be sent as labels.
  # string_as_label = true

  ## If set, enable TLS with the given certificate.
  # tls_cert = "/etc/ssl/telegraf.crt"
  # tls_key = "/etc/ssl/telegraf.key"
  
  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Use only tls ciphers defined in this list
  ## Possible values:
  ## TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
  ## TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
  ## TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
  ## TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
  ## TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
  ## TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
  ## TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
  ## TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
  ## TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
  ## TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA
  ## TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
  ## TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
  ## TLS_RSA_WITH_AES_128_GCM_SHA256
  ## TLS_RSA_WITH_AES_256_GCM_SHA384
  ## TLS_RSA_WITH_AES_128_CBC_SHA256
  ## TLS_RSA_WITH_AES_128_CBC_SHA
  ## TLS_RSA_WITH_AES_256_CBC_SHA
  ## TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA
  ## TLS_RSA_WITH_3DES_EDE_CBC_SHA
  ## TLS_RSA_WITH_RC4_128_SHA
  ## TLS_ECDHE_RSA_WITH_RC4_128_SHA
  ## TLS_ECDHE_ECDSA_WITH_RC4_128_SHA
  ## If value wasn't defined default will be used
  # tls_cipher_suites = ["TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305", "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256", "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA", "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA", "TLS_RSA_WITH_AES_128_GCM_SHA256", "TLS_RSA_WITH_AES_256_GCM_SHA384", "TLS_RSA_WITH_AES_128_CBC_SHA256", "TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_RSA_WITH_AES_256_CBC_SHA"]

  ## contains the minimum SSL/TLS version that is acceptable.
  ## If not set, then TLS 1.0 is taken as the minimum.
  # tls_min_version = "TLS11"

  ## contains the maximum SSL/TLS version that is acceptable.
  ## If not set, then the maximum version supported by this package is used,
  ## which is currently TLS 1.2 (for go < 1.12) or TLS 1.3 (for go >= 1.12).
  # tls_max_version = "TLS12"

  ## Export metric collection time.
  # export_timestamp = false
```
