// +build linux

// Package smartctl is a collector for S.M.A.R.T data for HDD, SSD + NVMe devices, linux only
// https://www.smartmontools.org/
package smartctl

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Disk is the struct to capture the specifics of a given device
// It will store some basic information plus anything extended queries return
type Disk struct {
	Name      string
	Vendor    string
	Product   string
	Block     string
	Serial    string
	Rotation  string
	Transport string
	Health    string
	ReadCache string
	Writeback string
	RawData   bytes.Buffer
	Stats     DiskStats
}

// DiskFail is the struct to return from our goroutine parseDisks
type DiskFail struct {
	Name  string
	Error error
}

// DiskStats is the set of fields we'll be collecting from smartctl
type DiskStats struct {
	CurrentTemp float64
	MaxTemp     float64
	ReadError   []float64
	WriteError  []float64
	VerifyError []float64
}

// SmartCtl is the struct that stores our disk paths for checking
type SmartCtl struct {
	Init       bool
	SudoPath   string
	CtlPath    string
	Include    []string
	Exclude    []string
	Disks      []string
	DiskOutput map[string]Disk
	DiskFailed map[string]error
}

// tagFinders is a global map of Disk struct elements to corresponding regexp
var tagFinders = map[string]*regexp.Regexp{
	"Vendor":    regexp.MustCompile(`Vendor:\s+(\w+)`),
	"Product":   regexp.MustCompile(`Product:\s+(\w+)`),
	"Block":     regexp.MustCompile(`Logical block size:\s+(\w+)`),
	"Serial":    regexp.MustCompile(`Serial number:\s+(\w+)`),
	"Rotation":  regexp.MustCompile(`Rotation Rate:\s+(\w+)`),
	"Transport": regexp.MustCompile(`Transport protocol:\s+(\w+)`),
	"Health":    regexp.MustCompile(`SMART Health Status:\s+(\w+)`),
	"ReadCache": regexp.MustCompile(`Read Cache is:\s+(\w+)`),
	"Writeback": regexp.MustCompile(`Writeback Cache is:\s+(\w+)`),
}

// fieldFinders is a global map of DiskStats struct elements
var fieldFinders = map[string]*regexp.Regexp{
	"CurrentTemp": regexp.MustCompile(`Current Drive Temperature:\s+([0-9]+)`),
	"MaxTemp":     regexp.MustCompile(`Drive Trip Temperature:\s+([0-9]+)`),
}

// sliceFinders is a global map of slices in the DiskStats struct
var sliceFinders = map[string]*regexp.Regexp{
	"ReadError":   regexp.MustCompile(`read:\s+(.*)\n`),
	"WriteError":  regexp.MustCompile(`write:\s+(.*)\n`),
	"VerifyError": regexp.MustCompile(`verify:\s+(.*)\n`),
}

var sampleConfig = `
  ## smartctl requires installation of the smartmontools for your distro (linux only)
  ## along with root permission to run. In this collector we presume sudo access to the 
  ## binary.
  ##
  ## Users have the ability to specify an list of disk name to include, to exclude, 
  ## or both. In this iteration of the collectors, you must specify the full smartctl
  ## path for the disk, we are not currently supporting regex. For example, to include/exclude
  ## /dev/sda from your list, you would specify:
  ## include = ["/dev/sda -d scsi"]
  ## exclude = ['/dev/sda -d scsi"]
  ## 
  ## NOTE: If you specify an include list, this will skip the smartctl --scan function
  ## and only collect for those you've requested (minus any exclusions).
  include = ["/dev/bus/0 -d megaraid,24"]
  exclude = ["/dev/sda -d scsi"]
`

// SampleConfig returns the preformatted string on how to use the smartctl collector
func (s *SmartCtl) SampleConfig() string {
	return sampleConfig
}

// Description returns a preformatted string outlining the smartctl collector
func (s *SmartCtl) Description() string {
	return "Use the linux smartmontool to determine HDD, SSD or NVMe physical status"
}

// ParseString is a generic function that takes a given regexp and applies to a buf, placing
// output into the dataVar
func (s *SmartCtl) ParseString(regexp *regexp.Regexp, buf *bytes.Buffer, dataVar *string) {
	str := regexp.FindStringSubmatch((*buf).String())

	if len(str) > 1 {
		*dataVar = str[1]
		return
	}

	*dataVar = "none"
}

// ParseStringSlice is a generic function that takes a given regexp and applies to a buf, placing
// output into the dataVar
func (s *SmartCtl) ParseStringSlice(regexp *regexp.Regexp, buf *bytes.Buffer, dataVar *[]string) {
	str := regexp.FindStringSubmatch((*buf).String())

	if len(str) > 1 {
		*dataVar = str[1:]
	}
}

