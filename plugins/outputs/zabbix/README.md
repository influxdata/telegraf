# Zabbix Output Plugin

This plugin writes metrics to [Zabbix][zabbix] via [traps][traps]. It has been
tested with versions v3.0, v4.0 and v6.0 but should work with newer versions
of Zabbix as long as the protocol doesn't change.

⭐ Telegraf v1.30.0
🏷️ datastore
💻 all

[zabbix]: https://www.zabbix.com/
[traps]: https://www.zabbix.com/documentation/current/en/manual/appendix/items/trapper

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send metrics to Zabbix
[[outputs.zabbix]]
  ## Address and (optional) port of the Zabbix server
  address = "zabbix.example.com:10051"

  ## Send metrics as type "Zabbix agent (active)"
  # agent_active = false

  ## Add prefix to all keys sent to Zabbix.
  # key_prefix = "telegraf."

  ## Name of the tag that contains the host name. Used to set the host in Zabbix.
  ## If the tag is not found, use the hostname of the system running Telegraf.
  # host_tag = "host"

  ## Skip measurement prefix to all keys sent to Zabbix.
  # skip_measurement_prefix = false

  ## This field will be sent as HostMetadata to Zabbix Server to autoregister the host.
  ## To enable this feature, this option must be set to a value other than "".
  # autoregister = ""

  ## Interval to resend auto-registration data to Zabbix.
  ## Only applies if autoregister feature is enabled.
  ## This value is a lower limit, the actual resend should be triggered by the next flush interval.
  # autoregister_resend_interval = "30m"

  ## Interval to send LLD data to Zabbix.
  ## This value is a lower limit, the actual resend should be triggered by the next flush interval.
  # lld_send_interval = "10m"

  ## Interval to delete stored LLD known data and start capturing it again.
  ## This value is a lower limit, the actual resend should be triggered by the next flush interval.
  # lld_clear_interval = "1h"
```

### agent_active

The `request` value in the package sent to Zabbix should be different if the
items configured in Zabbix are [Zabbix trapper][zabbixtrapper] or
[Zabbix agent (active)][zabbixagentactive].

`agent_active = false` will send data as _sender data_, expecting trapper items.

`agent_active = true` will send data as _agent data_, expecting active Zabbix
agent items.

[zabbixtrapper]: https://www.zabbix.com/documentation/6.4/en/manual/config/items/itemtypes/trapper?hl=Trapper
[zabbixagentactive]: https://www.zabbix.com/documentation/6.4/en/manual/config/items/itemtypes/zabbix_agent

### key_prefix

We can set a prefix that should be added to all Zabbix keys.

This is configurable with the option `key_prefix`, set by default to
`telegraf.`.

Example how the configuration `key_prefix = "telegraf."` will generate the
Zabbix keys given a Telegraf metric:

```diff
- measurement,host=hostname valueA=0,valueB=1
+ telegraf.measurement.valueA
+ telegraf.measurement.valueB
```

### skip_measurement_prefix

We can skip the measurement prefix added to all Zabbix keys.

Example with `skip_measurement_prefix = true"` and `prefix = "telegraf."`:

```diff
- measurement,host=hostname valueA=0,valueB=1
+ telegraf.valueA
+ telegraf.valueB
```

Example with `skip_measurement_prefix = true"` and `prefix = ""`:

```diff
- measurement,host=hostname valueA=0,valueB=1
+ valueA
+ valueB
```

### autoregister

If this field is active, Telegraf will send an
[autoregister request][autoregisterrequest] to Zabbix, using the content of
this field as the [HostMetadata][hostmetadata].

One request is sent for each of the different values seen by Telegraf for the
`host` tag.

[autoregisterrequest]: https://www.zabbix.com/documentation/current/en/manual/discovery/auto_registration?hl=autoregistration
[hostmetadata]: https://www.zabbix.com/documentation/current/en/manual/discovery/auto_registration?hl=autoregistration#using-host-metadata

### autoregister_resend_interval

If `autoregister` is defined, this field set the interval at which
autoregister requests are resend to Zabbix.

The [telegraf interval format][intervals_format] should be used.

