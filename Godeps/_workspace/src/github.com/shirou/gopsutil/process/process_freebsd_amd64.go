// +build freebsd
// +build amd64

package process

// copied from sys/sysctl.h
const (
	CTLKern          = 1  // "high kernel": proc, limits
	KernProc         = 14 // struct: process entries
	KernProcPID      = 1  // by process id
	KernProcProc     = 8  // only return procs
	KernProcPathname = 12 // path to executable
)

// copied from sys/user.h
type KinfoProc struct {
	KiStructsize   int32
	KiLayout       int32
	KiArgs         int64
	KiPaddr        int64
	KiAddr         int64
	KiTracep       int64
	KiTextvp       int64
	KiFd           int64
	KiVmspace      int64
	KiWchan        int64
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
	KiSize         int64
	KiRssize       int64
	KiSwrss        int64
	KiTsize        int64
	KiDsize        int64
	KiSsize        int64
	KiXstat        [2]byte
	KiAcflag       [2]byte
	KiPctcpu       int32
	KiEstcpu       int32
	KiSlptime      int32
	KiSwtime       int32
	KiCow          int32
	KiRuntime      int64
	KiStart        [16]byte
	KiChildtime    [16]byte
	KiFlag         int64
	KiKflag        int64
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
	KiRusage       [144]byte
	KiRusageCh     [144]byte
	KiPcb          int64
	KiKstack       int64
	KiUdata        int64
	KiTdaddr       int64
	KiSpareptrs    [48]byte
	KiSpareint64s  [96]byte
	KiSflag        int64
	KiTdflags      int64
}
