package snmp

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

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
		cv := fmt.Sprintf("%#v", cmd)[8:] // trim `[]string` prefix
		fmt.Fprintf(os.Stderr, "Unmocked command. Please add the following to `mockedCommands` in snmp_mocks_generate.go, and then run `go generate`:\n\t%s,\n", cv)
		os.Exit(1)
	}
	fmt.Printf("%s", mcr.stdout)
	fmt.Fprintf(os.Stderr, "%s", mcr.stderr)
	if mcr.exitError {
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	execCommand = mockExecCommand
}

// BEGIN GO GENERATE CONTENT
var mockedCommandResults = map[string]mockedCommandResult{
	"snmptranslate\x00-Td\x00-Ob\x00-m\x00all\x00.1.0.0.0":                    mockedCommandResult{stdout: "TEST::testTable\ntestTable OBJECT-TYPE\n  -- FROM\tTEST\n  MAX-ACCESS\tnot-accessible\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) 0 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00-m\x00all\x00.1.0.0.1.1":                  mockedCommandResult{stdout: "TEST::hostname\nhostname OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) 1 1 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00-m\x00all\x00.1.0.0.1.2":                  mockedCommandResult{stdout: "TEST::1.2\nanonymous#1 OBJECT-TYPE\n  -- FROM\tTEST\n::= { iso(1) 0 testOID(0) 1 2 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00-m\x00all\x001.0.0.1.1":                   mockedCommandResult{stdout: "TEST::hostname\nhostname OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) 1 1 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00-m\x00all\x00.1.0.0.0.1.1":                mockedCommandResult{stdout: "TEST::server\nserver OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) testTable(0) testTableEntry(1) 1 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00-m\x00all\x00.1.0.0.0.1.1.0":              mockedCommandResult{stdout: "TEST::server.0\nserver OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) testTable(0) testTableEntry(1) server(1) 0 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00-m\x00all\x00.999":                        mockedCommandResult{stdout: ".999\n [TRUNCATED]\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00TEST::server":                             mockedCommandResult{stdout: "TEST::server\nserver OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) testTable(0) testTableEntry(1) 1 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00TEST::server.0":                           mockedCommandResult{stdout: "TEST::server.0\nserver OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) testTable(0) testTableEntry(1) server(1) 0 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00TEST::testTable":                          mockedCommandResult{stdout: "TEST::testTable\ntestTable OBJECT-TYPE\n  -- FROM\tTEST\n  MAX-ACCESS\tnot-accessible\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) 0 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00TEST::connections":                        mockedCommandResult{stdout: "TEST::connections\nconnections OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tINTEGER\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) testTable(0) testTableEntry(1) 2 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00TEST::latency":                            mockedCommandResult{stdout: "TEST::latency\nlatency OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) testTable(0) testTableEntry(1) 3 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00-Ob\x00TEST::hostname":                           mockedCommandResult{stdout: "TEST::hostname\nhostname OBJECT-TYPE\n  -- FROM\tTEST\n  SYNTAX\tOCTET STRING\n  MAX-ACCESS\tread-only\n  STATUS\tcurrent\n::= { iso(1) 0 testOID(0) 1 1 }\n", stderr: "", exitError: false},
	"snmptranslate\x00-Td\x00TEST::testTable.1":                               mockedCommandResult{stdout: "TEST::testTableEntry\ntestTableEntry OBJECT-TYPE\n  -- FROM\tTEST\n  MAX-ACCESS\tnot-accessible\n  STATUS\tcurrent\n  INDEX\t\t{ server }\n::= { iso(1) 0 testOID(0) testTable(0) 1 }\n", stderr: "", exitError: false},
	"snmptable\x00-Ch\x00-Cl\x00-c\x00public\x00127.0.0.1\x00TEST::testTable": mockedCommandResult{stdout: "server connections latency \nTEST::testTable: No entries\n", stderr: "", exitError: false},
}
