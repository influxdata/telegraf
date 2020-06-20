// +build windows

package portname

import (
	"os"
	"path/filepath"
)

func servicesPath() string {
	return filepath.Join(os.Getenv("WINDIR"), `system32\drivers\etc\services`)
}
