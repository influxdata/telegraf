// +build freebsd

package process

import (
	"bytes"
	"encoding/binary"
	"unsafe"
	"strings"
	"syscall"

	cpu "github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/internal/common"
	net "github.com/shirou/gopsutil/net"
)

// MemoryInfoExStat is different between OSes
type MemoryInfoExStat struct {
}

type MemoryMapsStat struct {
}

func Pids() ([]int32, error) {
	var ret []int32
	procs, err := processes()
	if err != nil {
		return ret, nil
	}

	for _, p := range procs {
		ret = append(ret, p.Pid)
	}

	return ret, nil
}

func (p *Process) Ppid() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.KiPpid, nil
}
func (p *Process) Name() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	return string(k.KiComm[:]), nil
}
func (p *Process) Exe() (string, error) {
	return "", common.NotImplementedError
}
func (p *Process) Cmdline() (string, error) {
	mib := []int32{CTLKern, KernProc, KernProcArgs, p.Pid}
	buf, _, err := common.CallSyscall(mib)
	if err != nil {
		return "", err
	}
	ret := strings.FieldsFunc(string(buf), func(r rune) bool {
		if r == '\u0000' {
			return true
		}
		return false
	})

	return strings.Join(ret, " "), nil
}
func (p *Process) CreateTime() (int64, error) {
	return 0, common.NotImplementedError
}
func (p *Process) Cwd() (string, error) {
	return "", common.NotImplementedError
}
func (p *Process) Parent() (*Process, error) {
	return p, common.NotImplementedError
}
func (p *Process) Status() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	return string(k.KiStat[:]), nil
}
func (p *Process) Uids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	uids := make([]int32, 0, 3)

	uids = append(uids, int32(k.KiRuid), int32(k.KiUID), int32(k.KiSvuid))

	return uids, nil
}
func (p *Process) Gids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	gids := make([]int32, 0, 3)
	gids = append(gids, int32(k.KiRgid), int32(k.KiNgroups[0]), int32(k.KiSvuid))

	return gids, nil
}
func (p *Process) Terminal() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	ttyNr := uint64(k.KiTdev)

	termmap, err := getTerminalMap()
	if err != nil {
		return "", err
	}

	return termmap[ttyNr], nil
}
func (p *Process) Nice() (int32, error) {
	return 0, common.NotImplementedError
}
func (p *Process) IOnice() (int32, error) {
	return 0, common.NotImplementedError
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	var rlimit []RlimitStat
	return rlimit, common.NotImplementedError
}
func (p *Process) IOCounters() (*IOCountersStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}
	return &IOCountersStat{
		ReadCount:  uint64(k.KiRusage.Inblock),
		WriteCount: uint64(k.KiRusage.Oublock),
	}, nil
}
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	return nil, common.NotImplementedError
}
func (p *Process) NumFDs() (int32, error) {
	return 0, common.NotImplementedError
}
func (p *Process) NumThreads() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.KiNumthreads, nil
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, common.NotImplementedError
}
func (p *Process) CPUTimes() (*cpu.CPUTimesStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}
	return &cpu.CPUTimesStat{
		CPU:    "cpu",
		User:   float64(k.KiRusage.Utime.Sec) + float64(k.KiRusage.Utime.Usec)/1000000,
		System: float64(k.KiRusage.Stime.Sec) + float64(k.KiRusage.Stime.Usec)/1000000,
	}, nil
}
func (p *Process) CPUAffinity() ([]int32, error) {
	return nil, common.NotImplementedError
}
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}
	v, err := syscall.Sysctl("vm.stats.vm.v_page_size")
	if err != nil {
		return nil, err
	}
	pageSize := binary.LittleEndian.Uint16([]byte(v))

	return &MemoryInfoStat{
		RSS: uint64(k.KiRssize) * uint64(pageSize),
		VMS: uint64(k.KiSize),
	}, nil
}
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	return nil, common.NotImplementedError
}
func (p *Process) MemoryPercent() (float32, error) {
	return 0, common.NotImplementedError
}

func (p *Process) Children() ([]*Process, error) {
	pids, err := common.CallPgrep(invoke, p.Pid)
	if err != nil {
		return nil, err
	}
	ret := make([]*Process, 0, len(pids))
	for _, pid := range pids {
		np, err := NewProcess(pid)
		if err != nil {
			return nil, err
		}
		ret = append(ret, np)
	}
	return ret, nil
}

func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	return nil, common.NotImplementedError
}

func (p *Process) Connections() ([]net.NetConnectionStat, error) {
	return nil, common.NotImplementedError
}

func (p *Process) IsRunning() (bool, error) {
	return true, common.NotImplementedError
}
func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	var ret []MemoryMapsStat
	return &ret, common.NotImplementedError
}

func copyParams(k *KinfoProc, p *Process) error {

	return nil
}

func processes() ([]Process, error) {
	results := make([]Process, 0, 50)

	mib := []int32{CTLKern, KernProc, KernProcProc, 0}
	buf, length, err := common.CallSyscall(mib)
	if err != nil {
		return results, err
	}

	// get kinfo_proc size
	k := KinfoProc{}
	procinfoLen := int(unsafe.Sizeof(k))
	count := int(length / uint64(procinfoLen))

	// parse buf to procs
	for i := 0; i < count; i++ {
		b := buf[i*procinfoLen : i*procinfoLen+procinfoLen]
		k, err := parseKinfoProc(b)
		if err != nil {
			continue
		}
		p, err := NewProcess(int32(k.KiPid))
		if err != nil {
			continue
		}
		copyParams(&k, p)

		results = append(results, *p)
	}

	return results, nil
}

func parseKinfoProc(buf []byte) (KinfoProc, error) {
	var k KinfoProc
	br := bytes.NewReader(buf)
	err := binary.Read(br, binary.LittleEndian, &k)
	return k, err
}

func (p *Process) getKProc() (*KinfoProc, error) {
	mib := []int32{CTLKern, KernProc, KernProcPID, p.Pid}

	buf, length, err := common.CallSyscall(mib)
	if err != nil {
		return nil, err
	}
	procK := KinfoProc{}
	if length != uint64(unsafe.Sizeof(procK)) {
		return nil, err
	}

	k, err := parseKinfoProc(buf)
	if err != nil {
		return nil, err
	}

	return &k, nil
}

func NewProcess(pid int32) (*Process, error) {
	p := &Process{Pid: pid}

	return p, nil
}
