# Starlark Aggregator Plugin

This plugin allows to implement a custom aggregator plugin via a
[Starlark][starlark] script.

The Starlark language is a dialect of Python and will be familiar to those who
have experience with the Python language. However, there are major
[differences](#python-differences). Existing Python code is unlikely to work
unmodified.

> [!NOTE]
> The execution environment is sandboxed, and it is not possible to access the
> local filesystem or perfoming network operations. This is by design of the
> Starlark language as a configuration language.

The Starlark script used by this plugin needs to be composed of the three
methods defining an aggreagtor named `add`, `push` and `reset`.

The `add` method is called as soon as a new metric is added to the plugin the
metrics to the aggregator. After `period`, the `push` method is called to
output the resulting metrics and finally the aggregation is reset by using the
`reset` method of the Starlark script.

The Starlark functions might use the global function `state` to keep aggregation
information such as added metrics etc.

More details on the syntax and available functions can be found in the
[Starlark specification][spec].

⭐ Telegraf v1.21.0
🏷️ transformation
💻 all

[starlark]: https://github.com/google/starlark-go
[spec]: https://github.com/google/starlark-go/blob/d1966c6b9fcd/doc/spec.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Aggregate metrics using a Starlark script
[[aggregators.starlark]]
  ## The Starlark source can be set as a string in this configuration file, or
  ## by referencing a file containing the script.  Only one source or script
  ## should be set at once.
  ##
  ## Source of the Starlark script.
  source = '''
state = {}

def add(metric):
  state["last"] = metric

def push():
  return state.get("last")

def reset():
  state.clear()
'''

  ## File containing a Starlark script.
  # script = "/usr/local/bin/myscript.star"

  ## The constants of the Starlark script.
  # [aggregators.starlark.constants]
  #   max_size = 10
  #   threshold = 0.75
  #   default_name = "Julia"
  #   debug_mode = true
```

## Usage

The Starlark code should contain a function called `add` that takes a metric as
argument.  The function will be called with each metric to add, and doesn't
return anything.

```python
def add(metric):
  state["last"] = metric
```

The Starlark code should also contain a function called `push` that doesn't take
any argument.  The function will be called to compute the aggregation, and
returns the metrics to push to the accumulator.

```python
def push():
  return state.get("last")
```

The Starlark code should also contain a function called `reset` that doesn't
take any argument.  The function will be called to reset the plugin, and doesn't
return anything.

```python
def reset():
  state.clear()
```

For a list of available types and functions that can be used in the code, see
the [Starlark specification][spec].

## Python Differences

Refer to the section [Python
Differences](../../processors/starlark/README.md#python-differences) of the
documentation about the Starlark processor.

## Libraries available

Refer to the section [Libraries
available](../../processors/starlark/README.md#libraries-available) of the
documentation about the Starlark processor.

## Common Questions

Refer to the section [Common
Questions](../../processors/starlark/README.md#common-questions) of the
documentation about the Starlark processor.

## Examples

- [minmax](testdata/min_max.star)
- [merge](testdata/merge.star)

[All examples](testdata) are in the testdata folder.

Open a Pull Request to add any other useful Starlark examples.
