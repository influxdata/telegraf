# Probing plugins after startup

## Objective

Allow Telegraf to probe plugins during startup to enable enhanced plugin error 
detection like availability of hardware or services

## Keywords

inputs, outputs, startup, probe, error, ignore, behavior

## Overview

When plugins are first instantiated, Telegraf will call the plugin's `Start()` 
method (for inputs) or `Connect()` (for outputs) which will initialize its 
configuration based off of config options and the running environment. It is 
sometimes the case that while the initialization step succeeds, the upstream 
service in which the plugin relies on is not actually running, or is not capable 
of being communicated with due to incorrect configuration or environmental 
problems. In situations like this, Telegraf does not detect that the plugin's 
upstream service is not functioning properly, and thus it will continually call 
the plugin during each `Gather()` iteration. This often has the effect of 
polluting journald and system logs with voluminous error messages, which creates 
issues for system administrators who rely on such logs to identify other 
unrelated system problems.

More background discussion on this option, including other possible avenues, can 
be viewed [here](https://github.com/influxdata/telegraf/issues/16028).

## Probing

Probing is an action whereby the plugin should ensure that the plugin will be
fully functional on a best effort basis. This may comprise communicating with
its external service, trying to access required devices, entities or executables
etc  to ensure that the plugin will not produce errors during e.g. data collection
or data output. Probing must *not* produce, process or output any metrics.

Plugins that support probing must implement the `ProbePlugin` interface. Such 
plugins must behave in the following manner:

1. Return an error if the external dependencies (hardware, services, 
executables, etc.) of the plugin are not available.
2. Return an error if information cannot be gathered (in the case of inputs) or 
sent (in the case of outputs) due to unrecoverable issues. For example, invalid 
authentication, missing permissions, or non-existent endpoints.
3. Otherwise, return `nil` indicating the plugin will be fully functional.

## Configuration

The Telegraf project has already introduced the `startup_error_behavior` 
configuration which allows the user to define how Telegraf should behave if the 
plugin fails to start. The `ignore` value allows the plugin to be ignored if the 
`Start()`/`Connect()` method returns an error. We propose to add an additional 
value to this parameter called `probe`, that will behave as a superset of the 
`ignore` behavior. When `startup_error_behavior=probe`, Telegraf will perform 
the following steps on plugin startup:

1. Check if the plugin implements `ProbePlugin`. If it does not, Telegraf will 
fatally exit with a log message indicating that the supplied configuration is 
invalid for the plugin.
2. Call the `Start()` or `Connect()` method.
3. If an error is returned, cause the plugin to be ignored in the same manner as 
if `ignore` was specified.
4. Call `Probe()`. If `Probe()` returns an error, cause the plugin to be ignored.

## Plugin Requirements

As already stated, plugins participating in the `probe` scheme must implement 
the `ProbePlugin` interface. The exact way the plugin implements the behavior 
will depend on the plugin in question, but generally it should take the same 
actions as it would with `Gather()` or `Write()`, such that failures during 
`Gather()`/`Write()` would also imply a failure of `Probe()`. If `Probe()` 
returns an error, the plugin's `Close()` method should still be safe to call.

It should be noted that for output plugins, it's advisable to implement 
`Probe()` in a way that doesn't write real metrics to the backend if possible. 
How this might be done depends on the output in question.

Plugins should return `nil` upon successful probes or a retriable error (via 
predefined error type) to enable the defined ignoring behavior. A non-retriable 
error (via predefined error type) or a generic error will bypass the startup 
error behaviors and Telegraf must fail and exit in the startup phase.

## Related Issues

- [#16028](https://github.com/influxdata/telegraf/issues/16028)
- [#15916](https://github.com/influxdata/telegraf/pull/15916)
- [#16001](https://github.com/influxdata/telegraf/pull/16001)

