// +build !dragonfly,!linux,!netbsd,!openbsd,!solaris,!darwin,!freebsd

package postfix

func statCTime(_ interface{}) time.Time {
	return time.Time{}
}
