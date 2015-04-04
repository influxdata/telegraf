// +build windows

package disk

import (
	"bytes"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	common "github.com/shirou/gopsutil/common"
)

var (
	procGetDiskFreeSpaceExW     = common.Modkernel32.NewProc("GetDiskFreeSpaceExW")
	procGetLogicalDriveStringsW = common.Modkernel32.NewProc("GetLogicalDriveStringsW")
	procGetDriveType            = common.Modkernel32.NewProc("GetDriveTypeW")
	provGetVolumeInformation    = common.Modkernel32.NewProc("GetVolumeInformationW")
)

var (
	FileFileCompression = int64(16)     // 0x00000010
	FileReadOnlyVolume  = int64(524288) // 0x00080000
)

const WaitMSec = 500

func DiskUsage(path string) (DiskUsageStat, error) {
	ret := DiskUsageStat{}

	ret.Path = path
	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	diskret, _, err := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))
	if diskret == 0 {
		return ret, err
	}
	ret.Total = uint64(lpTotalNumberOfBytes)
	//	ret.Free = uint64(lpFreeBytesAvailable) // python psutil does not use this
	ret.Free = uint64(lpTotalNumberOfFreeBytes)
	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Used) / float64(ret.Total) * 100.0

	//TODO: implement inodes stat
	ret.InodesTotal = 0
	ret.InodesUsed = 0
	ret.InodesFree = 0
	ret.InodesUsedPercent = 0.0
	return ret, nil
}

func DiskPartitions(all bool) ([]DiskPartitionStat, error) {
	var ret []DiskPartitionStat
	lpBuffer := make([]byte, 254)
	diskret, _, err := procGetLogicalDriveStringsW.Call(
		uintptr(len(lpBuffer)),
		uintptr(unsafe.Pointer(&lpBuffer[0])))
	if diskret == 0 {
		return ret, err
	}
	for _, v := range lpBuffer {
		if v >= 65 && v <= 90 {
			path := string(v) + ":"
			if path == "A:" || path == "B:" { // skip floppy drives
				continue
			}
			typepath, _ := syscall.UTF16PtrFromString(path)
			typeret, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(typepath)))
			if typeret == 0 {
				return ret, syscall.GetLastError()
			}
			// 2: DRIVE_REMOVABLE 3: DRIVE_FIXED 5: DRIVE_CDROM

			if typeret == 2 || typeret == 3 || typeret == 5 {
				lpVolumeNameBuffer := make([]byte, 256)
				lpVolumeSerialNumber := int64(0)
				lpMaximumComponentLength := int64(0)
				lpFileSystemFlags := int64(0)
				lpFileSystemNameBuffer := make([]byte, 256)
				volpath, _ := syscall.UTF16PtrFromString(string(v) + ":/")
				driveret, _, err := provGetVolumeInformation.Call(
					uintptr(unsafe.Pointer(volpath)),
					uintptr(unsafe.Pointer(&lpVolumeNameBuffer[0])),
					uintptr(len(lpVolumeNameBuffer)),
					uintptr(unsafe.Pointer(&lpVolumeSerialNumber)),
					uintptr(unsafe.Pointer(&lpMaximumComponentLength)),
					uintptr(unsafe.Pointer(&lpFileSystemFlags)),
					uintptr(unsafe.Pointer(&lpFileSystemNameBuffer[0])),
					uintptr(len(lpFileSystemNameBuffer)))
				if driveret == 0 {
					return ret, err
				}
				opts := "rw"
				if lpFileSystemFlags&FileReadOnlyVolume != 0 {
					opts = "ro"
				}
				if lpFileSystemFlags&FileFileCompression != 0 {
					opts += ".compress"
				}

				d := DiskPartitionStat{
					Mountpoint: path,
					Device:     path,
					Fstype:     string(bytes.Replace(lpFileSystemNameBuffer, []byte("\x00"), []byte(""), -1)),
					Opts:       opts,
				}
				ret = append(ret, d)
			}
		}
	}
	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	ret := make(map[string]DiskIOCountersStat, 0)
	query, err := common.CreateQuery()
	if err != nil {
		return ret, err
	}

	drivebuf := make([]byte, 256)
	r, _, err := procGetLogicalDriveStringsW.Call(
		uintptr(len(drivebuf)),
		uintptr(unsafe.Pointer(&drivebuf[0])))

	if r == 0 {
		return ret, err
	}

	drivemap := make(map[string][]*common.CounterInfo, 0)
	for _, v := range drivebuf {
		if v >= 65 && v <= 90 {
			drive := string(v)
			r, _, err = procGetDriveType.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(drive + `:\`))))
			if r != common.DRIVE_FIXED {
				continue
			}
			drivemap[drive] = make([]*common.CounterInfo, 0, 2)
			var counter *common.CounterInfo

			counter, err = common.CreateCounter(query,
				"read",
				fmt.Sprintf(`\PhysicalDisk(0 %s:)\Disk Reads/sec`, drive))
			if err != nil {
				return nil, err
			}
			drivemap[drive] = append(drivemap[drive], counter)
			counter, err = common.CreateCounter(query,
				"write",
				fmt.Sprintf(`\PhysicalDisk(0 %s:)\Disk Writes/sec`, drive))
			if err != nil {
				return nil, err
			}
			drivemap[drive] = append(drivemap[drive], counter)
		}
	}
	r, _, err = common.PdhCollectQueryData.Call(uintptr(query))
	if r != 0 && err != nil {
		return nil, err
	}
	time.Sleep(time.Duration(WaitMSec) * time.Millisecond)
	r, _, err = common.PdhCollectQueryData.Call(uintptr(query))
	if r != 0 && err != nil {
		return nil, err
	}

	for drive, counters := range drivemap {
		stat := DiskIOCountersStat{}
		for _, v := range counters {
			var fmtValue common.PDH_FMT_COUNTERVALUE_LARGE
			r, _, err := common.PdhGetFormattedCounterValue.Call(uintptr(v.Counter), common.PDH_FMT_LARGE, uintptr(0), uintptr(unsafe.Pointer(&fmtValue)))
			if r != 0 && r != common.PDH_INVALID_DATA {
				return nil, err
			}

			switch v.PostName {
			case "read":
				stat.ReadCount = uint64(fmtValue.LargeValue)
			case "write":
				stat.WriteCount = uint64(fmtValue.LargeValue)
			default:
				return ret, fmt.Errorf("unknown postname: %s", v.PostName)
			}
			stat.Name = drive
		}
		ret[drive] = stat
	}

	return ret, nil
}
