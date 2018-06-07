# Telegraf Plugin: Tengine

### Configuration:

```
# Read Tengine's basic status information (ngx_http_reqstat_module)
[[inputs.tengine]]
  ## An array of Tengine reqstat module URI to gather stats.
  urls = ["http://127.0.0.1/us"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP response timeout (default: 5s)
  response_timeout = "5s"
```

### Measurements & Fields:

- Measurement
    - bytes_in total number of bytes received from client
    - bytes_out total number of bytes sent to client
    - conn_total total number of accepted connections
    - req_total total number of processed requests
    - http_2xx total number of 2xx requests
    - http_3xx total number of 3xx requests
    - http_4xx total number of 4xx requests
    - http_5xx total number of 5xx requests
    - http_other_status total number of other requests
    - rt accumulation or rt
    - ups_req total number of requests calling for upstream
    - ups_rt accumulation or upstream rt
    - ups_tries total number of times calling for upstream
    - http_200 total number of 200 requests
    - http_206 total number of 206 requests
    - http_302 total number of 302 requests
    - http_304 total number of 304 requests
    - http_403 total number of 403 requests
    - http_404 total number of 404 requests
    - http_416 total number of 416 requests
    - http_499 total number of 499 requests
    - http_500 total number of 500 requests
    - http_502 total number of 502 requests
    - http_503 total number of 503 requests
    - http_504 total number of 504 requests
    - http_508 total number of 508 requests
    - http_other_detail_status total number of requests of other status codes*http_ups_4xx total number of requests of upstream 4xx
    - http_ups_5xx total number of requests of upstream 5xx
### Tags:

- All measurements have the following tags:
    - port
    - server
    - server_name

### Example Output:

Using this configuration:
```
[[inputs.tengine]]
  ## An array of tengine req_status_show URI to gather stats.
  urls = ["http://127.0.0.1/us"]
```

When run with:
```
./telegraf --config telegraf.conf --input-filter tengine --test
```

It produces:
```
* Plugin: tengine, Collection 1
> tengine,host=gcp-thz-api-5,port=80,server=localhost,server_name=localhost bytes_in=9129i,bytes_out=56334i,conn_total=14i,http_200=90i,http_206=0i,http_2xx=90i,http_302=0i,http_304=0i,http_3xx=0i,http_403=0i,http_404=0i,http_416=0i,http_499=0i,http_4xx=0i,http_500=0i,http_502=0i,http_503=0i,http_504=0i,http_508=0i,http_5xx=0i,http_other_detail_status=0i,http_other_status=0i,http_ups_4xx=0i,http_ups_5xx=0i,req_total=90i,rt=0i,ups_req=0i,ups_rt=0i,ups_tries=0i 1526546308000000000
 tengine,host=gcp-thz-api-5,port=80,server=localhost,server_name=28.79.190.35.bc.googleusercontent.com bytes_in=1500i,bytes_out=3009i,conn_total=4i,http_200=1i,http_206=0i,http_2xx=1i,http_302=0i,http_304=0i,http_3xx=0i,http_403=0i,http_404=1i,http_416=0i,http_499=0i,http_4xx=3i,http_500=0i,http_502=0i,http_503=0i,http_504=0i,http_508=0i,http_5xx=0i,http_other_detail_status=0i,http_other_status=0i,http_ups_4xx=0i,http_ups_5xx=0i,req_total=4i,rt=0i,ups_req=0i,ups_rt=0i,ups_tries=0i 1526546308000000000
 tengine,host=gcp-thz-api-5,port=80,server=localhost,server_name=www.google.com bytes_in=372i,bytes_out=786i,conn_total=1i,http_200=1i,http_206=0i,http_2xx=1i,http_302=0i,http_304=0i,http_3xx=0i,http_403=0i,http_404=0i,http_416=0i,http_499=0i,http_4xx=0i,http_500=0i,http_502=0i,http_503=0i,http_504=0i,http_508=0i,http_5xx=0i,http_other_detail_status=0i,http_other_status=0i,http_ups_4xx=0i,http_ups_5xx=0i,req_total=1i,rt=0i,ups_req=0i,ups_rt=0i,ups_tries=0i 1526546308000000000
 tengine,host=gcp-thz-api-5,port=80,server=localhost,server_name=35.190.79.28 bytes_in=4433i,bytes_out=10259i,conn_total=5i,http_200=3i,http_206=0i,http_2xx=3i,http_302=0i,http_304=0i,http_3xx=0i,http_403=0i,http_404=11i,http_416=0i,http_499=0i,http_4xx=11i,http_500=0i,http_502=0i,http_503=0i,http_504=0i,http_508=0i,http_5xx=0i,http_other_detail_status=0i,http_other_status=0i,http_ups_4xx=0i,http_ups_5xx=0i,req_total=14i,rt=0i,ups_req=0i,ups_rt=0i,ups_tries=0i 1526546308000000000
 tengine,host=gcp-thz-api-5,port=80,server=localhost,server_name=tenka-prod-api.txwy.tw bytes_in=3014397400i,bytes_out=14279992835i,conn_total=36844i,http_200=3177339i,http_206=0i,http_2xx=3177339i,http_302=0i,http_304=0i,http_3xx=0i,http_403=0i,http_404=123i,http_416=0i,http_499=0i,http_4xx=123i,http_500=17214i,http_502=4453i,http_503=80i,http_504=0i,http_508=0i,http_5xx=21747i,http_other_detail_status=0i,http_other_status=0i,http_ups_4xx=123i,http_ups_5xx=21747i,req_total=3199209i,rt=245874536i,ups_req=2685076i,ups_rt=245858217i,ups_tries=2685076i 1526546308000000000
```
