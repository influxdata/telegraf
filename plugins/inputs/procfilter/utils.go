package procfilter

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)

/* A dual string/regexp object
A r suffix denotes a regular expression rather than a plain string.
'fu' matches exactly fu
'fu'r matches anything containing fu
'^fu'r matches anything starting with fu
*/
type stregexp struct {
	isRe   bool
	invert bool // Invert the match.
	pat    string
	re     *regexp.Regexp // the compiled version of the user string
}

func NewStregexp(pat string, isRe bool, invert bool) (*stregexp, error) {
	var sr stregexp

	if isRe {
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		sr.re = re
	}
	sr.isRe = isRe
	sr.invert = invert
	sr.pat = pat
	return &sr, nil
}

func (sr *stregexp) matchString(s string) bool {
	if !sr.isRe {
		// plain string compare
		return sr.invert != (sr.pat == s) // booleas XOR
	}
	// regexp match
	return sr.invert != sr.re.MatchString(s) // booleas XOR
}

func GIDName(gid int32) string {
	group, err := user.LookupGroupId(strconv.Itoa(int(gid)))
	if err != nil {
		return ""
	}
	return group.Name
}

func UIDName(uid int32) string {
	user, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		return ""
	}
	return user.Username
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func logErr(msg string) {
	log.Printf("E! procfilter: %s", msg)
}

func NYIError(msg string) error {
	return fmt.Errorf("%s not yet implemented", msg)
}

// pidFromFile read a PID from a file.
func pidFromFile(file string) (tPid, error) {
	pidString, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, fmt.Errorf("cannot get PID stored in file '%s', %s", file, err.Error())
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidString)))
	if err != nil {
		return 0, fmt.Errorf("cannot get PID stored in file '%s', %s", file, err.Error())
	}
	return tPid(pid), nil
}

func fileContent(file string) (string, error) {
	scriptString, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("cannot get the script stored in file '%s', %s", file, err.Error())
	}
	return string(scriptString), nil
}

// applyAll calls apply for all filters.
func applyAll(filters []filter) error {
	for _, f := range filters {
		err := f.Apply()
		if err != nil {
			return err
		}
	}
	return nil
}

// unpackSliceAsMap convert a slice of stats to an unpacked map of pid=>*procStat.If sm is not nil, add to this map.
func unpackSliceAsMap(stats []stat, sm map[tPid]*procStat) map[tPid]*procStat {
	if sm == nil {
		sm = map[tPid]*procStat{}
	}
	for _, s := range stats {
		id := s.PID()
		if id >= 0 {
			// A procStat
			sm[id] = s.(*procStat)
		} else {
			// This is a packed stat, need to unpack
			ps := s.(*packStat)
			if ps.other != "" {
				continue
			}
			for _, ss := range ps.elems {
				pid := ss.PID()
				sm[pid] = s.(*procStat)
			}
		}
	}
	return sm
}

// unpackSliceAsSlice convert a slice of stats to an unpacked slice of *procStat.. If ss is not nil, append to ss.
func unpackSliceAsSlice(stats []stat, ss []*procStat) []*procStat {
	if ss == nil {
		ss = []*procStat{}
	}
	for _, s := range stats {
		id := s.PID()
		if id >= 0 {
			// A procStat
			ss = append(ss, s.(*procStat))
		} else {
			// This is a packed stat, need to unpack
			ps := s.(*packStat)
			if ps.other != "" {
				continue
			}
			for _, subs := range ps.elems {
				ss = append(ss, subs)
			}
		}
	}
	return ss
}

// unpackMapAsMap convert a map of pid=>stats to an unpacked map of pid=>*procStat. If sm is not nil, add to this map.
func unpackMapAsMap(stats map[tPid]stat, sm map[tPid]*procStat) map[tPid]*procStat {
	if sm == nil {
		sm = map[tPid]*procStat{}
	}
	for id, s := range stats {
		if id >= 0 {
			// A procStat
			sm[id] = s.(*procStat)
		} else {
			// This is a packed stat, need to unpack
			ps := s.(*packStat)
			if ps.other != "" {
				continue
			}
			for _, ss := range ps.elems {
				pid := ss.PID()
				sm[pid] = ss
			}
		}
	}
	return sm
}

// unpackMapAsSlice convert a map of pid=>stats to an unpacked slice of *procStat. If ss is not nil, append to ss.
func unpackMapAsSlice(stats map[tPid]stat, ss []*procStat) []*procStat {
	if ss == nil {
		ss = []*procStat{}
	}
	for id, s := range stats {
		if id >= 0 {
			// A procStat
			ss = append(ss, s.(*procStat))
		} else {
			// This is a packed stat, need to unpack
			ps := s.(*packStat)
			if ps.other != "" {
				continue
			}
			for _, subs := range ps.elems {
				ss = append(ss, subs)
			}
		}
	}
	return ss
}

// unpackStatAsMap unpack a stat to an unpacked map of pid=>*procStat.If sm is not nil, add to this map.
func unpackStatAsMap(s stat, sm map[tPid]*procStat) map[tPid]*procStat {
	if sm == nil {
		sm = map[tPid]*procStat{}
	}
	id := s.PID()
	if id >= 0 {
		// A procStat
		sm[id] = s.(*procStat)
	} else {
		// This is a packed stat, need to unpack
		ps := s.(*packStat)
		if ps.other != "" {
			return sm
		}
		for _, ss := range ps.elems {
			pid := ss.PID()
			sm[pid] = s.(*procStat)
		}
	}
	return sm
}

// unpackStatAsSlice unpack a stat to a slice of *procStat. If ss is not nil, append to ss.
func unpackStatAsSlice(s stat, ss []*procStat) []*procStat {
	if ss == nil {
		ss = []*procStat{}
	}
	id := s.PID()
	if id >= 0 {
		// A procStat
		ss = append(ss, s.(*procStat))
	} else {
		// This is a packed stat, need to unpack
		ps := s.(*packStat)
		if ps.other != "" {
			return ss
		}
		for _, subs := range ps.elems {
			ss = append(ss, subs)
		}
	}
	return ss
}

func unpackFiltersAsMap(filters []filter, sm map[tPid]*procStat) map[tPid]*procStat {
	if sm == nil {
		sm = map[tPid]*procStat{}
	}
	for _, f := range filters {
		sm = f.Stats().unpackAsMap(sm)
	}
	return sm
}

func unpackFiltersAsSlice(filters []filter, ss []*procStat) []*procStat {
	if ss == nil {
		ss = []*procStat{}
	}
	// We use an intermediat map to be sure that no *procStat is twice in the final slice.
	sm := unpackFiltersAsMap(filters, nil)
	for _, ps := range sm {
		ss = append(ss, ps)
	}
	return ss
}
