//+build openbsd

package wgopenbsd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl/internal/wginternal"
	"golang.zx2c4.com/wireguard/wgctrl/internal/wgopenbsd/internal/wgh"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	// ifGroupWG is the WireGuard interface group name passed to the kernel.
	ifGroupWG = [16]byte{0: 'w', 1: 'g'}
)

var _ wginternal.Client = &Client{}

// A Client provides access to OpenBSD WireGuard ioctl information.
type Client struct {
	// Hooks which use system calls by default, but can also be swapped out
	// during tests.
	close           func() error
	ioctlIfgroupreq func(ifg *wgh.Ifgroupreq) error
	ioctlWGGetServ  func(wgs *wgh.WGGetServ) error
	ioctlWGGetPeer  func(wgp *wgh.WGGetPeer) error
}

// New creates a new Client and returns whether or not the ioctl interface
// is available.
func New() (*Client, bool, error) {
	// The OpenBSD ioctl interface operates on a generic AF_INET socket.
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return nil, false, err
	}

	// TODO(mdlayher): find a call to invoke here to probe for availability.
	// c.Devices won't work because it returns a "not found" error when the
	// kernel WireGuard implementation is available but the interface group
	// has no members.

	// By default, use system call implementations for all hook functions.
	return &Client{
		close:           func() error { return unix.Close(fd) },
		ioctlIfgroupreq: ioctlIfgroupreq(fd),
		ioctlWGGetServ:  ioctlWGGetServ(fd),
		ioctlWGGetPeer:  ioctlWGGetPeer(fd),
	}, true, nil
}

// Close implements wginternal.Client.
func (c *Client) Close() error {
	return c.close()
}

// Devices implements wginternal.Client.
func (c *Client) Devices() ([]*wgtypes.Device, error) {
	ifg := wgh.Ifgroupreq{
		// Query for devices in the "wg" group.
		Name: ifGroupWG,
	}

	// Determine how many device names we must allocate memory for.
	if err := c.ioctlIfgroupreq(&ifg); err != nil {
		return nil, err
	}

	// ifg.Len is size in bytes; allocate enough memory for the correct number
	// of wgh.Ifgreq and then store a pointer to the memory where the data
	// should be written (ifgrs) in ifg.Groups.
	//
	// From a thread in golang-nuts, this pattern is valid:
	// "It would be OK to pass a pointer to a struct to ioctl if the struct
	// contains a pointer to other Go memory, but the struct field must have
	// pointer type."
	// See: https://groups.google.com/forum/#!topic/golang-nuts/FfasFTZvU_o.
	ifgrs := make([]wgh.Ifgreq, ifg.Len/wgh.SizeofIfgreq)
	ifg.Groups = &ifgrs[0]

	// Now actually fetch the device names.
	if err := c.ioctlIfgroupreq(&ifg); err != nil {
		return nil, err
	}

	// Keep this alive until we're done doing the ioctl dance.
	runtime.KeepAlive(&ifg)

	devices := make([]*wgtypes.Device, 0, len(ifgrs))
	for _, ifgr := range ifgrs {
		// Remove any trailing NULL bytes from the interface names.
		d, err := c.Device(string(bytes.TrimRight(ifgr.Ifgrqu[:], "\x00")))
		if err != nil {
			return nil, err
		}

		devices = append(devices, d)
	}

	return devices, nil
}

// Device implements wginternal.Client.
func (c *Client) Device(name string) (*wgtypes.Device, error) {
	d, pkeys, err := c.getServ(name)
	if err != nil {
		return nil, err
	}

	d.Peers = make([]wgtypes.Peer, 0, len(pkeys))
	for _, pk := range pkeys {
		p, err := c.getPeer(d.Name, pk)
		if err != nil {
			return nil, err
		}

		d.Peers = append(d.Peers, *p)
	}

	return d, nil
}

// ConfigureDevice implements wginternal.Client.
func (c *Client) ConfigureDevice(name string, cfg wgtypes.Config) error {
	// Currently read-only: we must determine if a device belongs to this driver,
	// and if it does, return a sentinel so integration tests that configure a
	// device can be skipped.
	if _, err := c.Device(name); err != nil {
		return err
	}

	return wginternal.ErrReadOnly
}

