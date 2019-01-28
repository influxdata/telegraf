# Loggregator Forwarder Agent Telegraf Service Input Plugin

A [telegraf](https://github.com/influxdata/telegraf) service input plugin that supports the
[loggregator v2 api](https://github.com/cloudfoundry/loggregator-api/tree/master/v2). In more recent versions of the 
Cloud Foundry platform, a new pattern has been introduced known as the
[Loggregator Forwarder Agent](https://github.com/cloudfoundry/loggregator-agent-release). This architecture allows
logs and metrics coming off a single VM to be multiplexed to N number of downstream consumers who implement the
Loggregator API. This service will allow Telegraf to be downstream of the Loggregator Forwarder Agent.

### Configuration:

Both TLS and unsecured servers are supported. To disable TLS simply omit the certificate paths in the config.

```toml
[[inputs.loggregator]]
  ## A uint16 port for the LoggregatorInput Ingress server to listen on
  port = 13322

  ## A string path to the tls ca certificate
  tls_ca = "/path/to/tls_ca_cert.pem"

  ## A string path to the tls server certificate
  tls_cert = "/path/to/tls_cert.pem"

  ## A string path to the tls server private key
  tls_key = "/path/to/tls_cert.key"
```

### Known Issues

Counters with only delta values are currently dropped.
Discussions are ongoing about how to handle this case.
The plugin supports Counter envelopes with totals, Gauge envelopes, and HTTP Timer envelopes.
