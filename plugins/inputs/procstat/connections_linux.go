//go:build linux
// +build linux

package procstat

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/gosigar/sys/linux"
)

// connInfo represents a single proc's connection and the parent pid (for practical purpouses)
type connInfo struct {
	state   linux.TCPState
	srcIP   net.IP
	srcPort uint32
	dstIP   net.IP
	dstPort uint32
}

// IsAListenPort returns true if the port param is associated with a listener found in the host connections
func (n *networkInfo) IsAListenPort(port uint32) bool {
	_, ok := n.listenPorts[port]
	return ok
}

// Fetch fetches network info: TCP connections and hosts' IPs.
// Parameter getConnections is the function that will be used to obtain TCP connections
// Parameter getLocalIPs is the function that will be used to get IPs.
// It is passed as a parameter to facilitate testing
func (n *networkInfo) Fetch() error {
	var err error
	n.tcp, n.listenPorts, err = getTCPProcInfo()
	if err != nil {
		return fmt.Errorf("gathering host TCP info: %w", err)
	}

	// Get IPs, to be able to resolve procs listening in 0.0.0.0 or ::
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("getting network interfaces: %w", err)
	}

	n.publicIPs, n.privateIPs, err = getLocalIPs(ifaces)
	if err != nil {
		return fmt.Errorf("procstat getting local IPs: %w", err)
	}

	return nil
}

// GetConnectionsByPid return connection info for a particular PID
func (n *networkInfo) GetConnectionsByPid(pid uint32) (conn []connInfo, err error) {
	conn, ok := n.tcp[pid]
	if !ok {
		return conn, errPIDNotFound
	}
	return conn, nil
}

// GetPublicIPs return the list of public IPs (used to connect to others hosts)
func (n *networkInfo) GetPublicIPs() []net.IP {
	return n.publicIPs
}

// GetPrivateIPs return the list of private IPs (loopback devices, virtual, point2point)
func (n *networkInfo) GetPrivateIPs() []net.IP {
	return n.privateIPs
}

// IsPidListeningInAddr returns true if the pid has a listener in that ip and port
// Return false is pid=0
func (n *networkInfo) IsPidListeningInAddr(pid uint32, ip net.IP, port uint32) bool {
	if pid == 0 {
		return false
	}

	for _, c := range n.tcp[pid] {
		if c.srcIP.Equal(ip) && c.srcPort == port {
			return true
		}
	}

	return false
}

// getLocalIPs return the IPv4/v6 addresses active in the current host divided in two groups:
// "publicIPs" contains addresses to connect with other external hosts.
// "privateIPs" contains loopback addreses, virtual interfaces, etc.
// This division is a best effort and probably does not contains all the possibilities.
// It should extract the information from a list of interfaces passed as a parameter.
func getLocalIPs(ifaces []net.Interface) (publicIPs, privateIPs []net.IP, err error) {
	for _, i := range ifaces {
		// Ignore down interfaces
		if i.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := i.Addrs()
		if err != nil {
			return nil, nil, fmt.Errorf("getting addresses from interfaces: %w", err)
		}

		ips, err := extractIPs(addresses)
		if err != nil {
			return nil, nil, fmt.Errorf("getting IPs from interface addresses: %w", err)
		}

		if i.Flags&net.FlagLoopback != 0 || // Ignore loopback interfaces
			i.Flags&net.FlagPointToPoint != 0 || // ignore VPN interfaces
			len(i.HardwareAddr) != hardwareAddrLength || // ignore interfaces without a MAC address
			strings.HasPrefix(i.HardwareAddr.String(), dockerMACPrefix) || // ignore docker virtual interfaces
			strings.HasPrefix(i.HardwareAddr.String(), virtualBoxMACPrefix) { // ignore VirtualBox virtual interfaces
			privateIPs = append(privateIPs, ips...)
		} else {
			for _, i := range ips {
				if i.IsLinkLocalUnicast() {
					// Do not add link-local IPs: 169.254.0.0/16 or fe80::/10
					continue
				}
				publicIPs = append(publicIPs, i)
			}
		}
	}

	return publicIPs, privateIPs, nil
}

// getTCPProcInfo return the connections grouped by pid and a map of listening ports.
// Both results are for IPv4 and IPv6
func getTCPProcInfo() (connectionsByPid map[uint32][]connInfo, listeners map[uint32]interface{}, err error) {
	req := linux.NewInetDiagReq()
	var diagWriter io.Writer
	msgs, err := linux.NetlinkInetDiagWithBuf(req, nil, diagWriter)
	if err != nil {
		return nil, nil, fmt.Errorf("calling netlink to get sockets: %w", err)
	}

	listeners = map[uint32]interface{}{}
	connectionsByPid = map[uint32][]connInfo{}

	inodeToPid, err := mapInodesToPid()
	if err != nil {
		return nil, nil, fmt.Errorf("mapping inodes to pid: %w", err)
	}

	for _, diag := range msgs {
		inodeInfo := inodeToPid[diag.Inode]

		for _, proc := range inodeInfo {
			if linux.TCPState(diag.State) == linux.TCP_LISTEN {
				listeners[uint32(diag.SrcPort())] = nil
			}

			connectionsByPid[proc.pid] = append(connectionsByPid[proc.pid], connInfo{
				state:   linux.TCPState(diag.State),
				srcIP:   diag.SrcIP(),
				srcPort: uint32(diag.SrcPort()),
				dstIP:   diag.DstIP(),
				dstPort: uint32(diag.DstPort()),
			})
		}
	}

	return connectionsByPid, listeners, nil
}

// mapInodesToPid return a map with the procs associated to each inode.
func mapInodesToPid() (map[uint32][]inodeInfo, error) {
	ret := map[uint32][]inodeInfo{}

	fd, err := os.Open("/proc")
	if err != nil {
		return nil, fmt.Errorf("opening /proc: %w", err)
	}
	defer fd.Close()

	dirContents, err := fd.Readdirnames(0)
	if err != nil {
		return nil, fmt.Errorf("reading /proc files: %w", err)
	}

	for _, pidStr := range dirContents {
		readPidFDs(pidStr, ret)
	}

	return ret, nil
}

// readPidFDs given a PID, add to the ret map info about its inodes
func readPidFDs(pidStr string, ret map[uint32][]inodeInfo) {
	pid, err := strconv.ParseUint(pidStr, 10, 32)
	if err != nil {
		// exclude files with a not numeric name. We only want to access pid directories
		return
	}

	pidDir, err := os.Open("/proc/" + pidStr + "/fd/")
	if err != nil {
		// ignore errors:
		//   - missing directory, pid has already finished
		//   - permission denied
		return
	}
	defer pidDir.Close()

	fds, err := pidDir.Readdirnames(0)
	if err != nil {
		return
	}

	for _, fd := range fds {
		link, err := os.Readlink("/proc/" + pidStr + "/fd/" + fd)
		if err != nil {
			continue
		}

		var inode uint32

		_, err = fmt.Sscanf(link, "socket:[%d]", &inode)
		if err != nil {
			// this inode is not a socket
			continue
		}

		ret[inode] = append(ret[inode], inodeInfo{
			pid: uint32(pid),
		})
	}
}
