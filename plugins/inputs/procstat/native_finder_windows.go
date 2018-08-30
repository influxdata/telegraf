package procstat

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/process"
)

//Timeout is the timeout used when making wmi calls
var Timeout = 5 * time.Second

type queryType string

const (
	like     = queryType("LIKE")
	equals   = queryType("=")
	notEqual = queryType("!=")
)

//Pattern matches on the process name
func (pg *NativeFinder) Pattern(pattern string) ([]PID, error) {
	var pids []PID
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return pids, err
	}
	procs, err := process.Processes()
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			//skip, this can be caused by the pid no longer existing
			//or you having no permissions to access it
			continue
		}
		if regxPattern.MatchString(name) {
			pids = append(pids, PID(p.Pid))
		}
	}
	return pids, err
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
	// should look like "WHERE CommandLine LIKE "procstat"
	query = fmt.Sprintf("WHERE %s %s %q", variable, qType, value)
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
