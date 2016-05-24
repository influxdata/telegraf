# Example Input Plugin

The input plugin test UDP/TCP connections response time.
It can also check response text.

### Configuration:

```
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
```

### Measurements & Fields:

- net_response
    - response_time (float, seconds)
    - string_found (bool) # Only if "expected: option is set

### Tags:

- All measurements have the following tags:
    - server
    - port
    - protocol

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter net_response -test
net_response,server=192.168.2.2,port=22,protocol=tcp response_time=0.18070360500000002,string_found=true 1454785464182527094
net_response,server=192.168.2.2,port=2222,protocol=tcp response_time=1.090124776,string_found=false 1454784433658942325

```
