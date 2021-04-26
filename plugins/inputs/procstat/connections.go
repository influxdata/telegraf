package procstat

import (
	"fmt"
	"net"

	"github.com/influxdata/telegraf"
)

const (
	// dockerMACPrefix https://macaddress.io/faq/how-to-recognise-a-docker-container-by-its-mac-address
	dockerMACPrefix = "02:42"
	//nolint:lll // avoid splitting the link
	// virtualBoxMACPrefix https://github.com/mdaniel/virtualbox-org-svn-vbox-trunk/blob/2d259f948bc352ee400f9fd41c4a08710cd9138a/src/VBox/HostDrivers/VBoxNetAdp/VBoxNetAdp.c#L93
	virtualBoxMACPrefix = "0a:00:27"
	// hardwareAddrLength is the number of bytes of a MAC address
	hardwareAddrLength = 6
)

// errPIDNotFound is the error generated when the pid does not have network info
var errPIDNotFound = fmt.Errorf("pid not found")

// inodeInfo represents information of a proc associated with an inode
type inodeInfo struct {
	pid uint32
}

// networkInfo implements networkInfo using the netlink calls and parsing /proc to map sockets to PIDs
type networkInfo struct {
	log telegraf.Logger
	// tcp contains the connection info for each pid
	tcp map[uint32][]connInfo
	// listenPorts is a map with the listen ports in the host, used to ignore inbound connections
	listenPorts map[uint32]interface{}
	// publicIPs list of IPs considered "public" (used to connect to other hosts)
	publicIPs []net.IP
	// privateIPs list of IPs considered "private" (loopback, virtual interfaces, point2point, etc)
	privateIPs []net.IP
}
