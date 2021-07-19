# Elastic APM telegraf plugin

The `apm-server` Telegraf input collects data from [Elastic APM agents][apm_agents].

#### Supported APM HTTP endpoints
The [APM server][apm_endpoints] exposes endpoints for events intake, sourcemap upload, agent configuration and server information. 
The table below describes how this plugin handle them:

| APM Endpoint                                          | Path                                          | Response                                              |
|-------------------------------------------------------|-----------------------------------------------|-------------------------------------------------------|
| [Events intake][endpoint_events_intake]               | `/intake/v2/events`, `/intake/v2/rum/events`  | Serialize Events into LineProtocol. See detail below  |
| [Sourcemap upload][endpoint_sourcemap_upload]         | `/assets/v1/sourcemaps`                       | Accept all request without processing sources         |
| [Agent configuration][endpoint_agent_configuration]   | `/config/v1/agents`, `/config/v1/rum/agents`  | Configuration via APM Server is disabled              |
| [Server information][endpoint_server_information]     | `/`                                           | Returns Telegraf APM Server information               |


### Demo
Here is the demo how to use Elastic APM agents (Ruby, Java, JS RUM..) and send application metrics using telegraf-apm plugin into InfluxDB.
* [https://github.com/bonitoo-io/telegraf-apm](https://github.com/bonitoo-io/telegraf-apm)

### Example telegraf configuration:

```toml
[[inputs.apm_server]]
  ## http server bind address 
  service_address = ":8200"

  ## Http server timeouts
  idle_timeout  = "45s"
  read_timeout  = "30s"
  write_timeout = "30s"

  ## http request size limitations  
  # http server maximum header size
  # max_header_bytes = 1048576
  ## http server maximum request body size 	
  # max_body_size = 33554432

  ## Secret token used for APM agent authorization 
  secret_token = "my-secret-token"

  drop_unsampled_transactions = false
 
  ## fields excluded from intake event
  exclude_fields = [
        "exception_stacktrace*", "stacktrace*", "log_stacktrace*",
        "process_*",
        "service_language*",
        "service_runtime*",
        "service_agent_version",
        "service_framework*",
        "service_version",
        "service_agent_ephemeral_id",
        "system_architecture",
        "system_platform",
        "system_container_id",
        "span_count*",
        "context_request*",
        "context_response*",
        "context_destination*",
        "context_db_type",
        "context_db_statement",
        "id", "parent_id", "trace_id",
        "transaction_id",
        "sampled"
        ]
  ## list of fields that will be stored as tags
  tag_keys = ["result", "name", "transaction_type", "transaction_name", "type", "span_type", "span_subtype"]
```
### Metrics

Each incoming event from APM Agent contains two parts: `metadata` and `eventdata`. 
By default, The `metadata` are mapped to LineProtocol's as tags and `eventdata` are mapped to LineProtocol's fields.

These type of events are supported to transform to metrics:

* [Metadata][datamodel_metadata]
* [Metrics][datamodel_metrics]
* [Transactions][datamodel_transactions]
* [Spans][datamodel_spans]
* [Errors][datamodel_errors]

It is possible to specify which event fields should be stored as tags using `tag_keys` configuration. You can filter 
data to exclude specific fields using `exclude_fields` configuration.

#### common tags
* service_agent_name
* service_name
* system_hostname

#### apm_metricset
* samples_system.cpu.total.norm.pct
* samples_system.memory.actual.free
* samples_system.memory.total
* samples_system.process.cpu.total.norm.pct
* samples_system.process.memory.size
* samples_system.process.memory.rss.bytes    
    * _JVM specific_ 
        * samples_jvm.memory.heap.used
        * samples_jvm.memory.non_heap.committed
        * samples_jvm.memory.non_heap.max
        * samples_jvm.memory.non_heap.used
        * samples_jvm.thread.count
        
#### _Ruby_
* samples_ruby.heap.allocations.total
* samples_ruby.heap.slots.free
* samples_ruby.heap.slots.live
* samples_ruby.threads
* samples_ruby.gc.count
* samples_ruby.heap.allocations.total
* samples_ruby.heap.slots.free
* samples_ruby.heap.slots.live
* samples_ruby.threads

#### Transactions/Span samples

* transaction_name (tag)
* transaction_type (tag)
* span_subtype (tag)
* samples_transaction.breakdown.count
* samples_transaction.duration.count
* samples_transaction.duration.sum.us
* samples_span.self_time.count
* samples_span.self_time.sum.us
    
#### `apm_transaction`, `apm_span`
Transaction / span duration is stored in `duration` field.
Tags:
* name - name of transaction/span
* type - type of transaction/span
* result - transaction result

[datamodel_metadata]: https://www.elastic.co/guide/en/apm/get-started/7.6/metadata.html
[datamodel_spans]: https://www.elastic.co/guide/en/apm/get-started/current/transaction-spans.html
[datamodel_transactions]: https://www.elastic.co/guide/en/apm/get-started/current/transactions.html
[datamodel_metrics]: https://www.elastic.co/guide/en/apm/get-started/current/metrics.html
[datamodel_errors]: https://www.elastic.co/guide/en/apm/get-started/current/errors.html
[apm_endpoints]: https://www.elastic.co/guide/en/apm/server/current/intake-api.html
[endpoint_events_intake]: https://www.elastic.co/guide/en/apm/server/current/events-api.html
[endpoint_sourcemap_upload]: https://www.elastic.co/guide/en/apm/server/current/sourcemap-api.html
[endpoint_agent_configuration]: https://www.elastic.co/guide/en/apm/server/current/agent-configuration-api.html
[endpoint_server_information]: https://www.elastic.co/guide/en/apm/server/current/server-info.html
[apm_agents]: https://www.elastic.co/guide/en/apm/agent/index.html
