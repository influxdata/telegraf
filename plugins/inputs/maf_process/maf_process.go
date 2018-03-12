package maf_process

import (
	"regexp"
	"sync"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdata/maf"
	"github.com/influxdata/maf/plugins/inputs"
	"fmt"
)
	
type MafProcess struct {
	Name string
	CmdLine string
	Status string
	Count int

	KeyWords   []string `toml:"keywords"`
}

var sampleConfig = `
    ## process number check.
    ## R: Running S: Sleep T: Stop I: Idle Z: Zombie W: Wait L: Lock
    # keywords = ["sandboxav"]
`

func (_ *MafProcess) SampleConfig() string {
	return sampleConfig
}

func (_ *MafProcess) Description() string {
	return "MAF: Check process with specify keyword."
}

func (p *MafProcess) Gather(acc maf.Accumulator) error {
	if len(p.KeyWords) == 0 {
		p.KeyWords = []string{"sandboxav"}
	}

	procs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("Get pids error: %s", err)
	}

	var wg sync.WaitGroup

	for _, loop := range p.KeyWords {
		wg.Add(1)
		go func(keyWord string) {
			defer wg.Done()
			for _, proc := range procs {
				name, _ := proc.Name()
				status, _ := proc.Status()
				cmdline, _ := proc.Cmdline()
				//fmt.Println(name, status, cmdline)
				if match, _ := regexp.MatchString(string(`.*` + keyWord + `.*`), cmdline); match {
					p.Name = name
					p.Status = status
					p.CmdLine = cmdline
					p.Count++
				}
			}
			acc.AddFields("maf_process",
				map[string]interface{}{
					"name":         p.Name,
					"cmdline":      p.CmdLine,
					"status":       p.Status,
					"number":       p.Count,
				},
				map[string]string{
					"keyword":         keyWord,
				})
		}(loop)
	}
	wg.Wait()
	return nil
}

func init() {
	inputs.Add("maf_process", func() maf.Input {
		return &MafProcess{}
	})
}


