//go:build !linux

package system

import (
	"testing"

	"github.com/shirou/gopsutil/v4/host"
)

// setupOS cannot mock the os-group calls on non-Linux platforms because
// gopsutil reads from native APIs. Probe at runtime instead and return
// true only if the call succeeds.
func setupOS(t testing.TB) bool {
	t.Helper()
	_, err := host.KernelVersion()
	return err == nil
}
