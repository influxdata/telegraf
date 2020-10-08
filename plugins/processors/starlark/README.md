# Starlark Processor

The `starlark` processor calls a Starlark function for each matched metric,
allowing for custom programmatic metric processing.

The Starlark language is a dialect of Python, and will be familiar to those who
have experience with the Python language. However, there are major [differences](#python-differences).
Existing Python code is unlikely to work unmodified.  The execution environment
is sandboxed, and it is not possible to do I/O operations such as reading from
files or sockets.

The **[Starlark specification][]** has details about the syntax and available
functions.

Telegraf minimum version: Telegraf 1.15.0

### Configuration

```toml
[[processors.starlark]]
  ## The Starlark source can be set as a string in this configuration file, or
  ## by referencing a file containing the script.  Only one source or script
  ## should be set at once.

  ## Source of the Starlark script.
  source = '''
def apply(metric):
	return metric
'''

  ## File containing a Starlark script.
  # script = "/usr/local/bin/myscript.star"
```

### Usage

The Starlark code should contain a function called `apply` that takes a metric as
its single argument.  The function will be called with each metric, and can
return `None`, a single metric, or a list of metrics.

```python
def apply(metric):
	return metric
```

For a list of available types and functions that can be used in the code, see
the [Starlark specification][].

In addition to these, the following InfluxDB-specific
types and functions are exposed to the script.

- **Metric(*name*)**:
Create a new metric with the given measurement name.  The metric will have no
tags or fields and defaults to the current time.

- **name**:
The name is a [string][] containing the metric measurement name.

- **tags**:
A [dict-like][dict] object containing the metric's tags.

- **fields**:
A [dict-like][dict] object containing the metric's fields.  The values may be
of type int, float, string, or bool.

- **time**:
The timestamp of the metric as an integer in nanoseconds since the Unix
epoch.

- **deepcopy(*metric*)**: Make a copy of an existing metric.

### Python Differences

While Starlark is similar to Python, there are important differences to note:

- Starlark has limited support for error handling and no exceptions.  If an
  error occurs the script will immediately end and Telegraf will drop the
  metric.  Check the Telegraf logfile for details about the error.

- It is not possible to import other packages and the Python standard library
  is not available.

- It is not possible to open files or sockets.

- These common keywords are **not supported** in the Starlark grammar:
  ```
  as             finally        nonlocal
  assert         from           raise
  class          global         try
  del            import         with
  except         is             yield
  ```

### Libraries available

The ability to load external scripts other than your own is pretty limited. The following libraries are available for loading:

* json: `load("json.star", "json")` provides the following functions: `json.encode()`, `json.decode()`, `json.indent()`. See [json.star](/plugins/processors/starlark/testdata/json.star) for an example.

If you would like to see support for something else here, please open an issue.

### Common Questions

**How can I drop/delete a metric?**

If you don't return the metric it will be deleted.  Usually this means the
function should `return None`.

**How should I make a copy of a metric?**

Use `deepcopy(metric)` to create a copy of the metric.

**How can I return multiple metrics?**

You can return a list of metrics:

```python
def apply(metric):
    m2 = deepcopy(metric)
    return [metric, m2]
```

**What happens to a tracking metric if an error occurs in the script?**

The metric is marked as undelivered.

**How do I create a new metric?**

Use the `Metric(name)` function and set at least one field.

**What is the fastest way to iterate over tags/fields?**

The fastest way to iterate is to use a for-loop on the tags or fields attribute:

```python
def apply(metric):
    for k in metric.tags:
        pass
    return metric
```

When you use this form, it is not possible to modify the tags inside the loop,
if this is needed you should use one of the `.keys()`, `.values()`, or `.items()` methods:

```python
def apply(metric):
    for k, v in metric.tags.items():
        pass
    return metric
```

**How can I save values across multiple calls to the script?**

Telegraf freezes the global scope, which prevents it from being modified.
Attempting to modify the global scope will fail with an error.


### Examples

- [json](/plugins/processors/starlark/testdata/json.star) - an example of processing JSON from a field in a metric
- [number logic](/plugins/processors/starlark/testdata/number_logic.star) - transform a numerical value to another numerical value
- [pivot](/plugins/processors/starlark/testdata/pivot.star) - Pivots a key's value to be the key for another key.
- [ratio](/plugins/processors/starlark/testdata/ratio.star) - Compute the ratio of two integer fields
- [rename](/plugins/processors/starlark/testdata/rename.star) - Rename tags or fields using a name mapping.
- [scale](/plugins/processors/starlark/testdata/scale.star) - Multiply any field by a number
- [value filter](/plugins/processors/starlark/testdata/value_filter.star) - remove a metric based on a field value.

[All examples](/plugins/processors/starlark/testdata) are in the testdata folder.

Open a Pull Request to add any other useful Starlark examples.

[Starlark specification]: https://github.com/google/starlark-go/blob/master/doc/spec.md
[string]: https://github.com/google/starlark-go/blob/master/doc/spec.md#strings
[dict]: https://github.com/google/starlark-go/blob/master/doc/spec.md#dictionaries
