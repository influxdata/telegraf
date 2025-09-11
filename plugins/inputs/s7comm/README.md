# Siemens S7 Input Plugin

This plugin reads metrics from Siemens PLCs via the S7 protocol.

⭐ Telegraf v1.28.0
🏷️ hardware
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Startup error behavior options <!-- @/docs/includes/startup_error_behavior.md -->

In addition to the plugin-specific and global configuration settings the plugin
supports options for specifying the behavior when experiencing startup errors
using the `startup_error_behavior` setting. Available values are:

- `error`:  Telegraf with stop and exit in case of startup errors. This is the
            default behavior.
- `ignore`: Telegraf will ignore startup errors for this plugin and disables it
            but continues processing for all other plugins.
- `retry`:  Telegraf will try to startup the plugin in every gather or write
            cycle in case of startup errors. The plugin is disabled until
            the startup succeeds.
- `probe`:  Telegraf will probe the plugin's function (if possible) and disables
            the plugin in case probing fails. If the plugin does not support
            probing, Telegraf will behave as if `ignore` was set instead.

## Configuration

```toml @sample.conf
# Plugin for retrieving data from Siemens PLCs via the S7 protocol (RFC1006)
[[inputs.s7comm]]
  ## Parameters to contact the PLC (mandatory)
  ## The server is in the <host>[:port] format where the port defaults to 102
  ## if not explicitly specified.
  server = "127.0.0.1:102"
  rack = 0
  slot = 0

  ## Connection or drive type of S7 protocol
  ## Available options are "PD" (programming  device), "OP" (operator panel) or "basic" (S7 basic communication).
  # connection_type = "PD"

  ## Max count of fields to be bundled in one batch-request. (PDU size)
  # pdu_size = 20

  ## Timeout for requests
  # timeout = "10s"

  ## Log detailed connection messages for tracing issues
  # log_level = "trace"

  ## Metric definition(s)
  [[inputs.s7comm.metric]]
    ## Name of the measurement
    # name = "s7comm"

    ## Field definitions
    ## name    - field name
    ## address - indirect address "<area>.<type><address>[.extra]"
    ##           area    - e.g. be "DB1" for data-block one
    ##           type    - supported types are (uppercase)
    ##                     X  -- bit, requires the bit-number as 'extra'
    ##                           parameter
    ##                     B  -- byte (8 bit)
    ##                     C  -- character (8 bit)
    ##                     W  -- word (16 bit)
    ##                     DW -- double word (32 bit)
    ##                     I  -- integer (16 bit)
    ##                     DI -- double integer (32 bit)
    ##                     LI -- long integer (64 bit) only S7-1200 S7-1500 suported
    ##                     R  -- IEEE 754 real floating point number (32 bit)
    ##                     LR -- IEEE 754 long real floating point number (64 bit) only S7-1200 S7-1500 suported
    ##                     DT -- date-time, always converted to unix timestamp
    ##                           with nano-second precision
    ##                     S  -- string, requires the maximum length of the
    ##                           string as 'extra' parameter
    ##           address - start address to read if not specified otherwise
    ##                     in the type field
    ##           extra   - extra parameter e.g. for the bit and string type
    fields = [
      { name="rpm",             address="DB1.R4"    },
      { name="status_ok",       address="DB1.X2.1"  },
      { name="last_error",      address="DB2.S1.32" },
      { name="last_error_time", address="DB2.DT2"   },
      { name="long_counter",    address="DB3.LR12"  }
    ]

    ## Tags assigned to the metric
    # [inputs.s7comm.metric.tags]
    #   device = "compressor"
    #   location = "main building"
```

## Example Output

```text
s7comm,host=Hugin rpm=712i,status_ok=true,last_error="empty slot",last_error_time=1611319681000000000i 1611332164000000000
```

## Metrics

The format of metrics produced by this plugin depends on the metric
configuration(s).
