# Multiplier Processor

The multiplier processor is used to change the values of metrics with some multiplication factor.

### Examples:

* [input.system](./plugins/inputs/system) plugin produces metric [uptime](./plugins/inputs/system/SYSTEM_README.md#metrics) with amount of seconds after computer start. If it is needed to get value in hours multiplier configuration should contain
"system uptime=0.00027777777"

* Some **metric** can contain **value** in range [0;&nbsp;1]. If it is needed to get values in more human readable percentage format multiplier configuration should contain
"metric value=100"

**Note:** Multiplied values keep thier original types.

**Note:** Values that cannot be multiplied remain unchanged.

### Configuration:
```
# Multiply metrics values on some multiply factor
[[processors.multiplier]]
  ## Config can contain multiply factors for each metrics.
  ## Each config line should be the string in influx format.
  Config = [
    "mem used_percent=100,available_percent=100",
    "swap used_percent=100"
  ]

  # VerboseMode allows to print changes for debug purpose
  VerboseMode = false
```
