//go:build windows

package cpu

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var hasGroups = procNtQuerySystemInformationEx.Find() == nil

type stats struct {
	CPU    string
	User   uint64
	System uint64
	Idle   uint64
	Irq    uint64
}

func cpuTimes(perCPU, totalCPU bool) ([]stats, error) {
	if !perCPU && !totalCPU {
		return nil, nil
	}

	// Windows 11 and later splits more than 64 cores into processor groups.
	// Most API calls will only return processors of the group the current thread
	// is running on which leads to issues when Windows is switching the processor
	// group for this plugin as counters are then no longer monotonic!
	// Therefore, we need to use the (undocumented) NtQuerySystemInformationEx
	// function allowing to specify the processor group. However, on older
	// Windows versions this function is not available. The following logic uses
	// the extended function whenever available and falls back to the standard
	// version.
	var perfStats []systemProcessorPerformanceInformation
	var err error
	if hasGroups {
		perfStats, err = queryWithGroups()
	} else {
		perfStats, err = queryWithoutGroups()
	}
	if err != nil {
		return nil, err
	}

	var elements int
	if perCPU {
		elements += len(perfStats)
	}
	if totalCPU {
		elements++
	}

	// Accumulate the total-CPU stats from the individual ones as the gopsutil'result
	// cpu.Times() function is based on GetSystemTimes() which only report data
	// for the current processor group running the plugin. This leads to
	// non-monotonic counters when Windows switches the processor group for the
	// plugin.
	result := make([]stats, 0, elements)
	total := stats{CPU: "cpu-total"}
	for core, s := range perfStats {
		if perCPU {
			result = append(result, stats{
				CPU:    fmt.Sprintf("cpu%d", core),
				User:   uint64(s.UserTime),
				System: uint64(s.KernelTime - s.IdleTime),
				Idle:   uint64(s.IdleTime),
				Irq:    uint64(s.InterruptTime),
			})
		}

		if totalCPU {
			total.User += uint64(s.UserTime)
			total.System += uint64(s.KernelTime - s.IdleTime)
			total.Idle += uint64(s.IdleTime)
			total.Irq += uint64(s.InterruptTime)
		}
	}

	if totalCPU {
		result = append(result, total)
	}

	return result, nil
}

func (s *stats) cycleFields(active bool) map[string]interface{} {
	fields := map[string]interface{}{
		"time_user":       float64(s.User) / 10_000_000.0,   // 100ns to seconds
		"time_system":     float64(s.System) / 10_000_000.0, // 100ns to seconds
		"time_idle":       float64(s.Idle) / 10_000_000.0,   // 100ns to seconds
		"time_nice":       float64(0),
		"time_iowait":     float64(0),
		"time_irq":        float64(s.Irq) / 10_000_000.0, // 100ns to seconds
		"time_softirq":    float64(0),
		"time_steal":      float64(0),
		"time_guest":      float64(0),
		"time_guest_nice": float64(0),
	}
	if active {
		fields["time_active"] = float64(s.User+s.System+s.Irq) / 10_000_000.0 // 100ns to seconds
	}
	return fields
}

func (s *stats) percentageFields(last stats, active bool) (map[string]interface{}, error) {
	total := s.User + s.System + s.Irq + s.Idle
	lastTotal := last.User + last.System + last.Irq + last.Idle

	if total < lastTotal {
		return nil, errors.New("current total CPU time is less than previous total CPU time")
	}

	if total == lastTotal {
		return nil, nil
	}

	dT := float64(total - lastTotal)

	fields := map[string]interface{}{
		"usage_user":       100.0 * float64(s.User-last.User) / dT,
		"usage_system":     100.0 * float64(s.System-last.System) / dT,
		"usage_idle":       100.0 * float64(s.Idle-last.Idle) / dT,
		"usage_nice":       float64(0),
		"usage_iowait":     float64(0),
		"usage_irq":        100.0 * float64(s.Irq-last.Irq) / dT,
		"usage_softirq":    float64(0),
		"usage_steal":      float64(0),
		"usage_guest":      float64(0),
		"usage_guest_nice": float64(0),
	}
	if active {
		dActive := s.User + s.System + s.Irq
		dActive -= last.User + last.System + last.Irq
		fields["usage_active"] = 100.0 * float64(dActive) / dT
	}

	return fields, nil
}

// Windows API section
// The following code was copied from
// https://github.com/shirou/gopsutil/blob/master/cpu/cpu_windows.go
// under the BSD 3-Clause Clear License

var (
	modKernel32                      = windows.NewLazySystemDLL("kernel32.dll")
	modNt                            = windows.NewLazySystemDLL("ntdll.dll")
	procGetActiveProcessorGroupCount = modKernel32.NewProc("GetActiveProcessorGroupCount")
	procNtQuerySystemInformation     = modNt.NewProc("NtQuerySystemInformation")
	procNtQuerySystemInformationEx   = modNt.NewProc("NtQuerySystemInformationEx")
)

