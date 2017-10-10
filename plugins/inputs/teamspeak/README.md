# Teamspeak 3 Input Plugin

This plugin uses the Teamspeak 3 ServerQuery interface of the Teamspeak server to collect statistics of one or more
virtual servers. If you are querying an external Teamspeak server, make sure to add the host which is running Telegraf
to query_ip_whitelist.txt in the Teamspeak Server directory.

### Configuration:

```
[[inputs.teamspeak]]
## Server address for Teamspeak 3 ServerQuery
  server = "127.0.0.1:10011"
## Username for ServerQuery
  username = "serverqueryuser"
## Password for ServerQuery
  password = "secret"
## Array of virtual servers
  vservers = [1]
```

### Measurements:

- teamspeak
    - uptime
    - clients_online
    - total_ping
    - total_packet_loss
    - packets_sent_total
    - packets_received_total
    - bytes_sent_total
    - bytes_received_total

### Tags:

- The following tags are used:
    - v_server
    - name

### Example output:

```
teamspeak,v_server=1,name=LeopoldsServer,host=vm01 bytes_received_total=29638202639i,uptime=13567846i,total_ping=26.89,total_packet_loss=0,packets_sent_total=415821252i,packets_received_total=237069900i,bytes_sent_total=55309568252i,clients_online=11i 1507406561000000000
```