// getServ fetches a device and the public keys of its peers using an ioctl.
func (c *Client) getServ(name string) (*wgtypes.Device, []wgtypes.Key, error) {
	nb, err := deviceName(name)
	if err != nil {
		return nil, nil, err
	}

	// Fetch information for the specified device, and indicate that we have
	// pre-allocated room for peer public keys. 8 is the initial array size
	// value used by ncon's wg fork.
	wgs := wgh.WGGetServ{
		Name:      nb,
		Num_peers: 8,
	}

	var (
		// The number of peer public keys we should allocate space for, and
		// the slice where keys are allocated.
		n     int
		bkeys [][wgtypes.KeyLen]byte // []wgtypes.Key equivalent
	)

	for {
		// Updated on each loop iteration to provide enough space in case the
		// kernel tells us we need to provide more space.
		n = int(wgs.Num_peers)

		// See the comment in Devices about passing Go pointers within a
		// structure to ioctl.
		bkeys = make([][wgtypes.KeyLen]byte, n)
		wgs.Peers = &bkeys[0]

		// Query for a device by its name.
		if err := c.ioctlWGGetServ(&wgs); err != nil {

			// ioctl functions always return a wrapped unix.Errno value.
			// Conform to the wgctrl contract by converting "no such device" and
			// "inappropriate ioctl" to "not exist".
			switch err.(*os.SyscallError).Err {
			case unix.ENXIO, unix.ENOTTY:
				return nil, nil, os.ErrNotExist
			default:
				return nil, nil, err
			}
		}

		// Did the kernel tell us there are more peers than can fit in our
		// current memory? If not, we're done.
		if int(wgs.Num_peers) <= n {
			// Re-slice to the exact size needed.
			bkeys = bkeys[:wgs.Num_peers:wgs.Num_peers]
			break
		}
	}

	// wgtypes.Key has an identical memory layout with [wgtypes.KeyLen]byte, so
	// cast the slice directly.
	keys := *(*[]wgtypes.Key)(unsafe.Pointer(&bkeys))

	return &wgtypes.Device{
		Name:       name,
		Type:       wgtypes.OpenBSDKernel,
		PrivateKey: wgs.Privkey,
		PublicKey:  wgs.Pubkey,
		ListenPort: int(wgs.Port),
	}, keys, nil
}

// getPeer fetches a peer associated with a device and a public key.
func (c *Client) getPeer(device string, pubkey wgtypes.Key) (*wgtypes.Peer, error) {
	nb, err := deviceName(device)
	if err != nil {
		return nil, err
	}

	// The algorithm implemented here is the same as the one documented in
	// getServ, but we are fetching WGIP allowed IP arrays instead of peer
	// public keys. See the more in-depth documentation there.

	// 16 is the initial array size value used by ncon's wg fork.
	wgp := wgh.WGGetPeer{
		Name:    nb,
		Pubkey:  pubkey,
		Num_aip: 16,
	}

	var (
		n    int
		aips []wgh.WGCIDR
	)

	for {
		n = int(wgp.Num_aip)

		// See the comment in Devices about passing Go pointers within a
		// structure to ioctl.
		aips = make([]wgh.WGCIDR, n)
		wgp.Aip = &aips[0]

		// Query for a peer by its associated device and public key.
		if err := c.ioctlWGGetPeer(&wgp); err != nil {
			return nil, err
		}

		// Did the kernel tell us there are more allowed IPs than can fit in our
		// current memory? If not, we're done.
		if int(wgp.Num_aip) <= n {
			// Re-slice to the exact size needed.
			aips = aips[:wgp.Num_aip:wgp.Num_aip]
			break
		}
	}

	endpoint, err := parseEndpoint(wgp.Ip)
	if err != nil {
		return nil, err
	}

	allowedIPs, err := parseAllowedIPs(aips)
	if err != nil {
		return nil, err
	}

	return &wgtypes.Peer{
		PublicKey:                   pubkey,
		PresharedKey:                wgp.Psk,
		Endpoint:                    endpoint,
		PersistentKeepaliveInterval: time.Duration(wgp.Pka) * time.Second,
		LastHandshakeTime: time.Unix(
			wgp.Last_handshake.Sec,
			// Conversion required on openbsd/386.
			int64(wgp.Last_handshake.Nsec),
		),
		ReceiveBytes:  int64(wgp.Rx_bytes),
		TransmitBytes: int64(wgp.Tx_bytes),
		AllowedIPs:    allowedIPs,
	}, nil
}

