package ethtool

import "net"

type namespace interface {
	name() string
	interfaces() ([]namespacedInterface, error)
	driverName(intf namespacedInterface) (string, error)
	stats(intf namespacedInterface) (map[string]uint64, error)
	get(intf namespacedInterface) (map[string]uint64, error)
}

type namespacedInterface struct {
	net.Interface
	namespace namespace
}
