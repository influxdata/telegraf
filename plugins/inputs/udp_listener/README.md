# UDP listener service input plugin

The UDP listener is a service input plugin that listens for messages on a UDP
socket and adds those messages to InfluxDB.
The plugin expects messages in the
[Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

### Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.udp_listener]]
  ## Address and port to host UDP listener on
  service_address = ":8092"

  ## Number of UDP messages allowed to queue up. Once filled, the
  ## UDP listener will start dropping packets.
  allowed_pending_messages = 10000

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## A Note on UDP OS Buffer Sizes

Some OSes (most notably, Linux) place very restricive limits on the performance
of UDP protocols. It is _highly_ recommended that you increase these OS limits to
at least 8MB before trying to run large amounts of UDP traffic to your instance.
8MB is just a recommendation, and can be adjusted higher.

### Linux
Check the current UDP/IP receive buffer limit & default by typing the following
commands:

```
sysctl net.core.rmem_max
sysctl net.core.rmem_default
```

If the values are less than 8388608 bytes you should add the following lines to
the /etc/sysctl.conf file:

```
net.core.rmem_max=8388608
net.core.rmem_default=8388608
```

Changes to /etc/sysctl.conf do not take effect until reboot.
To update the values immediately, type the following commands as root:

```
sysctl -w net.core.rmem_max=8388608
sysctl -w net.core.rmem_default=8388608
```

### BSD/Darwin

On BSD/Darwin systems you need to add about a 15% padding to the kernel limit
socket buffer. Meaning if you want an 8MB buffer (8388608 bytes) you need to set
the kernel limit to `8388608*1.15 = 9646900`. This is not documented anywhere but
happens
[in the kernel here.](https://github.com/freebsd/freebsd/blob/master/sys/kern/uipc_sockbuf.c#L63-L64)

Check the current UDP/IP buffer limit by typing the following command:

```
sysctl kern.ipc.maxsockbuf
```

If the value is less than 9646900 bytes you should add the following lines
to the /etc/sysctl.conf file (create it if necessary):

```
kern.ipc.maxsockbuf=9646900
```

Changes to /etc/sysctl.conf do not take effect until reboot.
To update the values immediately, type the following commands as root:

```
sysctl -w kern.ipc.maxsockbuf=9646900
```
