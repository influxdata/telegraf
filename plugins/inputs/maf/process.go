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
    # keywords = ["sandboxmain.py", "reportingmain.py", "sbox_cc_watchdog"]
`

func (_ *MafProcess) SampleConfig() string {
	return sampleConfig
}

func (_ *MafProcess) Description() string {
	return "MAF: Check process number with specify keyword."
}

func (p *MafProcess) Gather(acc telegraf.Accumulator) error {
	if len(p.KeyWords) == 0 {
		p.KeyWords = []string{"sandboxmain.py"}
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

	var CmdLine string
	var Status string
	var CreateTime int64
	var Count int64
	for _, proc := range procs {
		status, _ := proc.Status()
		cmdline, _ := proc.Cmdline()
		createtime, _ := proc.CreateTime()
		if match, _ := regexp.MatchString(string(`.*` + keyWord + `.*`), cmdline); match {
			CmdLine = cmdline
			Status = status
			CreateTime = createtime
			Count++
			break
		}
	}
	acc.AddFields("maf_process",
		map[string]interface{}{
			"cmdline":      CmdLine,
			"status":       Status,
			"createtime":   CreateTime,
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


