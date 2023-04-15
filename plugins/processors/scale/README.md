# Scale Processor Plugin

The scale processor filters for a set of fields,
and scales the respective values from an input range into
the given output range according to this formula:

```math
\text{result}=(\text{value}-\text{input\_minimum})\cdot\frac{(\text{output\_maximum}-\text{output\_minimum})}
{(\text{input\_maximum}-\text{input\_minimum})} +
\text{output\_minimum}
```

Input fields are converted to floating point values.
If the conversion fails, those fields are ignored.

**Please note:** Neither the input nor the output values
are clipped to their respective ranges!

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Scale values with a predefined range to a different output range.
[[processors.scale]]
    ## It is possible to define multiple different scaling that can be applied
    ## do different sets of fields. Each scaling expects the following
    ## arguments:
    ##   - input_minimum: Minimum expected input value
    ##   - input_maximum: Maximum expected input value
    ##   - output_minimum: Minimum desired output value
    ##   - output_maximum: Maximum desired output value
    ##   - fields: a list of field names (or filters) to apply this scaling to
    
    ## Example: Define a scaling
    # [processors.scale.scaling]
    #    input_minimum = 0
    #    input_maximum = 1
    #    output_minimum = 0
    #    output_maximum = 100
    #    fields = ["temperature1", "temperature2"]
    
    ## Multiple scalings can be defined simultaneously
    ## Example: A second scaling. 
    # [processors.scale.scaling]
    #    input_minimum = 0
    #    input_maximum = 50
    #    output_minimum = 50
    #    output_maximum = 100
    #    fields = ["humidity*"]
```

## Example

```diff
- temperature, cpu=25
+ temperature, cpu=75.0
```
