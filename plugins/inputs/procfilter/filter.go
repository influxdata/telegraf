package procfilter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// TODO test for cycles if using filters()?
// TODO test for speed/optimization
// TODO test every filter/parameter

/* A filter will select a set of processes.
A filter can be used as input for other filters. Finally a filter can be used by a measurement that will output some tags/fields related to selected processes.
*/
type filter interface {
	Parse(p *Parser) error // Parse the parameters for this filer.
	Stats() *stats         // Get the concrete stats for this filter (what has been selected by the filter)
	Apply() error          // Apply the filter to its inputs (reccursively). Evaluation is lazy and done only once per sample.
}

// name2FuncFilter use a name to return an object of the proper concrete type matchingt the filter interface.
func name2FuncFilter(funcName string) filter {
	// sorry, tried to do that with reflct but did not manage to make it work
	fn := strings.ToLower(funcName)
	var f filter
	switch fn {
	case "all": // both syntax all and all() are allowed
		f = new(allFilter)
	case "top":
		f = new(topFilter)
	case "exceed":
		f = new(exceedFilter)
	case "user":
		f = new(userFilter)
	case "group":
		f = new(groupFilter)
	case "children":
		f = new(childrenFilter)
	case "command", "cmd":
		f = new(cmdFilter)
	case "exe":
		f = new(exeFilter)
	case "path":
		f = new(pathFilter)
	case "cmdline":
		f = new(cmdlineFilter)
	case "args":
		f = new(argsFilter)
	case "pid":
		f = new(pidFilter)
	case "or", "union":
		f = new(orFilter)
	case "and", "intersection":
		f = new(andFilter)
	case "not", "complement":
		f = new(notFilter)
	case "xor", "difference":
		f = new(notFilter)
	case "pack":
		f = new(packFilter)
	case "unpack":
		f = new(unpackFilter)
	case "packby", "by", "pack_by":
		f = new(packByFilter)
	case "filters":
		f = new(filtersFilter)
	default:
		f = nil
	}
	return f
}

// All processes on the server.
type allFilter struct {
	stats
}

func (f *allFilter) Apply() error {
	// This set of stats is updated once every sample by the Gather() method
	f.pid2Stat = allProcStats
	return nil
}

func (f *allFilter) Parse(p *Parser) error {
	// eg: all()
	// the all with not () is handled as a named filter that requires no special parsing
	return p.parseSymbol(')')
}

func (f *allFilter) Stats() *stats {
	return &f.stats
}

// Select the most (top) consuming given a criteria.
type topFilter struct {
	stats
	topNb int64  // how many process to keep
	crit  string // sort criteria
	input filter
}

