# HTTP Response Input Plugin

This input plugin checks HTTP/HTTPS connections.

### Configuration:

```
# HTTP/HTTPS request given an address a method and a timeout
[[inputs.http_response]]
  ## Deprecated in 1.12, use 'urls'
  ## Server address (default http://localhost)
  # address = "http://localhost"

  ## List of urls to query.
  # urls = ["http://localhost"]

  ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
  # http_proxy = "http://localhost:8888"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## HTTP Request Method
  # method = "GET"

  ## Whether to follow redirects from the server (defaults to false)
  # follow_redirects = false

  ## Optional HTTP Request Body
  # body = '''
  # {'fake':'data'}
  # '''

  ## Optional substring or regex match in body of the response (case sensitive)
  # response_string_match = "\"service_status\": \"up\""
  # response_string_match = "ok"
  # response_string_match = "\".*_status\".?:.?\"up\""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"

  ## Interface to use when dialing an address
  # interface = "eth0"
```

### Metrics:

- http_response
  - tags:
    - server (target URL)
    - method (request method)
    - status_code (response status code)
    - result ([see below](#result--result_code))
  - fields:
    - response_time (float, seconds)
    - response_string_match (int, 0 = mismatch / body read error, 1 = match)
    - http_response_code (int, response status code)
	- result_type (string, deprecated in 1.6: use `result` tag and `result_code` field)
    - result_code (int, [see below](#result--result_code))

#### `result` / `result_code`

Upon finishing polling the target server, the plugin registers the result of the operation in the `result` tag, and adds a numeric field called `result_code` corresponding with that tag value.

This tag is used to expose network and plugin errors. HTTP errors are considered a successful connection.

|Tag value                |Corresponding field value|Description|
--------------------------|-------------------------|-----------|
|success                  | 0                       |The HTTP request completed, even if the HTTP code represents an error|
|response_string_mismatch | 1                       |The option `response_string_match` was used, and the body of the response didn't match the regex. HTTP errors with content in their body (like 4xx, 5xx) will trigger this error|
|body_read_error          | 2                       |The option `response_string_match` was used, but the plugin wans't able to read the body of the response. Responses with empty bodies (like 3xx, HEAD, etc) will trigger this error|
|connection_failed        | 3                       |Catch all for any network error not specifically handled by the plugin|
|timeout                  | 4                       |The plugin timed out while awaiting the HTTP connection to complete|
|dns_error                | 5                       |There was a DNS error while attempting to connect to the host|


### Example Output:

```
http_response,method=GET,server=http://www.github.com,status_code=200,result=success http_response_code=200i,response_time=6.223266528,result_type="success",result_code=0i 1459419354977857955
```