The actual send of the autoregister request will happen in the next output flush
after this interval has been surpassed.

[intervals_format]: ../../../docs/CONFIGURATION.md#intervals

### lld_send_interval

To reduce the number of LLD requests sent to Zabbix (LLD processing is
[expensive][lldexpensive]), this plugin will send only one per
`lld_send_interval`.

When Telegraf is started, this plugin will start to collect the info needed to
generate this LLD packets (measurements, tags keys and values).

Once this interval is surpassed, the next flush of this plugin will add the
packet with the LLD data.

In the next interval, only new, or modified, LLDs will be sent.

[lldexpensive]: https://www.zabbix.com/documentation/4.2/en/manual/introduction/whatsnew420#:~:text=Daemons-,Separate%20processing%20for%20low%2Dlevel%20discovery,-Processing%20low%2Dlevel

### lld_clear_interval

When this interval is surpassed, the next flush will clear all the LLD data
collected.

This allows this plugin to forget about old data and resend LLDs to Zabbix, in
case the host has new discovery rules or the packet was lost.

If we have `flush_interval = "1m"`, `lld_send_interval = "10m"` and
`lld_clear_interval = "1h"` and Telegraf is started at 00:00, the first LLD will
be sent at 00:10. At 01:00 the LLD data will be deleted and at 01:10 LLD data
will be resent.

## Trap format

For each new metric generated by Telegraf, this output plugin will send one
trap for each field.

Given this Telegraf metric:

```text
measurement,host=hostname valueA=0,valueB=1
```

It will generate this Zabbix metrics:

```json
{"host": "hostname", "key": "telegraf.measurement.valueA", "value": "0"}
{"host": "hostname", "key": "telegraf.measurement.valueB", "value": "1"}
```

If the metric has tags (aside from `host`), they will be added, in alphabetical
order using the format for LLD metrics:

```text
measurement,host=hostname,tagA=keyA,tagB=keyB valueA=0,valueB=1
```

Zabbix generated metrics:

```json
{"host": "hostname", "key": "telegraf.measurement.valueA[keyA,keyB]", "value": "0"}
{"host": "hostname", "key": "telegraf.measurement.valueB[keyA,keyB]", "value": "1"}
```

This order is based on the tags keys, not the tag values, so, for example, this
Telegraf metric:

```text
measurement,host=hostname,aaaTag=999,zzzTag=111 value=0
```

Will generate this Zabbix metric:

```json
{"host": "hostname", "key": "telegraf.measurement.value[999,111]", "value": "0"}
```

## Zabbix low-level discovery

Zabbix needs an `item` created before receiving any metric. In some cases we do
not know in advance what are we going to send, for example, the name of a
container to send its cpu and memory consumption.

For this case Zabbix provides [low-level discovery][lld] that allow to create
new items dynamically based on the parameters sent by the trap.

As explained previously, this output plugin will format the Zabbix key using
the tags seen in the Telegraf metric following the LLD format.

To create those _discovered items_ this plugin uses the same mechanism as the
Zabbix agent, collecting information about which tags has been seen for each
measurement and periodically sending a request to a discovery rule with the
collected data.

Keep in mind that, for metrics in this category, Zabbix will discard them until
the low-level discovery (LLD) data is sent.
Sending LLD to Zabbix is a heavy-weight process and is only done at the interval
per the lld_send_interval setting.

[lld]: https://www.zabbix.com/documentation/current/manual/discovery/low_level_discovery

### Design

To explain how everything interconnects we will use an example with the
`net_response` input:

```toml
[[inputs.net_response]]
  protocol = "tcp"
  address = "example.com:80"
```

This input will generate this metric:

```text
$ telegraf -config example.conf -test
* Plugin: inputs.net_response, Collection 1
> net_response,server=example.com,port=80,protocol=tcp,host=myhost result_type="success",response_time=0.091026869 1522741063000000000
```

Here we have four tags: server, port, protocol and host (this one will be
assumed that is always present and treated differently).

The values those three parameters could take are unknown to Zabbix, so we
cannot create trappers items in Zabbix to receive that values (at least without
mixing that metric with another `net_response` metric with different tags).

