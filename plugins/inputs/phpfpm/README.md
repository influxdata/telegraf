# PHP-FPM Input Plugin

Get phpfpm stats using either HTTP status page or fpm socket.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics of phpfpm, via HTTP status page or socket
[[inputs.phpfpm]]
  ## An array of addresses to gather stats about. Specify an ip or hostname
  ## with optional port and path
  ##
  ## Plugin can be configured in three modes (either can be used):
  ##   - http: the URL must start with http:// or https://, ie:
  ##       "http://localhost/status"
  ##       "http://192.168.130.1/status?full"
  ##
  ##   - unixsocket: path to fpm socket, ie:
  ##       "/var/run/php5-fpm.sock"
  ##      or using a custom fpm status path:
  ##       "/var/run/php5-fpm.sock:fpm-custom-status-path"
  ##      glob patterns are also supported:
  ##       "/var/run/php*.sock"
  ##
  ##   - fcgi: the URL must start with fcgi:// or cgi://, and port must be present, ie:
  ##       "fcgi://10.0.0.12:9000/status"
  ##       "cgi://10.0.10.12:9001/status"
  ##
  ## Example of multiple gathering from local socket and remote host
  ## urls = ["http://192.168.1.20/status", "/tmp/fpm.sock"]
  urls = ["http://localhost/status"]

  ## Format of stats to parse, set to "status" or "json"
  ## If the user configures the URL to return JSON (e.g.
  ## http://localhost/status?json), set to JSON. Otherwise, will attempt to
  ## parse line-by-line. The JSON mode will produce additional metrics.
  # format = "status"

  ## Duration allowed to complete HTTP requests.
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

When using `unixsocket`, you have to ensure that telegraf runs on same
host, and socket path is accessible to telegraf user.

## Metrics

- phpfpm
  - tags:
    - pool
    - url
  - fields:
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
- phpfpm_process
  - tags:
    - pool
    - request method
    - request uri
    - script
    - url
    - user
  - fields:
    - pid
    - content length
    - last request cpu
    - last request memory
    - request duration
    - requests
    - start time
    - start since
    - state

## Example Output

```text
phpfpm,pool=www accepted_conn=13i,active_processes=2i,idle_processes=1i,listen_queue=0i,listen_queue_len=0i,max_active_processes=2i,max_children_reached=0i,max_listen_queue=0i,slow_requests=0i,total_processes=3i 1453011293083331187
phpfpm,pool=www2 accepted_conn=12i,active_processes=1i,idle_processes=2i,listen_queue=0i,listen_queue_len=0i,max_active_processes=2i,max_children_reached=0i,max_listen_queue=0i,slow_requests=0i,total_processes=3i 1453011293083691422
phpfpm,pool=www3 accepted_conn=11i,active_processes=1i,idle_processes=2i,listen_queue=0i,listen_queue_len=0i,max_active_processes=2i,max_children_reached=0i,max_listen_queue=0i,slow_requests=0i,total_processes=3i 1453011293083691658
```

With the JSON output, additional metrics around processes are generated:

```text
phpfpm,pool=www,url=http://127.0.0.1:44637?full&json accepted_conn=3879i,active_processes=1i,idle_processes=9i,listen_queue=0i,listen_queue_len=0i,max_active_processes=3i,max_children_reached=0i,max_listen_queue=0i,slow_requests=0i,start_since=4901i,total_processes=10i
phpfpm_process,pool=www,request_method=GET,request_uri=/fpm-status?json&full,script=-,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=583i,last_request_cpu=0,last_request_memory=0,request_duration=159i,requests=386i,start_time=1702044927i,state="Running"
phpfpm_process,pool=www,request_method=GET,request_uri=/fpm-status,script=-,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=584i,last_request_cpu=0,last_request_memory=2097152,request_duration=174i,requests=390i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/index.php,script=script.php,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=585i,last_request_cpu=104.93,last_request_memory=2097152,request_duration=9530i,requests=389i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/ping,script=-,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=586i,last_request_cpu=0,last_request_memory=2097152,request_duration=127i,requests=399i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/index.php,script=script.php,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=587i,last_request_cpu=0,last_request_memory=2097152,request_duration=9713i,requests=382i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/ping,script=-,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=588i,last_request_cpu=0,last_request_memory=2097152,request_duration=133i,requests=383i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/fpm-status?json,script=-,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=589i,last_request_cpu=0,last_request_memory=2097152,request_duration=154i,requests=381i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/ping,script=-,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=590i,last_request_cpu=0,last_request_memory=2097152,request_duration=108i,requests=397i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/index.php,script=script.php,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=591i,last_request_cpu=110.28,last_request_memory=2097152,request_duration=9068i,requests=381i,start_time=1702044927i,state="Idle"
phpfpm_process,pool=www,request_method=GET,request_uri=/index.php,script=script.php,url=http://127.0.0.1:44637?full&json,user=- content_length=0i,pid=592i,last_request_cpu=64.27,last_request_memory=2097152,request_duration=15559i,requests=391i,start_time=1702044927i,state="Idle"
```
