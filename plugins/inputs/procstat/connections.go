package procstat

import (
	"fmt"
	"net"
)

const (
	// DockerMACPrefix https://macaddress.io/faq/how-to-recognise-a-docker-container-by-its-mac-address
	DockerMACPrefix = "02:42"
	// VirtualBoxMACPrefix https://github.com/mdaniel/virtualbox-org-svn-vbox-trunk/blob/2d259f948bc352ee400f9fd41c4a08710cd9138a/src/VBox/HostDrivers/VBoxNetAdp/VBoxNetAdp.c#L93
	VirtualBoxMACPrefix = "0a:00:27"
	// HardwareAddrLength is the number of bytes of a MAC address
	HardwareAddrLength = 6
)

var (
	// ErrorPIDNotFound is the error generated when the pid does not have network info
	ErrorPIDNotFound = fmt.Errorf("pid not found")
)

// InodeInfo represents information of a proc associated with an inode
type InodeInfo struct {
	pid  uint32
	ppid uint32
}

// NetworkInfo implements NetworkInfo using the netlink calls and parsing /proc to map sockets to PIDs
type NetworkInfo struct {
	// tcp contains the connection info for each pid
	tcp map[uint32][]ConnInfo
	// listenPorts is a map with the listen ports in the host, used to ignore inbound connections
	listenPorts map[uint32]interface{}
	// publicIPs list of IPs considered "public" (used to connect to other hosts)
	publicIPs []net.IP
	// privateIPs list of IPs considered "private" (loopback, virtual interfaces, point2point, etc)
	privateIPs []net.IP
}
