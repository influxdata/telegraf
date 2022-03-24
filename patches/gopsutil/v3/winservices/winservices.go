//go:build windows
// +build windows

package winservices

import (
	"context"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// Service represent a windows service.
type Service struct {
	Name   string
	Config mgr.Config
	Status ServiceStatus
	srv    *mgr.Service
}

// ServiceStatus combines State and Accepted commands to fully describe running service.
type ServiceStatus struct {
	State         svc.State
	Accepts       svc.Accepted
	Pid           uint32
	Win32ExitCode uint32
}

// NewService create and return a windows Service
func NewService(name string) (*Service, error) {
	// call windows service function need to OpenService handler,
	// so first call func OpenService to get the specified service handler.
	service, err := getService(name)
	if err != nil {
		return nil, err
	}
	return &Service{
		Name: name,
		srv:  service,
	}, nil
}

// GetServiceDetail get a windows service by name
func (s *Service) GetServiceDetail() error {
	return s.GetServiceDetailWithContext(context.Background())
}

// GetServiceDetailWithContext get a windows service by name
func (s *Service) GetServiceDetailWithContext(ctx context.Context) error {
	config, err := s.QueryServiceConfigWithContext(ctx)
	if err != nil {
		return err
	}
	s.Config = config

	status, err := s.QueryStatusWithContext(ctx)
	if err != nil {
		return err
	}
	s.Status = status

	return nil
}

// QueryServiceConfig return the specified service config
func (s *Service) QueryServiceConfig() (mgr.Config, error) {
	return s.QueryServiceConfigWithContext(context.Background())
}

// QueryServiceConfigWithContext call QueryServiceConfig() and QueryServiceConfig2()
// implement windows https://msdn.microsoft.com/en-us/library/windows/desktop/ms684932(v=vs.85).aspx
func (s *Service) QueryServiceConfigWithContext(ctx context.Context) (mgr.Config, error) {
	return s.srv.Config()
}

// QueryStatus return the specified name service currentState and ControlsAccepted
func (s *Service) QueryStatus() (ServiceStatus, error) {
	return s.QueryStatusWithContext(context.Background())
}

// QueryStatusWithContext return the specified name service currentState and ControlsAccepted
func (s *Service) QueryStatusWithContext(ctx context.Context) (ServiceStatus, error) {
	var p *windows.SERVICE_STATUS_PROCESS
	var bytesNeeded uint32
	var buf []byte

	if err := windows.QueryServiceStatusEx(s.srv.Handle, windows.SC_STATUS_PROCESS_INFO, nil, 0, &bytesNeeded); err != windows.ERROR_INSUFFICIENT_BUFFER {
		return ServiceStatus{}, err
	}

	buf = make([]byte, bytesNeeded)
	p = (*windows.SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0]))
	if err := windows.QueryServiceStatusEx(s.srv.Handle, windows.SC_STATUS_PROCESS_INFO, &buf[0], uint32(len(buf)), &bytesNeeded); err != nil {
		return ServiceStatus{}, err
	}

	return ServiceStatus{
		State:         svc.State(p.CurrentState),
		Accepts:       svc.Accepted(p.ControlsAccepted),
		Pid:           p.ProcessId,
		Win32ExitCode: p.Win32ExitCode,
	}, nil
}

// ListServices return all windows service
// reference to golang.org/x/sys/windows/svc/mgr#ListServices()
func ListServices() ([]Service, error) {
	m, err := openSCManager()
	if err != nil {
		return nil, err
	}
	defer m.close()

	names, err := m.mgr.ListServices()
	if err != nil {
		return nil, err
	}

	services := make([]Service, 0)
	for _, name := range names {
		services = append(services, Service{Name: name})
	}

	return services, nil
}
