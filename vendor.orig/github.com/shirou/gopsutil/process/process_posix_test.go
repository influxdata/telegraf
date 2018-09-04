// +build linux freebsd

package process

import (
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func Test_SendSignal(t *testing.T) {
	checkPid := os.Getpid()

	p, _ := NewProcess(int32(checkPid))
	err := p.SendSignal(unix.SIGCONT)
	if err != nil {
		t.Errorf("send signal %v", err)
	}
}
