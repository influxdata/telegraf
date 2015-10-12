# PHPFPM plugin

Get phpfpm stat using either HTTP status page or fpm socket.

# Measurements

Meta:

- unit: int64
- tags: `host=ip pool=poolname`

Measurement names:

- accepted_conn
- listen_queue
- max_listen_queue
- listen_queue_len
- idle_processes
- active_processes
- total_processes
- max_active_processes
- max_children_reached
- slow_requests

# Example output

Using this configuration:

```
[phpfpm]
  # An array of address to gather stats about. Specify an ip on hostname
  # with optional port and path. ie localhost, 10.10.3.33/server-status, etc.
  #
  # We can configure int two modes:
  #   - unixsocket: the string is the path to fpm socket like
  #      /var/run/php5-fpm.sock
  #   - http: the URL has to start with http:// or https://
  #
  # If no servers are specified, then default to 127.0.0.1/server-status
  urls = ["http://localhost/status", "10.0.0.12:/var/run/php5-fpm-www2.sock"]
```

When run with:

```
./telegraf_linux_amd64 -config telegraf.conf -filter phpfpm -test
```

It produces:

```
* Plugin: phpfpm, Collection 1
> [host="10.0.0.12" pool="www"] phpfpm_idle_processes value=1
> [host="10.0.0.12" pool="www"] phpfpm_total_processes value=2
> [host="10.0.0.12" pool="www"] phpfpm_max_children_reached value=0
> [host="10.0.0.12" pool="www"] phpfpm_max_listen_queue value=0
> [host="10.0.0.12" pool="www"] phpfpm_listen_queue value=0
> [host="10.0.0.12" pool="www"] phpfpm_listen_queue_len value=0
> [host="10.0.0.12" pool="www"] phpfpm_active_processes value=1
> [host="10.0.0.12" pool="www"] phpfpm_max_active_processes value=2
> [host="10.0.0.12" pool="www"] phpfpm_slow_requests value=0
> [host="10.0.0.12" pool="www"] phpfpm_accepted_conn value=305

> [host="localhost" pool="www2"] phpfpm_max_children_reached value=0
> [host="localhost" pool="www2"] phpfpm_slow_requests value=0
> [host="localhost" pool="www2"] phpfpm_max_listen_queue value=0
> [host="localhost" pool="www2"] phpfpm_active_processes value=1
> [host="localhost" pool="www2"] phpfpm_listen_queue_len value=0
> [host="localhost" pool="www2"] phpfpm_idle_processes value=1
> [host="localhost" pool="www2"] phpfpm_total_processes value=2
> [host="localhost" pool="www2"] phpfpm_max_active_processes value=2
> [host="localhost" pool="www2"] phpfpm_accepted_conn value=306
> [host="localhost" pool="www2"] phpfpm_listen_queue value=0
```

# Test

We can easily invoke in test mode with

```
./telegraf -config telegraf.conf -filter phpfpm -test
```

to run only this plugin for quickly test it beside `go test`