// deviceName converts an interface name string to the format required to pass
// with wgh.WGGetServ.
func deviceName(name string) ([16]byte, error) {
	var out [unix.IFNAMSIZ]byte
	if len(name) > unix.IFNAMSIZ {
		return out, fmt.Errorf("wgopenbsd: interface name %q too long", name)
	}

	copy(out[:], name)
	return out, nil
}

// parseEndpoint parses a peer endpoint from a wgh.WGIP structure.
func parseEndpoint(ip wgh.WGIP) (*net.UDPAddr, error) {
	// sockaddr* structures have family at index 1.
	switch ip[1] {
	case unix.AF_INET:
		sa := *(*unix.RawSockaddrInet4)(unsafe.Pointer(&ip[0]))

		ep := &net.UDPAddr{
			IP:   make(net.IP, net.IPv4len),
			Port: bePort(sa.Port),
		}
		copy(ep.IP, sa.Addr[:])

		return ep, nil
	case unix.AF_INET6:
		sa := *(*unix.RawSockaddrInet6)(unsafe.Pointer(&ip[0]))

		// TODO(mdlayher): IPv6 zone?
		ep := &net.UDPAddr{
			IP:   make(net.IP, net.IPv6len),
			Port: bePort(sa.Port),
		}
		copy(ep.IP, sa.Addr[:])

		return ep, nil
	default:
		// No endpoint configured.
		return nil, nil
	}
}

// bePort interprets a port integer stored in native endianness as a big
// endian value. This is necessary for proper endpoint port handling on
// little endian machines.
func bePort(port uint16) int {
	b := *(*[2]byte)(unsafe.Pointer(&port))
	return int(binary.BigEndian.Uint16(b[:]))
}

// parseAllowedIPs parses allowed IPs from a slice of wgh.WGCIDR structures.
func parseAllowedIPs(aips []wgh.WGCIDR) ([]net.IPNet, error) {
	ipns := make([]net.IPNet, 0, len(aips))
	for _, aip := range aips {
		var size, masklen int
		switch aip.Af {
		case unix.AF_INET:
			size, masklen = net.IPv4len, 32
		case unix.AF_INET6:
			size, masklen = net.IPv6len, 128
		default:
			return nil, fmt.Errorf("wgopenbsd: unrecognized allowed IP address family: %d", aip.Af)
		}

		// Copy the array from aip to retain it.
		ip := make(net.IP, size)
		copy(ip, aip.Ip[:size])

		ipns = append(ipns, net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(int(aip.Mask), masklen),
		})
	}

	return ipns, nil
}

// ioctlIfgroupreq returns a function which performs the appropriate ioctl on
// fd to retrieve members of an interface group.
func ioctlIfgroupreq(fd int) func(*wgh.Ifgroupreq) error {
	return func(ifg *wgh.Ifgroupreq) error {
		return ioctl(fd, wgh.SIOCGIFGMEMB, unsafe.Pointer(ifg))
	}
}

// ioctlWGGetServ returns a function which performs the appropriate ioctl on
// fd to fetch information about a WireGuard device.
func ioctlWGGetServ(fd int) func(*wgh.WGGetServ) error {
	return func(wgs *wgh.WGGetServ) error {
		return ioctl(fd, wgh.SIOCGWGSERV, unsafe.Pointer(wgs))
	}
}

// ioctlWGGetPeer returns a function which performs the appropriate ioctl on
// fd to fetch information about a peer associated with a WireGuard device.
func ioctlWGGetPeer(fd int) func(*wgh.WGGetPeer) error {
	return func(wgp *wgh.WGGetPeer) error {
		return ioctl(fd, wgh.SIOCGWGPEER, unsafe.Pointer(wgp))
	}
}

// ioctl is a raw wrapper for the ioctl system call.
func ioctl(fd int, req uint, arg unsafe.Pointer) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(arg))
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}

	return nil
}
