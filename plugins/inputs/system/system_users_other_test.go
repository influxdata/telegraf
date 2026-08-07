//go:build !linux

package system

import (
	"testing"

	"github.com/shirou/gopsutil/v4/host"
)

// setupUsers cannot mock host.Users() on non-Linux platforms because gopsutil
// hardcodes the utmp path or returns ErrNotImplementedError. It probes the
// call at runtime and returns true only if users can actually be read.
func setupUsers(t *testing.T) bool {
	t.Helper()
	_, err := host.Users()
	return err == nil
}
