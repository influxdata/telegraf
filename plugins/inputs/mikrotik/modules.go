package mikrotik

import (
	"strings"
)

var modules = map[string]string{
	"interface":                       "/rest/interface",
	"interface_wireguard_peers":       "/rest/interface/wireguard/peers",
	"interface_wireless_registration": "/rest/interface/wireless/registration-table",
	"ip_dhcp_server_lease":            "/rest/ip/dhcp-server/lease",
	"ip_firewall_connection":          "/rest/ip/firewall/connection",
	"ip_firewall_filter":              "/rest/ip/firewall/filter",
	"ip_firewall_mangle":              "/rest/ip/firewall/mangle",
	"ip_firewall_nat":                 "/rest/ip/firewall/nat",
	"ipv6_firewall_connection":        "/rest/ipv6/firewall/connection",
	"ipv6_firewall_filter":            "/rest/ipv6/firewall/filter",
	"ipv6_firewall_mangle":            "/rest/ipv6/firewall/mangle",
	"ipv6_firewall_nat":               "/rest/ipv6/firewall/nat",
	"system_resourses":                "/rest/system/resource",
	"system_script":                   "/rest/system/script",
}

func getModuleNames() string {
	moduleNames := []string{}
	for k := range modules {
		moduleNames = append(moduleNames, k)
	}

	return strings.Join(moduleNames, ", ")
}
