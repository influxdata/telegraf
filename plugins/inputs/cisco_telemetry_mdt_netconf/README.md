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
  # Address and port
  server_address = "10.10.10.10:830"

  # Credentials
  username = "cisco"
  password = "cisco"

  # Enable check for authenticity of the NETCONF server
  # Unknown servers are ignored by default. Set ignore_server_authenticity
  # to true to disable the check for authenticity of a server's public key.
  # Optional, default value: false.
  ignore_server_authenticity = false

  # Public key of the NETCONF server
  # Mandatory if ignore_server_authenticity is set to false.
  # The public key should follow the format of the known_hosts file,
  # as documented in sshd(8) manual page.
  server_public_key = "[10.10.10.10]:830 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDXxWHGjcEcyEDw/YbJeB824husNnchKKbRtR5i9s+Y712kckQpkWScgwRJJsvneUg4Ztu4ZS8PPzlfiaoHAzOiKjuE7Ns+zklaPSwTj6hf6Sl0FuChWMXi/EchfPcUREQ9mlKL10oMD37W+m3vRUtmj/LM1gNHUSjp3Q1RsyfhLfxYw7I2RQXDfindwxxrX32iWWJdPMfY7PDRYpvh/xmyQVb9RdOhZ7qA/xkDc+SS1hZrzCkh2kaKTd4Glh76K58fEuQ2NFCRYztezWa7D61OiXIeWZJ4x2Utb8xH6wsGA5T0vBt89DB7EvF8xsnEdDtlMsI8L99JtGlNO3MXasdf"

  # Redial interval if the client fails to connect to the server
  # Optional, default value: 10s.
  redial = "10s"

  ## Telemetry streaming
  # IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/if:interfaces-state/interface"
    update_trigger = "periodic"
    period = "10s"

    # Leaves to be marked as keys in Influx LINE format.
    # They are valid throughout all other defined operations.
    keys = ["/if:interfaces-state/interface/name"]

  # IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/mdt-oper:mdt-oper-data/mdt-subscriptions"
    update_trigger = "periodic"
    period = "10s"
    keys = ["/mdt-oper:mdt-oper-data/mdt-subscriptions/subscription-id"]

  # IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/memory-ios-xe-oper:memory-statistics/memory-ios-xe-oper:memory-statistic"
    update_trigger = "periodic"
    period = "10s"
    keys = ["/memory-ios-xe-oper:memory-statistics/memory-ios-xe-oper:memory-statistic/memory-ios-xe-oper:name"]

  # IOS-XE Subscription - Xpath union for multiple subtrees
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/interfaces-ios-xe-oper:interfaces/interface/statistics/in-octets|/interfaces-ios-xe-oper:interfaces/interface/statistics/out-octets"
    update_trigger = "periodic"
    period = "5s"
    keys = ["/interfaces-ios-xe-oper:interfaces/interface/name"]

  # IOS-XE Subscription - on-change
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/cdp-ios-xe-oper:cdp-neighbor-details/cdp-neighbor-detail"
    update_trigger = "on-change"
    period = "0s"

  ## Get operations
  # IOS-XE Get Request with filter
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/interfaces-state/interface[name='GigabitEthernet1']/oper-status"
    period = "10s"
    keys = ["/interfaces-state/interface/name"]

  # IOS-XE Get Request with filter and multiple keys
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/interfaces-state/interface[name='GigabitEthernet2']"
    period = "10s"
    keys = ["/interfaces-state/interface/name", "/interfaces-state/interface/if-index"]

  # IOS-XE Get Request without filter
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/memory-statistics/memory-statistic"
    period = "10s"
    keys = ["/memory-statistics/memory-statistic/name"]

  ## Event notification subscription
  # NSO Event notfication subscription with a key
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.notification]]
    stream = "ncs-alarms"
    keys = ["alarm-notification/alarm-class"]
```

## Example Output

## Metrics
