# Telegraf plugin: phpfpm

Get phpfpm stat using either HTTP status page or fpm socket.

# Measurements

Meta:

- tags: `pool=poolname`

Measurement names:

- phpfpm

Measurement field:

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
  # We can configure in three modes:
  #   - unixsocket: the string is the path to fpm socket like
  #      /var/run/php5-fpm.sock
  #   - http: the URL has to start with http:// or https://
  #   - fcgi: the URL has to start with fcgi:// or cgi://, and socket port must present
  #
  # If no servers are specified, then default to 127.0.0.1/server-status
  urls = ["http://localhost/status", "10.0.0.12:/var/run/php5-fpm-www2.sock", "fcgi://10.0.0.12:9000/status"]
```

When run with:

```
./telegraf --config telegraf.conf --input-filter phpfpm --test
```

It produces:

```
* Plugin: phpfpm, Collection 1
> phpfpm,pool=www accepted_conn=13i,active_processes=2i,idle_processes=1i,listen_queue=0i,listen_queue_len=0i,max_active_processes=2i,max_children_reached=0i,max_listen_queue=0i,slow_requests=0i,total_processes=3i 1453011293083331187
> phpfpm,pool=www2 accepted_conn=12i,active_processes=1i,idle_processes=2i,listen_queue=0i,listen_queue_len=0i,max_active_processes=2i,max_children_reached=0i,max_listen_queue=0i,slow_requests=0i,total_processes=3i 1453011293083691422
> phpfpm,pool=www3 accepted_conn=11i,active_processes=1i,idle_processes=2i,listen_queue=0i,listen_queue_len=0i,max_active_processes=2i,max_children_reached=0i,max_listen_queue=0i,slow_requests=0i,total_processes=3i 1453011293083691658
```

## Note

When using `unixsocket`, you have to ensure that telegraf runs on same
host, and socket path is accessible to telegraf user.
