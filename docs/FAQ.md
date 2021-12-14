# Frequently Asked Questions

## Q: How can I monitor the Docker Engine Host from within a container?

You will need to setup several volume mounts as well as some environment
variables:

```shell
docker run --name telegraf \
    -v /:/hostfs:ro \
    -e HOST_ETC=/hostfs/etc \
    -e HOST_PROC=/hostfs/proc \
    -e HOST_SYS=/hostfs/sys \
    -e HOST_VAR=/hostfs/var \
    -e HOST_RUN=/hostfs/run \
    -e HOST_MOUNT_PREFIX=/hostfs \
    telegraf
```

## Q: Why do I get a "no such host" error resolving hostnames that other programs can resolve?

Go uses a pure Go resolver by default for [name resolution](https://golang.org/pkg/net/#hdr-Name_Resolution).
This resolver behaves differently than the C library functions but is more
efficient when used with the Go runtime.

If you encounter problems or want to use more advanced name resolution methods
that are unsupported by the pure Go resolver, you can switch to the cgo
resolver.

If running manually set:

```shell
export GODEBUG=netdns=cgo
```

If running as a service add the environment variable to `/etc/default/telegraf`:

```shell
GODEBUG=netdns=cgo
```

## Q: How can I manage series cardinality?

High [series cardinality][], when not properly managed, can cause high load on
your database.  Telegraf attempts to avoid creating series with high
cardinality, but some monitoring workloads such as tracking containers are are
inherently high cardinality.  These workloads can still be monitored, but care
must be taken to manage cardinality growth.

You can use the following techniques to avoid cardinality issues:

- Use [metric filtering][] options to exclude unneeded measurements and tags.
- Write to a database with an appropriate [retention policy][].
- Consider using the [Time Series Index][tsi].
- Monitor your databases using the [show cardinality][] commands.
- Consult the [InfluxDB documentation][influx docs] for the most up-to-date techniques.

[series cardinality]: https://docs.influxdata.com/influxdb/v1.7/concepts/glossary/#series-cardinality
[metric filtering]: https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#metric-filtering
[retention policy]: https://docs.influxdata.com/influxdb/latest/guides/downsampling_and_retention/
[tsi]: https://docs.influxdata.com/influxdb/latest/concepts/time-series-index/
[show cardinality]: https://docs.influxdata.com/influxdb/latest/query_language/spec/#show-cardinality
[influx docs]: https://docs.influxdata.com/influxdb/latest/
