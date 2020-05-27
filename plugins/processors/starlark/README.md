# Starlark Processor

https://github.com/google/starlark-go/blob/master/doc/spec.md

### Configuration

```toml
[[processors.starlark]]
```

### Gotchas

don't return two references to the same metric.

error line number

### TODO

how to delete a metric?
- must call Drop?
- don't return: check returned values and autodrop

how to copy a metric?
- must call deepcopy()
- returning multiple references is an error

how to return multiple metrics?
- return a list of metric

how to create a new metric?

fastest way to iterate?

how to modify while iterating

how to retain metrics/modify globals
- global scope is froze

### Example
