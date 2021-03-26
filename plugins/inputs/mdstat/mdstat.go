package mdstat

import (
	"bufio"
	//"fmt"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// #############################################################################
// Telegraf plugin stuff
// #############################################################################
type MDSTAT_PLUGIN struct {
	Mdstat_File string `toml:"mdstat_file"`
}

var sampleConfig = `
  # The location of the mdstat file to read.
  # mdstat_file = /proc/mdstat
`

func (s *MDSTAT_PLUGIN) SampleConfig() string {
	return sampleConfig
}

func (s *MDSTAT_PLUGIN) Description() string {
	return "A plugin to read metrics about mdadm managed RAID arrays."
}

func (s *MDSTAT_PLUGIN) Gather(acc telegraf.Accumulator) error {
	var MDSTAT_FILE string
	if s.Mdstat_File == "" {
		MDSTAT_FILE = "/proc/mdstat"
	} else {
		MDSTAT_FILE = s.Mdstat_File
	}

	// Lets open the file.
	f, err := os.Open(MDSTAT_FILE)
	if err != nil {
		return err
	}

	// Make sure we schedule the clean up.
	closingFunc := func() error {
		if err = f.Close(); err != nil {
			return err
		}
		return nil
	}
	defer closingFunc()

	result, err := parseFile(f)
	if err != nil {
		return err
	}

	for _, device := range result.devices {
		devTags := map[string]string{
			"device": device.name,
		}

		devFields := map[string]interface{}{
			"status":          device.status,
			"raidType":        device.raidType,
			"minDisks":        device.minDisks,
			"currDisks":       device.currDisks,
			"missingDisks":    device.missingDisks,
			"failedDisks":     device.failedDisks,
			"inRecovery":      device.inRecovery,
			"recoveryPercent": device.recoveryPercent,
		}

		// Add raid array stats per raid device
		acc.AddGauge("mdstat_device", devFields, devTags, time.Now())

		//Now add status for each disk in the array
		for _, disk := range device.diskList {
			diskTags := map[string]string{
				"device": device.name,
				"disk":   disk.name,
			}

			diskFields := map[string]interface{}{
				"role":   disk.role,
				"failed": disk.failed,
			}

			acc.AddGauge("mdstat_disk", diskFields, diskTags, time.Now())
		}

	}

	return nil
}

func init() {
	inputs.Add("mdstat", func() telegraf.Input { return &MDSTAT_PLUGIN{} })
}

// #############################################################################
// mdstat parsing stuff
// #############################################################################
type Personalities []string

type Disk struct {
	name   string
	role   int
	failed bool
}

type Device struct {
	name            string
	status          string
	raidType        string
	diskList        []Disk
	minDisks        int
	currDisks       int
	missingDisks    int
	failedDisks     int
	inRecovery      bool
	recoveryPercent float32
}

type MDSTAT struct {
	personalities Personalities
	devices       []Device
}

const PERSONALITY_PREFIX = "Personalities : "
const UNUSED_PREFIX = "unused"
const RECOVERY_STRING = "recovery"

func parseFile(r io.Reader) (MDSTAT, error) {
	var err error
	// This is a text file, so we can scan it line by line.
	// We will break the file up into it's parts so each can be parsed individually
	s := bufio.NewScanner(r)
	var parsedMap MDSTAT
	var deviceEntry []string
	for s.Scan() {
		line := s.Text()

		// If the line is a personality line
		if strings.HasPrefix(line, PERSONALITY_PREFIX) {
			parsedMap.personalities, err = parsePersonalities(line)
			if err != nil {
				return parsedMap, err
			}
			continue
		}

		// If there's an unused line.
		if strings.HasPrefix(line, UNUSED_PREFIX) {
			// Right now we don't use the "Unused" line
			continue
		}

		if strings.Compare("", line) == 0 && len(deviceEntry) > 0 {

			parsedDev, err := parseDeviceEntry(deviceEntry)
			if err != nil {
				return parsedMap, err
			}
			parsedMap.devices = append(parsedMap.devices, parsedDev)

			// Reset the device entry so it's ready for a new one.
			//Maybe change the if statement so it only happens on an empty line.
			deviceEntry = nil
			continue
		}

		// If it's an empty line, don't append it to the device entry.
		if strings.Compare("", line) != 0 {
			deviceEntry = append(deviceEntry, line)
		}
	}

	scanErr := s.Err()
	if scanErr != nil {
		err = scanErr
	}

	return parsedMap, err
}

func parsePersonalities(personalitiesLine string) (Personalities, error) {
	var result = strings.Fields(personalitiesLine)
	if len(result) <= 2 {
		err := errors.New("Personalities line not long enough")
		return nil, err
	}
	return result[2:], nil
}

func parseDeviceEntry(deviceEntry []string) (Device, error) {
	var parsedDevice Device
	var err error
	// The first line should be the device line.
	deviceLineFields := strings.Fields(deviceEntry[0])
	if len(deviceLineFields) <= 5 {
		err = errors.New("Not enough lines in the device field")
		return parsedDevice, err
	}

	//Get name and status from the md device line
	parsedDevice.name = deviceLineFields[0]
	parsedDevice.status = deviceLineFields[2]
	parsedDevice.raidType = deviceLineFields[3]

	// For each disk, parse it's information
	for _, disk := range deviceLineFields[4:] {
		var DISK_REGEX = "(?P<diskname>[a-zA-Z0-9]+)\\[(?P<diskrole>[0-9]+)\\](?:\\((?P<failedstatus>F)\\))?"
		re := regexp.MustCompile(DISK_REGEX)
		captures := re.FindStringSubmatch(disk)
		if captures == nil {
			err = errors.New("Malformed disk entry")
			return parsedDevice, err
		}
		var parsedDisk Disk
		// Capture groups start at 1 because index 0 is the full string
		parsedDisk.name = captures[1]
		parsedDisk.role, err = strconv.Atoi(captures[2])
		if err != nil {
			return parsedDevice, err
		}

		if captures[3] == "F" {
			parsedDevice.failedDisks++
			parsedDisk.failed = true
		} else {
			parsedDisk.failed = false
		}
		// Once we have the info for the disk, add it to the list.
		parsedDevice.diskList = append(parsedDevice.diskList, parsedDisk)
	}

	// Now for the config line
	CONFIG_LINE_REGEX := ".* \\[(?P<ndisk>[0-9]+)/(?P<mdisks>[0-9]+)\\] \\[(?P<arraystat>[U_]+)\\]"
	re := regexp.MustCompile(CONFIG_LINE_REGEX)
	captures := re.FindStringSubmatch(deviceEntry[1])
	if captures == nil {
		err = errors.New("Malformed device config line")
		return parsedDevice, err
	}
	parsedDevice.minDisks, err = strconv.Atoi(captures[1])
	if err != nil {
		return parsedDevice, err
	}
	parsedDevice.currDisks, err = strconv.Atoi(captures[2])
	if err != nil {
		return parsedDevice, err
	}

	//Since we already know the number of active disks, we don't need to count the U's
	// in the [UUU_U] field. Instead we only need to count the number of _'s that appear
	// since they represent the number of inactive disks. Subtracting the failed disks from
	// The number of _'s represents the number of missing disks.
	parsedDevice.missingDisks = strings.Count(captures[3], "_") - parsedDevice.failedDisks

	//Some device entries have a third line. We only check for the recovery and not the bitmap
	if len(deviceEntry) >= 3 {
		// Lets check for a recovery line.
		if strings.Contains(deviceEntry[2], RECOVERY_STRING) {
			RECOVERY_LINE_REGEX := "recovery = (?P<recoveryPercent>[0-9]+\\.[0-9]+)%"
			re := regexp.MustCompile(RECOVERY_LINE_REGEX)
			captures := re.FindStringSubmatch(deviceEntry[2])
			if captures != nil {
				parsedDevice.inRecovery = true
				value, err := strconv.ParseFloat(captures[1], 32)
				if err != nil {
					return parsedDevice, err
				}
				parsedDevice.recoveryPercent = float32(value)
			}
		}
	}

	return parsedDevice, nil
}
