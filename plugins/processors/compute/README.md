# Compute Processor

The compute processor is used to calculate field values of metrics by
specifying formulas for those target fields. These formulas can contain
constants, arithmetic operation as well as variables referencing to existing
field values.

When referencing fields in the formulas that do not exist, a strategy can be
provided to ignore or fill the missing values.

### Configuration
```toml
  # Compute values for a metric using the given formula(s)
  [[processors.compute]]
  ## Strategy to handle missing variables in cases where a formula refers to
  ## a non-existing field. Possible values are:
  ##		 ignore  - ignore formula for metric and do not update field
  ##     const   - target field will be replaced by the "constant" defined below
  ##     default - target field will be set to "default" defined below
  missing = "ignore"

  ## Constant to be used in the "const" strategy for missing fields
  # constant = 0

  ## Default value to be used in the "default" strategy for missing fields
  # default = 0

  ## Table of computations
  [processors.compute.fields]
    value = "(a + 3) / 4.3"
    x_sqr = "pow(a, 2)"
    x_abs = "abs(value)"
```

### Supported operations
- `+`, `-`, `/`, `*` basic arithmetic operations:
- `%` modulo operation (integer only)
- `abs(.)` absolute value function
- `pow(x,y)` `x` raised to the power of `y`

### Examples

Compute `deltaT` from sensor values `T_in` and `T_out`, convert the `operating`
time from seconds to hours and calcualte the `volumne` from a measured
`diameter`:
```toml

[processors.compute]]
  missing = "ignore"
  [processors.compute.fields]
    deltaT    = "T_out - T_in"
    operation = "operation / 3600.0"
    volume    = "4.0/3 * 3.1415 * pow(diameter, 3)"
```

```diff
- chp,host=telegraf diameter=31.5,T_in=19.8000000000001,operation=19625,T_out=63.7 1584111710000000000
+ chp,host=telegraf diameter=31.5,T_out=63.7,T_in=19.8000000000001,operation=5.451388888888889,deltaT=43.899999999999906,volume=130920.44174999998 1584111710000000000
```
*Please note*:  The `operation` fields gets overwritten by the computation.

Ignore all computations where the referenced `diameter` field is missing:
```toml
[processors.compute]]
  missing = "ignore"
  [processors.compute.fields]
    operation = "operation / 3600.0"
    volume    = "4.0/3 * 3.1415 * pow(diameter, 3)"
```

```diff
- machine1,host=telegraf operation=19625 1584111710000000000
+ machine1,host=telegraf operation=5.451388888888889, 1584111710000000000
```

Fill all computations where the referenced `diameter` field is missing with a
constant result:
```toml
[processors.compute]]
  missing = "const"
  constant = 42.0
  [processors.compute.fields]
    operation = "operation / 3600.0"
    volume    = "4.0/3 * 3.1415 * pow(diameter, 3)"
```

```diff
- machine1,host=telegraf operation=85786 1584111710000000000
+ machine1,host=telegraf operation=23.829444444444444,volume=42.0 1584111710000000000
```

Assume the missing `diameter` to be equal to the given `default` value during computation:
```toml
[processors.compute]]
  missing = "default"
  default = 3.0
  [processors.compute.fields]
    operation = "operation / 3600.0"
    volume    = "4.0/3 * 3.1415 * pow(diameter, 3)"
```

```diff
- machine1,host=telegraf operation=85786 1584111710000000000
+ machine1,host=telegraf operation=23.829444444444444,volume=113.094 1584111710000000000
```
