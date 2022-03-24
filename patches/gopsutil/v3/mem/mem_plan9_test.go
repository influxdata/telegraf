//go:build plan9
// +build plan9

package mem

import (
	"os"
	"reflect"
	"testing"
)

var virtualMemoryTests = []struct {
	mockedRootFS string
	stat         *VirtualMemoryStat
}{
	{
		"swap", &VirtualMemoryStat{
			Total:       1071185920,
			Available:   808370176,
			Used:        11436032,
			UsedPercent: 1.3949677238843257,
			Free:        808370176,
			SwapTotal:   655360000,
			SwapFree:    655360000,
		},
	},
}

func TestVirtualMemoryPlan9(t *testing.T) {
	origProc := os.Getenv("HOST_ROOT")
	t.Cleanup(func() {
		os.Setenv("HOST_ROOT", origProc)
	})

	for _, tt := range virtualMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			os.Setenv("HOST_ROOT", "testdata/plan9/virtualmemory/")

			stat, err := VirtualMemory()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("error %v", err)
			}
			if !reflect.DeepEqual(stat, tt.stat) {
				t.Errorf("got: %+v\nwant: %+v", stat, tt.stat)
			}
		})
	}
}

var swapMemoryTests = []struct {
	mockedRootFS string
	swap         *SwapMemoryStat
}{
	{
		"swap", &SwapMemoryStat{
			Total: 655360000,
			Used:  0,
			Free:  655360000,
		},
	},
}

func TestSwapMemoryPlan9(t *testing.T) {
	origProc := os.Getenv("HOST_ROOT")
	t.Cleanup(func() {
		os.Setenv("HOST_ROOT", origProc)
	})

	for _, tt := range swapMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			os.Setenv("HOST_ROOT", "testdata/plan9/virtualmemory/")

			swap, err := SwapMemory()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("error %v", err)
			}
			if !reflect.DeepEqual(swap, tt.swap) {
				t.Errorf("got: %+v\nwant: %+v", swap, tt.swap)
			}
		})
	}
}
