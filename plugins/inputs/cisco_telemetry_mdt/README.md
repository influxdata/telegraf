# Cisco Model-Driven Telemetry (MDT) Input Plugin

Cisco model-driven telemetry (MDT) is an input plugin that consumes telemetry
data from Cisco IOS XR, IOS XE and NX-OS platforms. It supports TCP & GRPC
dialout transports.  RPC-based transport can utilize TLS for authentication and
encryption.  Telemetry data is expected to be GPB-KV (self-describing-gpb)
encoded.

The GRPC dialout transport is supported on various IOS XR (64-bit) 6.1.x and
later, IOS XE 16.10 and later, as well as NX-OS 7.x and later platforms.

The TCP dialout transport is supported on IOS XR (32-bit and 64-bit) 6.1.x and
later.

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Cisco model-driven telemetry (MDT) input plugin for IOS XR, IOS XE and NX-OS platforms
[[inputs.cisco_telemetry_mdt]]
 ## Telemetry transport can be "tcp" or "grpc".  TLS is only supported when
 ## using the grpc transport.
 transport = "grpc"

 ## Address and port to host telemetry listener
 service_address = ":57000"

 ## Grpc Maximum Message Size, default is 4MB, increase the size.
 max_msg_size = 4000000

 ## Enable TLS; grpc transport only.
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"

 ## Enable TLS client authentication and define allowed CA certificates; grpc
 ##  transport only.
 # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

 ## Define (for certain nested telemetry measurements with embedded tags) which fields are tags
 # embedded_tags = ["Cisco-IOS-XR-qos-ma-oper:qos/interface-table/interface/input/service-policy-names/service-policy-instance/statistics/class-stats/class-name"]

  ## Include the delete field in every telemetry message.
  # include_delete_field = false

 ## Define aliases to map telemetry encoding paths to simple measurement names
 [inputs.cisco_telemetry_mdt.aliases]
   ifstats = "ietf-interfaces:interfaces-state/interface/statistics"
 ## Define Property Xformation, please refer README and https://pubhub.devnetcloud.com/media/dme-docs-9-3-3/docs/appendix/ for Model details.
 [inputs.cisco_telemetry_mdt.dmes]
#    Global Property Xformation.
#    prop1 = "uint64 to int"
#    prop2 = "uint64 to string"
#    prop3 = "string to uint64"
#    prop4 = "string to int64"
#    prop5 = "string to float64"
#    auto-prop-xfrom = "auto-float-xfrom" #Xform any property which is string, and has float number to type float64
#    Per Path property xformation, Name is telemetry configuration under sensor-group, path configuration "WORD         Distinguished Name"
#    Per Path configuration is better as it avoid property collision issue of types.
#    dnpath = '{"Name": "show ip route summary","prop": [{"Key": "routes","Value": "string"}, {"Key": "best-paths","Value": "string"}]}'
#    dnpath2 = '{"Name": "show processes cpu","prop": [{"Key": "kernel_percent","Value": "float"}, {"Key": "idle_percent","Value": "float"}, {"Key": "process","Value": "string"}, {"Key": "user_percent","Value": "float"}, {"Key": "onesec","Value": "float"}]}'
#    dnpath3 = '{"Name": "show processes memory physical","prop": [{"Key": "processname","Value": "string"}]}'

 ## Additional GRPC connection settings.
 [inputs.cisco_telemetry_mdt.grpc_enforcement_policy]
  ## GRPC permit keepalives without calls, set to true if your clients are
  ## sending pings without calls in-flight. This can sometimes happen on IOS-XE
  ## devices where the GRPC connection is left open but subscriptions have been
  ## removed, and adding subsequent subscriptions does not keep a stable session.
  # permit_keepalive_without_calls = false

  ## GRPC minimum timeout between successive pings, decreasing this value may
  ## help if this plugin is closing connections with ENHANCE_YOUR_CALM (too_many_pings).
  # keepalive_minimum_time = "5m"
