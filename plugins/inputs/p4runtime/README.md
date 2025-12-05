# P4 Runtime Input Plugin

This plugin collects metrics from the data plane of network devices, such as
Programmable Switches or Programmable Network Interface Cards by reading the
`Counter` values of the [P4 program][p4lang] running on the device.
Metrics are collected through a gRPC connection with the [P4 runtime][p4runtime]
server.

> [!TIP]
> If you want to gather information about the program name, please follow the
> instruction in [6.2.1. Annotating P4 code with PkgInfo][p4annotation] to
> modify your P4 program.

⭐ Telegraf v1.26.0
🏷️ network, applications
💻 all

[p4lang]: https://p4.org
[p4runtime]: https://github.com/p4lang/p4runtime
[p4annotation]: https://p4.org/p4-spec/p4runtime/main/P4Runtime-Spec.html#sec-annotating-p4-code-with-pkginfo

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# P4Runtime telemetry input plugin
[[inputs.p4runtime]]
  ## Define the endpoint of P4Runtime gRPC server to collect metrics.
  # endpoint = "127.0.0.1:9559"
  ## Set DeviceID required for Client Arbitration.
  ## https://p4.org/p4-spec/p4runtime/main/P4Runtime-Spec.html#sec-client-arbitration-and-controller-replication
  # device_id = 1
  ## Filter counters by their names that should be observed.
  ## Example: counter_names_include=["ingressCounter", "egressCounter"]
  # counter_names_include = []

  ## Optional TLS Config.
  ## Enable client-side TLS and define CA to authenticate the device.
  # enable_tls = false
  # tls_ca = "/etc/telegraf/ca.crt"
  ## Set minimal TLS version to accept by the client.
  # tls_min_version = "TLS12"
  ## Use TLS but skip chain & host verification.
  # insecure_skip_verify = true

  ## Define client-side TLS certificate & key to authenticate to the device.
  # tls_cert = "/etc/telegraf/client.crt"
  # tls_key = "/etc/telegraf/client.key"
```

## Metrics

The P4 runtime server communicates using the
[P4 Protocol Buffer definition][p4proto].  Static information about the program
loaded into programmable switch are collected by
`GetForwardingPipelineConfigRequest` message. The plugin gathers dynamic metrics
using the `Read` method. `Readrequest` is defined with single `Entity` of type
`CounterEntry`. Since a P4 Counter is an array type, the plugin collects values
of every cell of array by [wildcard queries][wildcards].

Counters defined in a P4 Program have a unique ID and name. The `counter_index`
defines which cell value of the counter array is used in the metric.

Tags are constructed in given manner:

- `p4program_name`: P4 program name provided by user.
- `counter_name`: Name of given counter in P4 program.
- `counter_type`: Type of counter (BYTES, PACKETS, BOTH).

Fields are constructed in given manner:

- `bytes`: Number of bytes gathered in counter.
- `packets` Number of packets gathered in counter.
- `counter_index`: Index at which metrics are collected in P4 counter.

## Example Output

```text
p4_runtime,counter_name=MyIngress.egressTunnelCounter,counter_type=BOTH,host=p4 bytes=408i,packets=4i,counter_index=200i 1675175030000000000
```

[p4proto]: https://github.com/p4lang/p4runtime/blob/main/proto/p4/v1/p4runtime.proto
[wildcards]: https://github.com/p4lang/p4runtime/blob/main/proto/p4/v1/p4runtime.proto#L379
