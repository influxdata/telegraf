package bond

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// default host proc path
const defaultHostProc = "/proc"
const defaultHostSys = "/sys"

// env host proc variable name
const envProc = "HOST_PROC"
const envProc = "HOST_SYS"

type Bond struct {
	HostProc       string   `toml:"host_proc"`
	HostSys        string   `toml:"host_sys"`
	SysDetails     bool     `toml:"collect_sys_details"`
	BondInterfaces []string `toml:"bond_interfaces"`
}

var sampleConfig = `
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"

  ## Sets 'sys' directory path
  ## If not specified, then default is /sys
  # host_proc = "/sys"

  ## Tries to collect additional bond details from /sys/class/net/{bond}
  ## currently only useful for LACP (mode 4) bonds
  # collect_sys_details = false

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
	bond.loadPaths()
	// list bond interfaces from bonding directory or gather all interfaces.
	bondNames, err := bond.listInterfaces()
	if err != nil {
		return err
	}
	for _, bondName := range bondNames {
		bondAbsPath := bond.HostProc + "/net/bonding/" + bondName
		file, err := os.ReadFile(bondAbsPath)
		if err != nil {
			acc.AddError(fmt.Errorf("error inspecting '%s' interface: %v", bondAbsPath, err))
			continue
		}
		rawProcFile := strings.TrimSpace(string(file))
		err = bond.gatherBondInterface(bondName, rawProcFile, acc)
		if err != nil {
			acc.AddError(fmt.Errorf("error inspecting '%s' interface: %v", bondName, err))
		}

		if bond.SysDetails {
			// Some details about bonds only exist in /sys/class/net/
			sysPath := bond.HostSys + "/class/net/" + bondName
			err = bond.gatherSysDetails(bondName, sysPath, acc)
			if err != nil {
				acc.AddError(fmt.Errorf("error inspecting '%s' interface: %v", bondName, err))
			}
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

func (bond *Bond) gatherSysDetails(bondName string, bondDir string, acc telegraf.Accumulator) error {
	/*
		Files we may need
		bonding/mode
		bonding/slaves
		bonding/ad_num_ports
	*/

	var mode string
	var slaves []string
	var ad_port_count int

	// To start with, we get the bond operating mode
	file, err := os.ReadFile(bondDir + "/bonding/mode")
	if err != nil {
		acc.AddError(fmt.Errorf("error inspecting '%s/bonding/mode' interface: %v", bondDir, err))
		continue
	}
	rawFile := strings.TrimSpace(string(file))
	scanner := bufio.NewScanner(strings.NewReader(rawFile))
	for scanner.Scan() {
		line := scanner.Text()
		mode = strings.Split(line, " ")[0]
	}

	tags := map[string]string{
		"bond": bondName,
		"mode": mode,
	}

	// Next we collect the number of bond slaves the system expects
	file, err := os.ReadFile(bondDir + "/bonding/slaves")
	if err != nil {
		acc.AddError(fmt.Errorf("error inspecting '%s/bonding/slaves' interface: %v", bondDir, err))
		continue
	}
	rawFile := strings.TrimSpace(string(file))
	scanner := bufio.NewScanner(strings.NewReader(rawFile))
	for scanner.Scan() {
		line := scanner.Text()
		slaves = strings.Split(line, " ")
	}

	if mode == "802.3ad" {
		/*
			If we're in LACP mode, we should check on how the bond ports are
			interacting with the upstream switch ports
		*/
		file, err := os.ReadFile(bondDir + "/bonding/ad_num_ports")
		if err != nil {
			acc.AddError(fmt.Errorf("error inspecting '%s/bonding/ad_num_ports' interface: %v", bondDir, err))
			continue
		}
		rawFile := strings.TrimSpace(string(file))
		scanner := bufio.NewScanner(strings.NewReader(rawFile))
		for scanner.Scan() {
			ad_port_count = int(scanner.Text())
		}
	} else {
		ad_port_count = len(slaves)
	}

	fields := map[string]interface{}{
		"slave_count":   len(slaves),
		"ad_port_count": ad_port_count,
	}
	acc.AddFields("bond_sys", fields, tags)
	return nil
}

func (bond *Bond) gatherSlavePart(bondName string, rawFile string, acc telegraf.Accumulator) error {
	var slave string
	var status int
	var slaveCount int
	var churned int

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
		if strings.Contains(name, "Actor Churned Count") || strings.Contains(name, "Partner Churned Count") {
			count, err := strconv.Atoi(value)
			churned += count
		}
		if strings.Contains(name, "Link Failure Count") {
			count, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			fields := map[string]interface{}{
				"status":   status,
				"failures": count,
				"churned":  churned,
			}
			tags := map[string]string{
				"bond":      bondName,
				"interface": slave,
			}
			acc.AddFields("bond_slave", fields, tags)
			slaveCount++
		}
	}
	fields := map[string]interface{}{
		"count": slaveCount,
	}
	tags := map[string]string{
		"bond": bondName,
	}
	acc.AddFields("bond_slave", fields, tags)

	return scanner.Err()
}

// loadPaths can be used to read path firstly from config
// if it is empty then try read from env variable
func (bond *Bond) loadPaths() {
	if bond.HostProc == "" {
		bond.HostProc = proc(envProc, defaultHostProc)
	}
	if bond.HostSys == "" {
		bond.HostSys = proc(envSys, defaultHostSys)
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
