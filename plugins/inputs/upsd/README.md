# UPSD Input Plugin

This plugin reads data of one or more Uninterruptible Power Supplies from a
[Network UPS Tools][upsd] daemon using its NUT network protocol.

⭐ Telegraf v1.24.0
🏷️ hardware, server
💻 all

[upsd]: https://networkupstools.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Monitor UPSes connected via Network UPS Tools
[[inputs.upsd]]
  ## A running NUT server to connect to.
  ## IPv6 addresses must be enclosed in brackets (e.g. "[::1]")
  # server = "127.0.0.1"
  # port = 3493
  # username = "user"
  # password = "password"

  ## Force parsing numbers as floats
  ## It is highly recommended to enable this setting to parse numbers
  ## consistently as floats to avoid database conflicts where some numbers are
  ## parsed as integers and others as floats.
  # force_float = false

  ## Emit vendor/product IDs as strings regardless of their value. Avoids
  ## type conflicts when some UPS devices report numeric-looking IDs and
  ## others report alphanumeric. See README for migration notes.
  # stringify_ids = false

  ## Collect additional fields if they are available for the UPS
  ## The fields need to be specified as NUT variable names, see
  ## https://networkupstools.org/docs/developer-guide.chunked/apas02.html
  ## Wildcards are accepted.
  # additional_fields = []

  ## Dump information for debugging
  ## Allows to print the raw variables (and corresponding types) as received
  ## from the NUT server ONCE for each UPS.
  ## Please attach this information when reporting issues!
  # log_level = "trace"
```

### Pitfalls

Please note that field types are automatically determined based on the values.
Especially the strings `enabled` and `disabled` are automatically converted to
`boolean` values. This might lead to trouble for fields that can contain
non-binary values like `enabled`, `disabled` and `muted` as the output field
will be `boolean` for the first two values but `string` for the latter. To
convert `enabled` and `disabled` values back to `string` for those fields, use
the [enum processor][enum_processor] with

```toml
[[processors.enum]]
  [[processors.enum.mapping]]
    field = "ups_beeper_status"
    [processors.enum.mapping.value_mappings]
      true = "enabled"
      false = "disabled"
```

Alternatively, you can also map the non-binary value to a `boolean`.

[enum_processor]: /plugins/processors/enum/README.md

### Vendor/Product ID types (`stringify_ids`)

The underlying NUT client library (`go.nut`) auto-detects numeric-looking values
and converts them to `int64`. This means a `vendorid` like `"0764"` becomes
`int64(764)` while a non-numeric `vendorid` like `"ABCD"` stays a string. When
multiple UPS devices of different vendors write to the same InfluxDB bucket,
this causes field type conflicts on `ups_vendorid`, `ups_productid`,
`driver_parameter_vendorid` and `driver_parameter_productid`.

Set `stringify_ids = true` to force these four fields to always be emitted as
strings. The default is currently `false` to preserve backwards-compatible
behavior, but will flip to `true` in a future release. If the option is left
unset, a warning is logged on startup.

> [!NOTE]
> The NUT library parses `"0764"` into `int64(764)` before Telegraf sees it,
> so the stringified value will be `"764"`; leading zeros are lost and cannot
> be recovered at this layer.

## Metrics

This implementation tries to maintain compatibility with the apcupsd metrics:

- upsd
  - tags:
    - serial
    - ups_name
    - model
  - fields:
    - status_flags ([status-bits][rfc9271-sec5.1])
    - input_voltage
    - load_percent
    - battery_charge_percent
    - time_left_ns
    - output_voltage
    - internal_temp
    - battery_voltage
    - input_frequency
    - battery_date
    - nominal_input_voltage
    - nominal_battery_voltage
    - nominal_power
    - firmware

With the exception of:

- tags:
  - status (string representing the set status_flags)
- fields:
  - time_on_battery_ns

[rfc9271-sec5.1]: https://www.rfc-editor.org/rfc/rfc9271.html#section-5.1

## Example Output

```text
upsd,serial=AS1231515,ups_name=name1 load_percent=9.7,time_left_ns=9800000,output_voltage=230.4,internal_temp=32.4,battery_voltage=27.4,input_frequency=50.2,input_voltage=230.4,battery_charge_percent=100,status_flags=8i 1490035922000000000
```
