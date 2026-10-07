# GroundWork Output Plugin

This plugin writes metrics to a [GroundWork Monitor][groundwork] instance.

> [!IMPORTANT]
> Plugin only supports GroundWork v8 or later.

⭐ Telegraf v1.21.0
🏷️ applications, messaging
💻 all

[groundwork]: https://www.gwos.com/product/groundwork-monitor/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret store support

This plugin supports secrets from secret stores for the `username` and
`password` option.
See the [secret store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Send telegraf metrics to GroundWork Monitor
[[outputs.groundwork]]
  ## URL of your groundwork instance.
  url = "https://groundwork.example.com"

  ## Agent uuid for GroundWork API Server.
  agent_id = ""

  ## Username and password to access GroundWork API.
  username = ""
  password = ""

  ## Default application type to use in GroundWork client
  # default_app_type = "TELEGRAF"

  ## Default display name for the host with services(metrics).
  # default_host = "telegraf"

  ## Default service state.
  # default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname.
  # resource_tag = "host"

  ## The name of the tag that contains the host group name.
  # group_tag = "group"

  ## The name of the tag that contains the host alias (display name).
  ## Empty by default, so no alias is sent.
  # alias_tag = ""

  ## The name of the tag that contains the service name.
  # service_tag = "service"

  ## Optional mappings building the values from tags and string fields.
  # [[outputs.groundwork.map_hostname]]
  #   tag = "node_name"
  #   matcher = "(.+)"
  #   template = "$1"
  # [[outputs.groundwork.map_hostname]]
  #   tag = "host"
  #   matcher = "^([^.]+)"
  #   template = "$1"
  ## Also map_hostgroup, map_hostalias, map_service, map_status, map_message.
```

## How metrics map to GroundWork

Each metric becomes one service of one host. The metrics of a write are
grouped by host into a single request, and hosts and services that GroundWork
does not know yet are created. Hosts are reported as up.

* The host is named by `resource_tag`, and gets the `Alias` property from
  `alias_tag` when it is set. The `group_tag` adds the host to a host group.
* The service is named by `service_tag`, or by the metric name.
* Every numeric or boolean field becomes a performance value of the service,
  with the metric time and the unit from the `unitType` tag. String fields are
  not sent as values, but `message` and `status` fields are used as described
  below, and the mappings can read any string field.
* Every other tag becomes a service property.
* Tags and fields named after the options above, `status`, `message`,
  `unitType`, `critical`, `warning`, or ending with `_cr` or `_wn` are used by
  the plugin, and are neither sent as values nor as properties.

### Resolution order

Each value is taken from the first source that provides it:

| Value | Sources, in order |
| --- | --- |
| Host | `map_hostname`, `resource_tag`, `default_host` |
| Host group | `map_hostgroup`, `group_tag`; none means no group |
| Host alias | `map_hostalias`, `alias_tag`; none means no alias |
| Service | `map_service`, `service_tag`, metric name |
| Status | `map_status`, `status` tag, `status` field, computed from thresholds |
| Message | `map_message`, `message` tag, `message` field |
| Critical threshold of field `F` | `F_cr` tag, `critical` tag, `F_cr` field |
| Warning threshold of field `F` | `F_wn` tag, `warning` tag, `F_wn` field |

A status that is not one of the supported statuses is skipped, so the next
source is tried. When no source gives a status, it is computed from the
performance values: a value with both thresholds set is critical when it is
at or above the critical threshold, and warning when it is at or above the
warning threshold. When the warning threshold is above the critical one, the
comparison is reversed, so lower values are worse. The worst status of all
values wins, values without both thresholds count as `SERVICE_OK`, and a
metric without any performance values gets `SERVICE_UNKNOWN`.

For example, this metric becomes the `disk` service of host `db1`, with status
`SERVICE_WARNING` since 85 is between the thresholds, and with the `path`
property:

```text
disk,host=db1,path=/var used_percent=85,used_percent_wn=80,used_percent_cr=90
```

## List of tags used by the plugin

* __group__ - to define the name of the group you want to monitor,
  can be changed with config.
* __host__ - to define the name of the host you want to monitor,
  can be changed with config.
* __service__ - to define the name of the service you want to monitor,
  can be changed with config.
* __status__ - to define the status of the service. Supported statuses:
  "SERVICE_OK", "SERVICE_WARNING", "SERVICE_UNSCHEDULED_CRITICAL",
  "SERVICE_PENDING", "SERVICE_SCHEDULED_CRITICAL", "SERVICE_UNKNOWN".
* __message__ - to provide any message you want,
  it overrides __message__ field value.
* __unitType__ - to use in monitoring contexts (subset of The Unified Code for
  Units of Measure standard). Supported types: "1", "%cpu", "KB", "GB", "MB".
* __critical__ - to define the default critical threshold value for all
  fields, it overrides the `<field>_cr` field value.
* __warning__ - to define the default warning threshold value for all
  fields, it overrides the `<field>_wn` field value.
* __&lt;field&gt;_cr__ - to define the critical threshold value of `<field>`,
  it overrides the __critical__ tag value and the `<field>_cr` field value.
* __&lt;field&gt;_wn__ - to define the warning threshold value of `<field>`,
  it overrides the __warning__ tag value and the `<field>_wn` field value.

## Mappings

The `map_hostgroup`, `map_hostalias`, `map_hostname`, `map_service`,
`map_status` and `map_message` options build the corresponding value from the
metric tags and string fields instead of the tags listed above. A tag takes
precedence over a string field with the same name, so a `message` field can be
matched as long as there is no `message` tag. For example, to name hosts by
the `node_name` tag, falling back to the short `host` tag name when `node_name`
is missing, to group hosts by cluster and namespace, and to set the service
status from the syslog `message` field:

```toml
[[outputs.groundwork.map_hostname]]
  tag = "node_name"
  matcher = "(.+)"
  template = "$1"

[[outputs.groundwork.map_hostname]]
  tag = "host"
  matcher = "^([^.]+)"
  template = "$1"

[[outputs.groundwork.map_hostgroup]]
  tag = "cluster,namespace"
  matcher = "^(.+),(.+)$"
  template = "$1-$2"

[[outputs.groundwork.map_status]]
  tag = "message"
  matcher = "Database is shutting down"
  template = "SERVICE_UNSCHEDULED_CRITICAL"
```

The template is expanded for the first match of the matcher. Write `${1}`
instead of `$1` when the group is followed by a letter, digit or `_`: `$1_host`
refers to a group named `1_host` and expands to nothing, while `${1}_host`
gives the intended result.

When no entry of a mapping matches, or the mapped status is not a supported
status, the value falls back to the tag-based logic.

To skip metrics for this output only, use the `metricpass` option of the
output, for example to drop syslog messages about client connections:

```toml
[[outputs.groundwork]]
  metricpass = '!("message" in fields) || !fields.message.matches("DB client connected")'
```

## NOTE

The current version of GroundWork Monitor does not support metrics whose values
are strings. Such metrics will be skipped and will not be added to the final
payload. You can find more context in this pull request: [#10255][].

[#10255]: https://github.com/influxdata/telegraf/pull/10255
