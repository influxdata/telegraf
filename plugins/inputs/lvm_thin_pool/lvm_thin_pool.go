package lvm_thin_pool

import (
	"os/exec"
	"fmt"
	"strings"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// LvmThinPool is a telegraf plugin to gather information about lvm thin pools
type LvmThinPool struct {
	UseSudo bool
	Path string
}

const measurement = "lvm_thin_pool"

// SampleConfig returns sample configuration options.
func (p *LvmThinPool) SampleConfig() string {
	return `
  ## Adjust your sudo settings appropriately if using this option
  use_sudo = false
  # set path to the thin pool and use it as tag
  path = my_volume_group/my_thin_pool
`
}

// Description returns a short description of the plugin
func (p *LvmThinPool) Description() string {
	return "Gather lvm thin pool size, used percentage and thin count by parsing 'lvdisplay --columns' output."
}

func (p *LvmThinPool) Gather(acc telegraf.Accumulator) error {
	// check if lvdisplay is available
	lvdisplayPath, err := exec.LookPath("lvdisplay")
	if err != nil {
		acc.AddError(err)
	}

	// compose command and arguments slice
	var cmdName string
	var args []string
	if p.UseSudo {
		cmdName = "sudo"
		args = append(args, lvdisplayPath)
	} else {
		cmdName = lvdisplayPath
	}

	var lv_attrs string = "lv_size,lv_metadata_size,data_percent,metadata_percent,thin_count"
	var lvdisplay_args = []string{"-C", "-o", lv_attrs, "--units", "m", "--separator", "':'", "--noheadings", p.Path}
	args = append(args, lvdisplay_args...)
		
	// execute lvdisplay
	cmd := exec.Command(cmdName, args...)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	data := strings.Split(string(out), ":")

	// extract values
	var dataParsed [4] float64
	var thinCount uint
	for i, d := range data {
		if i < 4 {
			d = strings.TrimPrefix(d, "  ")
			d = strings.TrimSuffix(d, "m")
			dp, err := strconv.ParseFloat(d, 64)
			dataParsed[i] = dp
			if err != nil {
				acc.AddError(err)
			}
		} else {
			d = strings.TrimSuffix(d, "\n")
			dp, err := strconv.ParseUint(d, 10, 64)
			thinCount = dp
			if err != nil {
				acc.AddError(err)
			}
		}
	}
	// lvSize, err := strconv.ParseFloat(strings.TrimSuffix(data[0], "m"), 64)
	// if err != nil {
	// 	acc.AddError(err)
	// }
	// lvMetadata, err := strconv.ParseFloat(strings.TrimSuffix(data[1], "m"), 64)
	// if err != nil {
	// 	acc.AddError(err)
	// }
	// usedData, err := strconv.ParseFloat(data[2], 64)
	// if err != nil {
	// 	acc.AddError(err)
	// }
	// usedMetadata, err := strconv.ParseFloat(data[3], 64)
	// if err != nil {
	// 	acc.AddError(err)
	// }
	// thinCount, err := strconv.ParseUint(data[4], 10, 64)
	// if err != nil {
	// 	acc.AddError(err)
	// }

	fields := map[string]interface{}{
		"lv_size": dataParsed[0],
		"lv_metadata": dataParsed[1],
		"data_percent": dataParsed[2],
		"metadata_percent": dataParsed[3],
		"thin_count": thinCount,
	}
	tags := map[string]string{
		"path":  p.Path,
	}
	acc.AddCounter(measurement, fields, tags)

	return nil
}

func init() {
	inputs.Add("lvm_thin_pool", func() telegraf.Input {
		return &LvmThinPool{}
	})
}
