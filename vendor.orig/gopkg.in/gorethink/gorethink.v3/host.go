package gorethink

import (
	"fmt"
)

// Host name and port of server
type Host struct {
	Name string
	Port int
}

// NewHost create a new Host
func NewHost(name string, port int) Host {
	return Host{
		Name: name,
		Port: port,
	}
}

// Returns host address (name:port)
func (h Host) String() string {
	return fmt.Sprintf("%s:%d", h.Name, h.Port)
}
