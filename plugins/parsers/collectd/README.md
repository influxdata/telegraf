# Collectd

The collectd format parses the collectd binary network protocol.  Tags are
created for host, instance, type, and type instance.  All collectd values are
added as float64 fields.

For more information about the binary network protocol see
[here](https://collectd.org/wiki/index.php/Binary_protocol).

You can control the cryptographic settings with parser options.  Create an
authentication file and set `collectd_auth_file` to the path of the file, then
set the desired security level in `collectd_security_level`.

Additional information including client setup can be found
[here](https://collectd.org/wiki/index.php/Networking_introduction#Cryptographic_setup).

You can also change the path to the typesdb or add additional typesdb using
`collectd_typesdb`.

## Configuration

```toml
[[inputs.socket_listener]]
  service_address = "udp://:25826"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "collectd"

  ## Authentication file for cryptographic security levels
  collectd_auth_file = "/etc/collectd/auth_file"
  ## One of none (default), sign, or encrypt
  collectd_security_level = "encrypt"
  ## Path of to TypesDB specifications
  collectd_typesdb = ["/usr/share/collectd/types.db"]

  ## Multi-value plugins can be handled two ways.
  ## "split" will parse and store the multi-value plugin data into separate measurements
  ## "join" will parse and store the multi-value plugin as a single multi-value measurement.
  ## "split" is the default behavior for backward compatibility with previous versions of influxdb.
  collectd_parse_multivalue = "split"
```

## Example Output

```text
memory,type=memory,type_instance=buffered value=2520051712 1560455990829955922
memory,type=memory,type_instance=used value=3710791680 1560455990829955922
memory,type=memory,type_instance=buffered value=2520047616 1560455980830417318
memory,type=memory,type_instance=cached value=9472626688 1560455980830417318
memory,type=memory,type_instance=slab_recl value=2088894464 1560455980830417318
memory,type=memory,type_instance=slab_unrecl value=146984960 1560455980830417318
memory,type=memory,type_instance=free value=2978258944 1560455980830417318
memory,type=memory,type_instance=used value=3707047936 1560455980830417318
```
