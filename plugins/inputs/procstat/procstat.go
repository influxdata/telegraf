//go:generate ../../../tools/readme_config_includer/generator
package procstat

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command
var availableTags = []string{"cmdline", "pid", "ppid", "status", "user"}

type PID int32

type Procstat struct {
	PidFinder              string          `toml:"pid_finder"`
	PidFile                string          `toml:"pid_file"`
	Exe                    string          `toml:"exe"`
	Pattern                string          `toml:"pattern"`
	Prefix                 string          `toml:"prefix"`
	CmdLineTag             bool            `toml:"cmdline_tag" deprecated:"1.29.0;use 'tag_with' instead"`
	ProcessName            string          `toml:"process_name"`
	User                   string          `toml:"user"`
	SystemdUnit            string          `toml:"systemd_unit"`
	SupervisorUnit         []string        `toml:"supervisor_unit" deprecated:"1.29.0;use 'supervisor_units' instead"`
	SupervisorUnits        []string        `toml:"supervisor_units"`
	IncludeSystemdChildren bool            `toml:"include_systemd_children"`
	CGroup                 string          `toml:"cgroup"`
	PidTag                 bool            `toml:"pid_tag" deprecated:"1.29.0;use 'tag_with' instead"`
	WinService             string          `toml:"win_service"`
	Mode                   string          `toml:"mode"`
	TagWith                []string        `toml:"tag_with"`
	Log                    telegraf.Logger `toml:"-"`

	solarisMode bool
	finder      PIDFinder
	processes   map[PID]Process
	tagging     map[string]bool

	createProcess func(PID) (Process, error)
}

type PidsTags struct {
	PIDs []PID
	Tags map[string]string
}

func (*Procstat) SampleConfig() string {
	return sampleConfig
}

func (p *Procstat) Init() error {
	// Check solaris mode
	p.solarisMode = strings.EqualFold(p.Mode, "solaris")

	// Keep the old settings for compatibility
	if p.PidTag && !choice.Contains("pid", p.TagWith) {
		p.TagWith = append(p.TagWith, "pid")
	}
	if p.CmdLineTag && !choice.Contains("cmdline", p.TagWith) {
		p.TagWith = append(p.TagWith, "cmdline")
	}

	// Check tagging and setup LUT
	if err := choice.CheckSlice(p.TagWith, availableTags); err != nil {
		return fmt.Errorf("invalid tag_with setting: %w", err)
	}
	p.tagging = make(map[string]bool, len(p.TagWith))
	for _, tag := range p.TagWith {
		p.tagging[tag] = true
	}

	// Keep the old settings for compatibility
	for _, u := range p.SupervisorUnit {
		if !choice.Contains(u, p.SupervisorUnits) {
			p.SupervisorUnits = append(p.SupervisorUnits, u)
		}
	}

	// Check filtering
	switch {
	case len(p.SupervisorUnits) > 0, p.SystemdUnit != "", p.WinService != "",
		p.CGroup != "", p.PidFile != "", p.Exe != "", p.Pattern != "",
		p.User != "":
		// Do nothing as those are valid settings
	default:
		return errors.New("require filter option but none set")
	}

	// Instantiate the finder
	switch p.PidFinder {
	case "", "pgrep":
		p.PidFinder = "pgrep"
		finder, err := newPgrepFinder()
		if err != nil {
			return fmt.Errorf("creating pgrep finder failed: %w", err)
		}
		p.finder = finder
	case "native":
		// gopsutil relies on pgrep when looking up children on darwin
		// see https://github.com/shirou/gopsutil/blob/v3.23.10/process/process_darwin.go#L235
		requiresChildren := len(p.SupervisorUnits) > 0 && p.Pattern != ""
		if requiresChildren && runtime.GOOS == "darwin" {
			return errors.New("configuration requires 'pgrep' finder on your OS")
		}
		p.finder = &NativeFinder{}
	case "test":
		p.Log.Warn("running in test mode")
	default:
		return fmt.Errorf("unknown pid_finder %q", p.PidFinder)
	}

	// Initialize the running process cache
	p.processes = make(map[PID]Process)

	return nil
}

