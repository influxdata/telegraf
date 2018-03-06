# Example Input Plugin

This input plugin will test HTTP/HTTPS connections.

### Configuration:

```
# HTTP/HTTPS request given an address a method and a timeout
[[inputs.http_response]]
  ## Server address (default http://localhost)
  # address = "http://localhost"

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

  ## Optional substring or regex match in body of the response
  # response_string_match = "\"service_status\": \"up\""
  # response_string_match = "ok"
  # response_string_match = "\".*_status\".?:.?\"up\""

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"
```

### Measurements & Fields:

- http_response
    - response_time (float, seconds) # Not set if target is unreachable for any reason
    - http_response_code (int) # The HTTP code received
	- result_type (string) # Legacy field mantained for backwards compatibility
    - result_code (int) # Details [here](#result-tag-and-result_code-field)


### Tags:

- All measurements have the following tags:
    - server # Server URL used
    - method # HTTP method used (GET, POST, PUT, etc)
    - status_code # String with the HTTP status code
    - result # Details [here](#result-tag-and-result_code-field)

### Result tag and Result_code field
Upon finishing polling the target server, the plugin registers the result of the operation in the `result` tag, and adds a numeric field called `result_code` corresponding with that tag value.

This tag is used to expose network and plugin errors. HTTP errors are considered a sucessful connection by the plugin.

|Tag value                |Corresponding field value|Description|
--------------------------|-------------------------|-----------|
|success                  | 0                       |The HTTP request completed, even if the HTTP code represents an error|
|response_string_mismatch | 1                       |The option `response_string_match` was used, and the body of the response didn't match the regex|
|body_read_error          | 2                       |The option `response_string_match` was used, but the plugin wans't able to read the body of the response. Responses with empty bodies (like 3xx, HEAD, etc) will trigger this error|
|connection_failed        | 3                       |Catch all for any network error not specifically handled by the plugin|
|timeout                  | 4                       |The plugin timed out while awaiting the HTTP connection to complete|
|dns_error                | 5                       |There was a DNS error while attempting to connect to the host|

NOTE: The error codes are derived from the error object returned by the `net/http` Go library, so the accuracy of the errors depends on the handling of error states by the `net/http` Go library. **If a more detailed error report is required use the `log_network_errors` setting.**

### Example Output:

```
http_response,method=GET,server=http://www.github.com,status_code="200",result="sucess" http_response_code=200i,response_time=6.223266528,result_type="sucess",result_code="0" 1459419354977857955
```
