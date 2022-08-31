# Network Response Input Plugin

The input plugin test UDP/TCP connections response time and can optional
verify text in the response.

## Configuration

```toml @sample.conf
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
```

## Metrics

- net_response
  - tags:
    - server
    - port
    - protocol
    - result
  - fields:
    - response_time (float, seconds)
    - result_code (int, success = 0, timeout = 1, connection_failed = 2, read_failed = 3, string_mismatch = 4)

## Example Output

```shell
net_response,port=8086,protocol=tcp,result=success,server=localhost response_time=0.000092948,result_code=0i,result_type="success" 1525820185000000000
net_response,port=8080,protocol=tcp,result=connection_failed,server=localhost result_code=2i,result_type="connection_failed" 1525820088000000000
net_response,port=8080,protocol=udp,result=read_failed,server=localhost result_code=3i,result_type="read_failed",string_found=false 1525820088000000000
```
