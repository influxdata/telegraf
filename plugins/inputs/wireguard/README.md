# Wireguard Input Plugin

This plugin provides information about Wireguard interfaces

### Configuration

Make sure that running `wg show` on the host works correctly before enabling this plugin

```toml

[[inputs.wireguard]]
  # if none are provided, all of the available ones will be scraped
  interfaces = ["wg0"]

```

### Metrics

- wireguard
  - tags:
    - name
    - type
    - serverpublikey
    - listenport
    - firewallmark
    - peerpublickey
    - endpoint
  - fiels:
    - received_bytes
    - transmit_bytes
    - protocol
    - last_hanshake_time
    - persisten_keepalive_interval

### Example Output:

```
% telegraf --config=plugins/inputs/wireguard/dev/telegraf.conf --test
> wireguard,endpoint=<nil>,firewallmark=0,host=france-vpn,listenport=51820,name=wg0,peerpublickey=C9uvgpq+kgUqRXIpLqLLHGon3VNus5F05p/UkqnsyGI=,serverpublikey=M8NxRDSd/UqMFweBbUsiFDrx9jiiX5nE+S53n8Ag2Rk=,type=Linux\ kernel last_hanshake_time=-62135596800i,persisten_keepalive_interval=0i,protocol=1i,received_bytes=0i,transmit_bytes=0i 1572285866000000000
> wireguard,endpoint=23.23.23.23:22211,firewallmark=0,host=france-vpn,listenport=51820,name=wg0,peerpublickey=epmacZ5LRUYhnIOWuombi1S/m69X92ixTi80+OSGviE=,serverpublikey=M8NxRDSd/UqMFweBbUsiFDrx9jiiX5nE+S53n8Ag2Rk=,type=Linux\ kernel last_hanshake_time=1572285849i,persisten_keepalive_interval=0i,protocol=1i,received_bytes=255536884i,transmit_bytes=1709755308i 1572285866000000000
> wireguard,endpoint=<nil>,firewallmark=0,host=france-vpn,listenport=51820,name=wg0,peerpublickey=z64T8lbZz9JNT/pKVv3cB+kx+j8fEpp/djQYZwc75mc=,serverpublikey=M8NxRDSd/UqMFweBbUsiFDrx9jiiX5nE+S53n8Ag2Rk=,type=Linux\ kernel last_hanshake_time=-62135596800i,persisten_keepalive_interval=0i,protocol=1i,received_bytes=0i,transmit_bytes=0i 1572285866000000000
```
