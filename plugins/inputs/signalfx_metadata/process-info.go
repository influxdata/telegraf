package signalfxmetadata

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/process"
)

// NewProcessInfo - returns a new ProcessInfo instance
func NewProcessInfo(bufferSize int, numWorkers int) *ProcessInfo {
	var s = &ProcessInfo{
		processes:  make(map[int32]*process.Process),
		processIn:  make(chan *workerInProcess, bufferSize),
		bufferSize: bufferSize,
		numWorkers: numWorkers,
	}
	// ensure that the number of workers is always 1
	if s.numWorkers < 1 {
		s.numWorkers = 1
	}
	for i := 0; i < s.numWorkers; i++ {
		newWorkerProcess(s.processIn)
	}

	return s
}

// ProcessInfo - list of processes
type ProcessInfo struct {
	processes  map[int32]*process.Process
	processIn  chan *workerInProcess
	bufferSize int
	numWorkers int
}

// GetTop - returns a map of process information
func (s *ProcessInfo) GetTop() (response string, err error) {
	start := time.Now()
	var pids []int32
	var compressed bytes.Buffer
	pids, err = process.Pids()
	// store process instances in s.processes to collect accurate %cpu info
	if err == nil {
		var pidList = make(map[int32]bool, len(pids))
		var top = make(map[string][]interface{}, len(pids))
		var output = make(chan *workerOutProcess, s.bufferSize)

		// Add missing processes to process list
		for _, pid := range pids {
			pidList[pid] = true
			if _, isIn := s.processes[pid]; !isIn {
				if proc, er := process.NewProcess(pid); er == nil {
					s.processes[pid] = proc
				}
			}
		}

		// use separate go routine to push processes on to worker threads
		for pid, proc := range s.processes {
			// Remove dead processes from process list
			if _, in := pidList[pid]; !in {
				delete(s.processes, pid)
			} else {
				s.processIn <- &workerInProcess{
					pid:  pid,
					proc: proc,
					f:    s.GetProcessInfo,
					out:  output,
				}
			}
		}

		// wait for all processes to return
		count := 0
		for msg := range output {
			top[strconv.FormatInt(int64(msg.pid), 10)] = msg.out
			count++
			if count == len(pids) {
				close(output)
			}
		}

		if js, er := json.Marshal(top); er == nil {
			compressed, err = compressByteArray(js)
		}
	}
	response = fmt.Sprintf("{\"t\":\"%s\",\"v\":\"%s\"}", base64.StdEncoding.EncodeToString(compressed.Bytes()), pluginVersion)
	log.Printf("D! Input [signalfx-metadata] process list collection took %s \n", time.Since(start))
	return
}

func compressByteArray(in []byte) (buf bytes.Buffer, err error) {
	compressor := zlib.NewWriter(&buf)
	_, err = compressor.Write(in)
	_ = compressor.Close()
	return
}

func getProcessCommand(proc *process.Process) (response string) {
	response = " "
	if val, err := proc.Name(); err == nil {
		response = val
	}
	return
}

func getProcessCPUNiceValue(proc *process.Process) (response int32) {
	if val, err := proc.Nice(); err == nil {
		response = val
	}
	return
}

func getProcessCPUPercent(proc *process.Process) (response float64) {
	if val, err := proc.Percent(time.Duration(0)); err == nil {
		response = val
	}
	return
}

func getProcessCPUTime(proc *process.Process) (response string) {
	response = " "
	if val, err := proc.Times(); err == nil {
		response = toTime(val.User + val.System)
	}
	return
}

type workerOutProcess struct {
	pid int32
	out []interface{}
}

type workerInProcess struct {
	pid  int32
	proc *process.Process
	f    func(*process.Process) []interface{}
	out  chan *workerOutProcess
}

func newWorkerProcess(in chan *workerInProcess) {
	go func() {
		for msg := range in {
			msg.out <- &workerOutProcess{
				pid: msg.pid,
				out: msg.f(msg.proc),
			}
		}
	}()
}

// GetProcessInfo returns the top styled process list encoded in base64 and compressed
func (s *ProcessInfo) GetProcessInfo(proc *process.Process) []interface{} {
	return []interface{}{
		getProcessUsername(proc),
		getProcessPriority(proc),
		getProcessCPUNiceValue(proc),
		getProcessVirtualMemoryInfo(proc),
		getProcessResidentMemoryInfo(proc),
		getProcessMemoryExInfo(proc),
		getProcessStatus(proc),
		getProcessCPUPercent(proc),
		getProcessMemoryPercent(proc),
		getProcessCPUTime(proc),
		getProcessCommand(proc),
	}
}

func getProcessMemoryExInfo(proc *process.Process) (response uint64) {
	if val, err := proc.MemoryInfoEx(); err == nil {
		// MemoryInfoEx is not implemented on mac so we must reflect it
		memEx := reflect.ValueOf(val)
		f := reflect.Indirect(memEx).FieldByName("Shared")
		v := f.Uint()
		response = v / 1024
	}
	return
}

func getProcessVirtualMemoryInfo(proc *process.Process) (virtualMemory uint64) {
	if val, err := proc.MemoryInfo(); err == nil {
		virtualMemory = val.VMS / 1024
	}
	return
}

func getProcessResidentMemoryInfo(proc *process.Process) (residentMemory uint64) {
	if val, err := proc.MemoryInfo(); err == nil {
		residentMemory = val.RSS / 1024
	}
	return
}

func getProcessMemoryPercent(proc *process.Process) (response float32) {
	if val, err := proc.MemoryPercent(); err == nil {
		response = val
	}
	return
}

func getProcessPriority(proc *process.Process) (response int32) {
	if val, err := proc.IOnice(); err == nil {
		response = val
	}
	return
}

func getProcessStatus(proc *process.Process) (response string) {
	response = "D"
	if val, err := proc.Status(); err == nil {
		response = val
	}
	return
}

func getProcessUsername(proc *process.Process) (response string) {
	response = " "
	if val, err := proc.Username(); err == nil {
		response = val
	}
	return
}

func toTime(secs float64) (response string) {
	minutes := int(secs / 60)
	seconds := int(math.Mod(secs, 60.0))
	sec := seconds
	dec := (seconds - sec) * 100
	response = fmt.Sprintf("%02d:%02d.%02d", minutes, sec, dec)
	return
}
