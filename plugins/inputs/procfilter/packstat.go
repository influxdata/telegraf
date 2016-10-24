package procfilter

// TODO aggregate UIDs/GIDs/groups/... for elems ?

import (
	"fmt"
)

var packStatId tPid

// resetAllGatherStats  prepare pack stats for a new sample.
func resetAllPackStats() {
	packStatId = 0
}

func NewPackStat(es []*procStat) *packStat {
	packStatId--
	s := packStat{pid: packStatId, elems: es}
	s.uid = -1 // flag as unintialized
	s.gid = -1
	return &s
}

/* Packed/aggregated stat for a set of stat (probably mainly procStats) as a single entity
This is used to gather a single value for a set of processes. Eg: The total RSS used by all tomcat processes
*/
type packStat struct {
	pid   tPid
	elems []*procStat
	other string // Special "other" case
	// pack criteria values
	uid int32
	gid int32
	cmd string
	// aggregated values
	procNbTs   tStamp
	procNb     int64
	threadNbTs tStamp
	threadNb   int64
	fdNbTs     tStamp
	fdNb       int64
	cpuTs      tStamp
	cpu        float32
	rssTs      tStamp
	rss        int64
	vszTs      tStamp
	vsz        int64
	swapTs     tStamp
	swap       int64
}

func (p *packStat) PID() tPid {
	return p.pid
}

func (p *packStat) ProcessNumber() int64 {
	if p.procNbTs == stamp {
		return p.procNb
	}
	var nb int64
	for _, s := range p.elems {
		nb += s.ProcessNumber()
	}
	p.procNb = nb
	p.procNbTs = stamp
	return nb
}

func (p *packStat) Args() ([]string, error) {
	return []string{}, nil
}

func (p *packStat) CPU() (float32, error) {
	if p.cpuTs == stamp {
		return p.cpu, nil
	}
	var cpu float64
	for _, s := range p.elems {
		c, _ := s.CPU()
		cpu += float64(c)
	}
	p.cpu = float32(cpu)
	p.cpuTs = stamp
	return p.cpu, nil
}

func (p *packStat) GIDs() ([]int32, error) {
	if p.gid == -2 {
		return []int32{-2}, nil // other
	}
	if p.gid == -1 {
		return nil, fmt.Errorf("this packStat has no GID.")
	}
	return []int32{p.gid}, nil
}

func (p *packStat) UIDs() ([]int32, error) {
	if p.uid == -2 {
		return []int32{-2}, nil // other
	}
	if p.uid == -1 {
		return nil, fmt.Errorf("this packStat has no UID.")
	}
	return []int32{p.uid}, nil
}

func (p *packStat) Groups() ([]string, error) {
	if p.other != "" {
		return []string{p.other}, nil
	}
	if p.gid == -1 {
		return nil, fmt.Errorf("this packStat has no group.")
	}
	return []string{GIDName(p.gid)}, nil
}

func (p *packStat) Users() ([]string, error) {
	if p.other != "" {
		return []string{p.other}, nil
	}
	if p.uid == -1 {
		return nil, fmt.Errorf("this packStat has no user.")
	}
	return []string{UIDName(p.uid)}, nil
}

func (p *packStat) Cmd() (string, error) {
	if p.other != "" {
		return p.other, nil
	}
	return p.cmd, nil
}

func (p *packStat) Exe() (string, error) {
	if p.other != "" {
		return p.other, nil
	}
	return "", nil
}

func (p *packStat) CmdLine() (string, error) {
	if p.other != "" {
		return p.other, nil
	}
	return "", nil
}

func (p *packStat) Path() (string, error) {
	if p.other != "" {
		return p.other, nil
	}
	return "", nil
}

func (p *packStat) RSS() (int64, error) {
	if p.rssTs == stamp {
		return p.rss, nil
	}
	var sum int64
	for _, s := range p.elems {
		v, _ := s.RSS()
		sum += v
	}
	p.rss = sum
	p.rssTs = stamp
	return sum, nil
}

func (p *packStat) VSZ() (int64, error) {
	if p.vszTs == stamp {
		return p.vsz, nil
	}
	var sum int64
	for _, s := range p.elems {
		v, _ := s.VSZ()
		sum += v
	}
	p.vsz = sum
	p.vszTs = stamp
	return sum, nil
}

func (p *packStat) Swap() (int64, error) {
	if p.swapTs == stamp {
		return p.swap, nil
	}
	var sum int64
	for _, s := range p.elems {
		v, _ := s.Swap()
		sum += v
	}
	p.swap = sum
	p.swapTs = stamp
	return sum, nil
}

func (p *packStat) ThreadNumber() (int64, error) {
	if p.threadNbTs == stamp {
		return p.threadNb, nil
	}
	var sum int64
	for _, s := range p.elems {
		v, _ := s.ThreadNumber()
		sum += v
	}
	p.threadNb = sum
	p.threadNbTs = stamp
	return sum, nil
}

func (p *packStat) FDNumber() (int64, error) {
	if p.fdNbTs == stamp {
		return p.fdNb, nil
	}
	var sum int64
	for _, s := range p.elems {
		v, _ := s.FDNumber()
		sum += v
	}
	p.fdNb = sum
	p.fdNbTs = stamp
	return sum, nil
}

func (p *packStat) ChildrenPIDs(depth int) []tPid {
	mall := map[tPid]interface{}{}
	// reccursively get all children as a slice of gopsutils processes
	for _, s := range p.elems {
		pids := s.ChildrenPIDs(depth)
		for _, pid := range pids {
			mall[pid] = nil
		}
	}
	all := make([]tPid, len(mall))
	for pid, _ := range mall {
		all = append(all, pid)
	}
	return all
}
