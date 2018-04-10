# Consul Input Plugin

This plugin will collect statistics about all helath checks registered in the Consul and some basic information about Consul status.
It uses [Consul Health API](https://www.consul.io/docs/agent/http/health.html#health_state) and [Consul Catalog API] (https://www.consul.io/docs/agent/http/catalog.html)
to query the data.

It **will not** report the [telemetry](https://www.consul.io/docs/agent/telemetry.html) but Consul can report those stats already using StatsD protocol if needed.

Additionaly it can be configured to gather service health checks (parameter: `service_health`) but since it may be time
consuming it is disabled by default.

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

  ## Optional data centre to query the health checks from (default: "")
  # datacentre = ""

  ## Optional should we gather service health checks (default: false)
  # service_health = false
  
  ## SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## If false, skip chain & host verification
  # insecure_skip_verify = true
```

### consul_health_checks:
Tags:
- node: on which node check/service is registered on
- service_name: name of the service (this is the service name not the service ID)
- check_id

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

### consul_server_stats
Tags:
  *None*

Fields:
- leader (0/1)
- nodes
- peers
- services

### consul_service_health
Tags:
- node
- service_name

Fields:
- healthy (0/1)

## Example output
```
$ telegraf --config ./telegraf.conf -input-filter consul -test
* Plugin: consul, Collection 1
> consul_health_checks,host=wolfpit,node=consul-server-node check_id="serfHealth",check_name="Serf Health Status",service_id="",status="passing" 1464698464486439902
> consul_health_checks,host=wolfpit,node=consul-server-node,service_name=www.example.com check_id="service:www-example-com.test01",check_name="Service 'www.example.com' check",service_id="www-example-com.test01",status="critical" 1464698464486519036
> consul_server_stats,host=wolfpit leader=0,nodes=130,peers=5,services=84 1480009729000000000
```

and with `serivce_health` set to `true` additional measurements are provided:

```
> consul_service_health,host=wolfpit,node=dev-db-f1,service_name=db-dev-db healthy=1 1480009729000000000
> consul_service_health,host=wolfpit,node=dev-db-f2,service_name=db-dev-db healthy=1 1480009729000000000
> consul_service_health,host=wolfpit,node=dev-mesos-1,service_name=example healthy=1 1480009730000000000
> consul_health_checks,host=wolfpit,node=consul-server-node,check_id="serfHealth" check_name="Serf Health Status",service_id="",status="passing" 1464698464486439902
> consul_health_checks,host=wolfpit,node=consul-server-node,service_name=www.example.com,check_id="service:www-example-com.test01" check_name="Service 'www.example.com' check",service_id="www-example-com.test01",status="critical" 1464698464486519036
> consul_health_checks,host=wolfpit,node=consul-server-node,check_id="serfHealth" check_name="Serf Health Status",service_id="",status="passing",passing=1i,critical=0i,warning=0i 1464698464486439902
> consul_health_checks,host=wolfpit,node=consul-server-node,service_name=www.example.com,check_id="service:www-example-com.test01" check_name="Service 'www.example.com' check",service_id="www-example-com.test01",status="critical",passing=0i,critical=1i,warning=0i 1464698464486519036
```
