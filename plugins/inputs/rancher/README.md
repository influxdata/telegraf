# Example Input Plugin

This plugin gathers pretty minimal state information from the Rancher api about Hosts, Stacks, Services.
This plugin also tries to read environment variables from you system or container. You can manually set them at runtime,
or if you are deploying on rancher set `io.rancher.container.create_agent: true` and `io.rancher.container.agent.role: environment` labels.
By setting these labels you will only get data for the environment the telegraf container is running in. 


### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage rancher`.

```toml
[[inputs.rancher]]
  ## You can skip the client setup portion of this config if one of two conditions are met:
  ## One, you set the following environemnt variables manually: CATTLE_URL, CATTLE_ACCESS_KEY, CATTLE_SECRET_KEY on your host or in a container
  ## Two, you set 'io.rancher.container.create_agent: true' and 'io.rancher.container.agent.role: environment' labels and run the container
  ## in a rancher environment. This will create a service account for the container and eliminate the need for managing the API keys.
  ## Very important note is that using these labels and not passing an account API creds will only gather information for the environment
  ## this container is deployed in.

  ## Specify the rancher Api Url. This can also be auto detected from  the env variable CATTLE_URL.
  api_url = "http://rancher-host:8080/v3"

  ## The api access key for the rancher API. This can also be extracted from CATTLE_ACCESS_KEY
  api_access_key = ""

  ## The api secret key for the rancher API. This can also be extracted from CATTLE_SECRET_KEY
  api_secret_key = ""

  ## Set host to true when you want to also obtain host state stats
  host_data = true

  ## Set stack_data to true when you want to also obtain stack state stats
  stack_data = true

  ## Set service_data to true when you want to also obtain service state stats
  service_data = true
```

### Metrics:

Host measurements:

- rancher_host
  - tags:
    - id
    - hostname
    - agentState
    - agentId
    - agentIp
    - state
    - name
    - clusterName
  - fields:
    - state (int, 0=ok, 1=warn, 2=err)
    - containers (int, count of containers on host)
    - agentState (int, 0=ok, 1=warn, 2=err)
    
- rancher_stack
  - tags:
    - id
    - name
    - state
    - healthState
    - clusterId
    - clusterName
  - fields:
    - state (int, 0=ok, 1=warn, 2=err)
    - containers (int, count of associated container ids in stack)
    - agentState (int, 0=ok, 1=warn, 2=err)

- rancher_service
  - tags:
    - id
    - name
    - state
    - clusterId
    - clusterName
  - fields:
    - state (int, 0=ok, 1=warn, 2=err)
    - containers (int, count of associated container ids in service)

### Sample Queries:

This section should contain some useful InfluxDB queries that can be used to
get started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min for the measurement in the last hour:
```
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```

### Example Output:

```
> rancher_service,stackId=1st4,clusterName=Default,host=dhendel-hp,id=1s8,state=active,name=dns,clusterId=1c1 state=0i,containers=1i,scale=1i,currentScale=0i 1511811984000000000
> rancher_service,clusterName=Default,host=dhendel-hp,id=1s9,state=active,name=proxy,clusterId=1c1,stackId=1st5 currentScale=0i,state=0i,containers=1i,scale=1i 1511811984000000000
> rancher_service,host=dhendel-hp,id=1s10,state=active,name=healthcheck,clusterId=1c1,stackId=1st4,clusterName=Default state=0i,containers=1i,scale=1i,currentScale=0i 1511811984000000000
> rancher_service,clusterName=Default,host=dhendel-hp,id=1s11,state=active,name=ipsec,clusterId=1c1,stackId=1st4 scale=1i,currentScale=0i,state=0i,containers=2i 1511811984000000000
> rancher_service,id=1s12,state=active,name=kibana,clusterId=1c1,stackId=1st7,clusterName=Default,host=dhendel-hp scale=1i,currentScale=0i,state=0i,containers=1i 1511811984000000000
> rancher_service,host=dhendel-hp,id=1s13,state=active,name=ls,clusterId=1c1,stackId=1st7,clusterName=Default state=0i,containers=1i,scale=1i,currentScale=0i 1511811984000000000
> rancher_service,host=dhendel-hp,id=1s14,state=active,name=elasticsearch,clusterId=1c1,stackId=1st7,clusterName=Default state=0i,containers=1i,scale=1i,currentScale=0i 1511811984000000000
> rancher_service,stackId=1st8,clusterName=Default,host=dhendel-hp,id=1s15,state=active,name=logspout2,clusterId=1c1 state=0i,containers=1i,scale=1i,currentScale=0i 1511811984000000000

```