func (f *topFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	input := f.input
	err := input.Apply()
	if err != nil {
		return err
	}
	iStats := input.Stats()
	stats := []stat{}
	// Convert the stat map to a slice without 0 as criteria.
	// (removing the 0 speed up the sort and stabilizes the top result when there are only 0 criteria)
	// TODO use function pointer?
	switch f.crit {
	case "rss":
		for _, s := range iStats.pid2Stat {
			v, err := s.RSS()
			if err != nil {
				continue
			}
			if v > 0 {
				stats = append(stats, s)
			}
		}
	case "vsz":
		for _, s := range iStats.pid2Stat {
			v, err := s.VSZ()
			if err != nil {
				continue
			}
			if v > 0 {
				stats = append(stats, s)
			}
		}
	case "swap":
		for _, s := range iStats.pid2Stat {
			v, err := s.Swap()
			if err != nil {
				continue
			}
			if v > 0 {
				stats = append(stats, s)
			}
		}
	case "thread_nb":
		for _, s := range iStats.pid2Stat {
			v, err := s.ThreadNumber()
			if err != nil {
				continue
			}
			if v > 0 {
				stats = append(stats, s)
			}
		}
	case "fd_nb":
		for _, s := range iStats.pid2Stat {
			v, err := s.FDNumber()
			if err != nil {
				continue
			}
			if v > 0 {
				stats = append(stats, s)
			}
		}
	case "process_nb":
		for _, s := range iStats.pid2Stat {
			v := s.ProcessNumber()
			if v > 0 {
				stats = append(stats, s)
			}
		}
	case "cpu":
		for _, s := range iStats.pid2Stat {
			v, err := s.CPU()
			if err != nil {
				continue
			}
			if v > 0 {
				stats = append(stats, s)
			}
		}
	default:
		return fmt.Errorf("unknow sort criteria %q", f.crit)
	}
	// sort it accoding to rss
	switch f.crit {
	case "rss":
		sort.Sort(byRSS(stats))
	case "vsz":
		sort.Sort(byVSZ(stats))
	case "swap":
		sort.Sort(bySwap(stats))
	case "thread_nb":
		sort.Sort(byThreadNumber(stats))
	case "fd_nb":
		sort.Sort(byFDNumber(stats))
	case "process_nb":
		sort.Sort(byProcessNumber(stats))
	case "cpu":
		sort.Sort(byCPU(stats))
	default:
		return fmt.Errorf("unknow sort criteria %q", f.crit)
	}
	// build this filter procstat map (a subset of the one in input filter)
	l := min(int(f.topNb), len(stats))
	m := make(map[tPid]stat, l+1) // keep room for the "other"  stat
	for i := tPid(0); i < tPid(l); i++ {
		s := stats[i]
		m[s.PID()] = s
	}
	// Add an "other" packStat with all procStat not in top.
	ss := unpackSliceAsSlice(stats[l:], nil)
	o := NewPackStat(ss)
	o.other = fmt.Sprintf("_other.top.%s.%d", f.crit, f.topNb)
	m[o.PID()] = o
	f.pid2Stat = m
	return nil
}

func (f *topFilter) Parse(p *Parser) error {
	// eg: [top(]rss,5,filter
	err := p.parseArgIdentifier(&f.crit)
	if err != nil {
		return err
	}
	err = p.parseArgInt(&f.topNb)
	if err != nil {
		return err
	}
	err = p.parseArgLastFilter(&f.input)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *topFilter) Stats() *stats {
	return &f.stats
}

// Select what exceeds a criteria/value limit.
type exceedFilter struct {
	stats
	rv    string // unparsed value
	iv    int64
	fv    float64
	crit  string
	input filter
}

func (f *exceedFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	input := f.input
	err := input.Apply()
	if err != nil {
		return err
	}
	iStats := input.Stats()
	eo := []*procStat{}
	m := map[tPid]stat{}
	for pid, s := range iStats.pid2Stat {
		switch f.crit { // TODO invert for<->switch for perfformance?
		case "rss":
			rss, err := s.RSS()
			if err != nil {
				continue
			}
			if rss > f.iv {
				m[pid] = s
			} else {
				eo = unpackStatAsSlice(s, eo)
			}
		case "vsz":
			vsz, err := s.VSZ()
			if err != nil {
				continue
			}
			if vsz > f.iv {
				m[pid] = s
			} else {
				eo = unpackStatAsSlice(s, eo)
			}
		case "swap":
			swap, err := s.Swap()
			if err != nil {
				continue
			}
			if swap > f.iv {
				m[pid] = s
			} else {
				eo = unpackStatAsSlice(s, eo)
			}
		case "thread_nb":
			tnb, err := s.ThreadNumber()
			if err != nil {
				continue
			}
			if tnb > f.iv {
				m[pid] = s
			} else {
				eo = unpackStatAsSlice(s, eo)
			}
		case "fd_nb":
			tnb, err := s.FDNumber()
			if err != nil {
				continue
			}
			if tnb > f.iv {
				m[pid] = s
			} else {
				eo = unpackStatAsSlice(s, eo)
			}
		case "process_nb":
			pnb := s.ProcessNumber()
			if pnb > f.iv {
				m[pid] = s
			} else {
				eo = unpackStatAsSlice(s, eo)
			}
		case "cpu":
			cpu, err := s.CPU()
			if err != nil {
				continue
			}
			if float64(cpu) > f.fv {
				m[pid] = s
			} else {
				eo = unpackStatAsSlice(s, eo)
			}
		default:
			return fmt.Errorf("unknown sort criteria %q", f.crit)
		}
	}
	// Pack all other procStat in one stat.
	o := NewPackStat(eo)
	o.other = fmt.Sprintf("_other.exceed.%s.%s", f.crit, f.rv)
	m[o.PID()] = o
	f.pid2Stat = m
	return nil
}

