package disk

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func skipIfNotImplementedErr(t *testing.T, err error) {
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
}

func TestDisk_usage(t *testing.T) {
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	v, err := Usage(path)
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Path != path {
		t.Errorf("error %v", err)
	}
}

func TestDisk_partitions(t *testing.T) {
	ret, err := Partitions(false)
	skipIfNotImplementedErr(t, err)
	if err != nil || len(ret) == 0 {
		t.Errorf("error %v", err)
	}
	t.Log(ret)

	if len(ret) == 0 {
		t.Errorf("ret is empty")
	}
	for _, disk := range ret {
		if disk.Device == "" {
			t.Errorf("Could not get device info %v", disk)
		}
	}
}

func TestDisk_io_counters(t *testing.T) {
	ret, err := IOCounters()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(ret) == 0 {
		t.Errorf("ret is empty")
	}
	empty := IOCountersStat{}
	for part, io := range ret {
		t.Log(part, io)
		if io == empty {
			t.Errorf("io_counter error %v, %v", part, io)
		}
	}
}

// https://github.com/shirou/gopsutil/issues/560 regression test
func TestDisk_io_counters_concurrency_on_darwin_cgo(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	var wg sync.WaitGroup
	const max = 1000
	for i := 1; i < max; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			IOCounters()
		}()
	}
	wg.Wait()
}

func TestDiskUsageStat_String(t *testing.T) {
	v := UsageStat{
		Path:              "/",
		Total:             1000,
		Free:              2000,
		Used:              3000,
		UsedPercent:       50.1,
		InodesTotal:       4000,
		InodesUsed:        5000,
		InodesFree:        6000,
		InodesUsedPercent: 49.1,
		Fstype:            "ext4",
	}
	e := `{"path":"/","fstype":"ext4","total":1000,"free":2000,"used":3000,"usedPercent":50.1,"inodesTotal":4000,"inodesUsed":5000,"inodesFree":6000,"inodesUsedPercent":49.1}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("DiskUsageStat string is invalid: %v", v)
	}
}

func TestDiskPartitionStat_String(t *testing.T) {
	v := PartitionStat{
		Device:     "sd01",
		Mountpoint: "/",
		Fstype:     "ext4",
		Opts:       []string{"ro"},
	}
	e := `{"device":"sd01","mountpoint":"/","fstype":"ext4","opts":["ro"]}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("DiskUsageStat string is invalid: %v", v)
	}
}

func TestDiskIOCountersStat_String(t *testing.T) {
	v := IOCountersStat{
		Name:         "sd01",
		ReadCount:    100,
		WriteCount:   200,
		ReadBytes:    300,
		WriteBytes:   400,
		SerialNumber: "SERIAL",
	}
	e := `{"readCount":100,"mergedReadCount":0,"writeCount":200,"mergedWriteCount":0,"readBytes":300,"writeBytes":400,"readTime":0,"writeTime":0,"iopsInProgress":0,"ioTime":0,"weightedIO":0,"name":"sd01","serialNumber":"SERIAL","label":""}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("DiskUsageStat string is invalid: %v", v)
	}
}
