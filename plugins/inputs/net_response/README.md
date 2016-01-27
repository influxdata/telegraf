# Example Input Plugin

The input plugin test UDP/TCP connections response time.
It can also check response text.

### Configuration:

```
# List of UDP/TCP connections you want to check
[[inputs.net_response]]
  protocol = "tcp"
  # Server address (default IP localhost)
  address = "github.com:80"
  # Set timeout (default 1.0)
  timeout = 1.0
  # Set read timeout (default 1.0)
  read_timeout = 1.0
  # String sent to the server
  send = "ssh"
  # Expected string in answer
  expect = "ssh"

[[inputs.net_response]]
  protocol = "tcp"
  address = ":80"

[[inputs.net_response]]
  protocol = "udp"
  # Server address (default IP localhost)
  address = "github.com:80"
  # Set timeout (default 1.0)
  timeout = 1.0
  # Set read timeout (default 1.0)
  read_timeout = 1.0
  # String sent to the server
  send = "ssh"
  # Expected string in answer
  expect = "ssh"

[[inputs.net_response]]
  protocol = "udp"
  address = "localhost:161"
  timeout = 2.0
```

### Measurements & Fields:

- net_response
    - response_time (float, seconds)
    - string_found (bool) # Only if "expected: option is set

### Tags:

- All measurements have the following tags:
    - host
    - port
    - protocol

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter net_response -test
net_response,host=127.0.0.1,port=22,protocol=tcp response_time=0.18070360500000002,string_found=true 1454785464182527094
net_response,host=127.0.0.1,port=2222,protocol=tcp response_time=1.090124776,string_found=false 1454784433658942325

```
