// +build windows

package procstat

import (
	"context"
	"fmt"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/process"
)

//Timeout is the timeout used when making wmi calls
var Timeout = 5 * time.Second

func init() {
	wmi.DefaultClient.AllowMissingFields = true
}

type queryType string

const (
	like     = queryType("LIKE")
	equals   = queryType("=")
	notEqual = queryType("!=")
)

//Pattern matches the process name on windows and will find a pattern using a WMI like query
func (pg *NativeFinder) Pattern(pattern string) ([]PID, error) {
	var pids []PID
	procs, err := getWin32ProcsByVariable("Name", like, pattern, Timeout)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

//FullPattern matches the cmdLine on windows and will find a pattern using a WMI like query
func (pg *NativeFinder) FullPattern(pattern string) ([]PID, error) {
	var pids []PID
	procs, err := getWin32ProcsByVariable("CommandLine", like, pattern, Timeout)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

//GetWin32ProcsByVariable allows you to query any variable with a like query
func getWin32ProcsByVariable(variable string, qType queryType, value string, timeout time.Duration) ([]process.Win32_Process, error) {
	var dst []process.Win32_Process
	var query string
	query = fmt.Sprint("WHERE ", variable, " ", qType, " \"", value, "\"")
	fmt.Println("Query: ", query)
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
