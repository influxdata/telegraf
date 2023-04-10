# Scaler Processor Plugin

The scaler processor filters for a set of fields,
and scales the respective values from an input range into
the given output range according to this formula:

(value - input_minimum) * (output_maximum - output_minimum)
/ (input_maximum - input_minimum) + output_maximum

Nither the input, nor the output values
are constrained into their respective ranges.
However, it is required, that input_minimum and
input_maximum do not have the same value.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Scale values with a predefined range to a different output range.
[[processors.scaler]]
    ## It is possible to define multiple different scaling that can be applied
    ## do different sets of fields. Each scaling expects the following
    ## arguments:
    ##   - input_minimum: Minimum expected input value
    ##   - input_maximum: Maximum expected input value
    ##   - output_minimum: Minimum desired output value
    ##   - output_maximum: Maximum desired output value
    ##   - fields: a list of field names (or filters) to apply this scaling to
    
    ## Example: convert Fahrenheit to Celsius
    # [processors.scaler.scaling]
    #    input_minimum = 32
    #    input_maximum = 212
    #    output_minimum = 0
    #    output_maximum = 100
    #    fields = ["temperature1", "temperature2"]
        
    ## Example: A second scaling. 
    # [processors.scaler.scaling]
    #    input_minimum = -2800
    #    input_maximum = 100
    #    output_minimum = -20
    #    output_maximum = 40
    #    fields = ["humidity1", "humidity2"]
