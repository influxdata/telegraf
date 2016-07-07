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
$ ./telegraf -config telegraf.conf -input-filter http_response -test
http_response,method=GET,server=http://www.github.com http_response_code=200i,response_time=6.223266528 1459419354977857955
```
