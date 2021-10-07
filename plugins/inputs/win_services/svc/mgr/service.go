//go:build windows
// +build windows

// https://github.com/golang/sys/blob/master/windows/svc/mgr/service.go

package mgr

import (
	"syscall"
	"unsafe"

	"github.com/influxdata/telegraf/plugins/inputs/win_services/svc"
	"golang.org/x/sys/windows"
)

// Service is used to access Windows service.
type Service struct {
	Name   string
	Handle windows.Handle
}

// Query returns current status of service s.
func (s *Service) Query() (svc.Status, error) {
	var t windows.SERVICE_STATUS_PROCESS
	var needed uint32
	err := windows.QueryServiceStatusEx(s.Handle, windows.SC_STATUS_PROCESS_INFO, (*byte)(unsafe.Pointer(&t)), uint32(unsafe.Sizeof(t)), &needed)
	if err != nil {
		return svc.Status{}, err
	}
	return svc.Status{
		State:                   svc.State(t.CurrentState),
		Accepts:                 svc.Accepted(t.ControlsAccepted),
		ProcessId:               t.ProcessId,
		Win32ExitCode:           t.Win32ExitCode,
		ServiceSpecificExitCode: t.ServiceSpecificExitCode,
	}, nil
}

// Close relinquish access to the service s.
func (s *Service) Close() error {
	return windows.CloseServiceHandle(s.Handle)
}

// Config retrieves service s configuration paramteres.
func (s *Service) Config() (Config, error) {
	var p *windows.QUERY_SERVICE_CONFIG
	n := uint32(1024)
	for {
		b := make([]byte, n)
		p = (*windows.QUERY_SERVICE_CONFIG)(unsafe.Pointer(&b[0]))
		err := windows.QueryServiceConfig(s.Handle, p, n, &n)
		if err == nil {
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_INSUFFICIENT_BUFFER {
			return Config{}, err
		}
		if n <= uint32(len(b)) {
			return Config{}, err
		}
	}

	b, err := s.queryServiceConfig2(windows.SERVICE_CONFIG_DESCRIPTION)
	if err != nil {
		return Config{}, err
	}
	p2 := (*windows.SERVICE_DESCRIPTION)(unsafe.Pointer(&b[0]))

	b, err = s.queryServiceConfig2(windows.SERVICE_CONFIG_DELAYED_AUTO_START_INFO)
	if err != nil {
		return Config{}, err
	}
	p3 := (*windows.SERVICE_DELAYED_AUTO_START_INFO)(unsafe.Pointer(&b[0]))
	delayedStart := false
	if p3.IsDelayedAutoStartUp != 0 {
		delayedStart = true
	}

	b, err = s.queryServiceConfig2(windows.SERVICE_CONFIG_SERVICE_SID_INFO)
	if err != nil {
		return Config{}, err
	}
	sidType := *(*uint32)(unsafe.Pointer(&b[0]))

	return Config{
		ServiceType:      p.ServiceType,
		StartType:        p.StartType,
		ErrorControl:     p.ErrorControl,
		BinaryPathName:   windows.UTF16PtrToString(p.BinaryPathName),
		LoadOrderGroup:   windows.UTF16PtrToString(p.LoadOrderGroup),
		TagId:            p.TagId,
		Dependencies:     toStringSlice(p.Dependencies),
		ServiceStartName: windows.UTF16PtrToString(p.ServiceStartName),
		DisplayName:      windows.UTF16PtrToString(p.DisplayName),
		Description:      windows.UTF16PtrToString(p2.Description),
		DelayedAutoStart: delayedStart,
		SidType:          sidType,
	}, nil
}

// queryServiceConfig2 calls Windows QueryServiceConfig2 with infoLevel parameter and returns retrieved service configuration information.
func (s *Service) queryServiceConfig2(infoLevel uint32) ([]byte, error) {
	n := uint32(1024)
	for {
		b := make([]byte, n)
		err := windows.QueryServiceConfig2(s.Handle, infoLevel, &b[0], n, &n)
		if err == nil {
			return b, nil
		}
		if err.(syscall.Errno) != syscall.ERROR_INSUFFICIENT_BUFFER {
			return nil, err
		}
		if n <= uint32(len(b)) {
			return nil, err
		}
	}
}

func toStringSlice(ps *uint16) []string {
	r := make([]string, 0)
	p := unsafe.Pointer(ps)

	for {
		s := windows.UTF16PtrToString((*uint16)(p))
		if len(s) == 0 {
			break
		}

		r = append(r, s)
		offset := unsafe.Sizeof(uint16(0)) * (uintptr)(len(s)+1)
		p = unsafe.Pointer(uintptr(p) + offset)
	}

	return r
}
