# Scale Processor Plugin

The scale processor filters for a set of fields,
and scales the respective values from an input range into
the given output range according to this formula:

$$\textnormal{result}=(\textnormal{value}-\textnormal{input\_minimum})\cdot
\frac{(\textnormal{\textnormal{output\_maximum}}-\textnormal{output\_minimum})}
{(\textnormal{input\_maximum}-\textnormal{input\_minimum})} +
\textnormal{output\_minimum}$$

The input fields are expected to be numeric.
Strings representing a single numer are also allowed.
The scaled result will always be a floating point value.

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
    
    ## Multiple scalings can be defined simoultaniously
    ## Example: A second scaling. 
    # [processors.scale.scaling]
    #    input_minimum = -2800
    #    input_maximum = 100
    #    output_minimum = -20
    #    output_maximum = 40
    #    fields = ["humidity1", "humidity2"]
