//go:build windows
// +build windows

// https://github.com/golang/sys/blob/master/windows/svc/mgr/mgr.go

package mgr

import (
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/influxdata/telegraf/plugins/inputs/win_services/unsafeheader"
	"golang.org/x/sys/windows"
)

type Mgr struct {
	Handle windows.Handle
}

// Connect establishes a connection to the service control manager.
func Connect(accessMask uint32) (*Mgr, error) {
	var s *uint16
	h, err := windows.OpenSCManager(s, nil, accessMask)
	if err != nil {
		return nil, err
	}
	return &Mgr{Handle: h}, nil
}

// Disconnect closes connection to the service control manager m.
func (m *Mgr) Disconnect() error {
	return windows.CloseServiceHandle(m.Handle)
}

// ListServices enumerates services in the specified
// service control manager database m.
// If the caller does not have the SERVICE_QUERY_STATUS
// access right to a service, the service is silently
// omitted from the list of services returned.
func (m *Mgr) ListServices() ([]string, error) {
	var err error
	var bytesNeeded, servicesReturned uint32
	var buf []byte
	for {
		var p *byte
		if len(buf) > 0 {
			p = &buf[0]
		}
		err = windows.EnumServicesStatusEx(m.Handle, windows.SC_ENUM_PROCESS_INFO,
			windows.SERVICE_WIN32, windows.SERVICE_STATE_ALL,
			p, uint32(len(buf)), &bytesNeeded, &servicesReturned, nil, nil)
		if err == nil {
			break
		}
		if err != syscall.ERROR_MORE_DATA {
			return nil, err
		}
		if bytesNeeded <= uint32(len(buf)) {
			return nil, err
		}
		buf = make([]byte, bytesNeeded)
	}
	if servicesReturned == 0 {
		return nil, nil
	}

	var services []windows.ENUM_SERVICE_STATUS_PROCESS
	hdr := (*unsafeheader.Slice)(unsafe.Pointer(&services))
	hdr.Data = unsafe.Pointer(&buf[0])
	hdr.Len = int(servicesReturned)
	hdr.Cap = int(servicesReturned)

	var names []string
	for _, s := range services {
		name := windows.UTF16PtrToString(s.ServiceName)
		names = append(names, name)
	}
	return names, nil
}

// OpenService retrieves access to service name, so it can
// be interrogated and controlled.
func (m *Mgr) OpenService(name string, accessMask uint32) (*Service, error) {
	namePtr, err := UTF16FromString(name)
	if err != nil {
		return nil, err
	}
	h, err := windows.OpenService(m.Handle, &namePtr[0], accessMask)
	if err != nil {
		return nil, err
	}
	return &Service{Name: name, Handle: h}, nil
}

// UTF16FromString returns the UTF-16 encoding of the UTF-8 string
// s, with a terminating NUL added. If s contains a NUL byte at any
// location, it returns (nil, syscall.EINVAL).
func UTF16FromString(s string) ([]uint16, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return nil, syscall.EINVAL
		}
	}
	return utf16.Encode([]rune(s + "\x00")), nil
}
