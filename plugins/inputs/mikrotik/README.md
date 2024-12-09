# Mikrotik Input Plugin

This plugin gathers metrics from [Mikrotik's RouterOS][mikrotik] such as
interface statistics, uptime etc

[mikrotik]: https://mikrotik.com/software

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
[[inputs.mikrotik]]
  ## Mikrotik's address to query. Make sure that REST API is enabled: https://help.mikrotik.com/docs/spaces/ROS/pages/47579162/REST+API
  address = "https://192.168.88.1"

  ## User to use. Read access rights will be enough
  username = "admin"
  password = "password"

  ## Mikrotik's entities whose comments contain this strings will be ignored
  # ignore_comments = [
  #     "block",
  #     "doNotGatherMetricsFromThis"
  # ]

  ## Modules available to use (default: system_resourses)
  # include_modules = [
  #     "interface",
  #     "interface_wireguard_peers",
  #     "interface_wireless_registration",
  #     "ip_dhcp_server_lease",
  #     "ip_firewall_connection",
  #     "ip_firewall_filter",
  #     "ip_firewall_nat",
  #     "ip_firewall_mangle",
  #     "ipv6_firewall_connection",
  #     "ipv6_firewall_filter",
  #     "ipv6_firewall_nat",
  #     "ipv6_firewall_mangle",
  #     "system_script",
  #     "system_resourses"
  # ]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP response timeout
  # response_timeout = "5s"
```

## Metrics

For each specific module, a unique set of metrics and tags will be provided
based on the JSON structure returned by the REST endpoint. You can refer to
the `tagFields` and `valueFields` lists in `types.go` for a full set of
tags and values.

When querying Mikrotik, all available fields across different metrics
will be requested. However, Mikrotik’s design only returns fields that are
present in the current module’s response, ignoring any fields that don’t
apply to the specific endpoint. Disabled entities in Mikrotik are
automatically excluded from the response.

## Example Output

```text
mikrotik,.id=*1,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,default-name=ether1,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:B6,model=RBD52G-5HacD2HnD,name=ether1,platform=MikroTik,running=true,serial-number=SERIALNUMBER,source-module=interface,type=ether,version=7.16\ (stable) fp-rx-byte=23815497595i,fp-rx-packet=18015083i,fp-tx-byte=0i,fp-tx-packet=0i,link-downs=1i,rx-byte=23887557927i,rx-drop=0i,rx-error=0i,rx-packet=18015083i,tx-byte=1129765037i,tx-drop=0i,tx-error=0i,tx-packet=5384706i,tx-queue-drop=0i 1730320979000000000
mikrotik,.id=*2,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,default-name=ether2,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:B7,model=RBD52G-5HacD2HnD,name=ether2,platform=MikroTik,running=false,serial-number=SERIALNUMBER,slave=true,source-module=interface,type=ether,version=7.16\ (stable) fp-rx-byte=0i,fp-rx-packet=0i,fp-tx-byte=0i,fp-tx-packet=0i,link-downs=0i,rx-byte=0i,rx-drop=0i,rx-error=0i,rx-packet=0i,tx-byte=0i,tx-drop=0i,tx-error=0i,tx-packet=0i,tx-queue-drop=0i 1730320979000000000
mikrotik,.id=*3,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,default-name=ether3,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:B8,model=RBD52G-5HacD2HnD,name=ether3,platform=MikroTik,running=true,serial-number=SERIALNUMBER,slave=true,source-module=interface,type=ether,version=7.16\ (stable) fp-rx-byte=91438550i,fp-rx-packet=1244518i,fp-tx-byte=4403947442i,fp-tx-packet=2914053i,link-downs=1i,rx-byte=96416622i,rx-drop=0i,rx-error=0i,rx-packet=1244518i,tx-byte=4422341564i,tx-drop=0i,tx-error=0i,tx-packet=2929010i,tx-queue-drop=0i 1730320979000000000
mikrotik,.id=*4,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,default-name=ether4,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:B9,model=RBD52G-5HacD2HnD,name=ether4,platform=MikroTik,running=false,serial-number=SERIALNUMBER,slave=true,source-module=interface,type=ether,version=7.16\ (stable) fp-rx-byte=0i,fp-rx-packet=0i,fp-tx-byte=0i,fp-tx-packet=0i,link-downs=0i,rx-byte=0i,rx-drop=0i,rx-error=0i,rx-packet=0i,tx-byte=0i,tx-drop=0i,tx-error=0i,tx-packet=0i,tx-queue-drop=0i 1730320979000000000
mikrotik,.id=*5,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,default-name=ether5,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:BA,model=RBD52G-5HacD2HnD,name=ether5,platform=MikroTik,running=false,serial-number=SERIALNUMBER,slave=true,source-module=interface,type=ether,version=7.16\ (stable) fp-rx-byte=0i,fp-rx-packet=0i,fp-tx-byte=0i,fp-tx-packet=0i,link-downs=0i,rx-byte=0i,rx-drop=0i,rx-error=0i,rx-packet=0i,tx-byte=0i,tx-drop=0i,tx-error=0i,tx-packet=0i,tx-queue-drop=0i 1730320979000000000
mikrotik,.id=*6,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,default-name=wlan1,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:BB,model=RBD52G-5HacD2HnD,name=wlan1,platform=MikroTik,running=false,serial-number=SERIALNUMBER,slave=true,source-module=interface,type=wlan,version=7.16\ (stable) fp-rx-byte=1984519i,fp-rx-packet=8740i,fp-tx-byte=0i,fp-tx-packet=0i,link-downs=12i,rx-byte=1984519i,rx-drop=0i,rx-error=0i,rx-packet=8740i,tx-byte=17087451i,tx-drop=0i,tx-error=0i,tx-packet=47921i,tx-queue-drop=1i 1730320979000000000
mikrotik,.id=*7,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,default-name=wlan2,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:BC,model=RBD52G-5HacD2HnD,name=wlan2,platform=MikroTik,running=true,serial-number=SERIALNUMBER,slave=true,source-module=interface,type=wlan,version=7.16\ (stable) fp-rx-byte=5525090211i,fp-rx-packet=8360162i,fp-tx-byte=87942233i,fp-tx-packet=1222212i,link-downs=0i,rx-byte=5525090211i,rx-drop=0i,rx-error=0i,rx-packet=8360162i,tx-byte=23832544176i,tx-drop=0i,tx-error=0i,tx-packet=19198103i,tx-queue-drop=52532i 1730320979000000000
mikrotik,.id=*8,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=11:22:33:44:55:BB,model=RBD52G-5HacD2HnD,name=lan,platform=MikroTik,running=true,serial-number=SERIALNUMBER,source-module=interface,type=bridge,version=7.16\ (stable) fp-rx-byte=1107864899i,fp-rx-packet=5393891i,fp-tx-byte=0i,fp-tx-packet=0i,link-downs=0i,rx-byte=1124747646i,rx-drop=0i,rx-error=0i,rx-packet=5463554i,tx-byte=23815054209i,tx-drop=0i,tx-error=0i,tx-packet=18013585i,tx-queue-drop=0i 1730320979000000000
mikrotik,.id=*14,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,disabled=false,firmware-type=ipq4000L,host=localhost,mac-address=00:00:00:00:00:00,model=RBD52G-5HacD2HnD,name=lo,platform=MikroTik,running=true,serial-number=SERIALNUMBER,source-module=interface,type=loopback,version=7.16\ (stable) fp-rx-byte=0i,fp-rx-packet=0i,fp-tx-byte=0i,fp-tx-packet=0i,link-downs=0i,rx-byte=491147i,rx-drop=0i,rx-error=0i,rx-packet=2839i,tx-byte=491147i,tx-drop=0i,tx-error=0i,tx-packet=2839i,tx-queue-drop=0i 1730320979000000000
mikrotik,architecture-name=arm,board-name=hAP\ ac^2,cpu=ARM,current-firmware=7.15.3,firmware-type=ipq4000L,host=localhost,model=RBD52G-5HacD2HnD,platform=MikroTik,serial-number=SERIALNUMBER,source-module=system_resourses,version=7.16\ (stable) cpu-frequency=672i,cpu-load=0i,free-hdd-space=1482752i,free-memory=55685120i,total-memory=134217728i,uptime=85201i,write-sect-since-reboot=2344i,write-sect-total=30313i 1730320979000000000

```
