# Wireguard Input Plugin

The Wireguard input plugin collects statistics on the local Wireguard server
using the [`wgctrl`](https://github.com/WireGuard/wgctrl-go) library. It
reports gauge metrics for Wireguard interface device(s) and its peers.

### Configuration

```toml
# Collect Wireguard server interface and peer statistics
[[inputs.wireguard]]
  ## Optional list of Wireguard device/interface names to query.
  ## If omitted, all Wireguard interfaces are queried.
  # devices = ["wg0"]
```

### Metrics

- `wireguard_device`
  - tags:
    - `name` (interface device name, e.g. `wg0`)
    - `type` (Wireguard tunnel type, e.g. `linux_kernel` or `userspace`)
  - fields:
    - `listen_port` (int, UDP port on which the interface is listening)
    - `firewall_mark` (int, device's current firewall mark)
    - `peers` (int, number of peers associated with the device)

- `wireguard_peer`
  - tags:
    - `device` (associated interface device name, e.g. `wg0`)
    - `public_key` (peer public key, e.g. `NZTRIrv/ClTcQoNAnChEot+WL7OH7uEGQmx8oAN9rWE=`)
  - fields:
    - `persistent_keepalive_interval` (int, keepalive interval in seconds; 0 if unset)
    - `protocol_version` (int, Wireguard protocol version number)
    - `allowed_ips` (int, number of allowed IPs for this peer)
    - `last_handshake_time` (int, Unix timestamp of the last handshake for this peer)
    - `rx_bytes` (int, number of bytes received from this peer)
    - `tx_bytes` (int, number of bytes transmitted to this peer)

### Troubleshooting

#### Error: `operation not permitted`

By default, Telegraf runs as the `telegraf` system user. Wireguard
implementations that run in kernelspace (as opposed to userspace) require
userspace programs to run as root in order to communicate with the module.
Either update the system udev rules for the `telegraf` user or run Telegraf as
root (not recommended).

#### Error: `error enumerating Wireguard devices`

This usually happens when the device names specified in config are invalid.
Ensure that `sudo wg show` succeeds, and that the device names in config match
those printed by this command.

### Example Output

```
wireguard_device,host=WGVPN,name=tun0,type=linux_kernel firewall_mark=0i,listen_port=51820i 1567976672000000000
wireguard_device,host=WGVPN,name=tun0,type=linux_kernel peers=1i 1567976672000000000
wireguard_peer,device=wg0,host=WGVPN,public_key=NZTRIrv/ClTcQoNAnChEot+WL7OH7uEGQmx8oAN9rWE= allowed_ips=1i,persistent_keepalive_interval=0i,protocol_version=1i 1567976672000000000
wireguard_peer,device=wg0,host=WGVPN,public_key=NZTRIrv/ClTcQoNAnChEot+WL7OH7uEGQmx8oAN9rWE= last_handshake_time=1567905087i,rx_bytes=261415128i,tx_bytes=334031704i 1567976672000000000
```
