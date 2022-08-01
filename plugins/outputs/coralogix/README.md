# Coralogix Output Plugin

This plugin sends metrics to [Coralogix](https://coralogix.com/) servers
and agents via gRPC.

## Configuration

```toml @sample.conf
# Send metrics over gRPC to Coralogix
[[outputs.coralogix]]
  ## Provide the Coralogix endpoint
  ## address:port
  # service_address = "otel-metrics.coralogix.com:443"

  ## Your Coralogix private key is sensitive
  # private_key: "xxx"
  
  ## Metrics emitted by this plugin should be tagged
  ## in Coralogix with the following application and subsystem names
  # application_name: "$NAMESPACE"
  # subsystem_name: "$HOSTNAME"

  ## Override the default (5s) request timeout
  # timeout = "5s"

  ## Optional TLS Config.
  ##
  ## Root certificates for verifying server certificates encoded in PEM format.
  # tls_ca = "/etc/telegraf/ca.pem"
  ## The public and private keypairs for the client encoded in PEM format.
  ## May contain intermediate certificates.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS, but skip TLS chain and host verification.
  # insecure_skip_verify = false
  ## Send the specified TLS server name via SNI.
  # tls_server_name = "foo.example.com"

  ## Override the default (gzip) compression used to send data.
  ## Supports: "gzip", "none"
  # compression = "gzip"

  ## Additional resource attributes
  # [outputs.coralogix.attributes]
  # "service.name" = "demo"

  ## Additional gRPC request metadata
  # [outputs.coralogix.headers]
  # key1 = "value1"
```

### Schema

The InfluxDB->Coralogix is using same conversion as Opentelemetry output plugin.
