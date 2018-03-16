package maf

import (
	"regexp"
	"sync"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)
	
type MafProcess struct {
	KeyWords   []string `toml:"keywords"`
}

var sampleConfig = `
    interval = "60s"
    ## process number check.
    ## R: Running S: Sleep T: Stop I: Idle Z: Zombie W: Wait L: Lock
    # keywords = ["sandboxav"]
`

func (_ *MafProcess) SampleConfig() string {
	return sampleConfig
}

func (_ *MafProcess) Description() string {
	return "MAF: Check process number with specify keyword."
}

func (p *MafProcess) Gather(acc telegraf.Accumulator) error {
	if len(p.KeyWords) == 0 {
		p.KeyWords = []string{"sandboxav"}
	}

	var wg sync.WaitGroup

	for _, loop := range p.KeyWords {
		wg.Add(1)
		go func(keyWord string) {
			defer wg.Done()
			acc.AddError(p.gatherProcess(keyWord, acc))
		}(loop)
	}
	wg.Wait()
	return nil
}

func (_ *MafProcess) gatherProcess(keyWord string, acc telegraf.Accumulator) error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	var Name []string
	var Status []string
	var CmdLine []string
	Count := 0
	for _, proc := range procs {
		name, _ := proc.Name()
		status, _ := proc.Status()
		cmdline, _ := proc.Cmdline()
		if match, _ := regexp.MatchString(string(`.*` + keyWord + `.*`), cmdline); match {
			Name = append(Name, name + ";")
			Status = append(Status, status + ";")
			CmdLine = append(CmdLine, cmdline + ";")
			Count++
		}
	}
	acc.AddFields("maf_process",
		map[string]interface{}{
			"name":         Name,
			"cmdline":      CmdLine,
			"status":       Status,
			"number":       Count,
		},
		map[string]string{
			"keyword":         keyWord,
		},
	)

	return nil
}

func init() {
	inputs.Add("maf_process", func() telegraf.Input {
		return &MafProcess{}
	})
}


