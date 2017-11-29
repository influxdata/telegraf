// +build !dragonfly,!linux,!netbsd,!openbsd,!solaris,!darwin,!freebsd

package postfix

import (
	"time"
)

func statCTime(_ interface{}) time.Time {
	return time.Time{}
}
