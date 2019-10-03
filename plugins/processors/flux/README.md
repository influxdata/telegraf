# Flux Processor

The Flux processor can transform input data using a Flux script.

The script can access input data via `telegraf.from()` and can access the input metric name via the `_measurement` column.

For example:

```flux
import "telegraf"

telegraf.from()
  |> map(fn: (r) => ({r with value: r.value - 1}))
  // This would change the output metric name.
  |> map(fn: (r) => ({r with _measurement: "changed"}))
```

When the debug mode is active, the script does not run, and the output is saved to the specified path.

### Configuration:
```toml
# # Runs a flux script to process inputs and produce outputs
# [[processors.flux]]
#   ## Path to a Flux script.
#   path = "path/to/script.flux"
#   ## Enables debug mode if an output file is specified.
#   debug = "path/to/out.csv"
```

### Examples:

```toml
[[processors.flux]]
    path = "path/to/script.flux"
```

Content of `path/to/script.flux`:

```flux
import "telegraf"

telegraf.from()
  |> map(fn: (r) => ({r with usage: r.usage * 2})))
```

```diff
- cpu,host=macbook usage=33 1502489900000000000
+ cpu,host=macbook usage=66 1502489900000000000
```
