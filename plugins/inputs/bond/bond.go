package bond

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// default host proc path
const defaultHostProc = "/proc"

// env host proc variable name
const envProc = "HOST_PROC"

type Bond struct {
	HostProc       string   `toml:"host_proc"`
	BondInterfaces []string `toml:"bond_interfaces"`
}

var sampleConfig = `
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"

  ## By default, telegraf gather stats for all bond interfaces
  ## Setting interfaces will restrict the stats to the specified
  ## bond interfaces.
  # bond_interfaces = ["bond0"]
`

func (bond *Bond) Description() string {
	return "Collect bond interface status, slaves statuses and failures count"
}

func (bond *Bond) SampleConfig() string {
	return sampleConfig
}

func (bond *Bond) Gather(acc telegraf.Accumulator) error {
	// load proc path, get default value if config value and env variable are empty
	bond.loadPath()
	// list bond interfaces from bonding directory or gather all interfaces.
	bondNames, err := bond.listInterfaces()
	if err != nil {
		return err
	}
	for _, bondName := range bondNames {
		bondAbsPath := bond.HostProc + "/net/bonding/" + bondName
		file, err := ioutil.ReadFile(bondAbsPath)
		if err != nil {
			acc.AddError(fmt.Errorf("error inspecting '%s' interface: %v", bondAbsPath, err))
			continue
		}
		rawFile := strings.TrimSpace(string(file))
		err = bond.gatherBondInterface(bondName, rawFile, acc)
		if err != nil {
			acc.AddError(fmt.Errorf("error inspecting '%s' interface: %v", bondName, err))
		}
	}
	return nil
}

func (bond *Bond) gatherBondInterface(bondName string, rawFile string, acc telegraf.Accumulator) error {
	splitIndex := strings.Index(rawFile, "Slave Interface:")
	if splitIndex == -1 {
		splitIndex = len(rawFile)
	}
	bondPart := rawFile[:splitIndex]
	slavePart := rawFile[splitIndex:]

	err := bond.gatherBondPart(bondName, bondPart, acc)
	if err != nil {
		return err
	}
	err = bond.gatherSlavePart(bondName, slavePart, acc)
	if err != nil {
		return err
	}
	return nil
}

func (bond *Bond) gatherBondPart(bondName string, rawFile string, acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := map[string]string{
		"bond": bondName,
	}

	scanner := bufio.NewScanner(strings.NewReader(rawFile))
	for scanner.Scan() {
		line := scanner.Text()
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			continue
		}
		name := strings.TrimSpace(stats[0])
		value := strings.TrimSpace(stats[1])
		if strings.Contains(name, "Currently Active Slave") {
			fields["active_slave"] = value
		}
		if strings.Contains(name, "MII Status") {
			fields["status"] = 0
			if value == "up" {
				fields["status"] = 1
			}
			acc.AddFields("bond", fields, tags)
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return fmt.Errorf("Couldn't find status info for '%s' ", bondName)
}

func (bond *Bond) gatherSlavePart(bondName string, rawFile string, acc telegraf.Accumulator) error {
	var slave string
	var status int

	scanner := bufio.NewScanner(strings.NewReader(rawFile))
	for scanner.Scan() {
		line := scanner.Text()
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			continue
		}
		name := strings.TrimSpace(stats[0])
		value := strings.TrimSpace(stats[1])
		if strings.Contains(name, "Slave Interface") {
			slave = value
		}
		if strings.Contains(name, "MII Status") {
			status = 0
			if value == "up" {
				status = 1
			}
		}
		if strings.Contains(name, "Link Failure Count") {
			count, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			fields := map[string]interface{}{
				"status":   status,
				"failures": count,
			}
			tags := map[string]string{
				"bond":      bondName,
				"interface": slave,
			}
			acc.AddFields("bond_slave", fields, tags)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// loadPath can be used to read path firstly from config
// if it is empty then try read from env variable
func (bond *Bond) loadPath() {
	if bond.HostProc == "" {
		bond.HostProc = proc(envProc, defaultHostProc)
	}
}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	// return default path
	return path
}

func (bond *Bond) listInterfaces() ([]string, error) {
	var interfaces []string
	if len(bond.BondInterfaces) > 0 {
		interfaces = bond.BondInterfaces
	} else {
		paths, err := filepath.Glob(bond.HostProc + "/net/bonding/*")
		if err != nil {
			return nil, err
		}
		for _, p := range paths {
			interfaces = append(interfaces, filepath.Base(p))
		}
	}
	return interfaces, nil
}

func init() {
	inputs.Add("bond", func() telegraf.Input {
		return &Bond{}
	})
}
