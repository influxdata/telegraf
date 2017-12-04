package postfix

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## Postfix queue directory. If not provided, telegraf will try to use
  ## 'postconf -h queue_directory' to determine it.
  # queue_directory = "/var/spool/postfix"
`

const description = "Measure postfix queue statistics"

func getQueueDirectory() (string, error) {
	qd, err := exec.Command("postconf", "-h", "queue_directory").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(qd)), nil
}

func qScan(path string) (int64, int64, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, 0, err
	}

	finfos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return 0, 0, 0, err
	}

	var length, size int64
	var oldest time.Time
	for _, finfo := range finfos {
		length++
		size += finfo.Size()

		ctime := statCTime(finfo.Sys())
		if ctime.IsZero() {
			continue
		}
		if oldest.IsZero() || ctime.Before(oldest) {
			oldest = ctime
		}
	}
	var age int64
	if !oldest.IsZero() {
		age = int64(time.Now().Sub(oldest) / time.Second)
	} else if len(finfos) != 0 {
		// system doesn't support ctime
		age = -1
	}
	return length, size, age, nil
}

type Postfix struct {
	QueueDirectory string
}

func (p *Postfix) Gather(acc telegraf.Accumulator) error {
	if p.QueueDirectory == "" {
		var err error
		p.QueueDirectory, err = getQueueDirectory()
		if err != nil {
			return fmt.Errorf("unable to determine queue directory: %s", err)
		}
	}

	for _, q := range []string{"active", "hold", "incoming", "maildrop"} {
		length, size, age, err := qScan(path.Join(p.QueueDirectory, q))
		if err != nil {
			acc.AddError(fmt.Errorf("error scanning queue %s: %s", q, err))
			continue
		}
		fields := map[string]interface{}{"length": length, "size": size}
		if age != -1 {
			fields["age"] = age
		}
		acc.AddFields("postfix_queue", fields, map[string]string{"queue": q})
	}

	var dLength, dSize int64
	dAge := int64(-1)
	for _, q := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"} {
		length, size, age, err := qScan(path.Join(p.QueueDirectory, "deferred", q))
		if err != nil {
			if os.IsNotExist(err) {
				// the directories are created on first use
				continue
			}
			acc.AddError(fmt.Errorf("error scanning queue deferred/%s: %s", q, err))
			return nil
		}
		dLength += length
		dSize += size
		if age > dAge {
			dAge = age
		}
	}
	fields := map[string]interface{}{"length": dLength, "size": dSize}
	if dAge != -1 {
		fields["age"] = dAge
	}
	acc.AddFields("postfix_queue", fields, map[string]string{"queue": "deferred"})

	return nil
}

func (p *Postfix) SampleConfig() string {
	return sampleConfig
}

func (p *Postfix) Description() string {
	return description
}

func init() {
	inputs.Add("postfix", func() telegraf.Input {
		return &Postfix{
			QueueDirectory: "/var/spool/postfix",
		}
	})
}