func (p *Procstat) Gather(acc telegraf.Accumulator) error {
	now := time.Now()
	results, err := p.findPids()
	if err != nil {
		// Add lookup error-metric
		fields := map[string]interface{}{
			"pid_count":   0,
			"running":     0,
			"result_code": 1,
		}
		tags := map[string]string{
			"pid_finder": p.PidFinder,
			"result":     "lookup_error",
		}
		for _, pidTag := range results {
			for key, value := range pidTag.Tags {
				tags[key] = value
			}
		}
		acc.AddFields("procstat_lookup", fields, tags, now)
		return err
	}

	var count int
	running := make(map[PID]bool)
	for _, r := range results {
		if len(r.PIDs) < 1 && len(p.SupervisorUnits) > 0 {
			continue
		}
		count += len(r.PIDs)
		for _, pid := range r.PIDs {
			// Use the cached processes as we need the existing instances
			// to compute delta-metrics (e.g. cpu-usage).
			proc, found := p.processes[pid]
			if !found {
				// We've found a process that was not recorded before so add it
				// to the list of processes
				proc, err = p.createProcess(pid)
				if err != nil {
					// No problem; process may have ended after we found it
					continue
				}
				// Assumption: if a process has no name, it probably does not exist
				if name, _ := proc.Name(); name == "" {
					continue
				}

				// Add initial tags
				for k, v := range r.Tags {
					proc.SetTag(k, v)
				}

				if p.ProcessName != "" {
					proc.SetTag("process_name", p.ProcessName)
				}
				p.processes[pid] = proc
			}
			running[pid] = true
			m := proc.Metric(p.Prefix, p.tagging, p.solarisMode)
			m.SetTime(now)
			acc.AddMetric(m)
		}
	}

	// Cleanup processes that are not running anymore
	for pid := range p.processes {
		if !running[pid] {
			delete(p.processes, pid)
		}
	}

	// Add lookup statistics-metric
	fields := map[string]interface{}{
		"pid_count":   count,
		"running":     len(running),
		"result_code": 0,
	}
	tags := map[string]string{
		"pid_finder": p.PidFinder,
		"result":     "success",
	}
	for _, pidTag := range results {
		for key, value := range pidTag.Tags {
			tags[key] = value
		}
	}
	if len(p.SupervisorUnits) > 0 {
		tags["supervisor_unit"] = strings.Join(p.SupervisorUnits, ";")
	}
	acc.AddFields("procstat_lookup", fields, tags, now)

	return nil
}

// Get matching PIDs and their initial tags
func (p *Procstat) findPids() ([]PidsTags, error) {
	switch {
	case len(p.SupervisorUnits) > 0:
		return p.findSupervisorUnits()
	case p.SystemdUnit != "":
		return p.systemdUnitPIDs()
	case p.WinService != "":
		pids, err := p.winServicePIDs()
		if err != nil {
			return nil, err
		}
		tags := map[string]string{"win_service": p.WinService}
		return []PidsTags{{pids, tags}}, nil
	case p.CGroup != "":
		return p.cgroupPIDs()
	case p.PidFile != "":
		pids, err := p.finder.PidFile(p.PidFile)
		if err != nil {
			return nil, err
		}
		tags := map[string]string{"pidfile": p.PidFile}
		return []PidsTags{{pids, tags}}, nil
	case p.Exe != "":
		pids, err := p.finder.Pattern(p.Exe)
		if err != nil {
			return nil, err
		}
		tags := map[string]string{"exe": p.Exe}
		return []PidsTags{{pids, tags}}, nil
	case p.Pattern != "":
		pids, err := p.finder.FullPattern(p.Pattern)
		if err != nil {
			return nil, err
		}
		tags := map[string]string{"pattern": p.Pattern}
		return []PidsTags{{pids, tags}}, nil
	case p.User != "":
		pids, err := p.finder.UID(p.User)
		if err != nil {
			return nil, err
		}
		tags := map[string]string{"user": p.User}
		return []PidsTags{{pids, tags}}, nil
	}
	return nil, errors.New("no filter option set")
}

func (p *Procstat) findSupervisorUnits() ([]PidsTags, error) {
	groups, groupsTags, err := p.supervisorPIDs()
	if err != nil {
		return nil, fmt.Errorf("getting supervisor PIDs failed: %w", err)
	}

	// According to the PID, find the system process number and get the child processes
	pidTags := make([]PidsTags, 0, len(groups))
	for _, group := range groups {
		grppid := groupsTags[group]["pid"]
		if grppid == "" {
			pidTags = append(pidTags, PidsTags{nil, groupsTags[group]})
			continue
		}

		pid, err := strconv.ParseInt(grppid, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("converting PID %q failed: %w", grppid, err)
		}

		// Get all children of the supervisor unit
		pids, err := p.finder.Children(PID(pid))
		if err != nil {
			return nil, fmt.Errorf("getting children for %d failed: %w", pid, err)
		}
		tags := map[string]string{"pattern": p.Pattern, "parent_pid": p.Pattern}

		// Handle situations where the PID does not exist
		if len(pids) == 0 {
			continue
		}

		// Merge tags map
		for k, v := range groupsTags[group] {
			_, ok := tags[k]
			if !ok {
				tags[k] = v
			}
		}
		// Remove duplicate pid tags
		delete(tags, "pid")
		pidTags = append(pidTags, PidsTags{pids, tags})
	}
	return pidTags, nil
}

