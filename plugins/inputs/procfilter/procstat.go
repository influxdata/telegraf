package procfilter

import (
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/shirou/gopsutil/process"
)

// Keep all real processes stats
var allProcStats = map[tPid]stat{}

/* Stats for a process */
type procStat struct {
	pid        tPid
	gproc      *process.Process
	cmdTs      tStamp
	cmd        string
	exe        string
	cmdLine    string
	argsTs     tStamp
	args       []string
	pathTs     tStamp
	path       string
	uidsTs     tStamp
	uids       []int32
	usersTs    tStamp
	users      []string
	gidsTs     tStamp
	gids       []int32
	groupsTs   tStamp
	groups     []string
	cpuTs      tStamp
	cpu        float32
	memTs      tStamp
	rss        int64
	vsz        int64
	swap       int64
	threadNbTs tStamp
	threadNb   int64
	fdNbTs     tStamp
	fdNb       int64
}

func (p *procStat) PID() tPid {
	return p.pid
}

func (p *procStat) ProcessNumber() int64 {
	return 1
}

func (p *procStat) Args() ([]string, error) {
	if p.argsTs != 0 {
		return p.args, nil
	}
	args, err := p.gproc.CmdlineSlice()
	if err != nil {
		return []string{}, err
	}
	p.args = args
	p.argsTs = stamp
	return args, nil
}

func (p *procStat) CPU() (float32, error) {
	if p.cpuTs == stamp {
		return p.cpu, nil
	}
	cpu, err := p.gproc.Percent(0)
	if err != nil {
		return 0, err
	}
	p.cpu = float32(cpu)
	p.cpuTs = stamp
	return p.cpu, nil
}

func (p *procStat) GIDs() ([]int32, error) {
	if p.gidsTs != 0 {
		return p.gids, nil
	}
	gids, err := p.gproc.Gids()
	if err != nil {
		return []int32{}, err
	}
	p.gids = gids
	p.gidsTs = stamp
	return gids, nil
}

func (p *procStat) UIDs() ([]int32, error) {
	if p.uidsTs != 0 {
		return p.uids, nil
	}
	uids, err := p.gproc.Uids()
	if err != nil {
		return []int32{}, err
	}
	p.uids = uids
	p.uidsTs = stamp
	return uids, nil
}

func (p *procStat) Groups() ([]string, error) {
	if p.groupsTs != 0 {
		return p.groups, nil
	}
	// gopsutils has (not yet?) a grouname method
	groups := []string{}
	gids, err := p.GIDs()
	if err != nil {
		return groups, err
	}
	for _, gid := range gids {
		group, err := user.LookupGroupId(strconv.Itoa(int(gid)))
		if err != nil {
			continue
		}
		groups = append(groups, group.Name)
	}
	p.groups = groups
	p.groupsTs = stamp
	return groups, nil
}

func (p *procStat) Users() ([]string, error) {
	if p.usersTs != 0 {
		return p.users, nil
	}
	users := []string{}
	uids, err := p.UIDs()
	if err != nil {
		return users, err
	}
	for _, uid := range uids {
		user, err := user.LookupId(strconv.Itoa(int(uid)))
		if err != nil {
			continue
		}
		users = append(users, user.Username)
	}
	p.users = users
	p.usersTs = stamp
	return users, nil
}

func (p *procStat) fillCmd() error {
	if p.cmdTs != 0 {
		return nil
	}
	cmd, _ := p.gproc.Name()
	p.cmd = cmd
	cmdLine, _ := p.gproc.Cmdline()
	p.cmdLine = cmdLine
	exe, _ := p.gproc.Exe()
	p.exe = exe
	p.cmdTs = stamp
	return nil
}

func (p *procStat) Cmd() (string, error) {
	err := p.fillCmd()
	return p.cmd, err
}

func (p *procStat) Exe() (string, error) {
	err := p.fillCmd()
	return p.exe, err
}

func (p *procStat) CmdLine() (string, error) {
	err := p.fillCmd()
	return p.cmdLine, err
}

func (p *procStat) Path() (string, error) {
	// Do not include path in the glocal fillCmd call, because path is seldom used, aoid storing unused data.
	if p.pathTs != 0 {
		return p.path, nil
	}
	exe, err := p.Exe()
	path := filepath.Dir(exe)
	p.path = path
	p.pathTs = stamp
	return path, err
}

func (p *procStat) fillMem() error {
	if p.memTs == stamp {
		return nil
	}
	mem, err := p.gproc.MemoryInfo()
	if err != nil {
		return err
	}
	p.rss = int64(mem.RSS)
	p.vsz = int64(mem.VMS)
	p.swap = int64(mem.Swap)
	p.memTs = stamp
	return nil
}

func (p *procStat) RSS() (int64, error) {
	err := p.fillMem()
	return p.rss, err
}

func (p *procStat) VSZ() (int64, error) {
	err := p.fillMem()
	return p.vsz, err
}
func (p *procStat) Swap() (int64, error) {
	err := p.fillMem()
	return p.swap, err
}

func (p *procStat) ThreadNumber() (int64, error) {
	if p.threadNbTs == stamp {
		return p.threadNb, nil
	}
	nb, err := p.gproc.NumThreads()
	if err != nil {
		return 0, err
	}
	p.threadNb = int64(nb)
	p.threadNbTs = stamp
	return p.threadNb, nil
}

func (p *procStat) FDNumber() (int64, error) {
	if p.fdNbTs == stamp {
		return p.fdNb, nil
	}
	nb, err := p.gproc.NumFDs()
	if err != nil {
		return 0, err
	}
	p.fdNb = int64(nb)
	p.fdNbTs = stamp
	return p.threadNb, nil
}

func (p *procStat) ChildrenPIDs(depth int) []tPid {
	all := []tPid{}
	// reccusively get all children as a slice of gopsutils processes
	var roots = make([]*process.Process, 1)
	roots[0] = p.gproc
	getGoprocChildren(depth, roots, &all)
	return all
}

func getProcStatChildren(depth int, psRoots []*procStat, pall *[]tPid) {
	var gpRoots = make([]*process.Process, 1)
	for _, p := range psRoots {
		gpRoots[0] = p.gproc
		getGoprocChildren(depth, gpRoots, pall)
	}
}

func getGoprocChildren(depth int, roots []*process.Process, pall *[]tPid) {
	for _, p := range roots {
		*pall = append(*pall, tPid(p.Pid))
		if depth > 0 {
			subRoots, err := p.Children()
			if err != nil {
				continue
			}
			getGoprocChildren(depth-1, subRoots, pall)
		}
	}
}

func (p *stats) getChildrenStats(depth int) map[tPid]stat {
	all := map[tPid]stat{}
	// reccusively get all children as a slice of gopsutils processes
	for _, s := range p.pid2Stat {
		pids := s.ChildrenPIDs(depth)
		for _, pid := range pids {
			all[pid] = allProcStats[pid]
		}
	}
	return all
}

func resetAllProcStats() {
	// Get all new proicesses
	pids, err := process.Pids()
	if err != nil {
		return
	}
	naps := map[tPid]stat{}
	for _, pid := range pids {
		if os, known := allProcStats[tPid(pid)]; known {
			// Old (known) but still alive process, keep it.
			naps[tPid(pid)] = os
		} else {
			// This is a new process (unknown PID).
			gproc, err := process.NewProcess(pid)
			if err != nil {
				continue // process terminated?
			}
			s := procStat{}
			s.pid = tPid(pid)
			s.gproc = gproc
			naps[tPid(pid)] = &s
		}
	}
	// naps contains only valid PIDs/procStats.
	allProcStats = naps
}
