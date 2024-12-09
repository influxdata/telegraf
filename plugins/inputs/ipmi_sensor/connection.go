package ipmi_sensor

import (
	"strconv"
	"strings"
)

// connection properties for a Client
type connection struct {
	hostname  string
	username  string
	password  string
	port      int
	intf      string
	privilege string
	hexKey    string
}

func newConnection(server, privilege, hexKey string) *connection {
	conn := &connection{
		privilege: privilege,
		hexKey:    hexKey,
	}
	inx1 := strings.LastIndex(server, "@")
	inx2 := strings.Index(server, "(")

	connstr := server

	if inx1 > 0 {
		security := server[0:inx1]
		connstr = server[inx1+1:]
		up := strings.SplitN(security, ":", 2)
		if len(up) == 2 {
			conn.username = up[0]
			conn.password = up[1]
		}
	}

	if inx2 > 0 {
		inx2 = strings.Index(connstr, "(")
		inx3 := strings.Index(connstr, ")")

		conn.intf = connstr[0:inx2]
		conn.hostname = connstr[inx2+1 : inx3]
	}

	return conn
}

func (c *connection) options() []string {
	intf := c.intf
	if intf == "" {
		intf = "lan"
	}

	options := []string{
		"-H", c.hostname,
		"-U", c.username,
		"-P", c.password,
		"-I", intf,
	}

	if c.hexKey != "" {
		options = append(options, "-y", c.hexKey)
	}
	if c.port != 0 {
		options = append(options, "-p", strconv.Itoa(c.port))
	}
	if c.privilege != "" {
		options = append(options, "-L", c.privilege)
	}
	return options
}
