# Example Input Plugin

This input plugin will test HTTP/HTTPS connections.

### Configuration:

```
# List of UDP/TCP connections you want to check
[[inputs.http_response]]
  # Server address (default http://localhost)
  address = "https://github.com"
  # Set http response timeout (default 10)
  response_timeout = 10
  # HTTP Method (default "GET")
  method = "GET"
```

### Measurements & Fields:

- http_response
    - response_time (float, seconds)
    - http_response_code (int) #The code received

### Tags:

- All measurements have the following tags:
    - server
    - port
    - protocol

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter http_response -test
http_response,server=http://192.168.2.2:2000,method=GET response_time=0.18070360500000002,http_response_code=200 1454785464182527094
```