func (f *exceedFilter) Parse(p *Parser) error {
	// eg: [exceed(]rss,1G)
	// TODO eg: [exceed(]rss,10%)
	// TODO eg: [exceed(]cpu,20%)
	var err error
	err = p.parseArgIdentifier(&f.crit)
	if err != nil {
		return err
	}
	// Keep a copy of the human provided value for later string output.
	_, lit := p.scanIgnoreWhitespace()
	p.unscan()
	f.rv = lit
	// Parse the value depending on the chosen criteria.
	switch f.crit {
	case "rss", "vsz", "swap", "thread_nb", "process_nb", "fd_nb":
		err := p.parseArgInt(&f.iv)
		if err != nil {
			return p.syntaxError(fmt.Sprintf("exceed with '%s; criteri requires an integer as threshold", f.crit))
		}
	case "cpu":
		var v int64
		err := p.parseArgInt(&v)
		if err != nil {
			return p.syntaxError(fmt.Sprintf("exceed with '%s; criteri requires an integer as threshold", f.crit))
		}
		f.fv = float64(v)
	default:
		return fmt.Errorf("unknow exceed criteria %q", f.crit)
	}

	err = p.parseArgLastFilter(&f.input)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *exceedFilter) Stats() *stats {
	return &f.stats
}

// Select matching PIDs.
type pidFilter struct {
	stats
	pid   tPid
	file  string
	input filter
}

func (f *pidFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	input := f.input
	err := input.Apply()
	if err != nil {
		return err
	}
	sm := map[tPid]stat{}
	f.pid2Stat = sm
	iStats := input.Stats()
	if len(iStats.pid2Stat) == 0 {
		return nil
	}
	if f.file != "" {
		// get the PID from a file
		pid, err := pidFromFile(f.file)
		if err != nil {
			return err
		}
		f.pid = pid
	}
	for pid, s := range iStats.pid2Stat {
		ipid := s.PID()
		if ipid == f.pid {
			sm[pid] = s
		}
	}
	return nil
}

func (f *pidFilter) Parse(p *Parser) error {
	// eg: [user(]"joe",all)
	tok, lit := p.scanIgnoreWhitespace()
	if tok == tTString {
		f.file = lit
	} else if tok == tTNumber {
		{
		}
		i, err := strconv.Atoi(lit)
		if err != nil {
			return p.syntaxError(fmt.Sprintf("found %q, expecting an integer", lit))
		}
		f.pid = tPid(i)
	} else {
		return p.syntaxError(fmt.Sprintf("found %q, expecting a string or number", lit))
	}
	err := p.parseArgLastFilter(&f.input)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *pidFilter) Stats() *stats {
	return &f.stats
}

// Select matching user.
type userFilter struct {
	stats
	name   *stregexp
	id     int32
	inputs []filter
}

