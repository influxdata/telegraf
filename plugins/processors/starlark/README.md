# Starlark Processor

https://github.com/google/starlark-go/blob/master/doc/spec.md

Retaining metrics:
- must use deepcopy function
- or copy to a new type
- i could also freeze and disallow retaining copies.  this is more in spirit of starlark's intentions.
- global scope is frozen

### Configuration

```toml
[[processors.starlark]]
```

### Gotchas

don't return two references to the same metric.

error line number

### TODO

what if a metric deleted
- must call Drop?
- check returned values and autodrop

what if a metric is copied
- must call deepcopy()
- returning multiple references is an error

disallow remove, add, clear during iteration

### Example
