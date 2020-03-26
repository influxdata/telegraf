# APM Server

APM Server is a input plugin that listens for requests sent by Elastic APM Agents. 
These type of events are supported to transform to metrics:

* [Metadata][datamodel_metadata]
* [Spans][datamodel_spans]
* [Transactions][datamodel_transactions]
* [Metrics][datamodel_metrics]
* [Errors][datamodel_errors]

### Supported APM HTTP endpoints
The [APM server specification][apm_endpoints] exposes endpoints for events intake, sourcemap upload, agent configuration and server information. 

The table below describe how this plugin conforms with them:

| APM Endpoint                                          | Path                                          | Response                          |
|-------------------------------------------------------|-----------------------------------------------|-----------------------------------|
| [Events intake][endpoint_events_intake]               | `/intake/v2/events`, `/intake/v2/rum/events`  | TODO  |
| [Sourcemap upload][endpoint_sourcemap_upload]         | `/assets/v1/sourcemaps`                       | TODO  |
| [Agent configuration][endpoint_agent_configuration]   | `/config/v1/agents`                           | `403` - disabled configuration    |
| [Server information][endpoint_server_information]     | `/`                                           | `200` - server information        |

### Configuration:

```toml
[[inputs.apm_server]]
  ## Address and port to list APM Agents
  service_address = ":8200"
```

### Agent Configuration
TODO

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
