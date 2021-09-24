//go:build linux
// +build linux

package dmesg

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDmesg(t *testing.T) {
	k := DmesgConf{
		Binary:  "/bin/cat",
		Options: []string{"dmesg-sample"},
		Filters: []regStrMap{
			{Filter: ".*hostname.*", Field: "hostname"},
			{Filter: ".*oom_reaper.*|.*Out of memory.*", Field: "oom.count"},
		},
	}
	fmt.Sprintf("Options: %v", k)
	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.NoError(t, err)
	fmt.Sprintf("DMESG OUTPUT: %v", t)

	fields := map[string]interface{}{
		"hostname":  int(1),
		"oom.count": int(0),
	}

	acc.AssertContainsFields(t, "dmesg", fields)
}