func (f *userFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	pss := unpackFiltersAsSlice(f.inputs, nil)
	sm := map[tPid]stat{}
	if f.name == nil { // Filter on numeric ID.
		for _, ps := range pss {
			ids, err := ps.UIDs()
			if err != nil {
				continue
			}
			for _, id := range ids {
				if id == f.id {
					sm[ps.pid] = stat(ps)
					break
				}
			}
		}
	} else { // Filter on name.
		for _, ps := range pss {
			names, err := ps.Users()
			if err != nil {
				continue
			}
			for _, name := range names {
				if f.name.matchString(name) {
					sm[ps.pid] = stat(ps)
					break
				}
			}
		}
	}
	f.stats.pid2Stat = sm
	return nil
}

func (f *userFilter) Parse(p *Parser) error {
	// eg: [user(]"foo",all)
	tok, lit := p.scanIgnoreWhitespace()
	p.unscan()
	switch tok {
	case tTString, tTRegexp:
		err := p.parseArgStregexp(&f.name)
		if err != nil {
			return p.syntaxError(err.Error())
		}
	case tTNumber:
		var i int64
		err := p.parseArgInt(&i)
		if err != nil {
			return err
		}
		f.id = int32(i)
	default:
		return p.syntaxError(fmt.Sprintf("found %q, expecting a string or number", lit))
	}
	err := p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *userFilter) Stats() *stats {
	return &f.stats
}

// Select matching group.
type groupFilter struct {
	stats
	name   *stregexp
	id     int32
	inputs []filter
}

func (f *groupFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	pss := unpackFiltersAsSlice(f.inputs, nil)
	sm := map[tPid]stat{}
	if f.name == nil { // Filter on numeric ID.
		for _, ps := range pss {
			ids, err := ps.GIDs()
			if err != nil {
				continue
			}
			for _, id := range ids {
				if id == f.id {
					sm[ps.pid] = stat(ps)
					break
				}
			}
		}
	} else { // Filter on name.
		for _, ps := range pss {
			names, err := ps.Groups()
			if err != nil {
				continue
			}
			for _, name := range names {
				if f.name.matchString(name) {
					sm[ps.pid] = stat(ps)
					break
				}
			}
		}
	}
	f.stats.pid2Stat = sm
	return nil
}

func (f *groupFilter) Parse(p *Parser) error {
	// eg: [group(]"foo",all)
	tok, lit := p.scanIgnoreWhitespace()
	p.unscan()
	switch tok {
	case tTString, tTRegexp:
		err := p.parseArgStregexp(&f.name)
		if err != nil {
			return p.syntaxError(err.Error())
		}
	case tTNumber:
		var i int64
		err := p.parseArgInt(&i)
		if err != nil {
			return err
		}
		f.id = int32(i)
	default:
		return p.syntaxError(fmt.Sprintf("found %q, ing a string or number", lit))
	}
	err := p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *groupFilter) Stats() *stats {
	return &f.stats
}

// Select children.
type childrenFilter struct {
	stats
	depth  int64
	inputs []filter
}

func (f *childrenFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	pss := unpackFiltersAsSlice(f.inputs, nil)
	pids := []tPid{}
	getProcStatChildren(int(f.depth), pss, &pids)
	for _, pid := range pids {
		f.pid2Stat[pid] = stat(allProcStats[pid])
	}
	return nil
}

func (f *childrenFilter) Parse(p *Parser) error {
	// eg: [childre(]f1,f2,5)
	err := p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	// optional depth
	err = p.parseArgInt(&f.depth)
	if err == nil {
		if f.depth <= 0 {
			return p.syntaxError(fmt.Sprintf("depth in children filter must be >= 1, found '%d'", f.depth))
		}
	}
	if f.depth == 0 {
		// A depth of 0 means we want to get all descendants
		// But we fix a limit to avoid any fishy cycle in the gopsutil code
		f.depth = 1024
	}
	return p.parseSymbol(')')
}

func (f *childrenFilter) Stats() *stats {
	return &f.stats
}

/* Filters related to the command line (exe, args, cmdline)
 */

// Select matching command name (basename).
type cmdFilter struct {
	stats
	pat    *stregexp
	inputs []filter
}

