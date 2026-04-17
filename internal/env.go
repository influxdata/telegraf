package internal

import (
	"os"
	"path/filepath"
)

// GetProcPath returns the path stored in HOST_PROC env variable, or /proc if HOST_PROC has not been set.
func GetProcPath() string {
	if hostProc := os.Getenv("HOST_PROC"); hostProc != "" {
		return hostProc
	}
	return "/proc"
}

// GetSysPath returns the path stored in HOST_SYS env variable, or /sys if HOST_SYS has not been set.
func GetSysPath() string {
	if hostSys := os.Getenv("HOST_SYS"); hostSys != "" {
		return hostSys
	}
	if prefix := os.Getenv("HOST_MOUNT_PREFIX"); prefix != "" {
		return filepath.Join(prefix, "sys")
	}
	return "/sys"
}

// GetDevPath returns the path of the host /dev tree when telegraf is
// running inside a container. It prefers the explicit HOST_DEV env var
// (matching the gopsutil convention), falls back to
// HOST_MOUNT_PREFIX/dev so a single HOST_MOUNT_PREFIX covers /dev,
// /proc, /sys, /run in one go, and finally defaults to /dev on bare
// metal. See influxdata/telegraf#18671.
func GetDevPath() string {
	if hostDev := os.Getenv("HOST_DEV"); hostDev != "" {
		return hostDev
	}
	if prefix := os.Getenv("HOST_MOUNT_PREFIX"); prefix != "" {
		return filepath.Join(prefix, "dev")
	}
	return "/dev"
}

// GetRunPath returns the path of the host /run tree. Mirrors GetDevPath
// and is used by plugins that read /run/udev entries.
func GetRunPath() string {
	if hostRun := os.Getenv("HOST_RUN"); hostRun != "" {
		return hostRun
	}
	if prefix := os.Getenv("HOST_MOUNT_PREFIX"); prefix != "" {
		return filepath.Join(prefix, "run")
	}
	return "/run"
}
