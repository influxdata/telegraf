package ethtool

import "net"

type Namespace interface {
	Name() string
	Interfaces() ([]NamespacedInterface, error)
	DriverName(intf NamespacedInterface) (string, error)
	Stats(intf NamespacedInterface) (map[string]uint64, error)
}

type NamespacedInterface struct {
	net.Interface
	Namespace Namespace
}