type systemProcessorPerformanceInformation struct {
	IdleTime       int64  // idle time in 100ns (this is not a filetime).
	KernelTime     int64  // kernel time in 100ns.  kernel time includes idle time. (this is not a filetime).
	UserTime       int64  // usertime in 100ns (this is not a filetime).
	DpcTime        int64  // dpc time in 100ns (this is not a filetime).
	InterruptTime  int64  // interrupt time in 100ns
	InterruptCount uint64 // ULONG needs to be uint64
}

const systemProcessorPerformanceInformationClass = 8

var systemProcessorPerformanceInfoSize = uint32(unsafe.Sizeof(systemProcessorPerformanceInformation{}))

// queryWithoutGroups queries SystemProcessorPerformanceInformation using the
// non-Ex NtQuerySystemInformation call. This is the legacy fallback for
// environments where NtQuerySystemInformationEx is not available.
// NOTE: This API call only returns data for the calling thread's processor group!
func queryWithoutGroups() ([]systemProcessorPerformanceInformation, error) {
	maxCores := 2056
	buf := make([]systemProcessorPerformanceInformation, maxCores)
	bufSize := uintptr(systemProcessorPerformanceInfoSize) * uintptr(maxCores)

	// Invoke windows API; the returned err from the windows dll proc will
	// always be non-nil even when successful. See
	// https://godoc.org/golang.org/x/sys/windows#LazyProc.Call for more info
	var retSize uint32
	//nolint:gosec // Unsafe calls required when using the Windows API
	retCode, _, err := procNtQuerySystemInformation.Call(
		uintptr(systemProcessorPerformanceInformationClass), // System Information Class -> SystemProcessorPerformanceInformation
		uintptr(unsafe.Pointer(&buf[0])),                    // pointer to first element in result buffer
		bufSize,                                             // size of the buffer in memory
		uintptr(unsafe.Pointer(&retSize)),                   // pointer to the size of the returned results the windows proc will set this
	)
	if retCode != 0 {
		return nil, fmt.Errorf("calling to NtQuerySystemInformation failed: [0x%x] %w", retCode, err)
	}

	// Trim results to the number of returned elements
	n := retSize / systemProcessorPerformanceInfoSize
	return buf[:n], nil
}

// queryWithGroups queries SystemProcessorPerformanceInformation for every active
// processor group via NtQuerySystemInformationEx and concatenates the results. The
// group index is passed as the InputBuffer per the Ex calling convention documented at
// https://www.geoffchappell.com/studies/windows/km/ntoskrnl/api/ex/sysinfo/queryex.htm
func queryWithGroups() ([]systemProcessorPerformanceInformation, error) {
	// Get the number of processor groups. This should be at least one, a zero
	// count indicates an error.
	r, _, err := procGetActiveProcessorGroupCount.Call()
	if r == 0 {
		return nil, fmt.Errorf("calling GetActiveProcessorGroupCount failed: %w", err)
	}
	groups := uint16(r)

	var result []systemProcessorPerformanceInformation
	for g := uint16(0); g < groups; g++ {
		processors := windows.GetActiveProcessorCount(g)
		if processors == 0 {
			return nil, fmt.Errorf("calling GetActiveProcessorCount for processor group %d failed", g)
		}
		// buffer sized exactly for this group's logical CPU count
		buf := make([]systemProcessorPerformanceInformation, processors)
		bufSize := uintptr(systemProcessorPerformanceInfoSize) * uintptr(processors)

		// InputBuffer is a USHORT (2 bytes) holding the target processor group index.
		var retSize uint32
		//nolint:gosec // Unsafe calls required when using the Windows API
		retCode, _, err := procNtQuerySystemInformationEx.Call(
			systemProcessorPerformanceInformationClass, // System Information Class -> SystemProcessorPerformanceInformation
			uintptr(unsafe.Pointer(&g)),                // InputBuffer: pointer to USHORT group index
			unsafe.Sizeof(g),                           // InputBufferLength: sizeof(USHORT) being 2
			uintptr(unsafe.Pointer(&buf[0])),           // pointer to first element in result buffer
			bufSize,                                    // size of the buffer in memory
			uintptr(unsafe.Pointer(&retSize)),          // pointer to the size of the returned results the windows proc will set this
		)
		if retCode != 0 {
			return nil, fmt.Errorf("calling NtQuerySystemInformationEx for processor group %d failed: [0x%x]: %w", g, retCode, err)
		}
		// Guard against a retSize that is not a whole number of entries or exceeds
		// the allocated buffer (e.g. CPU hot-add racing with GetActiveProcessorCount).
		if retSize%systemProcessorPerformanceInfoSize != 0 || uintptr(retSize) > bufSize {
			return nil, fmt.Errorf("calling NtQuerySystemInformationEx for processor group=%d returned unexpected size %d (bufSize=%d)", g, retSize, bufSize)
		}
		n := retSize / systemProcessorPerformanceInfoSize
		result = append(result, buf[:n]...)
	}

	return result, nil
}
