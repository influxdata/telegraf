//go:build windows
// +build windows

package port_name

import (
	"os"
	"path/filepath"
)

func servicesPath() string {
	return filepath.Join(os.Getenv("WINDIR"), `system32\drivers\etc\services`)
}
