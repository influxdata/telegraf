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
const envSys = "HOST_SYS"

type Bond struct {
	HostProc       string   `toml:"host_proc"`
	HostSys        string   `toml:"host_sys"`
	SysDetails     bool     `toml:"collect_sys_details"`
	BondInterfaces []string `toml:"bond_interfaces"`
	BondType       string
}

type sysFiles struct {
	ModeFile    string
	SlaveFile   string
	ADPortsFile string
}

var sampleConfig = `
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"

  ## Sets 'sys' directory path
  ## If not specified, then default is /sys
  # host_proc = "/sys"

  ## By default, telegraf gather stats for all bond interfaces
  ## Setting interfaces will restrict the stats to the specified
  ## bond interfaces.
  # bond_interfaces = ["bond0"]

  ## Tries to collect additional bond details from /sys/class/net/{bond}
  ## currently only useful for LACP (mode 4) bonds
  # collect_sys_details = false

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

		/*
			Some details about bonds only exist in /sys/class/net/
			In particular, LACP bonds track upstream port state here
		*/
		if bond.SysDetails {
			files, err := bond.readSysFiles(bond.HostSys + "/class/net/" + bondName)
			if err != nil {
				acc.AddError(err)
			}
			err = bond.gatherSysDetails(bondName, files, acc)
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
		if name == "Bonding Mode" {
			bond.BondType = value
		}
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

func readFile(filePath string) (string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error inspecting '%s' interface: %v", filePath, err)
	}
	rawFile := strings.TrimSpace(string(file))
	return rawFile, nil
}

func (bond *Bond) readSysFiles(bondDir string) (sysFiles, error) {
	/*
		Files we may need
		bonding/mode
		bonding/slaves
		bonding/ad_num_ports
	*/
	var err error
	var output sysFiles

	output.ModeFile, err = readFile(bondDir + "/bonding/mode")
	if err != nil {
		return sysFiles{}, err
	}
	output.SlaveFile, err = readFile(bondDir + "/bonding/slaves")
	if err != nil {
		return sysFiles{}, err
	}
	if bond.BondType == "IEEE 802.3ad Dynamic link aggregation" {
		output.ADPortsFile, err = readFile(bondDir + "/bonding/ad_num_ports")
		if err != nil {
			return sysFiles{}, err
		}
	}
	return output, nil
}

func (bond *Bond) gatherSysDetails(bondName string, files sysFiles, acc telegraf.Accumulator) error {
	var mode string
	var slaves []string
	var slavesTmp []string
	var adPortCount int

	// To start with, we get the bond operating mode
	scanner := bufio.NewScanner(strings.NewReader(files.ModeFile))
	for scanner.Scan() {
		line := scanner.Text()
		mode = strings.TrimSpace(strings.Split(line, " ")[0])
	}

	tags := map[string]string{
		"bond": bondName,
		"mode": mode,
	}

	// Next we collect the number of bond slaves the system expects
	scanner = bufio.NewScanner(strings.NewReader(files.SlaveFile))
	for scanner.Scan() {
		slavesTmp = strings.Split(scanner.Text(), " ")
	}
	for _, slave := range slavesTmp {
		if slave != "" {
			slaves = append(slaves, slave)
		}
	}
	fmt.Println("Slaves ", slaves)
	if mode == "802.3ad" {
		/*
			If we're in LACP mode, we should check on how the bond ports are
			interacting with the upstream switch ports
		*/
		scanner = bufio.NewScanner(strings.NewReader(files.ADPortsFile))
		for scanner.Scan() {
			fmt.Println("AD Ports ", scanner.Text())
			// a failed conversion can be treated as 0 ports
			adPortCount, _ = strconv.Atoi(strings.TrimSpace(scanner.Text()))
		}
	} else {
		adPortCount = len(slaves)
	}

	fields := map[string]interface{}{
		"slave_count":   len(slaves),
		"ad_port_count": ad_port_count,
	}
	acc.AddFields("bond_sys", fields, tags)
	return nil
}

func (bond *Bond) gatherSlavePart(bondName string, rawFile string, acc telegraf.Accumulator) error {
	var slaveCount int
	tags := map[string]string{
		"bond": bondName,
	}
	fields := map[string]interface{}{
		"status": 0,
	}
	var scanPast bool
	if bond.BondType == "IEEE 802.3ad Dynamic link aggregation" {
		scanPast = true
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
		if strings.Contains(name, "Slave Interface") {
			tags["interface"] = value
			slaveCount++
		}
		if strings.Contains(name, "MII Status") && value == "up" {
			fields["status"] = 1
		}
		if strings.Contains(name, "Link Failure Count") {
			count, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			fields["failures"] = count
			if !scanPast {
				acc.AddFields("bond_slave", fields, tags)
			}
		}
		if strings.Contains(name, "Actor Churned Count") {
			count, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			fields["actor_churned"] = count
		}
		if strings.Contains(name, "Partner Churned Count") {
			count, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			fields["partner_churned"] = count
			fields["total_churned"] = fields["actor_churned"].(int) + fields["partner_churned"].(int)
			acc.AddFields("bond_slave", fields, tags)
		}
	}
	tags = map[string]string{
		"bond": bondName,
	}
	fields = map[string]interface{}{
		"count": slaveCount,
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
