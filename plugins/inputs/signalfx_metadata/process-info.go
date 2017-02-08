package signalfxMetadata

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/process"
)

// NewProcessInfo - returns a new ProcessInfo instance
func NewProcessInfo() *ProcessInfo {
	return &ProcessInfo{make(map[int32]*process.Process)}
}

// ProcessInfo - list of processes
type ProcessInfo struct {
	processes map[int32]*process.Process
}

// GetTop - returns a map of process information
func (s *ProcessInfo) GetTop() ([]byte, error) {
	var response = map[string]string{
		"v": pluginVersion,
	}
	var byteResponse []byte
	var top = make(map[string][]interface{})
	var pids []int32
	var pidList map[int32]bool
	var err error

	pidList = make(map[int32]bool)

	pids, err = process.Pids()
	if err == nil {
		// Add missing processes to process list
		for _, pid := range pids {
			pidList[pid] = true
			if _, isIn := s.processes[pid]; !isIn {
				if proc, er := process.NewProcess(pid); er == nil {
					s.processes[pid] = proc
				}
			}
		}
		for pid, proc := range s.processes {
			// Remove dead processes from process list
			if _, in := pidList[pid]; !in {
				delete(s.processes, pid)
			} else {
				pid64 := int64(pid)
				stringPid := strconv.FormatInt(pid64, 10)
				top[stringPid] = GetProcessInfo(proc)
			}
		}
		if js, er := json.Marshal(top); er == nil {
			compressed := compressByteArray(js)
			base64ed := base64.StdEncoding.EncodeToString(compressed)
			response["t"] = base64ed

		}
		byteResponse, err = json.Marshal(response)
	}
	return byteResponse, err
}

func compressByteArray(in []byte) []byte {
	var buf bytes.Buffer
	compressor := zlib.NewWriter(&buf)
	if _, err := compressor.Write(in); err != nil {
		panic(err)
	}
	if err := compressor.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func getProcessCommand(proc *process.Process) string {
	var response = " "
	if val, err := proc.Name(); err == nil {
		response = val
	}
	return response
}

func getProcessCPUNiceValue(proc *process.Process) int32 {
	var response = int32(0)
	if val, err := proc.Nice(); err == nil {
		response = val
	}
	return response
}

func getProcessCPUPercent(proc *process.Process) float64 {
	var response = float64(0)
	if val, err := proc.Percent(time.Duration(0)); err == nil {
		response = val
	}
	return response
}

func getProcessCPUTime(proc *process.Process) string {
	var response = " "
	if val, err := proc.Times(); err == nil {
		response = toTime(val.User + val.System)
	}
	return response
}

// GetProcessInfo - returns an array of information about a process
func GetProcessInfo(proc *process.Process) []interface{} {
	var username = getProcessUsername(proc)
	var priority = getProcessPriority(proc)
	var cpuNiceValue = getProcessCPUNiceValue(proc)
	var virtualMemory, residentMemory = getProcessMemoryInfo(proc)
	var sharedMemory = getProcessMemoryExInfo(proc)
	var status = getProcessStatus(proc)
	var cpuPercent = getProcessCPUPercent(proc)
	var memPercent = getProcessMemoryPercent(proc)
	var cpuTime = getProcessCPUTime(proc)
	var commandValue = getProcessCommand(proc)

	return []interface{}{
		username,
		priority,
		cpuNiceValue,
		virtualMemory,
		residentMemory,
		sharedMemory,
		status,
		cpuPercent,
		memPercent,
		cpuTime,
		commandValue,
	}
}

func getProcessMemoryExInfo(proc *process.Process) uint64 {
	var response = uint64(0)
	if val, err := proc.MemoryInfoEx(); err == nil {
		// MemoryInfoEx is not implemented on mac so we must reflect it
		memEx := reflect.ValueOf(val)
		f := reflect.Indirect(memEx).FieldByName("Shared")
		v := f.Uint()
		response = v / 1024
	}
	return response
}

func getProcessMemoryInfo(proc *process.Process) (uint64, uint64) {
	var virtualMemory = uint64(0)
	var residentMemory = uint64(0)
	if val, err := proc.MemoryInfo(); err == nil {
		virtualMemory = val.VMS / 1024
		residentMemory = val.RSS / 1024
	}
	return virtualMemory, residentMemory
}

func getProcessMemoryPercent(proc *process.Process) float32 {
	var response = float32(0)
	if val, err := proc.MemoryPercent(); err == nil {
		response = val
	}
	return response
}

func getProcessPriority(proc *process.Process) int32 {
	var response = int32(0)
	if val, err := proc.IOnice(); err == nil {
		response = val
	}
	return response
}

func getProcessStatus(proc *process.Process) string {
	var response = "D"
	if val, err := proc.Status(); err == nil {
		response = val
	}
	return response
}

func getProcessUsername(proc *process.Process) string {
	var response = " "
	if val, err := proc.Username(); err == nil {
		response = val
	}
	return response
}

func toTime(secs float64) string {
	var response string
	minutes := int(secs / 60)
	seconds := int(math.Mod(secs, 60.0))
	sec := seconds
	dec := (seconds - sec) * 100
	response = fmt.Sprintf("%02d:%02d.%02d", minutes, sec, dec)
	return response
}
