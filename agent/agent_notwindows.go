// +build !windows

package agent

import "syscall"

const flushSignal = syscall.SIGUSR1
