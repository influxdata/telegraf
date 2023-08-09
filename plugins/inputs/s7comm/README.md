# Siemens S7 Input Plugin

This plugin reads metrics from Siemens PLCs via the S7 protocol.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

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

  ## Timeout for requests
  # timeout = "10s"

  ## Log detailed connection messages for debugging
  ## This option only has an effect when Telegraf runs in debug mode
  # debug_connection = false

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
    ##                     R  -- IEEE 754 real floating point number (32 bit)
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
      { name="last_error_time", address="DB2.DT2"   }
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
