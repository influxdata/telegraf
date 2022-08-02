# Cisco Model-Driven Telemetry (MDT) NETCONF Input Plugin

Cisco NETCONF telemetry is an input plugin that can consume telemetry data
from a [stream of NETCONF notifications](
    https://tools.ietf.org/html/rfc5277)
or from polling. The plugin offers three ways to receive data:

1. Subscription-based, periodic model-driven telemetry streaming through
   `ietf-yang-push` notifications (see [draft-ietf-netconf-yang-push-25](
       https://tools.ietf.org/html/draft-ietf-netconf-yang-push-25))
2. Subscription-based, event notification streaming through `notification`s (
   see [RFC5277](
       https://tools.ietf.org/html/rfc5277#page-9))
3. Periodic polling with `<get>` requests (see
   [rfc6241 - Network Configuration Protocol (NETCONF), section 7.7](
       https://tools.ietf.org/html/rfc6241))

The plugin requires the [netgonf library](github.com/cisco-ie/netgonf).

## Configuration

```toml @sample.conf
[[inputs.cisco_telemetry_mdt_netconf]]

  ## NETCONF over SSH connection
  ## Address and port
  server_address = "10.10.10.10:830"

  ## Credentials
  username = "cisco"
  password = "cisco"

  ## Enable check for authenticity of the NETCONF server
  ## Unknown servers are ignored by default. Set ignore_server_authenticity
  ## to true to disable the check for authenticity of a server's public key.
  ## Optional, default value: false.
  # ignore_server_authenticity = false

  ## Public key of the NETCONF server
  ## Mandatory if ignore_server_authenticity is set to false.
  ## The public key should follow the format of the known_hosts file,
  ## as documented in sshd(8) manual page.
  server_public_key = "[10.10.10.10]:830 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDXxWHGjcEcyEDw/YbJeB824husNnchKKbRtR5i9s+Y712kckQpkWScgwRJJsvneUg4Ztu4ZS8PPzlfiaoHAzOiKjuE7Ns+zklaPSwTj6hf6Sl0FuChWMXi/EchfPcUREQ9mlKL10oMD37W+m3vRUtmj/LM1gNHUSjp3Q1RsyfhLfxYw7I2RQXDfindwxxrX32iWWJdPMfY7PDRYpvh/xmyQVb9RdOhZ7qA/xkDc+SS1hZrzCkh2kaKTd4Glh76K58fEuQ2NFCRYztezWa7D61OiXIeWZJ4x2Utb8xH6wsGA5T0vBt89DB7EvF8xsnEdDtlMsI8L99JtGlNO3MXasdf"

  ## Redial interval if the client fails to connect to the server
  ## Optional, default value: 10s.
  # redial = "10s"

  ## Telemetry streaming
  ## IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/if:interfaces-state/interface"
    update_trigger = "periodic"
    period = "10s"

    ## Leaves to be marked as tags in Influx LINE format.
    ## They are valid throughout all other defined operations.
    tags = ["/if:interfaces-state/interface/name"]

  ## IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/mdt-oper:mdt-oper-data/mdt-subscriptions"
    update_trigger = "periodic"
    period = "10s"
    tags = ["/mdt-oper:mdt-oper-data/mdt-subscriptions/subscription-id"]

  ## IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/memory-ios-xe-oper:memory-statistics/memory-ios-xe-oper:memory-statistic"
    update_trigger = "periodic"
    period = "10s"
    tags = ["/memory-ios-xe-oper:memory-statistics/memory-ios-xe-oper:memory-statistic/memory-ios-xe-oper:name"]

  ## IOS-XE Subscription - Xpath union for multiple subtrees
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/interfaces-ios-xe-oper:interfaces/interface/statistics/in-octets|/interfaces-ios-xe-oper:interfaces/interface/statistics/out-octets"
    update_trigger = "periodic"
    period = "5s"
    tags = ["/interfaces-ios-xe-oper:interfaces/interface/name"]

  ## IOS-XE Subscription - on-change
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/cdp-ios-xe-oper:cdp-neighbor-details/cdp-neighbor-detail"
    update_trigger = "on-change"
    period = "0s"

  ## Get operations
  ## IOS-XE Get Request with filter
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/interfaces-state/interface[name='GigabitEthernet1']/oper-status"
    period = "10s"
    tags = ["/interfaces-state/interface/name"]

  ## IOS-XE Get Request with filter and multiple tags
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/interfaces-state/interface[name='GigabitEthernet2']"
    period = "10s"
    tags = ["/interfaces-state/interface/name", "/interfaces-state/interface/if-index"]

  ## IOS-XE Get Request without filter
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/memory-statistics/memory-statistic"
    period = "10s"
    tags = ["/memory-statistics/memory-statistic/name"]

  ## Event notification subscription
  ## NSO Event notfication subscription with a tag
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.notification]]
    stream = "ncs-alarms"
    tags = ["alarm-notification/alarm-class"]
```

## Example Output

```text
Cisco-IOS-XE-process-cpu-oper:cpu-usage/cpu-utilization,host=linux,path=Cisco-IOS-XE-process-cpu-oper:cpu-usage/cpu-utilization,source=3650,subscription=1 five_seconds=3i 1654898172903000000
Cisco-IOS-XE-memory-oper:memory-statistics/memory-statistic,Cisco-IOS-XE-memory-oper:memory-statistics/memory-statistic/name=reserve\ Processor,host=linux,source=3650 total-memory=102404i,used-memory=92i,free-memory=102312i,lowest-usage=102312i,highest-usage=102312i 1654898176722493661
Cisco-IOS-XE-interfaces-oper:interfaces/interface/statistics,host=linux,name=GigabitEthernet0/0,path=Cisco-IOS-XE-interfaces-oper:interfaces/interface/statistics,source=3650,subscription=2 discontinuity_time="2022-06-07T22:06:03.000763+00:00",in_octets=55043194i,in_unicast_pkts=580774i,in_broadcast_pkts=0i,in_multicast_pkts=0i,in_discards=0i,in_errors=0i,in_unknown_protos=0i,out_octets=9692848i,out_unicast_pkts=107693i,out_broadcast_pkts=0i,out_multicast_pkts=0i,out_discards=0i,out_errors=0i,rx_pps=5i,rx_kbps=18i,tx_pps=1i,tx_kbps=2i,num_flaps=0i,in_crc_errors=0i 1654898187900000000
```

## Metrics

The metrics collected by this input will depend on the path defined in subscriptions and get requests. Tags included in all metrics:

- source
- path
- subscription

Sample metrics include:

- Cisco-IOS-XE-process-cpu-oper:cpu-usage/cpu-utilization
  - tags:
    - host
    - path
    - source
    - subscription
  - fields:
    - five_seconds (int, percent)

- Cisco-IOS-XE-memory-oper:memory-statistics/memory-statistic
  - tags:
    - host
    - path
    - source
    - subscription
    - Cisco-IOS-XE-memory-oper:memory-statistics/memory-statistic/name (enforced tag)
  - fields:
    - total-memory (int, byte)
    - used-memory (int, byte)
    - free-memory (int, byte)
    - lowest-usage (int, byte)
    - highest-usage (int, byte)

- Cisco-IOS-XE-interfaces-oper:interfaces/interface/statistics
  - tags:
    - host
    - path
    - source
    - subscription
    - name (implicit tag: interface name)
  - fields:
    - discontinuity_time (string, date and time)
    - in_octets (int, byte)
    - in_unicast_pkts (int, packet)
    - in_broadcast_pkts (int, packet)
    - in_multicast_pkts (int, packet)
    - in_discards (int, packets)
    - in_errors (int, packets)
    - in_unknown_protos (int, packet)
    - out_octets (int, byte)
    - out_unicast_pkts (int, packet)
    - out_broadcast_pkts (int, packet)
    - out_multicast_pkts (int, packet)
    - out_discards (int, packet)
    - out_errors (int, packet)
    - rx_pps (int, pps)
    - rx_kbps (int, kbps)
    - tx_pps (int, pps)
    - tx_kbps (int, kbps)
    - num_flaps (int)
    - in_crc_errors (int)
