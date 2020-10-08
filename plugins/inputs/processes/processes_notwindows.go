// +build !windows

package processes

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/linux_sysctl_fs"
)

type Processes struct {
	execPS       func() ([]byte, error)
	readProcFile func(filename string) ([]byte, error)

	Log telegraf.Logger

	forcePS   bool
	forceProc bool
}

func (p *Processes) Gather(acc telegraf.Accumulator) error {
	// Get an empty map of metric fields
	fields := getEmptyFields()

	// Decide if we will use 'ps' to get stats (use procfs otherwise)
	usePS := true
	if runtime.GOOS == "linux" {
		usePS = false
	}
	if p.forcePS {
		usePS = true
	} else if p.forceProc {
		usePS = false
	}

	// Gather stats from 'ps' or procfs
	if usePS {
		if err := p.gatherFromPS(fields); err != nil {
			return err
		}
	} else {
		if err := p.gatherFromProc(fields); err != nil {
			return err
		}
	}

	acc.AddGauge("processes", fields, nil)
	return nil
}

// Gets empty fields of metrics based on the OS
func getEmptyFields() map[string]interface{} {
	fields := map[string]interface{}{
		"blocked":  int64(0),
		"zombies":  int64(0),
		"stopped":  int64(0),
		"running":  int64(0),
		"sleeping": int64(0),
		"total":    int64(0),
		"unknown":  int64(0),
	}
	switch runtime.GOOS {
	case "freebsd":
		fields["idle"] = int64(0)
		fields["wait"] = int64(0)
	case "darwin":
		fields["idle"] = int64(0)
	case "openbsd":
		fields["idle"] = int64(0)
	case "linux":
		fields["dead"] = int64(0)
		fields["paging"] = int64(0)
		fields["total_threads"] = int64(0)
		fields["idle"] = int64(0)
	}
	return fields
}

// exec `ps` to get all process states
func (p *Processes) gatherFromPS(fields map[string]interface{}) error {
	out, err := p.execPS()
	if err != nil {
		return err
	}

	for i, status := range bytes.Fields(out) {
		if i == 0 && string(status) == "STAT" {
			// This is a header, skip it
			continue
		}
		switch status[0] {
		case 'W':
			fields["wait"] = fields["wait"].(int64) + int64(1)
		case 'U', 'D', 'L':
			// Also known as uninterruptible sleep or disk sleep
			fields["blocked"] = fields["blocked"].(int64) + int64(1)
		case 'Z':
			fields["zombies"] = fields["zombies"].(int64) + int64(1)
		case 'X':
			fields["dead"] = fields["dead"].(int64) + int64(1)
		case 'T':
			fields["stopped"] = fields["stopped"].(int64) + int64(1)
		case 'R':
			fields["running"] = fields["running"].(int64) + int64(1)
		case 'S':
			fields["sleeping"] = fields["sleeping"].(int64) + int64(1)
		case 'I':
			fields["idle"] = fields["idle"].(int64) + int64(1)
		case '?':
			fields["unknown"] = fields["unknown"].(int64) + int64(1)
		default:
			p.Log.Infof("Unknown state %q from ps", string(status[0]))
		}
		fields["total"] = fields["total"].(int64) + int64(1)
	}
	return nil
}

// get process states from /proc/(pid)/stat files
func (p *Processes) gatherFromProc(fields map[string]interface{}) error {
	filenames, err := filepath.Glob(linux_sysctl_fs.GetHostProc() + "/[0-9]*/stat")

	if err != nil {
		return err
	}

	for _, filename := range filenames {
		_, err := os.Stat(filename)
		if err != nil {
			return err
		}
		data, err := p.readProcFile(filename)
		if err != nil {
			return err
		}
		if data == nil {
			continue
		}

		// Parse out data after (<cmd name>)
		i := bytes.LastIndex(data, []byte(")"))
		if i == -1 {
			continue
		}
		data = data[i+2:]

		stats := bytes.Fields(data)
		if len(stats) < 3 {
			return fmt.Errorf("Something is terribly wrong with %s", filename)
		}
		switch stats[0][0] {
		case 'R':
			fields["running"] = fields["running"].(int64) + int64(1)
		case 'S':
			fields["sleeping"] = fields["sleeping"].(int64) + int64(1)
		case 'D':
			fields["blocked"] = fields["blocked"].(int64) + int64(1)
		case 'Z':
			fields["zombies"] = fields["zombies"].(int64) + int64(1)
		case 'X':
			fields["dead"] = fields["dead"].(int64) + int64(1)
		case 'T', 't':
			fields["stopped"] = fields["stopped"].(int64) + int64(1)
		case 'W':
			fields["paging"] = fields["paging"].(int64) + int64(1)
		case 'I':
			fields["idle"] = fields["idle"].(int64) + int64(1)
		case 'P':
			if _, ok := fields["parked"]; ok {
				fields["parked"] = fields["parked"].(int64) + int64(1)
			}
			fields["parked"] = int64(1)
		default:
			p.Log.Infof("Unknown state %q in file %q", string(stats[0][0]), filename)
		}
		fields["total"] = fields["total"].(int64) + int64(1)

		threads, err := strconv.Atoi(string(stats[17]))
		if err != nil {
			p.Log.Infof("Error parsing thread count: %s", err.Error())
			continue
		}
		fields["total_threads"] = fields["total_threads"].(int64) + int64(threads)
	}
	return nil
}

func readProcFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		// Reading from /proc/<PID> fails with ESRCH if the process has
		// been terminated between open() and read().
		if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.ESRCH {
			return nil, nil
		}

		return nil, err
	}

	return data, nil
}

func execPS() ([]byte, error) {
	bin, err := exec.LookPath("ps")
	if err != nil {
		return nil, err
	}

	out, err := exec.Command(bin, "axo", "state").Output()
	if err != nil {
		return nil, err
	}

	return out, err
}

func init() {
	inputs.Add("processes", func() telegraf.Input {
		return &Processes{
			execPS:       execPS,
			readProcFile: readProcFile,
		}
	})
}
