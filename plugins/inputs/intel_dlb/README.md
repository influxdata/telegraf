# Intel® Dynamic Load Balancer (Intel® DLB)  Input Plugin

The `Intel DLB` plugin collects metrics exposed by applications built with
[Data Plane Development Kit](https://www.dpdk.org/) which is an extensive
set of open source libraries designed for accelerating packet processing
workloads, plugin is also using bifurcated driver. More specifically it's
targeted for applications that use Intel DLB as eventdev devices accessed
via bifurcated driver (allowing access from kernel and user-space).

## Metrics

There are two sources of metrics:

- DPDK-based app for detailed eventdev metrics per device, per port and per queue
- Sysfs entries from kernel driver for RAS metrics

## About Intel® Dynamic Load Balancer (Intel® DLB)

The Intel® Dynamic Load Balancer (Intel® DLB) is a PCIe device that provides
load-balanced, prioritized scheduling of events (that is, packets) across
CPU cores enabling efficient core-to-core communication. It is a hardware
accelerator located inside the latest Intel® Xeon® devices offered by Intel.
It supports the event-driven programming model of DPDK's Event Device Library.
This library is used in packet processing pipelines for multi-core scalability,
dynamic load-balancing, and variety of packet distribution and synchronization
schemes.

## About DPDK Event Device Library

The DPDK Event device library is an abstraction that provides the application
with features to schedule events. The eventdev framework introduces the event
driven programming model. In a polling model, lcores poll ethdev ports and
associated Rx queues directly to look for a packet. By contrast in an event
driven model, lcores call the scheduler that selects packets for them based on
programmer-specified criteria. The Eventdev library adds support for an event
driven programming model, which offers applications automatic multicore scaling,
dynamic load balancing, pipelining, packet ingress order maintenance and
synchronization services to simplify application packet processing.
By introducing an event driven programming model, DPDK can support
both polling and event driven programming models for packet processing,
and applications are free to choose whatever model (or combination of the two)
best suits their needs.

## Prerequisites

- [DLB >= v7.4](https://www.intel.com/content/www/us/en/download/686372/intel-dynamic-load-balancer.html)
- [DPDK >= 20.11.3](http://core.dpdk.org/download/)
- Linux kernel >= 5.12

> **NOTE:** It may happen that sysfs entries or the socket telemetry interface
> exposed by DPDK-based app will require root access. This means that either
> access permissions have to be adjusted for sysfs / socket telemetry
> interface to allow Telegraf to access it, or Telegraf should run with root
> privileges.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
## Reads metrics from DPDK using v2 telemetry interface.
[[inputs.intel_dlb]]
  ## Path to DPDK telemetry socket.
  # socket_path = "/var/run/dpdk/rte/dpdk_telemetry.v2"

  ## Default eventdev command list, it gathers metrics from socket by given commands.
  ## Supported options:
  ##   "/eventdev/dev_xstats", "/eventdev/port_xstats",
  ##   "/eventdev/queue_xstats", "/eventdev/queue_links"
  # eventdev_commands = ["/eventdev/dev_xstats", "/eventdev/port_xstats", "/eventdev/queue_xstats", "/eventdev/queue_links"]

  ## Detect DLB devices based on device id.
  ## Currently, only supported and tested device id is `0x2710`.
  ## Configuration added to support forward compatibility.
  # dlb_device_types = ["0x2710"]

  ## Specifies plugin behavior regarding unreachable socket (which might not have been initialized yet).
  ## Available choices:
  ##   - error: Telegraf will return an error on startup if socket is unreachable
  ##   - ignore: Telegraf will ignore error regarding unreachable socket on both startup and gather
  # unreachable_socket_behavior = "error"
```

Default configuration allows getting metrics for all metrics
reported via `/eventdev/` command:

- `/eventdev/dev_xstats`
- `/eventdev/port_xstats`
- `/eventdev/queue_xstats`
- `/eventdev/queue_links`

## Example Output

```text
intel_dlb,command=/eventdev/dev_xstats\,0,host=controller1 dev_dir_pool_size=0i,dev_inflight_events=8192i,dev_ldb_pool_size=8192i,dev_nb_events_limit=8192i,dev_pool_size=0i,dev_rx_drop=0i,dev_rx_interrupt_wait=0i,dev_rx_ok=463126660i,dev_rx_umonitor_umwait=0i,dev_total_polls=78422946i,dev_tx_nospc_dir_hw_credits=0i,dev_tx_nospc_hw_credits=584614i,dev_tx_nospc_inflight_credits=0i,dev_tx_nospc_inflight_max=0i,dev_tx_nospc_ldb_hw_credits=584614i,dev_tx_nospc_new_event_limit=59331982i,dev_tx_ok=694694059i,dev_zero_polls=29667908i 1641996791000000000
intel_dlb,command=/eventdev/queue_links\,0\,1,host=controller1 qid_0=128i,qid_1=128i 1641996791000000000
intel_dlb_ras,device=pci0000:6d,host=controller1,metric_file=aer_dev_correctable BadDLLP=0i,BadTLP=0i,CorrIntErr=0i,HeaderOF=0i,NonFatalErr=0i,Rollover=0i,RxErr=0i,TOTAL_ERR_COR=0i,Timeout=0i 1641996791000000000
intel_dlb_ras,device=pci0000:6d,host=controller1,metric_file=aer_dev_fatal ACSViol=0i,AtomicOpBlocked=0i,BlockedTLP=0i,CmpltAbrt=0i,CmpltTO=0i,DLP=0i,ECRC=0i,FCP=0i,MalfTLP=0i,PoisonTLPBlocked=0i,RxOF=0i,SDES=0i,TLP=0i,TLPBlockedErr=0i,TOTAL_ERR_FATAL=0i,UncorrIntErr=0i,Undefined=0i,UnsupReq=0i,UnxCmplt=0i 1641996791000000000
```
