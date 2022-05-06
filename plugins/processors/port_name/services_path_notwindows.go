//go:build !windows
// +build !windows

package port_name

func servicesPath() string {
	return "/etc/services"
}
