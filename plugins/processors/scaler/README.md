# Scaler Processor Plugin

The scaler processor filters for a set of fields and scales the respective values from an input range in to a given output range according to this formula:

$$\textnormal{result}=(\textnormal{value}-\textnormal{input\_minimum})\cdot \frac{(\textnormal{\textnormal{output\_maximum}}-\textnormal{output\_minimum})}{(\textnormal{input\_maximum}-\textnormal{input\_minimum})} + \textnormal{output\_minimum}$$

Nither the input, not the output values, are required to be in the respective ranges.
However, it is required, that input_minimum and input_maximum do not have the same value.

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

    # It is possible to define multiple different scalings that can be applied do different sets of fields
    # Each scaling expects five arguments:
    #   - input_minimum: Minimum expected input value
    #   - input_maximum: Maximum expected input value
    #   - output_minimum: Minimum desired output value
    #   - output_maximum: Maximum desired output value
    #   - fields: a list of field names (or filters) to apply this scaling to
    
    # Convert Fahrenheit to Celsius
    [processors.scaler.scaling]
        input_minimum = 32
        input_maximum = 212
        output_minimum = 0
        output_maximum = 100
        fields = ["temperature1", "temperature2"]
        

    # Defined a second scaling. 
    [processors.scaler.scaling]
        input_minimum = -2800
        input_maximum = 100
        output_minimum = -20
        output_maximum = 40
        fields = ["humidity1", "humidity2"]
```

## Tags

No tags are applied by this processor.
