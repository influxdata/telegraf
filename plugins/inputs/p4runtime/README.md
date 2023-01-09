# P4 Runtime Input Plugin

The `p4runtime` plugin gathers metrics about `Counter` values
present in P4 Program loaded onto Programmable Switch.
Metrics are collected through gRPC connection with
[P4Runtime](https://github.com/p4lang/p4runtime) server.

P4Runtime Plugin uses `PkgInfo.Name` field.
If user want to gather information about program name, please follow
[6.2.1. Annotating P4 code with PkgInfo] in P4 program.

[6.2.1. Annotating P4 code with PkgInfo]: https://p4.org/p4-spec/p4runtime/main/P4Runtime-Spec.html#sec-annotating-p4-code-with-pkginfo

## Configuration

```toml @sample.conf
# P4Runtime telemetry input plugin
[[inputs.p4runtime]]
  ## Define the address of P4Runtime gRPC server to collect metrics.
  # endpoint = "127.0.0.1:9559"
  ## Set DeviceID required for Client Arbitration.
  ## https://p4.org/p4-spec/p4runtime/main/P4Runtime-Spec.html#sec-client-arbitration-and-controller-replication 
  # device_id = "1"
  ## Filter counters by their names that should be observed.
  ## Example: counter_names=["ingressCounter", "egressCounter"]
  # counter_names = []

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

P4Runtime gRPC server communicates using [p4runtime.proto](
https://github.com/p4lang/p4runtime/blob/main/proto/p4/v1/p4runtime.proto)
Protocol Buffer.
Static information about P4 program loaded into programmable switch
are collected by `GetForwardingPipelineConfigRequest` message.
Plugin gathers dynamic metrics with `Read` method.
`Readrequest` is defined with single `Entity` of type `CounterEntry`.
Since P4 Counter is array, plugin collects values of every cell of array
by [wildcard query].

[wildcard query]: https://github.com/p4lang/p4runtime/blob/main/proto/p4/v1/p4runtime.proto#L379

Counters defined in P4 Program have unique ID and name.
Counters are arrays, thus `counter_index` informs
which cell value of array is described in metric.

Tags are constructed in given manner:

`p4program_name`: P4 program name provided by user.
Instruction [6.2.1. Annotating P4 code with PkgInfo]
`counter_name`: Name of given counter in P4 program.
`counter_index`: Index at which metrics are collected in P4 counter.
`counter_type`: Type of counter (BYTES, PACKETS, BOTH).

Fields are constructed in given manner:
`bytes`: Number of bytes gathered in counter.
`packets` Number of packets gathered in counter.

## Example Output

Expected output for p4runtime plugin instance
running on host named `p4runtime-host`:

```shell
p4_runtime,counter_index=10,counter_name=MyIngress.egressTunnelCounter,counter_type=BOTH,host=p4runtime-host bytes=20i,packets=0i 1664973560000000000
```
