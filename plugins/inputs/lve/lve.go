// +build linux

package lve

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Lve struct {
	lveListFile string
	MinUid      int `toml:"min_uid"`
	MaxUid      int `toml:"max_uid"`
}

func (k *Lve) Description() string {
	return "Get LVE usage statistics and limits from /proc/lve/list"
}

const sampleConfig = `
  ## Minimum LVE/User UID to emit
  # min_uid = 500
  ## Maximum LVE/User UID to emit
  # max_uid = 1000000000
`

func (k *Lve) SampleConfig() string {
	return sampleConfig
}

func (k *Lve) Gather(acc telegraf.Accumulator) error {
	if _, err := os.Stat(k.lveListFile); os.IsNotExist(err) {
		return fmt.Errorf("lve: %s does not exist!", k.lveListFile)
	} else if err != nil {
		return err
	}

	file, err := os.Open(k.lveListFile)
	if err != nil {
		return err
	}
	r := csv.NewReader(file)
	r.Comma = '\t'
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return nil
	}

	if len(records[0]) < 2 {
		return fmt.Errorf("lve: %s has unrecognized format, not enough columns", k.lveListFile)
	}

	// first row is header
	for _, columns := range records[1:] {
		// first column is like: 0,0
		col1 := strings.SplitN(columns[0], ",", 2)
		if len(col1) != 2 {
			continue
		}

		int_uid, err := strconv.Atoi(col1[1])
		if int_uid < k.MinUid || int_uid > k.MaxUid {
			continue
		}

		user, err := user.LookupId(col1[1])
		if err != nil {
			return fmt.Errorf("lve: %s has unrecognized format, not enough columns", k.lveListFile)
			//continue
		}

		fields := make(map[string]interface{})

		for i, column := range columns[1:] {
			// Convert the stat value into an integer.
			v, err := strconv.ParseInt(column, 10, 64)
			if err != nil {
				return err
			}
			fields[string(records[0][i+1])] = v
		}

		acc.AddFields("lve", fields, map[string]string{"username": user.Username})
	}
	return nil
}

func init() {
	inputs.Add("lve", func() telegraf.Input {
		return &Lve{
			lveListFile: "/proc/lve/list",
			MinUid:      500,
			MaxUid:      1000000000,
		}
	})
}
