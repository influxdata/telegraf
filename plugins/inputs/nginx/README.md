# Nginx Input Plugin

This plugin gathers basic status from the open source web server Nginx. Nginx
Plus is a commercial version. For more information about the differences between
Nginx (F/OSS) and Nginx Plus, see the Nginx [documentation][diff-doc].

[diff-doc]: https://www.nginx.com/blog/whats-difference-nginx-foss-nginx-plus/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read Nginx's basic status information (ngx_http_stub_status_module)
[[inputs.nginx]]
  ## An array of Nginx stub_status URI to gather stats.
  urls = ["http://localhost/server_status"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP response timeout (default: 5s)
  response_timeout = "5s"
```

## Metrics

- Measurement
  - accepts
  - active
  - handled
  - reading
  - requests
  - waiting
  - writing

## Tags

- All measurements have the following tags:
  - port
  - server

## Example Output

Using this configuration:

```toml
[[inputs.nginx]]
  ## An array of Nginx stub_status URI to gather stats.
  urls = ["http://localhost/status"]
```

When run with:

```sh
./telegraf --config telegraf.conf --input-filter nginx --test
```

It produces:

```text
nginx,port=80,server=localhost accepts=605i,active=2i,handled=605i,reading=0i,requests=12132i,waiting=1i,writing=1i 1456690994701784331
```
