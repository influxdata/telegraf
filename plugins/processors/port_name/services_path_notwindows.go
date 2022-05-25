//go:build !windows
// +build !windows

package port_name

import (
	"os"
)

// servicesPath tries to find the `services` file at the common
// place(s) on most systems and returns its path. If it can't
// find anything, it returns the common default `/etc/services`
func servicesPath() string {
	var files = []string{
		"/etc/services",
		"/usr/etc/services", // fallback on OpenSuSE
	}

	for i := range files {
		if _, err := os.Stat(files[i]); err == nil {
			return files[i]
		}
	}
	return files[0]
}
