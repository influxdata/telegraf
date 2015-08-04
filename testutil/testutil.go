package testutil

import (
	"net"
	"net/url"
	"os"
)

var localhost = "localhost"

// GetLocalHost returns the DOCKER_HOST environment variable, parsing
// out any scheme or ports so that only the IP address is returned.
func GetLocalHost() string {
	if dockerHostVar := os.Getenv("DOCKER_HOST"); dockerHostVar != "" {
		u, err := url.Parse(dockerHostVar)
		if err != nil {
			return dockerHostVar
		}

		// split out the ip addr from the port
		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return dockerHostVar
		}

		return host
	}
	return localhost
}
