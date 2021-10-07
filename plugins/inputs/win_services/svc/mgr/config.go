//go:build windows
// +build windows

// https://github.com/golang/sys/blob/master/windows/svc/mgr/config.go

package mgr

type Config struct {
	ServiceType      uint32
	StartType        uint32
	ErrorControl     uint32
	BinaryPathName   string // fully qualified path to the service binary file, can also include arguments for an auto-start service
	LoadOrderGroup   string
	TagId            uint32
	Dependencies     []string
	ServiceStartName string // name of the account under which the service should run
	DisplayName      string
	Password         string
	Description      string
	SidType          uint32 // one of SERVICE_SID_TYPE, the type of sid to use for the service
	DelayedAutoStart bool   // the service is started after other auto-start services are started plus a short delay
}
