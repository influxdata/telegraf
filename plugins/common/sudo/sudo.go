// Package sudo selects the privilege-escalation program used when a plugin
// sets use_sudo. OpenBSD ships doas in the base system, so plugins use that
// instead of sudo.
package sudo

import "runtime"

func Binary() string {
	if runtime.GOOS == "openbsd" {
		return "doas"
	}
	return "sudo"
}
