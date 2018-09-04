package mem

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVirtual_memory(t *testing.T) {
	if runtime.GOOS == "solaris" {
		t.Skip("Only .Total is supported on Solaris")
	}

	v, err := VirtualMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := &VirtualMemoryStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}

	assert.True(t, v.Total > 0)
	assert.True(t, v.Available > 0)
	assert.True(t, v.Used > 0)

	total := v.Used + v.Free + v.Buffers + v.Cached
	totalStr := "used + free + buffers + cached"
	if runtime.GOOS == "windows" {
		total = v.Used + v.Available
		totalStr = "used + available"
	}
	assert.Equal(t, v.Total, total,
		"Total should be computable (%v): %v", totalStr, v)

	assert.True(t, runtime.GOOS == "windows" || v.Free > 0)
	assert.True(t, v.Available > v.Free,
		"Free should be a subset of Available: %v", v)

	inDelta := assert.InDelta
	if runtime.GOOS == "windows" {
		inDelta = assert.InEpsilon
	}
	inDelta(t, v.UsedPercent,
		100*float64(v.Used)/float64(v.Total), 0.1,
		"UsedPercent should be how many percent of Total is Used: %v", v)
}

func TestSwap_memory(t *testing.T) {
	v, err := SwapMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := &SwapMemoryStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}
}

func TestVirtualMemoryStat_String(t *testing.T) {
	v := VirtualMemoryStat{
		Total:       10,
		Available:   20,
		Used:        30,
		UsedPercent: 30.1,
		Free:        40,
	}
	e := `{"total":10,"available":20,"used":30,"usedPercent":30.1,"free":40,"active":0,"inactive":0,"wired":0,"buffers":0,"cached":0,"writeback":0,"dirty":0,"writebacktmp":0,"shared":0,"slab":0,"pagetables":0,"swapcached":0,"commitlimit":0,"committedas":0,"hightotal":0,"highfree":0,"lowtotal":0,"lowfree":0,"swaptotal":0,"swapfree":0,"mapped":0,"vmalloctotal":0,"vmallocused":0,"vmallocchunk":0,"hugepagestotal":0,"hugepagesfree":0,"hugepagesize":0}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("VirtualMemoryStat string is invalid: %v", v)
	}
}

func TestSwapMemoryStat_String(t *testing.T) {
	v := SwapMemoryStat{
		Total:       10,
		Used:        30,
		Free:        40,
		UsedPercent: 30.1,
	}
	e := `{"total":10,"used":30,"free":40,"usedPercent":30.1,"sin":0,"sout":0}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("SwapMemoryStat string is invalid: %v", v)
	}
}
