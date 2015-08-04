// +build darwin

package disk

import (
	"syscall"
	"unsafe"

	"github.com/influxdb/telegraf/plugins/system/ps/common"
)

func DiskPartitions(fstypes []string) ([]DiskPartitionStat, error) {
	var ret []DiskPartitionStat

	set := make(map[string]struct{}, len(fstypes))
	for _, s := range fstypes {
		set[s] = struct{}{}
	}

	count, err := Getfsstat(nil, MntWait)
	if err != nil {
		return ret, err
	}
	fs := make([]Statfs_t, count)
	_, err = Getfsstat(fs, MntWait)
	for _, stat := range fs {
		opts := "rw"
		if stat.Flags&MntReadOnly != 0 {
			opts = "ro"
		}
		if stat.Flags&MntSynchronous != 0 {
			opts += ",sync"
		}
		if stat.Flags&MntNoExec != 0 {
			opts += ",noexec"
		}
		if stat.Flags&MntNoSuid != 0 {
			opts += ",nosuid"
		}
		if stat.Flags&MntUnion != 0 {
			opts += ",union"
		}
		if stat.Flags&MntAsync != 0 {
			opts += ",async"
		}
		if stat.Flags&MntSuidDir != 0 {
			opts += ",suiddir"
		}
		if stat.Flags&MntSoftDep != 0 {
			opts += ",softdep"
		}
		if stat.Flags&MntNoSymFollow != 0 {
			opts += ",nosymfollow"
		}
		if stat.Flags&MntGEOMJournal != 0 {
			opts += ",gjounalc"
		}
		if stat.Flags&MntMultilabel != 0 {
			opts += ",multilabel"
		}
		if stat.Flags&MntACLs != 0 {
			opts += ",acls"
		}
		if stat.Flags&MntNoATime != 0 {
			opts += ",noattime"
		}
		if stat.Flags&MntClusterRead != 0 {
			opts += ",nocluster"
		}
		if stat.Flags&MntClusterWrite != 0 {
			opts += ",noclusterw"
		}
		if stat.Flags&MntNFS4ACLs != 0 {
			opts += ",nfs4acls"
		}
		d := DiskPartitionStat{
			Device:     common.IntToString(stat.Mntfromname[:]),
			Mountpoint: common.IntToString(stat.Mntonname[:]),
			Fstype:     common.IntToString(stat.Fstypename[:]),
			Opts:       opts,
		}
		_, ok := set[d.Fstype]
		if ok || len(fstypes) == 0 {
			ret = append(ret, d)
		}
	}

	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	return nil, common.NotImplementedError
}

func Getfsstat(buf []Statfs_t, flags int) (n int, err error) {
	var _p0 unsafe.Pointer
	var bufsize uintptr
	if len(buf) > 0 {
		_p0 = unsafe.Pointer(&buf[0])
		bufsize = unsafe.Sizeof(Statfs_t{}) * uintptr(len(buf))
	}
	r0, _, e1 := syscall.Syscall(SYS_GETFSSTAT64, uintptr(_p0), bufsize, uintptr(flags))
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}
