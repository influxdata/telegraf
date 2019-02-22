# Loggregator Forwarder Agent Telegraf Input Plugin

A [telegraf](https://github.com/influxdata/telegraf) input plugin that supports the 
[loggregator v2 api](https://github.com/cloudfoundry/loggregator-api/tree/master/v2).  This agent 
relies on Loggregator's Forwarder Agent architecture to receive metrics within an individual VM. Use
this if you intend to colocate Telegraf on each VM.

### Configuration:

Both mutual TLS and unsecured servers are supported. To disable mTLS simply omit the certificate paths in the config.

```toml
[[inputs.loggregator_forwarder_agent]]
  ## A uint16 port for the LoggregatorInput Ingress server to listen on
  port = 13322

  ## A string path to the tls ca certificate
  tls_ca_path = "/path/to/tls_ca_cert.pem"

  ## A string path to the tls server certificate
  tls_cert_path = "/path/to/tls_cert.pem"

  ## A string path to the tls server private key
  tls_key_path = "/path/to/tls_cert.key"
```

### Known Issues

Counters with only delta values are currently dropped.
The plugin supports Counter envelopes with totals, Gauge envelopes and `http` Timer envelopes.