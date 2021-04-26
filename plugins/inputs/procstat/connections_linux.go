//go:build linux
// +build linux

/*
* The functions in this file are desgined to help input plugin procstat to gather informatión
* about procs networking.
* The main idea is to add outgoing and listeners for each process.
* Procs in others network namespaces other than the default are taked into account just for outgoing
* connections, as listeners are not accessible, so ignored.
* The main flow of this functions is:
*   - read all /proc/ pid directories to get inodes and network namespaces
*   - for each network NS, call Netlink to get all connections
*   - map those connections to procs (using inode)
*   - return also the list of listeners ports (only in the default network ns)
 */

package procstat

import (
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/elastic/gosigar/sys/linux"
	"github.com/vishvananda/netns"
)

// inode numeric representation of inodes
type inode uint32

// netNSProcInfo stores proc info grouped by network namespace
type netNSProcInfo struct {
	data map[string]netNSProcs
}

// netNSProcs store for each network namespace, a link to its path and the list
// of procs under it
type netNSProcs struct {
	id string
	// inodeToProc map each inode to a list of procs (pid + ppid)
	inodeToProc map[inode][]inodeInfo
}

// connInfo represents a single proc's connection and the parent pid (for
// practical purposes ).
type connInfo struct {
	netNs   string
	state   linux.TCPState
	srcIP   net.IP
	srcPort uint32
	dstIP   net.IP
	dstPort uint32
}

// IsAListenPort returns true if the port param is associated with a listener
// found in the host connections
func (n *networkInfo) IsAListenPort(port uint32) bool {
	_, ok := n.listenPorts[port]
	return ok
}

// Fetch fetches network info: TCP connections and hosts' IPs.
// Parameter getConnections is the function that will be used to obtain TCP
// connections.
// Parameter getLocalIPs is the function that will be used to get IPs.
// It is passed as a parameter to facilitate testing.
func (n *networkInfo) Fetch() error {
	var err error
	n.tcp, n.listenPorts, err = n.getTCPProcInfo()
	if err != nil {
		return fmt.Errorf("TCP info: %w", err)
	}

	// Get IPs, to be able to resolve procs listening in 0.0.0.0 or ::
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("network interfaces: %w", err)
	}

	n.publicIPs, n.privateIPs, err = getLocalIPs(ifaces)
	if err != nil {
		return fmt.Errorf("local IPs: %w", err)
	}

	return nil
}

