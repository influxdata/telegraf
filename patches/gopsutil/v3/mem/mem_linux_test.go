//go:build linux
// +build linux

package mem

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVirtualMemoryEx(t *testing.T) {
	v, err := VirtualMemoryEx()
	if err != nil {
		t.Error(err)
	}

	t.Log(v)
}

var virtualMemoryTests = []struct {
	mockedRootFS string
	stat         *VirtualMemoryStat
}{
	{
		"intelcorei5", &VirtualMemoryStat{
			Total:          16502300672,
			Available:      11495358464,
			Used:           3437277184,
			UsedPercent:    20.82907863769651,
			Free:           8783491072,
			Active:         4347392000,
			Inactive:       2938834944,
			Wired:          0,
			Laundry:        0,
			Buffers:        212496384,
			Cached:         4069036032,
			WriteBack:      0,
			Dirty:          176128,
			WriteBackTmp:   0,
			Shared:         1222402048,
			Slab:           253771776,
			Sreclaimable:   186470400,
			Sunreclaim:     67301376,
			PageTables:     65241088,
			SwapCached:     0,
			CommitLimit:    16509730816,
			CommittedAS:    12360818688,
			HighTotal:      0,
			HighFree:       0,
			LowTotal:       0,
			LowFree:        0,
			SwapTotal:      8258580480,
			SwapFree:       8258580480,
			Mapped:         1172627456,
			VmallocTotal:   35184372087808,
			VmallocUsed:    0,
			VmallocChunk:   0,
			HugePagesTotal: 0,
			HugePagesFree:  0,
			HugePageSize:   2097152,
		},
	},
	{
		"issue1002", &VirtualMemoryStat{
			Total:          260579328,
			Available:      215199744,
			Used:           34328576,
			UsedPercent:    13.173944481121694,
			Free:           124506112,
			Active:         108785664,
			Inactive:       8581120,
			Wired:          0,
			Laundry:        0,
			Buffers:        4915200,
			Cached:         96829440,
			WriteBack:      0,
			Dirty:          0,
			WriteBackTmp:   0,
			Shared:         0,
			Slab:           9293824,
			Sreclaimable:   2764800,
			Sunreclaim:     6529024,
			PageTables:     405504,
			SwapCached:     0,
			CommitLimit:    130289664,
			CommittedAS:    25567232,
			HighTotal:      134217728,
			HighFree:       67784704,
			LowTotal:       126361600,
			LowFree:        56721408,
			SwapTotal:      0,
			SwapFree:       0,
			Mapped:         38793216,
			VmallocTotal:   1996488704,
			VmallocUsed:    0,
			VmallocChunk:   0,
			HugePagesTotal: 0,
			HugePagesFree:  0,
			HugePageSize:   0,
		},
	},
}

func TestVirtualMemoryLinux(t *testing.T) {
	origProc := os.Getenv("HOST_PROC")
	defer os.Setenv("HOST_PROC", origProc)

	for _, tt := range virtualMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			os.Setenv("HOST_PROC", filepath.Join("testdata/linux/virtualmemory/", tt.mockedRootFS, "proc"))

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

const validFile = `Filename				Type		Size		Used		Priority
/dev/dm-2                               partition	67022844	490788		-2
/swapfile                               file		2		1		-3
`

const invalidFile = `INVALID				Type		Size		Used		Priority
/dev/dm-2                               partition	67022844	490788		-2
/swapfile                               file		1048572		0		-3
`

func TestParseSwapsFile_ValidFile(t *testing.T) {
	assert := assert.New(t)
	stats, err := parseSwapsFile(strings.NewReader(validFile))
	assert.NoError(err)

	assert.Equal(*stats[0], SwapDevice{
		Name:      "/dev/dm-2",
		UsedBytes: 502566912,
		FreeBytes: 68128825344,
	})

	assert.Equal(*stats[1], SwapDevice{
		Name:      "/swapfile",
		UsedBytes: 1024,
		FreeBytes: 1024,
	})
}

func TestParseSwapsFile_InvalidFile(t *testing.T) {
	_, err := parseSwapsFile(strings.NewReader(invalidFile))
	assert.Error(t, err)
}

func TestParseSwapsFile_EmptyFile(t *testing.T) {
	_, err := parseSwapsFile(strings.NewReader(""))
	assert.Error(t, err)
}
