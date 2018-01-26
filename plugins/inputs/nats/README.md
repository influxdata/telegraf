# NATS Monitoring Input Plugin

The [NATS](http://www.nats.io/about/) monitoring plugin reads from
specified NATS instance and submits metrics to InfluxDB. 

## Configuration

```toml
[[inputs.nats]]
  ## The address of the monitoring end-point of the NATS server
  server = "http://localhost:8222"
```