// GetConnectionsByPid return connection info for a particular PID
func (n *networkInfo) GetConnectionsByPid(pid uint32) ([]connInfo, error) {
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

// GetPrivateIPs return the list of private IPs (loopback devices, virtual,
// point2point)
func (n *networkInfo) GetPrivateIPs() []net.IP {
	return n.privateIPs
}

// IsPidListeningInAddr returns true if the pid has a listener in that ip and
// port.
// Return false is pid=0.
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

// getLocalIPs return the IPv4/v6 addresses active in the current host divided
// in two groups:
// "publicIPs" contains addresses to connect with other external hosts.
// "privateIPs" contains loopback addreses, virtual interfaces, etc.
// This division is a best effort and probably does not contains all the
// possibilities.
// It should extract the information from a list of interfaces passed as a
// parameter.
func getLocalIPs(ifaces []net.Interface) (
	publicIPs,
	privateIPs []net.IP,
	err error,
) {
	for _, i := range ifaces {
		// Ignore down interfaces
		if i.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := i.Addrs()
		if err != nil {
			return nil, nil,
				fmt.Errorf("getting addresses from interfaces: %w", err)
		}

		ips, err := extractIPs(addresses)
		if err != nil {
			return nil, nil,
				fmt.Errorf("getting IPs from interface addresses: %w", err)
		}

		// Ignore loopback interfaces
		if i.Flags&net.FlagLoopback != 0 ||
			// ignore VPN interfaces
			i.Flags&net.FlagPointToPoint != 0 ||
			// ignore interfaces without a MAC address
			len(i.HardwareAddr) != hardwareAddrLength ||
			// ignore docker virtual interfaces
			strings.HasPrefix(i.HardwareAddr.String(), dockerMACPrefix) ||
			// ignore VirtualBox virtual interfaces
			strings.HasPrefix(i.HardwareAddr.String(), virtualBoxMACPrefix) {
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

// getTCPProcInfo return the connections grouped by pid and a map of listening
// ports.
// Both results are for IPv4 and IPv6
// Ignore listeners inside non-default network namespace
func (n *networkInfo) getTCPProcInfo() (map[uint32][]connInfo, map[uint32]interface{}, error) {
	var diagWriter io.Writer

	netNSProcInfo := netNSProcInfo{}
	err := netNSProcInfo.gatherData()
	if err != nil {
		return nil, nil, fmt.Errorf("gathering net info: %w", err)
	}

	defaultNetNs, err := netns.Get()
	if err != nil {
		return nil, nil,
			fmt.Errorf("unable to get current net namespace: %w", err)
	}
	defer defaultNetNs.Close()

	listeners := map[uint32]interface{}{}
	connectionsByPid := map[uint32][]connInfo{}

	for _, netNs := range netNSProcInfo.GetNS() {
		nsHandle, err := netNs.GetNSHandle()
		if err != nil {
			// ignore error, unable to get the handler for the net ns, probably
			// it was a short lived proc
			continue
		}

		// We force this goroutine to an OS thread.
		// It also prevents other threads to inherit the namespace change.
		// https://github.com/golang/go/commit/2595fe7fb6f272f9204ca3ef0b0c55e66fb8d90f
		// What we want to achieve is to get the connections from another
		// network namespace without affecting the operation of the rest of the
		// goroutines.
		// While we are within the modified network ns, we prevent this
		// goroutine from creating others threads, which would inherit the
		// namespace change.
		// We must also leave the default ns in this thread, so that other
		// goroutines using this thread do not have the modified network ns.
		runtime.LockOSThread()

		// Move the process to a different ns to get its connections
		err = netns.Set(nsHandle)
		if err != nil {
			nsHandle.Close()
			n.log.Errorf("unable to change net namespace: %v\n", err)
			continue
		}

		req := linux.NewInetDiagReq()
		msgs, err := linux.NetlinkInetDiagWithBuf(req, nil, diagWriter)
		if err != nil {
			nsHandle.Close()
			n.log.Errorf("calling netlink to get sockets: %v\n", err)
			continue
		}

		// Move back the process to its original net ns
		err = netns.Set(defaultNetNs)
		if err != nil {
			nsHandle.Close()
			n.log.Errorf("unable to change net namespace back to the default: %v\n", err)
			continue
		}
		runtime.UnlockOSThread()

		for _, diag := range msgs {
			inodeInfo := netNs.inodeToProc[inode(diag.Inode)]

			for _, proc := range inodeInfo {
				if linux.TCPState(diag.State) == linux.TCP_LISTEN {
					if !nsHandle.Equal(defaultNetNs) {
						// Ignore listener outside of the default net NS
						nsHandle.Close()
						continue
					}
					listeners[uint32(diag.SrcPort())] = nil
				}

				connectionsByPid[proc.pid] = append(connectionsByPid[proc.pid], connInfo{
					netNs:   netNs.id,
					state:   linux.TCPState(diag.State),
					srcIP:   diag.SrcIP(),
					srcPort: uint32(diag.SrcPort()),
					dstIP:   diag.DstIP(),
					dstPort: uint32(diag.DstPort()),
				})
			}
		}
		nsHandle.Close()
	}

	return connectionsByPid, listeners, nil
}

// gatherData parse /proc to get procs info grouped by network namespace
func (n *netNSProcInfo) gatherData() error {
	n.data = map[string]netNSProcs{}

	fd, err := os.Open("/proc")
	if err != nil {
		return fmt.Errorf("opening /proc: %w", err)
	}
	defer fd.Close()

	dirContents, err := fd.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("reading /proc: %w", err)
	}

	for _, pidStr := range dirContents {
		n.readPidInfo(pidStr)
	}

	return nil
}

// readPidInfo for each PID, get network namespace and associated inodes
func (n *netNSProcInfo) readPidInfo(pidStr string) {
	pid, err := strconv.ParseUint(pidStr, 10, 32)
	if err != nil {
		// exclude files with a not numeric name. We only want to access pid directories
		return
	}

	// Get process net namespace
	netNs, err := os.Readlink("/proc/" + pidStr + "/ns/net")
	if err != nil {
		// ignore errors:
		//   - missing directory, pid has already finished
		//   - permission denied
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
		// ignore errors:
		//   - missing directory, pid has already finished
		return
	}

	for _, fd := range fds {
		link, err := os.Readlink("/proc/" + pidStr + "/fd/" + fd)
		if err != nil {
			// ignore errors:
			//   - missing file, fd disappeared
			continue
		}

		var procInode uint32

		_, err = fmt.Sscanf(link, "socket:[%d]", &procInode)
		if err != nil {
			// this inode is not a socket
			continue
		}

		n.AddProc(netNs, uint32(pid), inode(procInode))
	}
}

// AddProc store pid with its associated network namespace an inode
func (n *netNSProcInfo) AddProc(netNs string, pid uint32, procInode inode) {
	netNsData, ok := n.data[netNs]
	if !ok {
		netNsData = netNSProcs{
			id:          netNs,
			inodeToProc: map[inode][]inodeInfo{},
		}
	}

	procInfo, ok := netNsData.inodeToProc[procInode]
	if ok {
		procInfo = append(procInfo, inodeInfo{pid: pid})
		netNsData.inodeToProc[procInode] = procInfo
	} else {
		netNsData.inodeToProc[procInode] = []inodeInfo{{pid: pid}}
	}

	n.data[netNs] = netNsData
}

// GetNS return the list of network namespaces discovered in the /proc
func (n *netNSProcInfo) GetNS() (ret []netNSProcs) {
	for _, ns := range n.data {
		ret = append(ret, ns)
	}
	return ret
}

// GetNSHandle return the file descriptor for the given network namespace.
// It will use the first proc path available in the procs under this net ns
func (n *netNSProcs) GetNSHandle() (ret netns.NsHandle, err error) {
	for _, procs := range n.inodeToProc {
		for _, p := range procs {
			ret, err = netns.GetFromPath(fmt.Sprintf("/proc/%d/ns/net", p.pid))
			if err == nil {
				return ret, nil
			}
		}
	}

	return ret, fmt.Errorf("net ns not longer available in analyzed procs")
}
