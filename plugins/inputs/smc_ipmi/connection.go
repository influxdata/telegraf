package smc_ipmi

import "strings"

// Connection options
type Connection struct {
	Hostname string
	Username string
	Password string
	//Port     int
}

// NewConnection constructor
func NewConnection(server string) *Connection {
	conn := &Connection{}
	//conn.Privilege = privilege
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

		//conn.Interface = connstr[0:inx2]
		conn.Hostname = connstr[inx2+1 : inx3]
	}

	return conn
}

func (t *Connection) options() []string {
	options := []string{
		t.Hostname,
		t.Username,
		t.Password,
	}
	return options
}