func (f *cmdFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	for _, input := range f.inputs {
		iStats := input.Stats()
		for pid, s := range iStats.pid2Stat {
			pat, _ := s.Cmd()
			if err != nil {
				continue
			}
			if f.pat.matchString(pat) {
				f.pid2Stat[pid] = s
			}
		}
	}
	return nil
}

func (f *cmdFilter) Parse(p *Parser) error {
	// eg: name("apa.*",all)
	err := p.parseArgStregexp(&f.pat)
	if err != nil {
		return err
	}
	err = p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *cmdFilter) Stats() *stats {
	return &f.stats
}

// Select matching exe name (full path to command with dirname and basename).
type exeFilter struct {
	stats
	pat    *stregexp
	inputs []filter
}

func (f *exeFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	for _, input := range f.inputs {
		iStats := input.Stats()
		for pid, s := range iStats.pid2Stat {
			pat, _ := s.Exe()
			if err != nil {
				continue
			}
			if f.pat.matchString(pat) {
				f.pid2Stat[pid] = s
			}
		}
	}
	return nil
}

func (f *exeFilter) Parse(p *Parser) error {
	// eg: name("apa.*",all)
	err := p.parseArgStregexp(&f.pat)
	if err != nil {
		return err
	}
	err = p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *exeFilter) Stats() *stats {
	return &f.stats
}

// Select matching arguments (must match one of the arguments).
type argsFilter struct {
	stats
	pat    *stregexp
	inputs []filter
}

func (f *argsFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	for _, input := range f.inputs {
		iStats := input.Stats()
		for pid, s := range iStats.pid2Stat {
			args, _ := s.Args()
			if err != nil {
				continue
			}
			for _, pat := range args {
				if f.pat.matchString(pat) {
					f.pid2Stat[pid] = s
					break
				}
			}
		}
	}
	return nil
}

func (f *argsFilter) Parse(p *Parser) error {
	// eg: name("apa.*",all)
	err := p.parseArgStregexp(&f.pat)
	if err != nil {
		return err
	}
	err = p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *argsFilter) Stats() *stats {
	return &f.stats
}

// Select matching path (dirname of the command).
type pathFilter struct {
	stats
	pat    *stregexp
	inputs []filter
}

func (f *pathFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	for _, input := range f.inputs {
		iStats := input.Stats()
		for pid, s := range iStats.pid2Stat {
			pat, _ := s.Path()
			if err != nil {
				continue
			}
			if f.pat.matchString(pat) {
				f.pid2Stat[pid] = s
			}
		}
	}
	return nil
}

func (f *pathFilter) Parse(p *Parser) error {
	// eg: name("apa.*",all)
	err := p.parseArgStregexp(&f.pat)
	if err != nil {
		return err
	}
	err = p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *pathFilter) Stats() *stats {
	return &f.stats
}

// Select matching command line (the full command as one big string)
type cmdlineFilter struct {
	stats
	pat    *stregexp
	inputs []filter
}

func (f *cmdlineFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	for _, input := range f.inputs {
		iStats := input.Stats()
		for pid, s := range iStats.pid2Stat {
			pat, _ := s.CmdLine()
			if err != nil {
				continue
			}
			if f.pat.matchString(pat) {
				f.pid2Stat[pid] = s
			}
		}
	}
	return nil
}

func (f *cmdlineFilter) Parse(p *Parser) error {
	// eg: name("apa.*",all)
	err := p.parseArgStregexp(&f.pat)
	if err != nil {
		return err
	}
	err = p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *cmdlineFilter) Stats() *stats {
	return &f.stats
}

/* Filters based on set algebra.
 */

// Select and unpack the union of input filers.
type orFilter struct {
	stats
	inputs []filter
}

func (f *orFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	pss := f.stats.unpackAsSlice(nil)
	for _, ps := range pss {
		f.pid2Stat[ps.pid] = ps
	}
	return nil
}

