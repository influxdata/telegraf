package fail2ban

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

// By all rights, we should use `string literal`, but the string contains "`".
var execStatusOutput = "Status\n" +
	"|- Number of jail:\t3\n" +
	"`- Jail list:\tdovecot, postfix, sshd"
var execStatusDovecotOutput = "Status for the jail: dovecot\n" +
	"|- Filter\n" +
	"|  |- Currently failed:\t11\n" +
	"|  |- Total failed:\t22\n" +
	"|  `- File list:\t/var/log/maillog\n" +
	"`- Actions\n" +
	"   |- Currently banned:\t0\n" +
	"   |- Total banned:\t100\n" +
	"   `- Banned IP list:"
var execStatusPostfixOutput = "Status for the jail: postfix\n" +
	"|- Filter\n" +
	"|  |- Currently failed:\t4\n" +
	"|  |- Total failed:\t10\n" +
	"|  `- File list:\t/var/log/maillog\n" +
	"`- Actions\n" +
	"   |- Currently banned:\t3\n" +
	"   |- Total banned:\t60\n" +
	"   `- Banned IP list:\t192.168.10.1 192.168.10.3"
var execStatusSshdOutput = "Status for the jail: sshd\n" +
	"|- Filter\n" +
	"|  |- Currently failed:\t0\n" +
	"|  |- Total failed:\t5\n" +
	"|  `- File list:\t/var/log/secure\n" +
	"`- Actions\n" +
	"   |- Currently banned:\t2\n" +
	"   |- Total banned:\t50\n" +
	"   `- Banned IP list:\t192.168.0.1 192.168.1.1"

func TestGather(t *testing.T) {
	f := Fail2ban{
		path: "/usr/bin/fail2ban-client",
	}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	var acc testutil.Accumulator
	err := f.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}

	fields1 := map[string]interface{}{
		"banned": 2,
		"failed": 0,
	}
	tags1 := map[string]string{
		"jail": "sshd",
	}

	fields2 := map[string]interface{}{
		"banned": 3,
		"failed": 4,
	}
	tags2 := map[string]string{
		"jail": "postfix",
	}

	fields3 := map[string]interface{}{
		"banned": 0,
		"failed": 11,
	}
	tags3 := map[string]string{
		"jail": "dovecot",
	}

	acc.AssertContainsTaggedFields(t, "fail2ban", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "fail2ban", fields2, tags2)
	acc.AssertContainsTaggedFields(t, "fail2ban", fields3, tags3)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	cmd, args := args[3], args[4:]

	if !strings.HasSuffix(cmd, "fail2ban-client") {
		fmt.Fprint(os.Stdout, "command not found")
		os.Exit(1)
	}

	if len(args) == 1 && args[0] == "status" {
		fmt.Fprint(os.Stdout, execStatusOutput)
		os.Exit(0)
	} else if len(args) == 2 && args[0] == "status" {
		if args[1] == "sshd" {
			fmt.Fprint(os.Stdout, execStatusSshdOutput)
			os.Exit(0)
		} else if args[1] == "postfix" {
			fmt.Fprint(os.Stdout, execStatusPostfixOutput)
			os.Exit(0)
		} else if args[1] == "dovecot" {
			fmt.Fprint(os.Stdout, execStatusDovecotOutput)
			os.Exit(0)
		}
	}
	fmt.Fprint(os.Stdout, "invalid argument")
	os.Exit(1)
}
