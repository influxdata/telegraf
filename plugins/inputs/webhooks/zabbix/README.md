# zabbix connector webhooks

You should configure your Zabbix Connecotr to point at the `webhooks` service. `Data type` is `Item values` and you can use `Baisc` as `HTTP authentication`.

## Events

The titles of the following sections are links to the full payloads and details for each event. The body contains what information from the event is persisted. The format is as follows:

```toml
# TAGS
* 'tagKey' = `tagValue` type
# FIELDS
* 'fieldKey' = `fieldValue` type
```

The tag values and field values show the place on the incoming JSON object where the data is sourced from.

See [Zabbix Conectors](https://www.zabbix.com/documentation/current/en/manual/config/export/streaming)

### `Zabbix Item Value` event

**Metric Name**
The metric name is composed from a specific tag value, default is "component"

Zabbix Tags:
  `"component" = "health"`
  `"component" = "network"`

Metric Name: zabbix_component_health_network

**Tags:**

* 'item' = `item.name` string
* 'host_raw' = `item.host.host` string
* 'hostname' = `item.host.name` string
* 'houstgroups = `item.groups` string
* 'itemid' = `item.itemid` int

Zabbix Tags are trasformed as follow:

Zabbix Tags:
  `"component" = "health"`
  `"component" = "network"`

Telegraf Tags:
 'tag_component' = `health,network`

**Fields:**

* 'value' = `item.value` int/float/string

