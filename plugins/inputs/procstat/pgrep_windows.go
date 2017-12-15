package procstat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/process"
)

var ErrorNotImplemented = errors.New("not implemented in windows")

//Timeout is the timeout used when making wmi calls
var Timeout = 5 * time.Second

// Implemention of PIDGatherer that execs pgrep to find processes
type Pgrep struct {
}

func NewPgrep() (PIDFinder, error) {
	return &Pgrep{}, nil
}

func (pg *Pgrep) PidFile(path string) ([]PID, error) {
	return nil, ErrorNotImplemented
}

func (pg *Pgrep) Pattern(pattern string) ([]PID, error) {
	pids := make([]PID, 0)
	procs, err := GetWin32ProcsByName(pattern, Timeout)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

func (pg *Pgrep) FullPattern(pattern string) ([]PID, error) {
	pids := make([]PID, 0)
	procs, err := GetWin32ProcsByCmdLine(pattern, Timeout)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

func (pg *Pgrep) Uid(user string) ([]PID, error) {
	return nil, ErrorNotImplemented
}

func GetWin32ProcsByCmdLine(cmdLine string, timeout time.Duration) ([]process.Win32_Process, error) {
	var dst []process.Win32_Process
	query := fmt.Sprint("WHERE CommandLine LIKE \"%", cmdLine, "%\"")
	q := wmi.CreateQuery(&dst, query)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := WMIQueryWithContext(ctx, q, &dst)
	if err != nil {
		return []process.Win32_Process{}, fmt.Errorf("could not get win32Proc: %s", err)
	}
	return dst, nil
}

func GetWin32ProcsByName(name string, timeout time.Duration) ([]process.Win32_Process, error) {
	var dst []process.Win32_Process
	query := fmt.Sprint("WHERE Name LIKE \"%", name, "%\"")
	q := wmi.CreateQuery(&dst, query)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := WMIQueryWithContext(ctx, q, &dst)
	if err != nil {
		return []process.Win32_Process{}, fmt.Errorf("could not get win32Proc: %s", err)
	}
	return dst, nil
}

// WMIQueryWithContext - wraps wmi.Query with a timed-out context to avoid hanging
func WMIQueryWithContext(ctx context.Context, query string, dst interface{}, connectServerArgs ...interface{}) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- wmi.Query(query, dst, connectServerArgs...)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
