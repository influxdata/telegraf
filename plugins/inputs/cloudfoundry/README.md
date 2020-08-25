# Cloudfoundry Input Plugin

The `cloudfoundry` plugin gather metrics and logs from the Cloudfoundry platform.

#### Configuration

```toml
[[inputs.cloudfoundry]]
  ## HTTP gateway URL to the cloudfoundry reverse log proxy gateway
  gateway_address = "https://log-stream.your-cloudfoundry-system-domain"

  ## API URL to the cloudfoundry API endpoint for your platform
  api_address = "https://api.your-cloudfoundry-system-domain"

  ## All instances with the same shard_id will receive an exclusive
  ## subset of the data. Use this to avoid duplication of metrics
  shard_id = "telegraf"

  ## Username and password for user authentication
  # username = ""
  # password = ""

  ## Client ID and secret for client authentication
  # client_id = ""
  # client_secret ""

  ## Skip verification of TLS certificates (insecure!)
  # insecure_skip_verify = false

  ## retry_interval sets the delay between reconnecting failed stream
  retry_interval = "1s"

  ## Source ID is the GUID of the application or component to collect
  ## metrics from. If unset (default) metrics from ALL platform components
  ## will be collected.
  ##
  ## Note: If you do not have UAA client_id/secret with the
  ## "doppler.firehose" or "logs.admin" scope you MUST set a source_id.
  ##
  source_id = ""

  ## Limit which types of events to collect (default: ALL)
  # types = ["counter", "timer", "gauge", "event", "log"]
```

### Metrics

The metrics emitted are dependent on the source stream, but will conform to the
following types.

#### Timer

Enable with `types = ["timer"]`

Timer metrics emit `_start`, `_stop` and `_duration` suffixed fields. For
example when collecting from an application source the `http` timer metric will
emit `http_start`, `http_stop` and `http_duration` fields.

### Logs

Enable with `types = ["log"]`

Logs will be collected into the `syslog` measurement with syslog-compatible fields.

### Counters

Enabled with = `types = ["counter"]`

Counter metrics emit `_total` and `_delta` suffixed fields.

### Gauges

Enabled with `types = ["gauge"]`

Gauge metrics are fields with a simple numeric value.

### Events

Enabled with `type = ["event"]`

Event metrics are similar to logs, they have a `title` and a `body` field and
represent an action that has taken place.

### Common tags

Most metrics will have the following tags set where applicable:

* `source_id`: the component guid the metric originated from
* `app_id`: the guid of the cloudfoundry application
* `app_name`: the name of the cloudfoundry application
* `space_id`: the guid of the cloudfoundry space the source is from
* `space_name`: the name of the cloudfoundry space the source is from
* `organization_id`: the guid of the cloudfoundry org the source is from
* `organization_name`: the name of cloudfoundry org the source is from

### Example Output

```
> syslog,app_name=myapplication,appname=myapplication,deployment=prod,facility=user,host=c0b9bd29-99b4-4df9-9665-781095c9d501,hostname=gds-tech-ops.sandbox.myapplication,instance_id=0,ip=10.0.34.16,job=diego-cell,organization_id=b92cf390-3dbb-4a6e-a24d-04a811c4624b,organization_name=gds-tech-ops,origin=rep,process_id=c0b9bd29-99b4-4df9-9665-781095c9d501,process_instance_id=c11a7e5d-36d8-4614-5a54-2b39,process_type=web,severity=err,source_id=c0b9bd29-99b4-4df9-9665-781095c9d501,source_type=APP/PROC/WEB/SIDECAR/PROCESS,space_id=f523b565-a298-4efb-994b-b637dd97ace2,space_name=sandbox app_id="c0b9bd29-99b4-4df9-9665-781095c9d501",facility_code=1i,index="11b3054c-47ca-4970-b1be-918b6ddc5b3c",message="time=\"2020-07-06T08:36:56Z\" level=info msg=\"Response: OK\" component=server method=POST remote_addr=\"10.255.146.55:55232\" response_time=2.253966ms status=200",procid="APP/PROC/WEB/SIDECAR/PROCESS",severity_code=3i,timestamp=1594024616353350411i,version=1i 1594024616353350411
> cloudfoundry,deployment=prod,host=c0b9bd29-99b4-4df9-9665-781095c9d501,ip=10.0.48.101,job=router,origin=gorouter,routing_instance_id=c11a7e5d-36d8-4614-5a54-2b39,status_code=200 content_length="79",forwarded="172.16.100.21
10.0.2.143",http_duration=12024493i,http_start=1594024616329959073i,http_stop=1594024616341983566i,index="096319a7-b7d7-472b-b62e-360268af8df9",method="POST",peer_type="Client",remote_address="10.0.2.143:34436",request_id="66f9344b-b3ea-4c2d-7742-22db7b061ffb",uri="https://myapplication-test.cloudapps.digital/chronograf/v1/sources/10000/proxy",user_agent="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36" 1594024616341992281
> cloudfoundry,app_name=myapplication,component=route-emitter,deployment=prod,host=c0b9bd29-99b4-4df9-9665-781095c9d501,instance_id=0,ip=10.0.48.102,job=router,organization_id=b92cf390-3dbb-4a6e-a24d-04a811c4624b,organization_name=gds-tech-ops,origin=gorouter,process_id=c0b9bd29-99b4-4df9-9665-781095c9d501,process_instance_id=c11a7e5d-36d8-4614-5a54-2b39,process_type=web,space_id=f523b565-a298-4efb-994b-b637dd97ace2,space_name=sandbox,status_code=200 app_id="c0b9bd29-99b4-4df9-9665-781095c9d501",content_length="79",forwarded="172.16.100.21",http_duration=11537986i,http_start=1594024616334984488i,http_stop=1594024616346522474i,index="ebd04ace-7969-4022-8488-986effd75069",method="POST",peer_type="Server",remote_address="10.0.2.143:38976",request_id="e458d9ee-2016-4e13-5c20-4ba049d3bef3",uri="https://myapplication-test.cloudapps.digital/chronograf/v1/sources/10000/proxy",user_agent="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36" 1594024616346529662
```
