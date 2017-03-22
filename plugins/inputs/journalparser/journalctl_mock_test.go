package journalparser

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

//TODO mocking execs would be much easier to work with if instead every time a command was executed, we created some pipes for STDIN/STDOUT/STDERR, and allowed the test code to fetch the pipes.

type mockedCommandResult struct {
	stdout    string
	stderr    string
	exitError bool
}

func mockExecCommand(arg0 string, args ...string) *exec.Cmd {
	args = append([]string{"-test.run=TestMockExecCommand", "--", arg0}, args...)
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stderr = os.Stderr // so the test output shows errors
	return cmd
}

// This is not a real test. This is just a way of mocking out commands.
//
// Idea based on https://github.com/golang/go/blob/7c31043/src/os/exec/exec_test.go#L568
func TestMockExecCommand(t *testing.T) {
	var cmd []string
	for _, arg := range os.Args {
		if string(arg) == "--" {
			cmd = []string{}
			continue
		}
		if cmd == nil {
			continue
		}
		cmd = append(cmd, string(arg))
	}
	if cmd == nil {
		return
	}

	cmd0 := strings.Join(cmd, "\000")
	mcr, ok := mockedCommandResults[cmd0]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unmocked command. Please add the following to `mockedCommandResults`:\n\t%v\n", cmd0)
		os.Exit(1)
	}
	fmt.Printf("%s", mcr.stdout)
	fmt.Fprintf(os.Stderr, "%s", mcr.stderr)
	if mcr.exitError {
		os.Exit(1)
	}
	time.Sleep(time.Second * 60)
	os.Exit(0)
}

func init() {
	execCommand = mockExecCommand
}

var mockedCommandResults = map[string]mockedCommandResult{
	"journalctl\x00-o\x00export\x00-f\x00-n\x000\x00FOO=fooval": mockedCommandResult{
		stdout: `__REALTIME_TIMESTAMP=1492979183630000
FOO=fooval
N=1

__REALTIME_TIMESTAMP=x
FOO=fooval

__REALTIME_TIMESTAMP=1492979183630001
FOO=fooval
N=2

__REALTIME_TIMESTAMP=1492979183630002
BAR=barval
N=3

`,
	},
	"journalctl\x00-o\x00export\x00-f\x00-n\x000\x00FOO=fooval\x00+\x00BAR=barval": mockedCommandResult{
		stdout: `__REALTIME_TIMESTAMP=1492979183630003
FOO=fooval
N=4

__REALTIME_TIMESTAMP=1492979183630004
BAR=barval
N=5

`,
	},
}
