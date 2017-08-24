# Telegraf Input Plugin: Consul

This plugin will collect statistics about all health checks registered in the Consul. It uses [Consul API](https://www.consul.io/docs/agent/http/health.html#health_state)
to query the data. It will not report the [telemetry](https://www.consul.io/docs/agent/telemetry.html) but Consul can report those stats already using StatsD protocol if needed.

## Configuration:

```
# Gather health check statuses from services registered in Consul
[[inputs.consul]]
  ## Most of these values defaults to the one configured on a Consul's agent level.
  ## Optional Consul server address (default: "")
  # address = ""
  ## Optional URI scheme for the Consul server (default: "")
  # scheme = ""
  ## Optional ACL token used in every request (default: "")
  # token = ""
  ## Optional username used for request HTTP Basic Authentication (default: "")
  # username = ""
  ## Optional password used for HTTP Basic Authentication (default: "")
  # password = ""
  ## Optional data centre to query the health checks from (default: "")
  # datacentre = ""
```

## Measurements:

### Consul:
Tags:
- node: on which node check/service is registered on
- service_name: name of the service (this is the service name not the service ID)
- check_id

Fields:
- check_name
- service_id
- status
- passing
- critical
- warning

`passing`, `critical`, and `warning` are integer representations of the health
check state. A value of `1` represents that the status was the state of the
the health check at this sample.

## Example output

```
$ telegraf --config ./telegraf.conf --input-filter consul --test
* Plugin: consul, Collection 1
> consul_health_checks,host=wolfpit,node=consul-server-node,check_id="serfHealth" check_name="Serf Health Status",service_id="",status="passing",passing=1i,critical=0i,warning=0i 1464698464486439902
> consul_health_checks,host=wolfpit,node=consul-server-node,service_name=www.example.com,check_id="service:www-example-com.test01" check_name="Service 'www.example.com' check",service_id="www-example-com.test01",status="critical",passing=0i,critical=1i,warning=0i 1464698464486519036
```
