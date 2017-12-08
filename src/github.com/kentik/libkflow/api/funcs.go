package api

import (
	"fmt"
	"net"
	"strings"
)

func NormalizeName(name string) string {
	return strings.Replace(name, ".", "_", -1)
}

func GetInterfaceUpdates(nif *net.Interface) (map[string]InterfaceUpdate, error) {
	addrs, err := interfaceAddrs(nif)
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 || nif.HardwareAddr == nil {
		return nil, fmt.Errorf("interface details unavailable")
	}

	addr := addrs[0]
	if len(addrs) > 1 {
		addrs = addrs[1:]
	} else {
		addrs = nil
	}

	n := len(nif.HardwareAddr)
	a := int(nif.HardwareAddr[n-2])
	b := int(nif.HardwareAddr[n-1])

	return map[string]InterfaceUpdate{
		nif.Name: InterfaceUpdate{
			Index:   uint64(a<<8 | b),
			Alias:   "",
			Desc:    nif.Name,
			Address: addr.Address,
			Netmask: addr.Netmask,
			Addrs:   addrs,
		},
	}, nil
}

func interfaceAddrs(nif *net.Interface) ([]Addr, error) {
	all, err := nif.Addrs()
	if err != nil {
		return nil, err
	}

	var addrs []Addr
	for _, a := range all {
		if a, ok := a.(*net.IPNet); ok && a.IP.IsGlobalUnicast() {
			addrs = append(addrs, Addr{
				Address: a.IP.String(),
				Netmask: net.IP(a.Mask).String(),
			})
		}
	}

	return addrs, nil
}
