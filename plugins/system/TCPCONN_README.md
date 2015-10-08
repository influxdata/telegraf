Telegraf plugin: TCPCONN

#### Description

The TCPCONN plugin collects TCP connections state by using `lsof`. 

Supported TCP Connection states are follows. 

- established
- syn_sent
- syn_recv
- fin_wait1
- fin_wait2
- time_wait
- close
- close_wait
- last_ack
- listen
- closing
- none


# Measurements:
### TCP Connections measurements:

Meta:
- units: counts

Measurement names:
- established
- syn_sent
- syn_recv
- fin_wait1
- fin_wait2
- time_wait
- close
- close_wait
- last_ack
- listen
- closing
- none

If there are no connection on the state, the metric is not counted.
