package procstat

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf/filter"
	"github.com/shirou/gopsutil/v3/process"
)

type Filter struct {
	Name            string   `toml:"name"`
	PidFiles        []string `toml:"pid_files"`
	Patterns        []string `toml:"patterns"`
	Users           []string `toml:"users"`
	CGroups         []string `toml:"cgroups"`
	SystemdUnits    []string `toml:"systemd_units"`
	SupervisorUnits []string `toml:"supervisor_units"`
	WinService      []string `toml:"win_services"`
	RecursionDepth  int      `toml:"recursion_depth"`

	filterCmds           []*regexp.Regexp
	filterUser           filter.Filter
	filterSupervisorUnit string
}

func (f *Filter) Init() error {
	if f.Name == "" {
		return errors.New("filter must be named")
	}

	// Check for only one service selector being active
	var active []string
	if len(f.PidFiles) > 0 {
		active = append(active, "pid_files")
	}
	if len(f.CGroups) > 0 {
		active = append(active, "cgroups")
	}
	if len(f.SystemdUnits) > 0 {
		active = append(active, "systemd_units")
	}
	if len(f.SupervisorUnits) > 0 {
		active = append(active, "supervisor_units")
	}
	if len(f.WinService) > 0 {
		active = append(active, "win_services")
	}
	if len(active) > 1 {
		return fmt.Errorf("cannot select multiple services %q", strings.Join(active, ", "))
	}

	// Prepare the filters
	f.filterCmds = make([]*regexp.Regexp, 0, len(f.Patterns))
	for _, p := range f.Patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("compiling pattern %q of filter %q failed: %w", p, f.Name, err)
		}
		f.filterCmds = append(f.filterCmds, re)
	}

	var err error
	if f.filterUser, err = filter.Compile(f.Users); err != nil {
		return fmt.Errorf("compiling users of filter %q failed: %w", f.Name, err)
	}
	f.filterSupervisorUnit = strings.TrimSpace(strings.Join(f.SupervisorUnits, " "))

	return nil
}

func (f *Filter) ApplyFilter() ([]processGroup, error) {
	// Determine processes on service level. if there is no constraint on the
	// services, use all processes for matching.
	var groups []processGroup
	switch {
	case len(f.PidFiles) > 0:
		g, err := findByPidFiles(f.PidFiles)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g...)
	case len(f.CGroups) > 0:
		g, err := findByCgroups(f.CGroups)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g...)
	case len(f.SystemdUnits) > 0:
		g, err := findBySystemdUnits(f.CGroups)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g...)
	case f.filterSupervisorUnit != "":
		g, err := findBySupervisorUnits(f.filterSupervisorUnit)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g...)
	case len(f.WinService) > 0:
		g, err := findByWindowsServices(f.WinService)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g...)
	default:
		procs, err := process.Processes()
		if err != nil {
			return nil, err
		}
		groups = append(groups, processGroup{processes: procs, tags: make(map[string]string)})
	}

	// Filter by additional properties such as users, patterns etc
	result := make([]processGroup, 0, len(groups))
	for _, g := range groups {
		var matched []*process.Process
		for _, p := range g.processes {
			// Users
			if f.filterUser != nil {
				username, err := p.Username()
				if err != nil {
					// This can happen if we don't have permissions or the process no longer exists
					continue
				}
				if !f.filterUser.Match(username) {
					continue
				}
			}

			// Patterns
			if len(f.filterCmds) > 0 {
				cmd, err := p.Cmdline()
				if err != nil {
					// This can happen if we don't have permissions or the process no longer exists
					continue
				}
				var found bool
				for _, re := range f.filterCmds {
					if re.MatchString(cmd) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			matched = append(matched, p)
		}
		result = append(result, processGroup{processes: matched, tags: g.tags})
	}

	// Resolve children down to the requested depth
	previous := result
	for depth := 0; depth < f.RecursionDepth || f.RecursionDepth < 0; depth++ {
		children := make([]processGroup, 0, len(previous))
		for _, group := range previous {
			for _, p := range group.processes {
				c, err := getChildren(p)
				if err != nil {
					return nil, fmt.Errorf("unable to get children of process %d: %w", p.Pid, err)
				}
				if len(c) == 0 {
					continue
				}

				tags := make(map[string]string, len(group.tags)+1)
				for k, v := range group.tags {
					tags[k] = v
				}
				tags["parent_pid"] = strconv.FormatInt(int64(p.Pid), 10)
				children = append(children, processGroup{
					processes: c,
					tags:      tags,
				})
			}
		}
		if len(children) == 0 {
			break
		}
		result = append(result, children...)
		previous = children
	}

	return result, nil
}

func getChildren(p *process.Process) ([]*process.Process, error) {
	children, err := p.Children()
	// Check for cases that do not really mean error but rather means that there
	// is no match.
	switch {
	case err == nil,
		errors.Is(err, process.ErrorNoChildren),
		strings.Contains(err.Error(), "exit status 1"):
		return children, nil
	}
	return nil, fmt.Errorf("unable to get children of process %d: %w", p.Pid, err)
}
