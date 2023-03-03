//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package slab

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type SlabStats struct {
	Log telegraf.Logger `toml:"-"`

	statFile string
	useSudo  bool
}

func (*SlabStats) SampleConfig() string {
	return sampleConfig
}

func (ss *SlabStats) Init() error {
	return nil
}

func (ss *SlabStats) Gather(acc telegraf.Accumulator) error {
	fields, err := ss.getSlabStats()
	if err != nil {
		return err
	}

	acc.AddGauge("slab", fields, nil)
	return nil
}

func (ss *SlabStats) getSlabStats() (map[string]interface{}, error) {
	fields := map[string]interface{}{}

	out, err := ss.runCmd("/bin/cat", []string{ss.statFile})
	if err != nil {
		return nil, err
	}

	bytesReader := bytes.NewReader(out)
	scanner := bufio.NewScanner(bytesReader)

	// Read header rows
	scanner.Scan() // for "slabinfo - version: 2.1"
	scanner.Scan() // for "# name <active_objs> <num_objs> <objsize> ..."

	// Read data rows
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Fields(line)

		if len(cols) < 4 {
			return nil, errors.New("the content of /proc/slabinfo is invalid")
		}

		var numObj, sizObj int

		numObj, err = strconv.Atoi(cols[2])
		if err != nil {
			return nil, err
		}

		sizObj, err = strconv.Atoi(cols[3])
		if err != nil {
			return nil, err
		}

		fields[normalizeName(cols[0])] = numObj * sizObj
	}
	return fields, nil
}

func (ss *SlabStats) runCmd(cmd string, args []string) ([]byte, error) {
	execCmd := exec.Command(cmd, args...)
	if os.Geteuid() != 0 && ss.useSudo {
		execCmd = exec.Command("sudo", append([]string{"-n", cmd}, args...)...)
	}

	out, err := internal.StdOutputTimeout(execCmd, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to run command %q: %w - %v", execCmd.Args, err, out)
	}

	return out, nil
}

func getHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

func normalizeName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "-", "_") + "_size"
}

func init() {
	inputs.Add("slab", func() telegraf.Input {
		return &SlabStats{
			statFile: path.Join(getHostProc(), "slabinfo"),
			useSudo:  true,
		}
	})
}
