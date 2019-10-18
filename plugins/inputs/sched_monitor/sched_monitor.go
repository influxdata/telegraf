package sched_monitor

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	globTaskPath = "/proc/[0-9]*/task/*"
	taskStatFmt  = "/proc/%d/task/%d/stat"
	taskSchedFmt = "/proc/%d/task/%d/sched"

	sampleConfig = `
	  ## A list of cpus to collect data from. heT format of this list is:
	  ## <cpu number>,...,<cpu number> 
	  ## or: 
	  ## <cpu number>-<cpu number> (must be a positive range in ascending order)
	  ## or a mixture:
	  ## <cpu number>,...,<cpu number>-<cpu number>
	  ## For example:
	  ## cpu_list = "1,3,5-8,12"
	  ##
	  ## The default is empty list
	  # cpu_list = 

	  ## To filter out kernel threads set the following to true (the default is false):
	  # exclude_kernel = false
	`
)

var (
	taskPathExpr  = regexp.MustCompile(`/proc/(?P<pid>\d+)/task/(?P<tid>\d+)`)
	tasks         = map[int]taskInfo{}
	currentTids   = map[int]bool{}
	monitoredCPUS = map[int]bool{}
	initialized   = false
)

type SchedMonitor struct {
	CPUList       string `toml:"cpu_list"`
	ExcludeKernel bool   `toml:"exclude_kernel"`
}

type taskInfo struct {
	name           string
	pid            int
	tid            int
	cpuTimeNanos   int64
	voluntaryCtx   int64
	involuntaryCtx int64
}

func (s *SchedMonitor) Gather(acc telegraf.Accumulator) error {
	if !initialized {
		initializeState(s.CPUList, s.ExcludeKernel)
	}

	collectData(s.ExcludeKernel, acc)
	return nil
}

func (s *SchedMonitor) Description() string {
	return "This plugin collects per-thread scheduling statistics (cpu time, voluntary/involuntary context switches) for configured CPUs"
}

func (s *SchedMonitor) SampleConfig() string {
	return sampleConfig
}

// iterates over all threads running on the host and collects their sched stats
func collectData(excludeKernel bool, acc telegraf.Accumulator) {
	iterateTasks(excludeKernel, func(pid int, tid int, cmd string, cpu int, cpuTime int64, ctxSwitch int64, invlCtxSwitch int64) {
		if task, found := tasks[tid]; found {
			if cpuTime == task.cpuTimeNanos {
				return // task hasn't been scheduled since the last check; nothing to report
			}

			// report deltas
			fields := map[string]interface{}{
				"cpu_time":       cpuTime - task.cpuTimeNanos,
				"ctx_swtch":      ctxSwitch - task.voluntaryCtx,
				"invl_ctx_swtch": invlCtxSwitch - task.involuntaryCtx,
			}

			tags := map[string]string{
				"cmd": cmd,
				"cpu": strconv.Itoa(cpu),
			}

			acc.AddFields("sched_monitor", fields, tags)

			// update task stats
			task.cpuTimeNanos = cpuTime
			task.voluntaryCtx = ctxSwitch
			task.involuntaryCtx = invlCtxSwitch
		}
	})

	// remove dead tasks
	for tid := range tasks {
		if _, found := currentTids[tid]; !found {
			delete(tasks, tid)
		}
	}

}

func initializeState(cpuList string, excludeKernel bool) {
	parseCPUList(cpuList)
	initialSnapshot(excludeKernel)

	initialized = true
}

// parses string cpu list and collects initial snapshot of sched stats for all threads
func initialSnapshot(excludeKernel bool) {
	iterateTasks(excludeKernel, func(pid int, tid int, cmd string, cpu int, cpuTime int64, ctxSwitch int64, invlCtxSwitch int64) {
		tasks[tid] = taskInfo{
			name:           cmd,
			pid:            pid,
			tid:            tid,
			cpuTimeNanos:   cpuTime,
			voluntaryCtx:   ctxSwitch,
			involuntaryCtx: invlCtxSwitch,
		}
	})
}

// iterates over all threads and executes a processing function for each one
func iterateTasks(excludeKernel bool, process func(int, int, string, int, int64, int64, int64)) {
	taskPaths, _ := filepath.Glob(globTaskPath)

	for _, taskPath := range taskPaths {
		pid, tid := decodeTaskPath(taskPath)
		currentTids[tid] = true

		cpu := lastUsedCPU(pid, tid)
		if cpu < 0 {
			continue
		}

		if _, found := monitoredCPUS[cpu]; found {
			schedFile := openSchedFile(pid, tid)
			cmd, cpuTime, ctxSwitch, invlCtxSwitch := parseSchedStats(schedFile)
			if (excludeKernel && isKernelTask(cmd)) || cmd == "" {
				continue
			}

			process(pid, tid, cmd, cpu, cpuTime, ctxSwitch, invlCtxSwitch)
		}
	}
}

func openSchedFile(pid int, tid int) io.Reader {
	schedFile, _ := os.Open(fmt.Sprintf(taskSchedFmt, pid, tid))
	defer schedFile.Close()

	return schedFile
}

func isKernelTask(cmd string) bool {
	return strings.HasPrefix(cmd, "[") && strings.HasSuffix(cmd, "]")
}

func parseCPUList(cpuList string) {
	tokens := strings.Split(cpuList, ",")

	for _, s := range tokens {
		if strings.Contains(s, "-") {
			fromTo := strings.Split(s, "-")
			fromCPU, _ := strconv.Atoi(fromTo[0])
			toCPU, _ := strconv.Atoi(fromTo[1])
			for i := fromCPU; i <= toCPU; i++ {
				monitoredCPUS[i] = true
			}
		} else {
			cpu, _ := strconv.Atoi(s)
			monitoredCPUS[cpu] = true
		}
	}
}

func decodeTaskPath(taskPath string) (int, int) {
	match := taskPathExpr.FindStringSubmatch(taskPath)
	pid, _ := strconv.Atoi(match[1])
	tid, _ := strconv.Atoi(match[2])

	return pid, tid
}

// parses sched file for a given thread
func parseSchedStats(file io.Reader) (cmd string, cpuTime int64, ctxSwich int64, invlCtxSwitch int64) {
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		cmd = strings.Fields(scanner.Text())[0]
	}

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 2 {
			continue
		}

		metric := strings.TrimSpace(fields[0])
		strValue := strings.TrimSpace(fields[1])

		if strings.Contains(metric, "sum_exec_runtime") {
			value, _ := strconv.ParseFloat(strValue, 64)
			cpuTime = int64(value * 1_000_000)
			continue
		} else if strings.Contains(metric, "nr_voluntary") {
			value, _ := strconv.ParseInt(strValue, 10, 64)
			ctxSwich = value
			continue
		} else if strings.Contains(metric, "nr_involuntary") {
			value, _ := strconv.ParseInt(strValue, 10, 64)
			invlCtxSwitch = value
			break
		}
	}

	return
}

func lastUsedCPU(pid int, tid int) int {
	f, _ := os.Open(fmt.Sprintf(taskStatFmt, pid, tid))
	defer f.Close()

	buf, _ := ioutil.ReadAll(f)
	if len(buf) == 0 {
		return -1
	}

	columns := strings.Split(string(buf), " ")
	cpu, _ := strconv.Atoi(columns[38])

	return cpu
}

func init() {
	inputs.Add("sched_monitor", func() telegraf.Input {
		return &SchedMonitor{}
	})
}