func (p *Procstat) supervisorPIDs() ([]string, map[string]map[string]string, error) {
	out, err := execCommand("supervisorctl", "status", strings.Join(p.SupervisorUnits, " ")).Output()
	if err != nil {
		if !strings.Contains(err.Error(), "exit status 3") {
			return nil, nil, err
		}
	}
	lines := strings.Split(string(out), "\n")
	// Get the PID, running status, running time and boot time of the main process:
	// pid 11779, uptime 17:41:16
	// Exited too quickly (process log may have details)
	mainPids := make(map[string]map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}

		kv := strings.Fields(line)
		if len(kv) < 2 {
			// Not a key-value pair
			continue
		}
		name := kv[0]

		statusMap := map[string]string{
			"supervisor_unit": name,
			"status":          kv[1],
		}

		switch kv[1] {
		case "FATAL", "EXITED", "BACKOFF", "STOPPING":
			statusMap["error"] = strings.Join(kv[2:], " ")
		case "RUNNING":
			statusMap["pid"] = strings.ReplaceAll(kv[3], ",", "")
			statusMap["uptimes"] = kv[5]
		case "STOPPED", "UNKNOWN", "STARTING":
			// No additional info
		}
		mainPids[name] = statusMap
	}

	return p.SupervisorUnits, mainPids, nil
}

func (p *Procstat) systemdUnitPIDs() ([]PidsTags, error) {
	if p.IncludeSystemdChildren {
		p.CGroup = fmt.Sprintf("systemd/system.slice/%s", p.SystemdUnit)
		return p.cgroupPIDs()
	}

	var pidTags []PidsTags
	pids, err := p.simpleSystemdUnitPIDs()
	if err != nil {
		return nil, err
	}
	tags := map[string]string{"systemd_unit": p.SystemdUnit}
	pidTags = append(pidTags, PidsTags{pids, tags})
	return pidTags, nil
}

func (p *Procstat) simpleSystemdUnitPIDs() ([]PID, error) {
	out, err := execCommand("systemctl", "show", p.SystemdUnit).Output()
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(out, []byte{'\n'})
	pids := make([]PID, 0, len(lines))
	for _, line := range lines {
		kv := bytes.SplitN(line, []byte{'='}, 2)
		if len(kv) != 2 {
			continue
		}
		if !bytes.Equal(kv[0], []byte("MainPID")) {
			continue
		}
		if len(kv[1]) == 0 || bytes.Equal(kv[1], []byte("0")) {
			return nil, nil
		}
		pid, err := strconv.ParseInt(string(kv[1]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid pid %q", kv[1])
		}
		pids = append(pids, PID(pid))
	}

	return pids, nil
}

func (p *Procstat) cgroupPIDs() ([]PidsTags, error) {
	procsPath := p.CGroup
	if procsPath[0] != '/' {
		procsPath = "/sys/fs/cgroup/" + procsPath
	}

	items, err := filepath.Glob(procsPath)
	if err != nil {
		return nil, fmt.Errorf("glob failed: %w", err)
	}

	pidTags := make([]PidsTags, 0, len(items))
	for _, item := range items {
		pids, err := p.singleCgroupPIDs(item)
		if err != nil {
			return nil, err
		}
		tags := map[string]string{"cgroup": p.CGroup, "cgroup_full": item}
		pidTags = append(pidTags, PidsTags{pids, tags})
	}

	return pidTags, nil
}

func (p *Procstat) singleCgroupPIDs(path string) ([]PID, error) {
	ok, err := isDir(path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not a directory %s", path)
	}
	procsPath := filepath.Join(path, "cgroup.procs")
	out, err := os.ReadFile(procsPath)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(out, []byte{'\n'})
	pids := make([]PID, 0, len(lines))
	for _, pidBS := range lines {
		if len(pidBS) == 0 {
			continue
		}
		pid, err := strconv.ParseInt(string(pidBS), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid pid %q", pidBS)
		}
		pids = append(pids, PID(pid))
	}

	return pids, nil
}

func isDir(path string) (bool, error) {
	result, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return result.IsDir(), nil
}

func (p *Procstat) winServicePIDs() ([]PID, error) {
	var pids []PID

	pid, err := queryPidWithWinServiceName(p.WinService)
	if err != nil {
		return pids, err
	}

	pids = append(pids, PID(pid))

	return pids, nil
}

func init() {
	inputs.Add("procstat", func() telegraf.Input {
		return &Procstat{createProcess: newProc}
	})
}
