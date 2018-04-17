# Consul Input Plugin

This plugin will collect statistics about all health checks registered in the
Consul. It uses [Consul API](https://www.consul.io/docs/agent/http/health.html#health_state)
to query the data. It will not report the
[telemetry](https://www.consul.io/docs/agent/telemetry.html) but Consul can
report those stats already using StatsD protocol if needed.

### Configuration:

```toml
# Gather health check statuses from services registered in Consul
[[inputs.consul]]
  ## Consul server address
  # address = "localhost"

  ## URI scheme for the Consul server, one of "http", "https"
  # scheme = "http"

  ## ACL token used in every request
  # token = ""

  ## HTTP Basic Authentication username and password.
  # username = ""
  # password = ""

  ## Data centre to query the health checks from
  # datacentre = ""

  ## SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## If false, skip chain & host verification
  # insecure_skip_verify = true
```

### Metrics:

- consul_health_checks
  - tags:
  	- node (node that check/service is registred on)
  	- service_name
  	- check_id
  - fields:
    - check_name
    - service_id
    - status
    - passing (integer)
    - critical (integer)
    - warning (integer)

`passing`, `critical`, and `warning` are integer representations of the health
check state. A value of `1` represents that the status was the state of the
the health check at this sample.

## Example output

```
consul_health_checks,host=wolfpit,node=consul-server-node,check_id="serfHealth" check_name="Serf Health Status",service_id="",status="passing",passing=1i,critical=0i,warning=0i 1464698464486439902
consul_health_checks,host=wolfpit,node=consul-server-node,service_name=www.example.com,check_id="service:www-example-com.test01" check_name="Service 'www.example.com' check",service_id="www-example-com.test01",status="critical",passing=0i,critical=1i,warning=0i 1464698464486519036
```
