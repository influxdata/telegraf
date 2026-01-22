# Nginx Input Plugin

This plugin gathers metrics from the open source [Nginx web server][nginx].
Nginx Plus is a commercial version. For more information about differences
between Nginx (F/OSS) and Nginx Plus, see the Nginx [documentation][diff_doc].

⭐ Telegraf v0.1.5
🏷️ server, web
💻 all

[nginx]: https://www.nginx.com
[diff_doc]: https://www.nginx.com/blog/whats-difference-nginx-foss-nginx-plus/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read Nginx's basic status information (ngx_http_stub_status_module)
[[inputs.nginx]]
  ## An array of Nginx stub_status URI to gather stats.
  urls = ["http://localhost/server_status", "http+unix:///var/run/nginx.sock:/server_status"]

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
