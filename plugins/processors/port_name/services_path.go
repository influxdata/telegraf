// +build windows

package portname

import (
	"os"
)

func servicesPath() string {
	return os.Getenv("WINDIR") + `\system32\drivers\etc\services`
}
