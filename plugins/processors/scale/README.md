# Scale Processor Plugin

The scale processor filters for a set of fields,
and scales the respective values from an input range into
the given output range according to this formula:

```math
\text{result}=(\text{value}-\text{input\_minimum})\cdot\frac{(\text{output\_maximum}-\text{output\_minimum})}
{(\text{input\_maximum}-\text{input\_minimum})} +
\text{output\_minimum}
```

Alternatively, you can apply a factor and offset to the input according to
this formula

```math
\text{result}=\text{factor} \cdot \text{value} + \text{offset}
```

Input fields are converted to floating point values if possible. Otherwise,
fields that cannot be converted are ignored and keep their original value.

**Please note:** Neither the input nor the output values are clipped to their
                 respective ranges!

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
    ## alternatively you can specify a scaling with factor and offset
    ##   - factor: factor to scale the input value with
    ##   - offset: additive offset for value after scaling
    ##   - fields: a list of field names (or filters) to apply this scaling to

    ## Example: Scaling with minimum and maximum values
    # [processors.scale.scaling]
    #    input_minimum = 0
    #    input_maximum = 1
    #    output_minimum = 0
    #    output_maximum = 100
    #    fields = ["temperature1", "temperature2"]

    ## Example: Scaling with factor and offset
    # [processors.scale.scaling]
    #    factor = 10.0
    #    offset = -5.0
    #    fields = ["voltage*"]
```

## Example

The example below uses these scaling values:

```toml
[processors.scale.scaling]
    input_minimum = 0
    input_maximum = 50
    output_minimum = 50
    output_maximum = 100
    fields = ["cpu"]
```

```diff
- temperature, cpu=25
+ temperature, cpu=75.0
```
