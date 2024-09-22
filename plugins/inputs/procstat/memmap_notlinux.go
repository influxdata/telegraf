//go:build !linux

package procstat

func collectMemmap(Process, string, map[string]any) {}
