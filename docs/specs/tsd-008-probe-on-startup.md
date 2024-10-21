# `probe` Option For `startup_error_behavior`

## Objective

Creation of a scheme that allows Telegraf to probe a plugin during its initialization phase to enable enhanced plugin error detection due to configuration or environmental problems.

## Keywords

inputs, outputs, startup, probe, error, ignore, behavior

## Overview

When plugins are first instantiated, Telegraf will call the plugin's `Start()` method (for inputs) or `Connect()` (for outputs) which will initialize its configuration based off of config options and the running environment. It is sometimes the case that while the initialization step succeeds, the upstream service in which the plugin relies on is not actually running, or is not capable of being communicated with due to incorrect configuration or environmental problems. In situations like this, Telegraf does not detect that the plugin's upstream service is not functioning properly, and thus it will continually call the plugin during each `Gather()` iteration. This often has the effect of polluting journald and system logs with voluminous error messages, which creates issues for system administrators who rely on such logs to identify other unrelated system problems.

More background discussion on this option, including other possible avenues, can be viewed [here](https://github.com/influxdata/telegraf/issues/16028).

## Probing

First, we must define what it means for a plugin to `Probe`. Probing is an action whereby the plugin will attempt to communicate with its external service, device, entity, or executable as if it were gathering real metrics. The payload of this communication attempt is not used to record any metrics. The result of a `Probe` shall be an error that indicates whether the communication was successful or not.

We define an interface for this action:

```go
type Prober interface {
    Probe() error
}
```

In almost all respects, a `Probe` is similar to a `Gather`, the only difference being that `Probe` does not record any metrics and only returns information on the success or failure of the action.

## Configuration

The Telegraf project has already introduced the `startup_error_behavior` configuration which allows the user to define how Telegraf should behave if the plugin fails to start. The `ignore` value allows the plugin to be ignored if the `Start()`/`Connect()` method returns an error. We propose to add an additional value to this parameter called `probe`, that will behave as a superset of the `ignore` behavior. When `startup_error_behavior=probe`, Telegraf will perform the following steps on plugin startup:

1. Call the `Start()` or `Connect()` method.
2. If an error is returned, cause the plugin to be ignored in the same manner as if `ignore` was specified.
3. If the plugin implements the `Prober` interface:
   1. Call `Probe()`.
   2. If `Probe()` returns an error, cause the plugin to be ignored.


## Plugin Requirements

As already stated, plugins participating in the `probe` scheme must implement the `Prober` interface. The exact way the plugin implements the behavior will depend on the plugin in question, but generally it should take the same actions as it would with `Gather()` or `Write()`, such that failures during `Gather()`/`Write()` would also imply a failure of `Probe()`. If `Probe()` returns an error, the plugin's `Close()` method should still be safe to call.

It should be noted that for output plugins, it's advisable to implement `Probe()` in a way that doesn't write real metrics to the backend if possible. How this might be done depends on the output in question.

Plugins should return `nil` upon successful probes or a retriable error (via predefined error type) to enable the defined ignoring behavior. A non-retriable error (via predefined error type) or a generic error will bypass the startup error behaviors and Telegraf must fail and exit in the startup phase.

## Related Issues

- [#16028](https://github.com/influxdata/telegraf/issues/16028)
- [#15916](https://github.com/influxdata/telegraf/pull/15916)
- [#16001](https://github.com/influxdata/telegraf/pull/16001)