// ParseFloat is a generic function that takes a given regexp and applies to a buf, placing
// output into the dataVar
func (s *SmartCtl) ParseFloat(regexp *regexp.Regexp, buf *bytes.Buffer, dataVar *float64) (err error) {
	str := regexp.FindStringSubmatch((*buf).String())

	if len(str) > 1 {
		*dataVar, err = strconv.ParseFloat(str[1], 64)
		if err != nil {
			return fmt.Errorf("[ERROR] Could not convert string (%s) to float64: %v\n", str[1], err)
		}
	}

	return nil
}

// ParseFloatSlice is a generic function that takes a given regexp and applies to a buf, placing
// output into the dataVar
func (s *SmartCtl) ParseFloatSlice(regexp *regexp.Regexp, buf *bytes.Buffer, dataVar *[]float64) (err error) {
	var errors []string
	var values []float64
	var val float64

	str := regexp.FindStringSubmatch((*buf).String())

	if len(str) > 1 {
		for _, each := range strings.Split(str[1], " ") {
			if len(each) <= 0 {
				continue
			} else if val, err = strconv.ParseFloat(each, 64); err != nil {
				errors = append(errors, fmt.Sprintf("[ERROR] Could not parse string (%s) into float64: %v", each, err))
				continue
			}
			values = append(values, val)
		}
	}

	*dataVar = values
	if len(errors) > 0 {
		return fmt.Errorf("%s\n", strings.Join(errors, "\n"))
	}

	return nil
}

// parseDisks is a private function that we call for our goroutine
func (s *SmartCtl) parseDisks(each string, c chan<- Disk, e chan<- DiskFail) {
	var out []byte
	var data Disk
	var stats DiskStats
	var err error

	var tagMap = map[string]*string{
		"Vendor":    &data.Vendor,
		"Product":   &data.Product,
		"Block":     &data.Block,
		"Serial":    &data.Serial,
		"Rotation":  &data.Rotation,
		"Transport": &data.Transport,
		"Health":    &data.Health,
		"ReadCache": &data.ReadCache,
		"Writeback": &data.Writeback,
	}

	var fieldMap = map[string]*float64{
		"CurrentTemp": &stats.CurrentTemp,
		"MaxTemp":     &stats.MaxTemp,
	}

	var sliceMap = map[string]*[]float64{
		"ReadError":   &stats.ReadError,
		"WriteError":  &stats.WriteError,
		"VerifyError": &stats.VerifyError,
	}

	disk := strings.Split(each, " ")
	cmd := []string{s.CtlPath, "-x"}
	cmd = append(cmd, disk...)

	data = Disk{Name: "empty"}
	if out, err = exec.Command(s.SudoPath, cmd...).CombinedOutput(); err != nil {
		e <- DiskFail{Name: each, Error: fmt.Errorf("[ERROR] could not collect (%s), err: %v\n", each, err)}
		return
	}

	if _, err = data.RawData.Write(out); err != nil {
		e <- DiskFail{Name: each, Error: fmt.Errorf("[ERROR] could not commit raw data to struct (%s): %v\n", each, err)}
		return
	}

	if len(disk) > 2 {
		data.Name = strings.Replace(fmt.Sprintf("%s_%s", disk[0], disk[2]), ",", "_", -1)
	} else {
		data.Name = strings.Replace(disk[0], ",", "_", -1)
	}

	// NOTE: for this loop to work you must keep the idx + Disk element names equal
	for idx := range tagFinders {
		s.ParseString(tagFinders[idx], &data.RawData, tagMap[idx])
	}

	stats = DiskStats{}
	for idx := range fieldFinders {
		if err = s.ParseFloat(fieldFinders[idx], &data.RawData, fieldMap[idx]); err != nil {
			fmt.Printf("[ERROR] ParseFloat: %v\n", err)
		}
	}

	for idx := range sliceFinders {
		if err = s.ParseFloatSlice(sliceFinders[idx], &data.RawData, sliceMap[idx]); err != nil {
			fmt.Printf("[ERROR] ParseFloatSlice: %v\n", err)
		}
	}

	data.Stats = stats
	c <- data
}