func (f *orFilter) Parse(p *Parser) error {
	// eg: [or(]f1,f2,f3)
	err := p.parseArgFilterList(&f.inputs, 2)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *orFilter) Stats() *stats {
	return &f.stats
}

// Select and unpack the inersection of input filers.
type andFilter struct {
	stats
	inputs []filter
}

func (f *andFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	counts := map[*procStat]int{} // for
	for _, input := range f.inputs {
		pss := input.Stats().unpackAsSlice(nil)
		for _, ps := range pss {
			counts[ps]++
		}
	}
	intersect := map[tPid]stat{}
	li := len(f.inputs)
	for ps, count := range counts {
		if count == li {
			// This PID appears once per input filter.
			intersect[ps.pid] = stat(ps)
		}
	}
	f.pid2Stat = intersect
	return nil
}

func (f *andFilter) Parse(p *Parser) error {
	// eg: [and(]f1,f2,f3)
	err := p.parseArgFilterList(&f.inputs, 2)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *andFilter) Stats() *stats {
	return &f.stats
}

// Select and unpack the complement of input filers.
type notFilter struct {
	stats
	inputs []filter
}

func (f *notFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	psm := unpackFiltersAsMap(f.inputs, nil)
	complement := map[tPid]stat{}
	// All pids that are in allProcStats and not in input filter
	for pid, ps := range allProcStats {
		if _, in := psm[pid]; !in {
			complement[pid] = ps
		}
	}
	f.pid2Stat = complement
	return nil
}

func (f *notFilter) Parse(p *Parser) error {
	// eg: [not(]f1,f2,f3)
	err := p.parseArgFilterList(&f.inputs, 1)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *notFilter) Stats() *stats {
	return &f.stats
}

// Select and unpack the synthetic difference of input filers. (what is in exactly one input filer)
type differenceFilter struct {
	stats
	inputs []filter
}

func (f *differenceFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	counts := map[*procStat]int{}
	for _, input := range f.inputs {
		pss := input.Stats().unpackAsSlice(nil)
		for _, ps := range pss {
			counts[ps]++
		}
	}
	difference := map[tPid]stat{}
	for ps, count := range counts {
		if count == 1 {
			// This PID appears only once so it belongs ony to one filter.
			difference[ps.pid] = stat(ps)
		}
	}
	f.pid2Stat = difference
	return nil
}

func (f *differenceFilter) Parse(p *Parser) error {
	// eg: [difference(]f1,f2,f3)
	err := p.parseArgFilterList(&f.inputs, 2)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *differenceFilter) Stats() *stats {
	return &f.stats
}

// Aggregate/gather all input filter as one sythetic workload
type packFilter struct {
	stats  //
	inputs []filter
}

func (f *packFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	// Pack/Gather all input procStat.
	pss := unpackFiltersAsSlice(f.inputs, nil)
	s := NewPackStat(pss)
	// a pakcStat contains only one ID (but packed stats are in .elems)
	f.pid2Stat = map[tPid]stat{}
	f.pid2Stat[s.PID()] = s
	return nil
}

func (f *packFilter) Parse(p *Parser) error {
	// eg: [pack(]f1,f2,f3)
	err := p.parseArgFilterList(&f.inputs, 1)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *packFilter) Stats() *stats {
	return &f.stats
}

// Unpack the content of input filters. (get only real procStats by unpacking the packStat)
type unpackFilter struct {
	stats
	inputs []filter
}

func (f *unpackFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	for _, input := range f.inputs {
		iStats := input.Stats()
		for id, s := range iStats.pid2Stat {
			if id > 0 {
				f.pid2Stat[id] = s
			} else {
				// This is a packed stat, need to unpack
				ps := s.(*packStat)
				for _, ss := range ps.elems {
					pid := ss.PID()
					f.pid2Stat[pid] = s
				}
			}
		}
	}
	return nil
}

