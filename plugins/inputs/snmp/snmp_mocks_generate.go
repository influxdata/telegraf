// +build generate

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// This file is a generator used to generate the mocks for the commands used by the tests.

// These are the commands to be mocked.
var mockedCommands = [][]string{
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.0.0.0"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.0.0.1.1"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.0.0.1.2"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", "1.0.0.1.1"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.0.0.0.1.1"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.0.0.0.1.1.0"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.0.0.0.1.5"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.2.3"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".1.0.0.0.1.7"},
	{"snmptranslate", "-Td", "-Ob", ".iso.2.3"},
	{"snmptranslate", "-Td", "-Ob", "-m", "all", ".999"},
	{"snmptranslate", "-Td", "-Ob", "TEST::server"},
	{"snmptranslate", "-Td", "-Ob", "TEST::server.0"},
	{"snmptranslate", "-Td", "-Ob", "TEST::testTable"},
	{"snmptranslate", "-Td", "-Ob", "TEST::connections"},
	{"snmptranslate", "-Td", "-Ob", "TEST::latency"},
	{"snmptranslate", "-Td", "-Ob", "TEST::description"},
	{"snmptranslate", "-Td", "-Ob", "TEST::hostname"},
	{"snmptranslate", "-Td", "-Ob", "IF-MIB::ifPhysAddress.1"},
	{"snmptranslate", "-Td", "-Ob", "BRIDGE-MIB::dot1dTpFdbAddress.1"},
	{"snmptranslate", "-Td", "-Ob", "TCP-MIB::tcpConnectionLocalAddress.1"},
	{"snmptranslate", "-Td", "TEST::testTable.1"},
	{"snmptable", "-Ch", "-Cl", "-c", "public", "127.0.0.1", "TEST::testTable"},
}

type mockedCommandResult struct {
	stdout    string
	stderr    string
	exitError bool
}

func main() {
	if err := generate(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func generate() error {
	f, err := os.OpenFile("snmp_mocks_test.go", os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	br := bufio.NewReader(f)
	var i int64
	for l, err := br.ReadString('\n'); err == nil; l, err = br.ReadString('\n') {
		i += int64(len(l))
		if l == "// BEGIN GO GENERATE CONTENT\n" {
			break
		}
	}
	f.Truncate(i)
	f.Seek(i, 0)

	fmt.Fprintf(f, "var mockedCommandResults = map[string]mockedCommandResult{\n")

	for _, cmd := range mockedCommands {
		ec := exec.Command(cmd[0], cmd[1:]...)
		out := bytes.NewBuffer(nil)
		err := bytes.NewBuffer(nil)
		ec.Stdout = out
		ec.Stderr = err
		ec.Env = []string{
			"MIBDIRS=+./testdata",
		}

		var mcr mockedCommandResult
		if err := ec.Run(); err != nil {
			if err, ok := err.(*exec.ExitError); !ok {
				mcr.exitError = true
			} else {
				return fmt.Errorf("executing %v: %s", cmd, err)
			}
		}
		mcr.stdout = string(out.Bytes())
		mcr.stderr = string(err.Bytes())
		cmd0 := strings.Join(cmd, "\000")
		mcrv := fmt.Sprintf("%#v", mcr)[5:] // trim `main.` prefix
		fmt.Fprintf(f, "%#v: %s,\n", cmd0, mcrv)
	}
	f.Write([]byte("}\n"))
	f.Close()

	return exec.Command("gofmt", "-w", "snmp_mocks_test.go").Run()
}