// ParseDisks takes in a list of Disks and accumulates the smartctl info where possible for each entry
func (s *SmartCtl) ParseDisks() (err error) {
	c := make(chan Disk, len(s.Disks))
	e := make(chan DiskFail, len(s.Disks))
	var a int

	for _, each := range s.Disks {
		go s.parseDisks(each, c, e)
	}

	for {
		if a == len(s.Disks) {
			break
		}

		select {
		case data := <-c:
			if len(data.Name) > 0 && data.Name != "empty" {
				s.DiskOutput[data.Name] = data
			}
			a++
		case err := <-e:
			s.DiskFailed[err.Name] = err.Error
			a++
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	return err
}

// splitDisks is a private helper function to parse out the disks we care about
func (s *SmartCtl) splitDisks(out string) (disks []string) {
	for _, each := range strings.Split(out, "\n") {
		if len(each) > 0 {
			disks = append(disks, strings.Split(each, " #")[0])
		}
	}
	return disks
}

func (s *SmartCtl) gatherDisks() (err error) {
	var out []byte

	cmd := []string{s.CtlPath, "--scan"}

	if out, err = exec.Command(s.SudoPath, cmd...).CombinedOutput(); err != nil {
		return fmt.Errorf("[ERROR] Could not gather disks from smartctl --scan: %v\n", err)
	}

	s.Disks = s.splitDisks(string(out))

	return nil
}

// ExcludeDisks is a private function to reduce the set of disks to query against
func (s *SmartCtl) ExcludeDisks() (disks []string) {
	elems := make(map[string]bool)

	for _, each := range s.Disks {
		elems[each] = false
	}

	for _, each := range s.Exclude {
		if _, ok := elems[each]; ok {
			delete(elems, each)
		}
	}

	for key := range elems {
		disks = append(disks, key)
	}

	return disks
}

// initStruct is a private function to confirm we have smartctl reqs installed/accessible
func (s *SmartCtl) initStruct() (err error) {
	if s.SudoPath, err = exec.LookPath("sudo"); err != nil {
		s.Init = false
		return fmt.Errorf("could not pull path for 'sudo': %v\n", err)
	}

	if s.CtlPath, err = exec.LookPath("smartctl"); err != nil {
		s.Init = false
		return fmt.Errorf("could not pull path for 'smartctl': %v\n", err)
	}

	// NOTE: if we specify the Include list in the config, this will skip the smartctl --scan
	if len(s.Include) > 0 {
		s.Disks = s.splitDisks(strings.Join(s.Include, "\n"))
	} else if err = s.gatherDisks(); err != nil {
		return err
	}

	if len(s.Exclude) > 0 {
		s.Disks = s.ExcludeDisks()
	}

	s.DiskOutput = make(map[string]Disk, len(s.Disks))
	s.DiskFailed = make(map[string]error)

	s.Init = true

	return nil
}

// init adds the smartctl collector as an input to telegraf
func init() {
	inputs.Add("smartctl", func() telegraf.Input { return &SmartCtl{} })
}

// Gather is the primary function to collect smartctl data
func (s *SmartCtl) Gather(acc telegraf.Accumulator) (err error) {
	var health float64

	if !s.Init {
		if err = s.initStruct(); err != nil {
			return fmt.Errorf("could not initialize smartctl plugin: %v\n", err)
		}
	}

	// actually gather the stats
	if err = s.ParseDisks(); err != nil {
		return fmt.Errorf("could not parse all the disks in our list: %v\n", err)
	}

	for _, each := range s.DiskOutput {
		tags := map[string]string{
			"name":       each.Name,
			"vendor":     each.Vendor,
			"product":    each.Product,
			"block_size": each.Block,
			"serial":     each.Serial,
			"rpm":        each.Rotation,
			"transport":  each.Transport,
			"read_cache": each.ReadCache,
			"writeback":  each.Writeback,
		}

		if each.Health == "OK" {
			health = 1.0
		} else {
			health = 0.0
		}

		fields := make(map[string]interface{})
		fields["health"] = health
		fields["current_temp"] = each.Stats.CurrentTemp
		fields["max_temp"] = each.Stats.MaxTemp

		// add the read error row
		if len(each.Stats.ReadError) == 7 {
			fields["ecc_corr_fast_read"] = each.Stats.ReadError[0]
			fields["ecc_corr_delay_read"] = each.Stats.ReadError[1]
			fields["ecc_reread"] = each.Stats.ReadError[2]
			fields["total_err_corr_read"] = each.Stats.ReadError[3]
			fields["corr_algo_read"] = each.Stats.ReadError[4]
			fields["data_read"] = each.Stats.ReadError[5]
			fields["uncorr_err_read"] = each.Stats.ReadError[6]
		}

		// add the write error row
		if len(each.Stats.WriteError) == 7 {
			fields["ecc_corr_fast_write"] = each.Stats.WriteError[0]
			fields["ecc_corr_delay_write"] = each.Stats.WriteError[1]
			fields["ecc_rewrite"] = each.Stats.WriteError[2]
			fields["total_err_corr_write"] = each.Stats.WriteError[3]
			fields["corr_algo_write"] = each.Stats.WriteError[4]
			fields["data_write"] = each.Stats.WriteError[5]
			fields["uncorr_err_write"] = each.Stats.WriteError[6]
		}

		// add the verify error row
		if len(each.Stats.VerifyError) == 7 {
			fields["ecc_corr_fast_verify"] = each.Stats.VerifyError[0]
			fields["ecc_corr_delay_verify"] = each.Stats.VerifyError[1]
			fields["ecc_reverify"] = each.Stats.VerifyError[2]
			fields["total_err_corr_verify"] = each.Stats.VerifyError[3]
			fields["corr_algo_verify"] = each.Stats.VerifyError[4]
			fields["data_verify"] = each.Stats.VerifyError[5]
			fields["uncorr_err_verify"] = each.Stats.VerifyError[6]
		}

		acc.AddFields("smartctl", fields, tags)
	}

	return nil
}
