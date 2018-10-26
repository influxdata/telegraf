package pf

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const measurement = "pf"
const pfctlCommand = "pfctl"

type PF struct {
	PfctlCommand string
	PfctlArgs    []string
	UseSudo      bool
	StateTable   []*Entry
	infoFunc     func() (string, error)
}

func (pf *PF) Description() string {
	return "Gather counters from PF"
}

func (pf *PF) SampleConfig() string {
	return `
  ## PF require root access on most systems.
  ## Setting 'use_sudo' to true will make use of sudo to run pfctl.
  ## Users must configure sudo to allow telegraf user to run pfctl with no password.
  ## pfctl can be restricted to only list command "pfctl -s info".
  use_sudo = false
`
}

// Gather is the entrypoint for the plugin.
func (pf *PF) Gather(acc telegraf.Accumulator) error {
	if pf.PfctlCommand == "" {
		var err error
		if pf.PfctlCommand, pf.PfctlArgs, err = pf.buildPfctlCmd(); err != nil {
			acc.AddError(fmt.Errorf("Can't construct pfctl commandline: %s", err))
			return nil
		}
	}

	o, err := pf.infoFunc()
	if err != nil {
		acc.AddError(err)
		return nil
	}

	if perr := pf.parsePfctlOutput(o, acc); perr != nil {
		acc.AddError(perr)
	}
	return nil
}

var errParseHeader = fmt.Errorf("Cannot find header in %s output", pfctlCommand)

func errMissingData(tag string) error {
	return fmt.Errorf("struct data for tag \"%s\" not found in %s output", tag, pfctlCommand)
}

type pfctlOutputStanza struct {
	HeaderRE  *regexp.Regexp
	ParseFunc func([]string, map[string]interface{}) error
	Found     bool
}

var pfctlOutputStanzas = []*pfctlOutputStanza{
	{
		HeaderRE:  regexp.MustCompile("^State Table"),
		ParseFunc: parseStateTable,
	},
	{
		HeaderRE:  regexp.MustCompile("^Counters"),
		ParseFunc: parseCounterTable,
	},
}

var anyTableHeaderRE = regexp.MustCompile("^[A-Z]")

func (pf *PF) parsePfctlOutput(pfoutput string, acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	scanner := bufio.NewScanner(strings.NewReader(pfoutput))
	for scanner.Scan() {
		line := scanner.Text()
		for _, s := range pfctlOutputStanzas {
			if s.HeaderRE.MatchString(line) {
				var stanzaLines []string
				scanner.Scan()
				line = scanner.Text()
				for !anyTableHeaderRE.MatchString(line) {
					stanzaLines = append(stanzaLines, line)
					more := scanner.Scan()
					if more {
						line = scanner.Text()
					} else {
						break
					}
				}
				if perr := s.ParseFunc(stanzaLines, fields); perr != nil {
					return perr
				}
				s.Found = true
			}
		}
	}
	for _, s := range pfctlOutputStanzas {
		if !s.Found {
			return errParseHeader
		}
	}

	acc.AddFields(measurement, fields, make(map[string]string))
	return nil
}

type Entry struct {
	Field      string
	PfctlTitle string
	Value      int64
}

var StateTable = []*Entry{
	{"entries", "current entries", -1},
	{"searches", "searches", -1},
	{"inserts", "inserts", -1},
	{"removals", "removals", -1},
}

var stateTableRE = regexp.MustCompile(`^  (.*?)\s+(\d+)`)

func parseStateTable(lines []string, fields map[string]interface{}) error {
	return storeFieldValues(lines, stateTableRE, fields, StateTable)
}

var CounterTable = []*Entry{
	{"match", "match", -1},
	{"bad-offset", "bad-offset", -1},
	{"fragment", "fragment", -1},
	{"short", "short", -1},
	{"normalize", "normalize", -1},
	{"memory", "memory", -1},
	{"bad-timestamp", "bad-timestamp", -1},
	{"congestion", "congestion", -1},
	{"ip-option", "ip-option", -1},
	{"proto-cksum", "proto-cksum", -1},
	{"state-mismatch", "state-mismatch", -1},
	{"state-insert", "state-insert", -1},
	{"state-limit", "state-limit", -1},
	{"src-limit", "src-limit", -1},
	{"synproxy", "synproxy", -1},
}

var counterTableRE = regexp.MustCompile(`^  (.*?)\s+(\d+)`)

func parseCounterTable(lines []string, fields map[string]interface{}) error {
	return storeFieldValues(lines, counterTableRE, fields, CounterTable)
}

func storeFieldValues(lines []string, regex *regexp.Regexp, fields map[string]interface{}, entryTable []*Entry) error {

	for _, v := range lines {
		entries := regex.FindStringSubmatch(v)
		if entries != nil {
			for _, f := range entryTable {
				if f.PfctlTitle == entries[1] {
					var err error
					if f.Value, err = strconv.ParseInt(entries[2], 10, 64); err != nil {
						return err
					}
				}
			}
		}
	}

	for _, v := range entryTable {
		if v.Value == -1 {
			return errMissingData(v.PfctlTitle)
		}
		fields[v.Field] = v.Value
	}

	return nil
}

func (pf *PF) callPfctl() (string, error) {
	cmd := execCommand(pf.PfctlCommand, pf.PfctlArgs...)
	out, oerr := cmd.Output()
	if oerr != nil {
		ee, ok := oerr.(*exec.ExitError)
		if !ok {
			return string(out), fmt.Errorf("error running %s: %s: (unable to get stderr)", pfctlCommand, oerr)
		}
		return string(out), fmt.Errorf("error running %s: %s: %s", pfctlCommand, oerr, ee.Stderr)
	}
	return string(out), oerr
}

var execLookPath = exec.LookPath
var execCommand = exec.Command

func (pf *PF) buildPfctlCmd() (string, []string, error) {
	cmd, err := execLookPath(pfctlCommand)
	if err != nil {
		return "", nil, fmt.Errorf("can't locate %s: %v", pfctlCommand, err)
	}
	args := []string{"-s", "info"}
	if pf.UseSudo {
		args = append([]string{cmd}, args...)
		cmd, err = execLookPath("sudo")
		if err != nil {
			return "", nil, fmt.Errorf("can't locate sudo: %v", err)
		}
	}
	return cmd, args, nil
}

func init() {
	inputs.Add("pf", func() telegraf.Input {
		pf := new(PF)
		pf.infoFunc = pf.callPfctl
		return pf
	})
}
