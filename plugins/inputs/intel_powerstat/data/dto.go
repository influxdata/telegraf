// +build linux

package data

// MsrData holds data calculated from MSR.
type MsrData struct {
	Mperf        uint64
	Aperf        uint64
	Tsc          uint64
	C3           uint64
	C6           uint64
	C7           uint64
	ThrottleTemp uint64
	Temp         uint64
	MperfDelta   uint64
	AperfDelta   uint64
	TscDelta     uint64
	C3Delta      uint64
	C6Delta      uint64
	C7Delta      uint64
	ReadDate     int64
}

// RaplData holds data calculated from RAPL.
type RaplData struct {
	DramCurrentEnergy   float64
	SocketCurrentEnergy float64
	SocketEnergy        float64
	DramEnergy          float64
	ReadDate            int64
}

// CPUInfo contains information about cpu cores.
type CPUInfo struct {
	PhysicalID string
	CoreID     string
	CPUID      string
	VendorID   string
	CPUFamily  string
	Model      string
	Flags      string
}