func (f *unpackFilter) Parse(p *Parser) error {
	// eg: [unpack(]f1,f2,f3)
	err := p.parseArgFilterList(&f.inputs, 1)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *unpackFilter) Stats() *stats {
	return &f.stats
}

// Pack processes using a group criteria. eg packby(user) will create one packStat per user that will contain all processes for this user.
type packByFilter struct {
	stats
	inputs []filter
	by     string // criteria to pack by
}

func (f *packByFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	applyAll(f.inputs)
	pss := unpackFiltersAsSlice(f.inputs, nil)
	switch f.by {
	case "user":
		mby := map[int32]*packStat{}
		for _, ps := range pss {
			ids, err := ps.UIDs()
			if err != nil {
				continue
			}
			v := ids[0]
			if packStat, known := mby[v]; known {
				// Already have a packStat for this user. Append to it.
				packStat.elems = append(packStat.elems, ps)
			} else {
				// New value, create a new packStat for all procStats with that value.
				packStat = NewPackStat([]*procStat{ps})
				packStat.uid = v
				mby[v] = packStat
				f.pid2Stat[packStat.pid] = stat(packStat)
			}
		}
	case "group":
		mby := map[int32]*packStat{}
		for _, ps := range pss {
			ids, err := ps.GIDs()
			if err != nil {
				continue
			}
			v := ids[0]
			if packStat, known := mby[v]; known {
				// Already have a packStat for this user. Append to it.
				packStat.elems = append(packStat.elems, ps)
			} else {
				// New value, create a new packStat for all procStats with that value.
				packStat = NewPackStat([]*procStat{ps})
				packStat.gid = v
				mby[v] = packStat
				f.pid2Stat[packStat.pid] = stat(packStat)
			}
		}
	case "cmd":
		mby := map[string]*packStat{}
		for _, ps := range pss {
			v, err := ps.Cmd()
			if err != nil {
				continue
			}
			if packStat, known := mby[v]; known {
				// Already have a packStat for this user. Append to it.
				packStat.elems = append(packStat.elems, ps)
			} else {
				// New value, create a new packStat for all procStats with that value.
				packStat = NewPackStat([]*procStat{ps})
				packStat.cmd = v
				mby[v] = packStat
				f.pid2Stat[packStat.pid] = stat(packStat)
			}
		}
	default:
		return fmt.Errorf("unknow packby criteria %q", f.by)

	}
	return nil
}

func (f *packByFilter) Parse(p *Parser) error {
	// eg: [packby(]user,f1,f2,f3)
	err := p.parseArgIdentifier(&f.by)
	if err != nil {
		return err
	}
	err = p.parseArgFilterList(&f.inputs, 0)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *packByFilter) Stats() *stats {
	return &f.stats
}

// Select filters by their name.
type filtersFilter struct {
	stats
	pat    *stregexp
	inputs []filter
}

func (f *filtersFilter) Apply() error {
	if !f.stats.reset() {
		return nil
	}
	inputs := []filter{}
	for name, filter := range currentParser.n2f {
		if f.pat.matchString(name) {
			inputs = append(inputs, filter)
		}
	}
	f.inputs = inputs
	err := applyAll(f.inputs)
	if err != nil {
		return err
	}
	// Collect all stats in all inputs (no unpack but filter out the special "other" packStas
	for _, input := range f.inputs {
		iStats := input.Stats()
		for pid, s := range iStats.pid2Stat {
			if pid < 0 && s.(*packStat).other != "" {
				continue
			}
			f.pid2Stat[pid] = s
		}
	}
	return nil
}

func (f *filtersFilter) Parse(p *Parser) error {
	err := p.parseArgStregexp(&f.pat)
	if err != nil {
		return err
	}
	return p.parseSymbol(')')
}

func (f *filtersFilter) Stats() *stats {
	return &f.stats
}
