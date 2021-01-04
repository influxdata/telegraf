# Derivative Aggregator Plugin
The Derivative Aggregator Plugin estimates the derivative for all fields of the
aggregated metrics.

### Time Derivatives

In its default configuration it determines the first and last measurement of
the period. From these measurements the time difference in seconds is
calculated. This time difference is than used to divide the difference of each
field using the following formula:
```
              field_last - field_first
derivative = --------------------------
                  time_difference
```
For each field the derivative is emitted with a naming pattern
`fieldname_by_seconds`.

### Custom Derivation Variable

The plugin supports to use a field of the aggregated measurements as derivation
variable in the denominator. This variable is assumed to be a monotonously
increasing value. In this feature the following formula is used:
```
                 field_last - field_first
derivative = --------------------------------
              variable_last - variable_first
```
For each field the derivative then is emitted with a naming pattern
`fieldname_by_variablename`.

### Roll-Over to next Period

By default the last measurement is used as first measurement in the next
aggregation period. This enables a continuous calculation of the derivative. If
within the next period an earlier timestamp is encountered this measurement will
replace the roll-over metric. A main benefit of this roll-over is the ability to
cope with multiple "quiet" periods, where no new measurement is pushed to the
aggregator. The roll-over will take place at most `max_roll_over`times.

### Configuration

```toml
[[aggregators.derivative]]
  ## Specific Derivative Aggregator Arguments:

  ## Configure a custom derivation variable. Timestamp is used if none is given.
  # variable = ""

  ## Infix to separate field name from derivation variable.
  # infix = "_by_"

  ## Roll-Over last measurement to first measurement of next period
  # max_roll_over = 10

  ## General Aggregator Arguments:

  ## calculate derivative every 30 seconds
  period = "30s"

  ## Fields for which the derivative should be calculated
  ## Important: The derivation variable must be contained in that list, if used.
  fieldpass = ["field1", "field2", "variable"]

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
```

### Tags:
No tags are applied by this aggregator.
Existing tags are passed throug the aggregator untouched.

### Example Output

```
net bytes_recv=15409i,packets_recv=164i,bytes_sent=16649i,packets_sent=120i 1508843640000000000
net bytes_recv=73987i,packets_recv=364i,bytes_sent=87328i,packets_sent=452i 1508843660000000000
net bytes_recv_by_packets_recv=292.89 1508843660000000000
net packets_sent_by_seconds=16.6,bytes_sent_by_seconds=3533.95 1508843660000000000
```