To solve this problem we use a discovery rule in Zabbix, that will receive the
different groups of tag values and create the traps to gather the metrics.

This plugin knows about three tags (excluding host) for the input
`net_response`, therefore it will generate this new Telegraf metric:

```text
lld.host=myhost net_response.port.protocol.server="{\"data\":[{\"{#PORT}\":\"80\",\"{#PROTOCOL}\":\"tcp\",\"{#SERVER}\":\"example.com\"}]}"
```

Once sent, the final package will be:

```json
{
  "request":"sender data",
  "data":[
    {
      "host":"myhost",
      "key":"telegraf.lld.net_response.port.protocol.server",
      "value":"{\"data\":[{\"{#PORT}\":\"80\",\"{#PROTOCOL}\":\"tcp\",\"{#SERVER}\":\"example.com\"}]}",
      "clock":1519043805
    }
  ],
  "clock":1519043805
}
```

The Zabbix key is generated joining `lld`, the input name and tags (keys)
alphabetically sorted.
Some inputs could use different groups of tags for different fields, that is
why the tags are added to the key, to allow having different discovery rules
for the same input.

The tags used in `value` are changed to uppercase to match the format of Zabbix.

In the Zabbix server we should have a discovery rule associated with that key
(telegraf.lld.net_response.port.protocol.server) and one item prototype for
each field, in this case `result_type` and `response_time`.

The item prototypes will be Zabbix trappers with keys (data type should also
match and some values will be better stored as _delta_):

```text
telegraf.net_response.response_time[{#PORT},{#PROTOCOL},{#SERVER}]
telegraf.net_response.result_type[{#PORT},{#PROTOCOL},{#SERVER}]
```

The macros in the item prototypes keys should be alphabetically sorted so they
can match the keys generated by this plugin.

With that keys and the example trap, the host `myhost` will have two new items:

```text
telegraf.net_response.response_time[80,tcp,example.com]
telegraf.net_response.result_type[80,tcp,example.com]
```

This plugin, for each metric, will send traps to the Zabbix server following
the same structure (INPUT.FIELD[tags sorted]...), filling the items created by
the discovery rule.

In summary:

- we need a discovery rule with the correct key and one item prototype for each
field
- this plugin will generate traps to create items based on the metrics seen in
Telegraf
- it will also send the traps to fill the new created items

### Reducing the number of LLDs

This plugin remembers which LLDs has been sent to Zabbix and avoid generating
the same metrics again, to avoid the cost of LLD processing in Zabbix.

It will only send LLD data each `lld_send_interval`.

But, could happen that package is lost or some host get new discovery rules, so
each `lld_clear_interval` the plugin will forget about the known data and start
collecting again.

### Note on inputs configuration

Which tags should expose each input should be controlled, because an unexpected
tag could modify the trap key and will not match the trapper defined in Zabbix.

For example, in the docker input, each container label is a new tag.

To control this we can add to the input a config like:

```toml
taginclude = ["host", "container_name"]
```

Allowing only the tags "host" and "container_name" to be used to generate the
key (and loosing the information provided in the others tags).

## Examples of metrics converted to traps

### Without tags

```text
mem,host=myHost available_percent=14.684620843239944,used=14246531072i 152276442800000000
```

```json
{
  "request":"sender data",
  "data":[
    {
      "host":"myHost",
      "key":"telegraf.mem.available_percent",
      "value":"14.382719",
      "clock":1522764428
    },
    {
      "host":"myHost",
      "key":"telegraf.mem.used",
      "value":"14246531072",
      "clock":1522764428
    }
  ]
}
```

### With tags

```text
docker_container_net,host=myHost,container_name=laughing_babbage rx_errors=0i,tx_errors=0i 1522764038000000000
```

```json
{
  "request":"sender data",
  "data": [
    {
      "host":"myHost",
      "key":"telegraf.docker_container_net.rx_errors[laughing_babbage]",
      "value":"0",
      "clock":15227640380
    },
    {
      "host":"myHost",
      "key":"telegraf.docker_container_net.tx_errors[laughing_babbage]",
      "value":"0",
      "clock":15227640380
    }
  ]
}
```
