package intel_powerstat

type msrData struct {
	mperf                 uint64
	aperf                 uint64
	timeStampCounter      uint64
	c3                    uint64
	c6                    uint64
	c7                    uint64
	throttleTemp          uint64
	temp                  uint64
	mperfDelta            uint64
	aperfDelta            uint64
	timeStampCounterDelta uint64
	c3Delta               uint64
	c6Delta               uint64
	c7Delta               uint64
	readDate              int64
}

type raplData struct {
	dramCurrentEnergy   float64
	socketCurrentEnergy float64
	socketEnergy        float64
	dramEnergy          float64
	readDate            int64
}

type cpuInfo struct {
	physicalID string
	coreID     string
	cpuID      string
	vendorID   string
	cpuFamily  string
	model      string
	flags      string
}
