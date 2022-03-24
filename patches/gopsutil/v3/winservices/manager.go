//go:build windows
// +build windows

package winservices

import (
	"golang.org/x/sys/windows/svc/mgr"
)

type scmanager struct {
	mgr *mgr.Mgr
}

func openSCManager() (*scmanager, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, err
	}
	return &scmanager{m}, nil
}

func (sc *scmanager) close() error {
	return sc.mgr.Disconnect()
}

func getService(serviceName string) (*mgr.Service, error) {
	m, err := openSCManager()
	if err != nil {
		return nil, err
	}
	defer m.close()
	return m.mgr.OpenService(serviceName)
}
