package pf

import (
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const measurement = "pf"
const pfctlCommand = "pfctl"

type PF struct {
	UseSudo    bool
	StateTable StateTable
	infoFunc   func() (string, error)
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

var stateTableHeaderRE = regexp.MustCompile("^State Table")
var anyTableHeaderRE = regexp.MustCompile("^[A-Z]")

func (pf *PF) parsePfctlOutput(pfoutput string, acc telegraf.Accumulator) error {
	lines := strings.Split(pfoutput, "\n")
	stateTableFound := false
	for i, line := range lines {
		if stateTableHeaderRE.MatchString(line) {
			endline := len(lines)
			for j, el := range lines[i+1:] {
				if anyTableHeaderRE.MatchString(el) {
					endline = j + i + 1
					break
				}
			}
			if perr := pf.parseStateTable(lines[i+1:endline], acc); perr != nil {
				return perr
			}
			stateTableFound = true
		}
	}
	if !stateTableFound {
		return errParseHeader
	}
	return nil
}

var stateTableRE = regexp.MustCompile(`^  (.*?)\s+(\d+)`)

func (pf *PF) parseStateTable(lines []string, acc telegraf.Accumulator) error {
	st := StateTable{}
	tags, err := st.getTags()
	if err != nil {
		return fmt.Errorf("Can't retrieve struct tags: %v", err)
	}
	fMap := make(map[string]bool)
	for i := 0; i < len(tags); i++ {
		fMap[tags[i]] = false
	}

	for _, v := range lines {
		entries := stateTableRE.FindStringSubmatch(v)
		if entries != nil {
			fs, err := st.setByTag(entries[1], entries[2])
			if err != nil {
				return errors.New("can't set statetable field from tag")
			}
			if fs {
				fMap[entries[1]] = true
			}
		}
	}

	for k, v := range fMap {
		if !v {
			return errMissingData(k)
		}
	}

	fields := make(map[string]interface{})
	fields["entries"] = st.CurrentEntries
	fields["searches"] = st.Searches
	fields["inserts"] = st.Inserts
	fields["removals"] = st.Removals
	acc.AddFields(measurement, fields, make(map[string]string))
	return nil
}

func (pf *PF) callPfctl() (string, error) {
	c, err := pf.buildPfctlCmd()
	if err != nil {
		return "", fmt.Errorf("Can't construct commandline: %s", err)
	}
	out, oerr := c.Output()
	if oerr != nil {
		return string(out), fmt.Errorf("error running %s: %s: %s", pfctlCommand, oerr, oerr.(*exec.ExitError).Stderr)
	}
	return string(out), oerr
}

var execLookPath = exec.LookPath
var execCommand = exec.Command

func (pf *PF) buildPfctlCmd() (*exec.Cmd, error) {
	cmd, err := execLookPath(pfctlCommand)
	if err != nil {
		return nil, fmt.Errorf("can't locate %s: %v", pfctlCommand, err)
	}
	args := []string{"-s", "info"}
	if pf.UseSudo {
		args = append([]string{cmd}, args...)
		cmd, err = execLookPath("sudo")
		if err != nil {
			return nil, fmt.Errorf("can't locate sudo: %v", err)
		}
	}
	c := execCommand(cmd, args...)
	return c, nil
}

type StateTable struct {
	CurrentEntries uint32 `pfctl:"current entries"`
	Searches       uint64 `pfctl:"searches"`
	Inserts        uint64 `pfctl:"inserts"`
	Removals       uint64 `pfctl:"removals"`
}

func (pf *StateTable) getTags() ([]string, error) {
	tags := []string{}
	structVal := reflect.ValueOf(pf).Elem()
	for i := 0; i < structVal.NumField(); i++ {
		tags = append(tags, structVal.Type().Field(i).Tag.Get(pfctlCommand))
	}
	return tags, nil
}

// setByTag sets val for a struct field given the tag. returns false if tag not found.
func (pf *StateTable) setByTag(tag string, val string) (bool, error) {
	structVal := reflect.ValueOf(pf).Elem()

	for i := 0; i < structVal.NumField(); i++ {
		tagField := structVal.Type().Field(i).Tag.Get(pfctlCommand)
		if tagField == tag {
			valueField := structVal.Field(i)
			switch valueField.Type().Kind() {
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				iVal, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					return false, fmt.Errorf("Error parsing \"%s\" into uint: %s", val, err)
				}
				valueField.SetUint(iVal)
				return true, nil
			default:
				return false, fmt.Errorf("unhandled struct type %s", valueField.Type())
			}
		}
	}
	return false, nil
}

func init() {
	inputs.Add("pf", func() telegraf.Input {
		pf := new(PF)
		pf.infoFunc = pf.callPfctl
		return pf
	})
}
