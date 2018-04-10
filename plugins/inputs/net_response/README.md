# Network Response Input Plugin

The input plugin test UDP/TCP connections response time.
It can also check response text.

### Configuration:

```
[[inputs.net_response]]
  ## Protocol, must be "tcp" or "udp"
  ## NOTE: because the "udp" protocol does not respond to requests, it requires
  ## a send/expect string pair (see below).
  protocol = "tcp"
  ## Server address (default localhost)
  address = "localhost:80"
  ## Set timeout
  timeout = "1s"

  ## Set read timeout (only used if expecting a response)
  read_timeout = "1s"

  ## The following options are required for UDP checks. For TCP, they are
  ## optional. The plugin will send the given string to the server and then
  ## expect to receive the given 'expect' string back.
  ## string sent to the server
  # send = "ssh"
  ## expected string in answer
  # expect = "ssh"

[[inputs.net_response]]
  protocol = "tcp"
  address = ":80"

# TCP or UDP 'ping' given url and collect response time in seconds
[[inputs.net_response]]
  ## Protocol, must be "tcp" or "udp"
  protocol = "tcp"
  ## Server address (default localhost)
  address = "github.com:80"
  ## Set timeout
  timeout = "1s"

  ## Optional string sent to the server
  send = "ssh"
  ## Optional expected string in answer
  expect = "ssh"
  ## Set read timeout (only used if expecting a response)
  read_timeout = "1s"

[[inputs.net_response]]
  protocol = "udp"
  address = "localhost:161"
  timeout = "2s"
  send = "hello server"
  expect = "hello client"
```

### Measurements & Fields:

- net_response
    - response_time (float, seconds)
    - result_type (string) # success, timeout, connection_failed, read_failed, string_mismatch
    - [**DEPRECATED**] string_found (boolean)

### Tags:

- All measurements have the following tags:
    - server
    - port
    - protocol

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter net_response --test
net_response,server=influxdata.com,port=8080,protocol=tcp,host=localhost result_type="timeout" 1499310361000000000
net_response,server=influxdata.com,port=443,protocol=tcp,host=localhost result_type="success",response_time=0.088703864 1499310361000000000
net_response,protocol=tcp,host=localhost,server=this.domain.does.not.exist,port=443 result_type="connection_failed" 1499310361000000000
net_response,protocol=udp,host=localhost,server=influxdata.com,port=8080 result_type="read_failed" 1499310362000000000
net_response,port=31338,protocol=udp,host=localhost,server=localhost result_type="string_mismatch",string_found=false,response_time=0.00242682 1499310362000000000
net_response,protocol=udp,host=localhost,server=localhost,port=31338 response_time=0.001128598,result_type="success",string_found=true 1499310362000000000
net_response,server=this.domain.does.not.exist,port=443,protocol=udp,host=localhost result_type="connection_failed" 1499310362000000000
```
