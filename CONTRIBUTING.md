### Contributing

1. [Sign the CLA][cla].
1. Open a [new issue][] to discuss the changes you would like to make.  This is
   not strictly required but it may help reduce the amount of rework you need
   to do later.
1. Make changes or write plugin using the guidelines in the following
   documents:
   - [Input Plugins][inputs]
   - [Processor Plugins][processors]
   - [Aggregator Plugins][aggregators]
   - [Output Plugins][outputs]
1. Ensure you have added proper unit tests and documentation.
1. Open a new [pull request][].

### GoDoc

Public interfaces for inputs, outputs, processors, aggregators, metrics,
and the accumulator can be found in the GoDoc:

[![GoDoc](https://godoc.org/github.com/influxdata/telegraf?status.svg)](https://godoc.org/github.com/influxdata/telegraf)

### Common development tasks

**Adding a dependency:**

Assuming you can already build the project, run these in the telegraf directory:

1. `dep ensure -vendor-only`
2. `dep ensure -add github.com/[dependency]/[new-package]`

**Unit Tests:**

Before opening a pull request you should run the linter checks and
the short tests.

**Run static analysis:**

```
make check
```

**Run short tests:**

```
make test
```

**Execute integration tests:**

Running the integration tests requires several docker containers to be
running.  You can start the containers with:
```
make docker-run
```

And run the full test suite with:
```
make test-all
```

Use `make docker-kill` to stop the containers.


[cla]: https://www.influxdata.com/legal/cla/
[new issue]: https://github.com/influxdata/telegraf/issues/new/choose
[pull request]: https://github.com/influxdata/telegraf/compare
[inputs]: /docs/INPUTS.md
[processors]: /docs/PROCESSORS.md
[aggregators]: /docs/AGGREGATORS.md
[outputs]: /docs/OUTPUTS.md