```

## Metrics

Metrics are named by the encoding path that generated the data, or by the alias
if the `inputs.cisco_telemetry_mdt.aliases` config section is defined.
Metric fields are dependent on the device type and path.

Tags included in all metrics:

- source
- path
- subscription

Additional tags (such as interface_name) may be included depending on the path.

## Example Output

```text
ifstats,path=ietf-interfaces:interfaces-state/interface/statistics,host=linux,name=GigabitEthernet2,source=csr1kv,subscription=101 in-unicast-pkts=27i,in-multicast-pkts=0i,discontinuity-time="2019-05-23T07:40:23.000362+00:00",in-octets=5233i,in-errors=0i,out-multicast-pkts=0i,out-discards=0i,in-broadcast-pkts=0i,in-discards=0i,in-unknown-protos=0i,out-unicast-pkts=0i,out-broadcast-pkts=0i,out-octets=0i,out-errors=0i 1559150462624000000
ifstats,path=ietf-interfaces:interfaces-state/interface/statistics,host=linux,name=GigabitEthernet1,source=csr1kv,subscription=101 in-octets=3394770806i,in-broadcast-pkts=0i,in-multicast-pkts=0i,out-broadcast-pkts=0i,in-unknown-protos=0i,out-octets=350212i,in-unicast-pkts=9477273i,in-discards=0i,out-unicast-pkts=2726i,out-discards=0i,discontinuity-time="2019-05-23T07:40:23.000363+00:00",in-errors=30i,out-multicast-pkts=0i,out-errors=0i 1559150462624000000
```

### NX-OS Configuration Example

```text
Requirement      DATA-SOURCE   Configuration
-----------------------------------------
Environment      DME           path sys/ch query-condition query-target=subtree&target-subtree-class=eqptPsuSlot,eqptFtSlot,eqptSupCSlot,eqptPsu,eqptFt,eqptSensor,eqptLCSlot
                 DME           path sys/ch depth 5  (Another configuration option)
Environment      NXAPI         show environment power
                 NXAPI         show environment fan
                 NXAPI         show environment temperature
Interface Stats  DME           path sys/intf query-condition query-target=subtree&target-subtree-class=rmonIfIn,rmonIfOut,rmonIfHCIn,rmonIfHCOut,rmonEtherStats
Interface State  DME           path sys/intf depth unbounded query-condition query-target=subtree&target-subtree-class=l1PhysIf,pcAggrIf,l3EncRtdIf,l3LbRtdIf,ethpmPhysIf
VPC              DME           path sys/vpc query-condition query-target=subtree&target-subtree-class=vpcDom,vpcIf
Resources cpu    DME           path sys/procsys query-condition query-target=subtree&target-subtree-class=procSystem,procSysCore,procSysCpuSummary,procSysCpu,procIdle,procIrq,procKernel,procNice,procSoftirq,procTotal,procUser,procWait,procSysCpuHistory,procSysLoad
Resources Mem    DME           path sys/procsys/sysmem/sysmemused
                               path sys/procsys/sysmem/sysmemusage
                               path sys/procsys/sysmem/sysmemfree
Per Process cpu  DME           path sys/proc depth unbounded query-condition rsp-foreign-subtree=ephemeral
vxlan(svi stats) DME           path sys/bd query-condition query-target=subtree&target-subtree-class=l2VlanStats
BGP              DME           path sys/bgp query-condition query-target=subtree&target-subtree-class=bgpDom,bgpPeer,bgpPeerAf,bgpDomAf,bgpPeerAfEntry,bgpOperRtctrlL3,bgpOperRttP,bgpOperRttEntry,bgpOperAfCtrl
mac dynamic      DME           path sys/mac query-condition query-target=subtree&target-subtree-class=l2MacAddressTable
bfd              DME           path sys/bfd/inst depth unbounded
lldp             DME           path sys/lldp depth unbounded
urib             DME           path sys/urib depth unbounded query-condition rsp-foreign-subtree=ephemeral
u6rib            DME           path sys/u6rib depth unbounded query-condition rsp-foreign-subtree=ephemeral
multicast flow   DME           path sys/mca/show/flows depth unbounded
multicast stats  DME           path sys/mca/show/stats depth unbounded
multicast igmp   NXAPI         show ip igmp groups vrf all
multicast igmp   NXAPI         show ip igmp interface vrf all
multicast igmp   NXAPI         show ip igmp snooping
multicast igmp   NXAPI         show ip igmp snooping groups
multicast igmp   NXAPI         show ip igmp snooping groups detail
multicast igmp   NXAPI         show ip igmp snooping groups summary
multicast igmp   NXAPI         show ip igmp snooping mrouter
multicast igmp   NXAPI         show ip igmp snooping statistics
multicast pim    NXAPI         show ip pim interface vrf all
multicast pim    NXAPI         show ip pim neighbor vrf all
multicast pim    NXAPI         show ip pim route vrf all
multicast pim    NXAPI         show ip pim rp vrf all
multicast pim    NXAPI         show ip pim statistics vrf all
multicast pim    NXAPI         show ip pim vrf all
```
