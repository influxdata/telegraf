// +build linux

package entropy

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const inputName = "entropy"

type Entropy struct {
	Proc string
}

var dfltProc = "/proc/sys/kernel/random/entropy_avail"

func (e *Entropy) Description() string {
	return fmt.Sprintf("Reads the available entropy from %s", dfltProc)
}

var sampleConfig string = fmt.Sprintf(`
  ## (Optional) Override the file from which to collect entropy stats. Default is:
  proc = "%s"
`, dfltProc)

func (e *Entropy) SampleConfig() string {
	return sampleConfig
}

func (e *Entropy) Gather(acc telegraf.Accumulator) error {
	ent := 0
	proc := e.Proc

	if proc == "" {
		proc = dfltProc
	}

	if _, err := os.Stat(proc); err != nil {
		return fmt.Errorf("could not stat proc file '%s': %v", proc, err)
	}

	content, err := ioutil.ReadFile(proc)
	if err != nil {
		return fmt.Errorf("failed to read proc file '%s': %v", proc, err)
	}

	ent, err = strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return fmt.Errorf("expected integer content but found %s: %v", content, err)
	}
	acc.AddFields(inputName, map[string]interface{}{"available": ent}, nil)
	return nil
}

func init() {
	inputs.Add("entropy", func() telegraf.Input { return &Entropy{} })
}
