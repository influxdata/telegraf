# P4 Runtime Input Plugin

P4 is a language for programming the data plane of network devices,
such as Programmable Switches or Programmable Network Interface Cards.
The P4Runtime API is a control plane specification to manage
the data plane elements of those devices dynamically by a P4 program.

The `p4runtime` plugin gathers metrics about `Counter` values
present in P4 Program loaded onto networking device.
Metrics are collected through gRPC connection with
[P4Runtime](https://github.com/p4lang/p4runtime) server.

P4Runtime Plugin uses `PkgInfo.Name` field.
If user wants to gather information about program name, please follow
[6.2.1. Annotating P4 code with PkgInfo] instruction and apply changes
to your P4 program.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

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

P4Runtime gRPC server communicates using [p4runtime.proto] Protocol Buffer.
Static information about P4 program loaded into programmable switch
are collected by `GetForwardingPipelineConfigRequest` message.
Plugin gathers dynamic metrics with `Read` method.
`Readrequest` is defined with single `Entity` of type `CounterEntry`.
Since P4 Counter is array, plugin collects values of every cell of array
by [wildcard query].

Counters defined in P4 Program have unique ID and name.
Counters are arrays, thus `counter_index` informs
which cell value of array is described in metric.

Tags are constructed in given manner:

- `p4program_name`: P4 program name provided by user.
If user wants to gather information about program name, please follow
[6.2.1. Annotating P4 code with PkgInfo] instruction and apply changes
to your P4 program.
- `counter_name`: Name of given counter in P4 program.
- `counter_type`: Type of counter (BYTES, PACKETS, BOTH).

Fields are constructed in given manner:

- `bytes`: Number of bytes gathered in counter.
- `packets` Number of packets gathered in counter.
- `counter_index`: Index at which metrics are collected in P4 counter.

## Example Output

Expected output for p4runtime plugin instance
running on host named `p4runtime-host`:

```text
p4_runtime,counter_name=MyIngress.egressTunnelCounter,counter_type=BOTH,host=p4 bytes=408i,packets=4i,counter_index=200i 1675175030000000000
```

[6.2.1. Annotating P4 code with PkgInfo]: https://p4.org/p4-spec/p4runtime/main/P4Runtime-Spec.html#sec-annotating-p4-code-with-pkginfo
[p4runtime.proto]: https://github.com/p4lang/p4runtime/blob/main/proto/p4/v1/p4runtime.proto
[wildcard query]: https://github.com/p4lang/p4runtime/blob/main/proto/p4/v1/p4runtime.proto#L379
