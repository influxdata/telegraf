package ipmi_sensor

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Connection properties for a Client
type Connection struct {
	Hostname  string
	Username  string
	Password  string
	Port      int
	Interface string
	Privilege string
}

func NewConnection(server string, privilege string) *Connection {
	conn := &Connection{}
	conn.Privilege = privilege
	inx1 := strings.LastIndex(server, "@")
	inx2 := strings.Index(server, "(")
	inx3 := strings.Index(server, ")")

	connstr := server

	if inx1 > 0 {
		security := server[0:inx1]
		connstr = server[inx1+1:]
		up := strings.SplitN(security, ":", 2)
		conn.Username = up[0]
		conn.Password = up[1]
	}

	if inx2 > 0 {
		inx2 = strings.Index(connstr, "(")
		inx3 = strings.Index(connstr, ")")

		conn.Interface = connstr[0:inx2]
		conn.Hostname = connstr[inx2+1 : inx3]
	}

	return conn
}

func (t *Connection) options() []string {
	intf := t.Interface
	if intf == "" {
		intf = "lan"
	}

	options := []string{
		"-H", t.Hostname,
		"-U", t.Username,
		"-P", t.Password,
		"-I", intf,
	}

	if t.Port != 0 {
		options = append(options, "-p", strconv.Itoa(t.Port))
	}
	if t.Privilege != "" {
		options = append(options, "-L", t.Privilege)
	}
	return options
}

// RemoteIP returns the remote (bmc) IP address of the Connection
func (c *Connection) RemoteIP() string {
	if net.ParseIP(c.Hostname) == nil {
		addrs, err := net.LookupHost(c.Hostname)
		if err != nil && len(addrs) > 0 {
			return addrs[0]
		}
	}
	return c.Hostname
}

// LocalIP returns the local (client) IP address of the Connection
func (c *Connection) LocalIP() string {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.Hostname, c.Port))
	if err != nil {
		// don't bother returning an error, since this value will never
		// make it to the bmc if we can't connect to it.
		return c.Hostname
	}
	_ = conn.Close()
	host, _, _ := net.SplitHostPort(conn.LocalAddr().String())
	return host
}
