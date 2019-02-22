# Loggregator Telegraf Input Plugin

A [telegraf](https://github.com/influxdata/telegraf) input plugin that supports the 
[loggregator v2 api](https://github.com/cloudfoundry/loggregator-api/tree/master/v2). This agent 
relies on Loggregator's Reverse Log Proxy. It creates a streaming connection to the RLP endpoint
using gRPC. This is meant to be deployed to a single VM.  Scaling to multiple VMs will cause
inconsistencies in your data.

### Configuration:

The provided certificates must be signed by a CA trusted by your Reverse Log Proxy.

```toml
  ## A string path to the tls ca certificate
  tls_ca_path = "/path/to/tls_ca_cert.pem"

  ## A string path to the tls server certificate
  tls_cert_path = "/path/to/tls_cert.pem"

  ## A string path to the tls server private key
  tls_key_path = "/path/to/tls_cert.key"

  ## Boolean value indicating whether or not to skip SSL verification
  insecure_skip_verify = false

  ## A string server name that the certificate is valid for
  rlp_common_name = "foo"
  
  ## A string address of the RLP server to get logs from
  rlp_address = "bar"

  ## A string duration for how frequently to report internal metrics
  internal_metrics_interval = "30s"
```

### Known Issues

Counters with only delta values are currently dropped.
The plugin supports Counter envelopes with totals, Gauge envelopes and `http` Timer envelopes.