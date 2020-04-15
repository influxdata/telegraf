## ping

[![GoDoc](https://godoc.org/github.com/glinton/ping?status.svg)](https://godoc.org/github.com/glinton/ping)

Simple (ICMP) library patterned after net/http.

Originally inspired by [sparrc/go-ping](https://github.com/sparrc/go-ping)


### Installation

```sh
go get -du github.com/glinton/ping
```

To install an all-go iputils patterned `ping` binary
```sh
go get github.com/glinton/ping/cmd/ping
# to run:
$GOPATH/bin/ping localhost
```

### Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/glinton/ping"
)

func main() {
	res, err := ping.IPv4(context.Background(), "google.com")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Completed one ping to google.com with %d bytes in %v\n",
		res.TotalLength, res.RTT)
}
```


### Notes Regarding ICMP Socket Permissions

System installed `ping` binaries generally have `setuid` attributes set, thus allowing them to utilize privileged ICMP sockets. This should work for applications built with this library as well, but a better approach would be to give the application the capability to create privileged ICMP sockets. To do so, run the following as the `root` user (not applicable to Windows).

```
setcap cap_net_raw=eip /path/to/your/application
```

This library tries to initialize a privileged ICMP socket before falling back to unprivileged raw sockets on Linux and Darwin (`udp4` or `udp6` as the network). On Linux, the system group of the user running the application must be allowed to create unprivileged ICMP sockets if desired. [See man pages icmp(7) for `ping_group_range`](http://man7.org/linux/man-pages/man7/icmp.7.html).

To allow a range of groups access to create unprivileged icmp sockets on linux (ipv4 or ipv6), run:

```
sudo sysctl -w net.ipv4.ping_group_range="GROUPID_START GROUPID_END"
```

If you plan to run your application as `root`, the aforementioned commmand is not necessary.

On Windows, running a terminal as admin should not be necessary.


### Known Issues

There is currently no support for TTL on windows, track progress at https://github.com/golang/go/issues/7175 and https://github.com/golang/go/issues/7174

Report any other issues you may find [here](https://github.com/glinton/ping/issues/new)
