// +build linux

package mdadm

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	statusLineRE      = regexp.MustCompile(`(\d+) blocks .*\[(\d+)/(\d+)\] \[[U_]+\]`)
	recoveryLineBlocksRE = regexp.MustCompile(`\((\d+)/\d+\)`)
	recoveryLinePctRE = regexp.MustCompile(`= (.+)%`)
	recoveryLineFinishRE = regexp.MustCompile(`finish=(.+)\w`)
	recoveryLineSpeedRE = regexp.MustCompile(`speed=(.+)[A-Z]`)
	componentDeviceRE = regexp.MustCompile(`(.*)\[\d+\]`)
)

type mdadmStat struct {
	statFile string
}

func (k *mdadmStat) Description() string {
	return "Get md array statistics from /proc/mdstat"
}

var mdSampleConfig = `
	## No configuration required for this collector
`

func (k *mdadmStat) SampleConfig() string {
	return mdSampleConfig
}

func evalStatusLine(deviceLine, statusLine string) (active, total, size int64, err error) {

	sizeStr := strings.Fields(statusLine)[0]
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("unexpected statusLine %q: %w", statusLine, err)
	}

	if strings.Contains(deviceLine, "raid0") || strings.Contains(deviceLine, "linear") {
		// In the device deviceLine, only disks have a number associated with them in [].
		total = int64(strings.Count(deviceLine, "["))
		return total, total, size, nil
	}

	if strings.Contains(deviceLine, "inactive") {
		return 0, 0, size, nil
	}

	matches := statusLineRE.FindStringSubmatch(statusLine)
	if len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("couldn't find all the substring matches: %s", statusLine)
	}

	total, err = strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("unexpected statusLine %q: %w", statusLine, err)
	}

	active, err = strconv.ParseInt(matches[3], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("unexpected statusLine %q: %w", statusLine, err)
	}

	return active, total, size, nil
}

func evalRecoveryLine(recoveryLine string) (syncedBlocks int64, pct float64, finish float64, speed float64, err error) {
	// Get count of completed vs. total blocks
	matches := recoveryLineBlocksRE.FindStringSubmatch(recoveryLine)
	if len(matches) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("unexpected recoveryLine: %s", recoveryLine)
	}
	syncedBlocks, err = strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error parsing int from recoveryLine %q: %w", recoveryLine, err)
	}

	// Get percentage complete
	matches := recoveryLinePctRE.FindStringSubmatch(recoveryLine)
	if len(matches) != 1 {
		return 0, 0, 0, 0, fmt.Errorf("unexpected recoveryLine: %s", recoveryLine)
	}
	pct, err = strconv.ParseFloat(matches[0], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error parsing int from recoveryLine %q: %w", recoveryLine, err)
	}

	// Get time expected left to complete
	matches := recoveryLineFinishRE.FindStringSubmatch(recoveryLine)
	if len(matches) != 1 {
		return 0, 0, 0, 0, fmt.Errorf("unexpected recoveryLine: %s", recoveryLine)
	}
	finish, err = strconv.ParseFloat(matches[0], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error parsing int from recoveryLine %q: %w", recoveryLine, err)
	}

	// Get recovery speed
	matches := recoveryLineSpeedRE.FindStringSubmatch(recoveryLine)
	if len(matches) != 1 {
		return 0, 0, 0, 0, fmt.Errorf("unexpected recoveryLine: %s", recoveryLine)
	}
	speed, err = strconv.ParseFloat(matches[0], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error parsing int from recoveryLine %q: %w", recoveryLine, err)
	}

	return syncedBlocks, pct, finish, speed, nil
}

func evalComponentDevices(deviceFields []string) []string {
	mdComponentDevices := make([]string, 0)
	if len(deviceFields) > 3 {
		for _, field := range deviceFields[4:] {
			match := componentDeviceRE.FindStringSubmatch(field)
			if match == nil {
				continue
			}
			mdComponentDevices = append(mdComponentDevices, match[1])
		}
	}

	// Ensure no churn on tag ordering change
	return strings.Join(sort.Strings(mdComponentDevices), ",")
}

func (k *mdadmStat) Gather(acc telegraf.Accumulator) error {
	data, err := k.getProcMdstat()
	if err != nil {
		return err
	}
	lines = strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" || line[0] == ' ' ||
			strings.HasPrefix(line, "Personalities") ||
			strings.HasPrefix(line, "unused") {
			continue
		}
		deviceFields := strings.Fields(line)
		if len(deviceFields) < 3 {
			return nil, fmt.Errorf("not enough fields in mdline (expected at least 3): %s", line)
		}
		mdName := deviceFields[0] // mdx
		state := deviceFields[2]  // active or inactive

		if len(lines) <= i+3 {
			return nil, fmt.Errorf("error parsing %q: too few lines for md device", mdName)
		}

		// Failed disks have the suffix (F) & Spare disks have the suffix (S).
		fail := int64(strings.Count(line, "(F)"))
		spare := int64(strings.Count(line, "(S)"))

		active, total, size, err := evalStatusLine(lines[i], lines[i+1])
		if err != nil {
			return nil, fmt.Errorf("error parsing md device lines: %w", err)
		}

		syncLineIdx := i + 2
		if strings.Contains(lines[i+2], "bitmap") { // skip bitmap line
			syncLineIdx++
		}

		// If device is syncing at the moment, get the number of currently
		// synced bytes, otherwise that number equals the size of the device.
		syncedBlocks := size
		speed := float64(0)
		finish := float64(0)
		pct := float64(0)
		recovering := strings.Contains(lines[syncLineIdx], "recovery")
		resyncing := strings.Contains(lines[syncLineIdx], "resync")
		checking := strings.Contains(lines[syncLineIdx], "check")

		// Append recovery and resyncing state info.
		if recovering || resyncing || checking {
			if recovering {
				state = "recovering"
			} else if checking {
				state = "checking"
			} else {
				state = "resyncing"
			}

			// Handle case when resync=PENDING or resync=DELAYED.
			if strings.Contains(lines[syncLineIdx], "PENDING") ||
				strings.Contains(lines[syncLineIdx], "DELAYED") {
				syncedBlocks = 0
			} else {
				syncedBlocks, pct, finish, speed, err = evalRecoveryLine(lines[syncLineIdx])
				if err != nil {
					return nil, fmt.Errorf("error parsing sync line in md device %q: %w", mdName, err)
				}

			}
		}
		fields := map[string]interface{} {
			"DisksActive": active,
			"DisksFailed": fail,
			"DisksSpare": spare,
			"DisksTotal": total,
			"BlocksTotal": size,
			"BlocksSynced": syncedBlocks,
			"BlocksSyncedPct": pct,
			"BlocksSyncedFinishTime": finish,
			"BlocksSyncedSpeed": speed,
		}
		tags := map[string]string {
			"Name": mdName
			"ActivityState": state,
			"Devices": evalComponentDevices(deviceFields),
		}
		acc.AddFields("mdadm", fields, tags)
	}

	return nil
}

func (k *mdadmStat) getProcMdstat() ([]byte, error) {
	if _, err := os.Stat(k.statFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("mdadm: %s does not exist", k.statFile)
	} else if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(k.statFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func init() {
	inputs.Add("mdadm", func() telegraf.Input {
		return &mdadmStat{
			statFile: "/proc/mdstat",
		}
	})
}
