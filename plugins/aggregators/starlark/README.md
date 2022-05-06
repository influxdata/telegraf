# Starlark Aggregator

The `starlark` aggregator allows to implement a custom aggregator plugin with a Starlark script. The Starlark
script needs to be composed of the three methods defined in the Aggregator plugin interface which are `add`, `push` and `reset`.

The Starlark Aggregator plugin calls the Starlark function `add` to add the metrics to the aggregator, then calls the Starlark function `push` to push the resulting metrics into the accumulator and finally calls the Starlark function `reset` to reset the entire state of the plugin.

The Starlark functions can use the global function `state` to keep temporary the metrics to aggregate.

The Starlark language is a dialect of Python, and will be familiar to those who
have experience with the Python language. However, there are major [differences](#python-differences).
Existing Python code is unlikely to work unmodified.  The execution environment
is sandboxed, and it is not possible to do I/O operations such as reading from
files or sockets.

The **[Starlark specification][]** has details about the syntax and available
functions.

## Configuration

```toml
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

The Starlark code should contain a function called `add` that takes a metric as argument.
The function will be called with each metric to add, and doesn't return anything.

```python
def add(metric):
  state["last"] = metric
```

The Starlark code should also contain a function called `push` that doesn't take any argument.
The function will be called to compute the aggregation, and returns the metrics to push to the accumulator.

```python
def push():
  return state.get("last")
```

The Starlark code should also contain a function called `reset` that doesn't take any argument.
The function will be called to reset the plugin, and doesn't return anything.

```python
def push():
  state.clear()
```

For a list of available types and functions that can be used in the code, see
the [Starlark specification][].

## Python Differences

Refer to the section [Python Differences](plugins/processors/starlark/README.md#python-differences) of the documentation about the Starlark processor.

## Libraries available

Refer to the section [Libraries available](plugins/processors/starlark/README.md#libraries-available) of the documentation about the Starlark processor.

## Common Questions

Refer to the section [Common Questions](plugins/processors/starlark/README.md#common-questions) of the documentation about the Starlark processor.

## Examples

- [minmax](/plugins/aggregators/starlark/testdata/min_max.star) - A minmax aggregator implemented with a Starlark script.
- [merge](/plugins/aggregators/starlark/testdata/merge.star) - A merge aggregator implemented with a Starlark script.

[All examples](/plugins/aggregators/starlark/testdata) are in the testdata folder.

Open a Pull Request to add any other useful Starlark examples.

[Starlark specification]: https://github.com/google/starlark-go/blob/d1966c6b9fcd/doc/spec.md
[dict]: https://github.com/google/starlark-go/blob/d1966c6b9fcd/doc/spec.md#dictionaries
