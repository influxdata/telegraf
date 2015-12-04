// +build freebsd
// +build 386

package process

// copied from sys/sysctl.h
const (
	CTLKern          = 1  // "high kernel": proc, limits
	KernProc         = 14 // struct: process entries
	KernProcPID      = 1  // by process id
	KernProcProc     = 8  // only return procs
	KernProcPathname = 12 // path to executable
	KernProcArgs     = 7  // get/set arguments/proctitle
)

type Timespec struct {
	Sec  int32
	Nsec int32
}

type Timeval struct {
	Sec  int32
	Usec int32
}

type Rusage struct {
	Utime    Timeval
	Stime    Timeval
	Maxrss   int32
	Ixrss    int32
	Idrss    int32
	Isrss    int32
	Minflt   int32
	Majflt   int32
	Nswap    int32
	Inblock  int32
	Oublock  int32
	Msgsnd   int32
	Msgrcv   int32
	Nsignals int32
	Nvcsw    int32
	Nivcsw   int32
}

// copied from sys/user.h
type KinfoProc struct {
	KiStructsize   int32
	KiLayout       int32
	KiArgs         int32
	KiPaddr        int32
	KiAddr         int32
	KiTracep       int32
	KiTextvp       int32
	KiFd           int32
	KiVmspace      int32
	KiWchan        int32
	KiPid          int32
	KiPpid         int32
	KiPgid         int32
	KiTpgid        int32
	KiSid          int32
	KiTsid         int32
	KiJobc         [2]byte
	KiSpareShort1  [2]byte
	KiTdev         int32
	KiSiglist      [16]byte
	KiSigmask      [16]byte
	KiSigignore    [16]byte
	KiSigcatch     [16]byte
	KiUID          int32
	KiRuid         int32
	KiSvuid        int32
	KiRgid         int32
	KiSvgid        int32
	KiNgroups      [2]byte
	KiSpareShort2  [2]byte
	KiGroups       [64]byte
	KiSize         int32
	KiRssize       int32
	KiSwrss        int32
	KiTsize        int32
	KiDsize        int32
	KiSsize        int32
	KiXstat        [2]byte
	KiAcflag       [2]byte
	KiPctcpu       int32
	KiEstcpu       int32
	KiSlptime      int32
	KiSwtime       int32
	KiCow          int32
	KiRuntime      int64
	KiStart        [8]byte
	KiChildtime    [8]byte
	KiFlag         int32
	KiKflag        int32
	KiTraceflag    int32
	KiStat         [1]byte
	KiNice         [1]byte
	KiLock         [1]byte
	KiRqindex      [1]byte
	KiOncpu        [1]byte
	KiLastcpu      [1]byte
	KiOcomm        [17]byte
	KiWmesg        [9]byte
	KiLogin        [18]byte
	KiLockname     [9]byte
	KiComm         [20]byte
	KiEmul         [17]byte
	KiSparestrings [68]byte
	KiSpareints    [36]byte
	KiCrFlags      int32
	KiJid          int32
	KiNumthreads   int32
	KiTid          int32
	KiPri          int32
	KiRusage       Rusage
	KiRusageCh     [72]byte
	KiPcb          int32
	KiKstack       int32
	KiUdata        int32
	KiTdaddr       int32
	KiSpareptrs    [24]byte
	KiSpareint64s  [48]byte
	KiSflag        int32
	KiTdflags      int32
}
