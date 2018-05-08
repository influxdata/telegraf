# Network Response Input Plugin

The input plugin test UDP/TCP connections response time and can optional
verify text in the response.

### Configuration:

```toml
# Collect response time of a TCP or UDP connection
[[inputs.net_response]]
  ## Protocol, must be "tcp" or "udp"
  ## NOTE: because the "udp" protocol does not respond to requests, it requires
  ## a send/expect string pair (see below).
  protocol = "tcp"
  ## Server address (default localhost)
  address = "localhost:80"

  ## Set timeout
  # timeout = "1s"

  ## Set read timeout (only used if expecting a response)
  # read_timeout = "1s"

  ## The following options are required for UDP checks. For TCP, they are
  ## optional. The plugin will send the given string to the server and then
  ## expect to receive the given 'expect' string back.
  ## string sent to the server
  # send = "ssh"
  ## expected string in answer
  # expect = "ssh"

  ## Uncomment to remove deprecated fields; recommended for new deploys
  # fieldexclude = ["result_type", "string_found"]
```

### Metrics:

- net_response
  - tags:
    - server
    - port
    - protocol
    - result
  - fields:
    - response_time (float, seconds)
    - success (int) # success 0, failure 1
    - result_code (int, success = 0, timeout = 1, connection_failed = 2, read_failed = 3, string_mismatch = 4)
    - result_type (string) **DEPRECATED in 1.7; use result tag**
    - string_found (boolean) **DEPRECATED in 1.4; use result tag**
    - response_time (float, seconds)
    - success (int) # success 0, failure 1
    - result_code (int) # success = 0, Failure = 1
    - [**DEPRECATED**] result_type (string) # success, timeout, connection_failed, read_failed, string_mismatch
    - [**DEPRECATED**] string_found (boolean)

### Tags:

- All measurements have the following tags:
    - server
    - port
    - protocol
    - result
  - fields:
    - response_time (float, seconds)
    - success (int) # success 0, failure 1
    - result_code (int, success = 0, timeout = 1, connection_failed = 2, read_failed = 3, string_mismatch = 4)
    - result_type (string) **DEPRECATED in 1.7; use result tag**
    - string_found (boolean) **DEPRECATED in 1.4; use result tag**

### Example Output:

```
net_response,port=8086,protocol=tcp,result=success,server=localhost response_time=0.000092948,result_code=0i,result_type="success" 1525820185000000000
net_response,port=8080,protocol=tcp,result=connection_failed,server=localhost result_code=2i,result_type="connection_failed" 1525820088000000000
net_response,port=8080,protocol=udp,result=read_failed,server=localhost result_code=3i,result_type="read_failed",string_found=false 1525820088000000000
net_response,server=influxdata.com,port=8080,protocol=tcp,host=localhost,result_text="timeout" result_code=1i,result_type="timeout" 1499310361000000000
net_response,server=influxdata.com,port=443,protocol=tcp,host=localhost,result_text="success" result_code=0i,result_type="success",response_time=0.088703864 1499310361000000000
net_response,protocol=tcp,host=localhost,server=this.domain.does.not.exist,port=443,result_text="connection_failed" result_code=1i,result_type="connection_failed" 1499310361000000000
net_response,protocol=udp,host=localhost,server=influxdata.com,port=8080,result_text="read_failed" result_code=1i,result_type="read_failed" 1499310362000000000
net_response,port=31338,protocol=udp,host=localhost,server=localhost,result_text="string_mismatch" result_code=1i,result_type="string_mismatch",string_found=false,response_time=0.00242682 1499310362000000000
net_response,protocol=udp,host=localhost,server=localhost,port=31338,result_text="success" response_time=0.001128598,result_code=0i,result_type="success",string_found=true 1499310362000000000
net_response,server=this.domain.does.not.exist,port=443,protocol=udp,host=localhost,result_text="connection_failed" result_code=1i,result_type="connection_failed" 1499310362000000000
```
