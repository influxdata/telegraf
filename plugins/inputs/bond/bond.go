package bond

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// default bond directory path
const (
	BOND_PATH = "/proc/net/bonding"
)

type Bond struct {
	BondPath       string   `toml:"bond_path"`
	BondInterfaces []string `toml:"bond_interfaces"`
}

var sampleConfig = `
  ## Sets bonding directory path
  ## If not specified, then default is:
  bond_path = "/proc/net/bonding"

  ## By default, telegraf gather stats for all bond interfaces
  ## Setting interfaces will restrict the stats to the specified
  ## bond interfaces.
  bond_interfaces = ["bond0"]
`

func (bond *Bond) Description() string {
	return "Collect bond interface status, slaves statuses and failures count"
}

func (bond *Bond) SampleConfig() string {
	return sampleConfig
}

func (bond *Bond) Gather(acc telegraf.Accumulator) error {
	// load path, get default value if config value and env variables are empty;
	// list bond interfaces from bonding directory or gather all interfaces.
	err := bond.listInterfaces()
	if err != nil {
		return err
	}
	for _, bondName := range bond.BondInterfaces {
		file, err := ioutil.ReadFile(bond.BondPath + "/" + bondName)
		if err != nil {
			acc.AddError(fmt.Errorf("E! error due inspecting '%s' interface: %v", bondName, err))
			continue
		}
		rawFile := strings.TrimSpace(string(file))
		err = bond.gatherBondInterface(bondName, rawFile, acc)
		if err != nil {
			acc.AddError(fmt.Errorf("E! error due inspecting '%s' interface: %v", bondName, err))
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

	lines := strings.Split(rawFile, "\n")
	for _, line := range lines {
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			continue
		}
		name := strings.ToLower(strings.Replace(strings.TrimSpace(stats[0]), " ", "_", -1))
		value := strings.TrimSpace(stats[1])
		if strings.Contains(name, "mii_status") {
			fields["status"] = 0
			if value == "up" {
				fields["status"] = 1
			}
			acc.AddFields("bond", fields, tags)
			return nil
		}
	}
	return fmt.Errorf("E! Couldn't find status info for '%s' ", bondName)
}

func (bond *Bond) gatherSlavePart(bondName string, rawFile string, acc telegraf.Accumulator) error {
	var slave string
	var status int

	lines := strings.Split(rawFile, "\n")
	for _, line := range lines {
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			continue
		}
		name := strings.ToLower(strings.Replace(strings.TrimSpace(stats[0]), " ", "_", -1))
		value := strings.TrimSpace(stats[1])
		if strings.Contains(name, "slave_interface") {
			slave = value
		}
		if strings.Contains(name, "mii_status") {
			status = 0
			if value == "up" {
				status = 1
			}
		}
		if strings.Contains(name, "link_failure_count") {
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
	return nil
}

func (bond *Bond) listInterfaces() error {
	if bond.BondPath == "" {
		bond.BondPath = BOND_PATH
	}
	if len(bond.BondInterfaces) == 0 {
		paths, err := filepath.Glob(bond.BondPath + "/*")
		if err != nil {
			return err
		}
		var interfaces []string
		for _, p := range paths {
			interfaces = append(interfaces, filepath.Base(p))
		}
		bond.BondInterfaces = interfaces
	}
	return nil
}

func init() {
	inputs.Add("bond", func() telegraf.Input {
		return &Bond{}
	})
}
