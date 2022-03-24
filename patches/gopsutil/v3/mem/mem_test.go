package mem

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/internal/common"
	"github.com/stretchr/testify/assert"
)

func skipIfNotImplementedErr(t *testing.T, err error) {
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
}

func TestVirtual_memory(t *testing.T) {
	if runtime.GOOS == "solaris" {
		t.Skip("Only .Total is supported on Solaris")
	}

	v, err := VirtualMemory()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := &VirtualMemoryStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}
	t.Log(v)

	assert.True(t, v.Total > 0)
	assert.True(t, v.Available > 0)
	assert.True(t, v.Used > 0)

	total := v.Used + v.Free + v.Buffers + v.Cached
	totalStr := "used + free + buffers + cached"
	switch runtime.GOOS {
	case "windows":
		total = v.Used + v.Available
		totalStr = "used + available"
	case "darwin", "openbsd":
		total = v.Used + v.Free + v.Cached + v.Inactive
		totalStr = "used + free + cached + inactive"
	case "freebsd":
		total = v.Used + v.Free + v.Cached + v.Inactive + v.Laundry
		totalStr = "used + free + cached + inactive + laundry"
	}
	assert.Equal(t, v.Total, total,
		"Total should be computable (%v): %v", totalStr, v)

	assert.True(t, runtime.GOOS == "windows" || v.Free > 0)
	assert.True(t, runtime.GOOS == "windows" || v.Available > v.Free,
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
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := &SwapMemoryStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}

	t.Log(v)
}

func TestVirtualMemoryStat_String(t *testing.T) {
	v := VirtualMemoryStat{
		Total:       10,
		Available:   20,
		Used:        30,
		UsedPercent: 30.1,
		Free:        40,
	}
	e := `{"total":10,"available":20,"used":30,"usedPercent":30.1,"free":40,"active":0,"inactive":0,"wired":0,"laundry":0,"buffers":0,"cached":0,"writeBack":0,"dirty":0,"writeBackTmp":0,"shared":0,"slab":0,"sreclaimable":0,"sunreclaim":0,"pageTables":0,"swapCached":0,"commitLimit":0,"committedAS":0,"highTotal":0,"highFree":0,"lowTotal":0,"lowFree":0,"swapTotal":0,"swapFree":0,"mapped":0,"vmallocTotal":0,"vmallocUsed":0,"vmallocChunk":0,"hugePagesTotal":0,"hugePagesFree":0,"hugePageSize":0}`
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
		Sin:         1,
		Sout:        2,
		PgIn:        3,
		PgOut:       4,
		PgFault:     5,
		PgMajFault:  6,
	}
	e := `{"total":10,"used":30,"free":40,"usedPercent":30.1,"sin":1,"sout":2,"pgIn":3,"pgOut":4,"pgFault":5,"pgMajFault":6}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("SwapMemoryStat string is invalid: %v", v)
	}
}

func TestSwapDevices(t *testing.T) {
	v, err := SwapDevices()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("error calling SwapDevices: %v", err)
	}

	t.Logf("SwapDevices() -> %+v", v)

	if len(v) == 0 {
		t.Fatalf("no swap devices found. [this is expected if the host has swap disabled]")
	}

	for _, device := range v {
		if device.Name == "" {
			t.Fatalf("deviceName not set in %+v", device)
		}
		if device.FreeBytes == 0 {
			t.Logf("[WARNING] free-bytes is zero in %+v. This might be expected", device)
		}
		if device.UsedBytes == 0 {
			t.Logf("[WARNING] used-bytes is zero in %+v. This might be expected", device)
		}
	}
}
