# Starlark Processor

The `starlark` processor calls a Starlark function for each matched metric,
allowing for custom programmatic metric processing.

The Starlark language is a dialect of Python, and will be familiar to those who
have experience with the Python language.  However, keep in mind that it is not
Python and that there are major syntax [differences][#Python Differences].
Existing Python code is unlikely to work unmodified.  The execution environment
is sandboxed, and it is not possible to do I/O operations such as reading from
files or sockets.

The Starlark [specification][] has details about the syntax and available
functions.

### Configuration

```toml
[[processors.starlark]]
  ## The Starlark source can be set as a string in this configuration file, or
  ## by referencing a file containing the script.  Only one source or script
  ## should be set at once.
  ##
  ## Source of the Starlark script.
  source = '''
def apply(metric):
	return metric
'''

  ## File containing a Starlark script.
  # script = "/usr/local/bin/myscript.star"
```

### Usage

The script should contain a function called `apply` that takes the metric as
its single argument.  The function will be called with each metric, and can
return `None`, a single metric, or a list of metrics.
```python
def apply(metric):
	return metric
```

Reference the Starlark [specification][] to see the list of available types and
functions that can be used in the script.  In addition to these the following
types and functions are exposed to the script.

**Metric(*name*)**:
Create a new metric with the given measurement name.  The metric will have no
tags or fields and defaults to the current time.

- **name**:
The name is a [string][string] containing the metric measurement name.

- **tags**:
A [dict-like][dict] object containing the metric's tags.


- **fields**:
A [dict-like][dict] object containing the metric's fields.  The values may be
of type int, float, string, or bool.

- **time**:
The timestamp of the metric as an integer in nanoseconds since the Unix
epoch.

**deepcopy(*metric*)**: Make a copy of an existing metric.

### Python Differences

While Starlark is similar to Python it is not the same.

- Starlark has limited support for error handling and no exceptions.  If an
  error occurs the script will immediately end and Telegraf will drop the
  metric.  Check the Telegraf logfile for details about the error.

- It is not possible to import other packages and the Python standard library
  is not available.  As such, it is not possible to open files or sockets.

- These common keywords are **not supported** in the Starlark grammar:
  ```
  as             finally        nonlocal
  assert         from           raise
  class          global         try
  del            import         with
  except         is             yield
  ```

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
if this is needed you should use the `.keys()`, `.values()`, or `.items()` forms:
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

- [ratio](/plugins/processors/starlark/testdata/ratio.star)
- [rename](/plugins/processors/starlark/testdata/rename.star)
- [scale](/plugins/processors/starlark/testdata/scale.star)

[specification]: https://github.com/google/starlark-go/blob/master/doc/spec.md
[string]: https://github.com/google/starlark-go/blob/master/doc/spec.md#strings
[dict]: https://github.com/google/starlark-go/blob/master/doc/spec.md#dictionaries
