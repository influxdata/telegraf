# Apache Zookeeper Input Plugin

This plugin collects variables from [Zookeeper][zookeeper] instances using the
[`mntr` command][admin_guide].

> [!NOTE]
> If the Prometheus Metric provider is enabled in Zookeeper use the
> [prometheus plugin][prometheus] instead with `http://<ip>:7000/metrics`.

⭐ Telegraf v0.2.0
🏷️ applications
💻 all

[zookeeper]: https://zookeeper.apache.org
[admin_guide]: https://zookeeper.apache.org/doc/current/zookeeperAdmin.html#sc_zkCommands
[prometheus]: /plugins/inputs/prometheus/README.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Reads 'mntr' stats from one or many zookeeper servers
[[inputs.zookeeper]]
  ## An array of address to gather stats about. Specify an ip or hostname
  ## with port. ie localhost:2181, 10.0.0.1:2181, etc.

  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 2181 is used
  servers = [":2181"]

  ## Timeout for metric collections from all servers.  Minimum timeout is "1s".
  # timeout = "5s"

  ## Float Parsing - the initial implementation forced any value unable to be
  ## parsed as an int to be a string. Setting this to "float" will attempt to
  ## parse float values as floats and not strings. This would break existing
  ## metrics and may cause issues if a value switches between a float and int.
  # parse_floats = "string"

  ## Optional TLS Config
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Password for the key file if it is encrypted
  # tls_key_pwd = ""
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## List of ciphers to accept, by default all secure ciphers will be accepted
  ## See https://pkg.go.dev/crypto/tls#pkg-constants for supported values.
  ## Use "all", "secure" and "insecure" to add all support ciphers, secure
  ## suites or insecure suites respectively.
  # tls_cipher_suites = ["secure"]
  ## Renegotiation method, "never", "once" or "freely"
  # tls_renegotiation_method = "never"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Troubleshooting

If you have any issues please check the direct Zookeeper output using netcat:

```sh
$ echo mntr | nc localhost 2181
zk_version      3.4.9-3--1, built on Thu, 01 Jun 2017 16:26:44 -0700
zk_avg_latency  0
zk_max_latency  0
zk_min_latency  0
zk_packets_received     8
zk_packets_sent 7
zk_num_alive_connections        1
zk_outstanding_requests 0
zk_server_state standalone
zk_znode_count  129
zk_watch_count  0
zk_ephemerals_count     0
zk_approximate_data_size        10044
zk_open_file_descriptor_count   44
zk_max_file_descriptor_count    4096
```

## Metrics

Exact field names are based on Zookeeper response and may vary between
configuration, platform, and version.

- zookeeper
  - tags:
    - server
    - port
    - state
  - fields:
    - approximate_data_size (integer)
    - avg_latency (integer)
    - ephemerals_count (integer)
    - max_file_descriptor_count (integer)
    - max_latency (integer)
    - min_latency (integer)
    - num_alive_connections (integer)
    - open_file_descriptor_count (integer)
    - outstanding_requests (integer)
    - packets_received (integer)
    - packets_sent (integer)
    - version (string)
    - watch_count (integer)
    - znode_count (integer)
    - followers (integer, leader only)
    - synced_followers (integer, leader only)
    - pending_syncs (integer, leader only)

## Example Output

```text
zookeeper,server=localhost,port=2181,state=standalone ephemerals_count=0i,approximate_data_size=10044i,open_file_descriptor_count=44i,max_latency=0i,packets_received=7i,outstanding_requests=0i,znode_count=129i,max_file_descriptor_count=4096i,version="3.4.9-3--1",avg_latency=0i,packets_sent=6i,num_alive_connections=1i,watch_count=0i,min_latency=0i 1522351112000000000
```
