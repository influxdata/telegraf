# Example Input Plugin

This input plugin will test HTTP/HTTPS connections.

### Configuration:

```
# HTTP/HTTPS request given an address a method and a timeout
[[inputs.http_response]]
  ## Server address (default http://localhost)
  address = "http://github.com"
  ## Set response_timeout (default 5 seconds)
  response_timeout = "5s"
  ## HTTP Request Method
  method = "GET"
  ## Whether to follow redirects from the server (defaults to false)
  follow_redirects = true
  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"
  ## Optional HTTP Request Body
  # body = '''
  # {'fake':'data'}
  # '''

  ## Optional substring or regex match in body of the response
  ## response_string_match = "\"service_status\": \"up\""
  ## response_string_match = "ok"
  ## response_string_match = "\".*_status\".?:.?\"up\""

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

### Measurements & Fields:

- http_response
    - response_time (float, seconds)
    - http_response_code (int) #The code received

### Tags:

- All measurements have the following tags:
    - server
    - method

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter http_response --test
http_response,method=GET,server=http://www.github.com http_response_code=200i,response_time=6.223266528 1459419354977857955
```
