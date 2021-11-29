# Hashicorp Nomad Input Plugin

The Nomad plugin must grab metrics from every Nomad agent of the cluster. Telegraf may be present in every node and connect to the agent locally. In this case should be something like `http://127.0.0.1:4646`.

> Tested on Nomad 1.1.6

## Configuration

```toml
[[inputs.nomad]]
  ## URL for the Nomad agent
  # url = "http://127.0.0.1:4646"

  ## Use auth token for authorization. 
  ## If both are set, an error is thrown.
  ## If both are empty, no token will be used.
  # nomad_token = "/path/to/auth/token"
  ## OR
  # nomad_token_string = "a1234567-40c7-9048-7bae-378687048181"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
```

## Metrics

Both Nomad servers and agents collect various metrics. For every details, please have a look at Nomad following documentation:

- [https://www.nomadproject.io/docs/operations/metrics](https://www.nomadproject.io/docs/operations/metrics)
- [https://www.nomadproject.io/docs/operations/telemetry](https://www.nomadproject.io/docs/operations/telemetry)
