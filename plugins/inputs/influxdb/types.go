package influxdb

type point struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Values map[string]interface{} `json:"values"`
}

type memstats struct {
	Alloc         int64      `json:"Alloc"`
	TotalAlloc    int64      `json:"TotalAlloc"`
	Sys           int64      `json:"Sys"`
	Lookups       int64      `json:"Lookups"`
	Mallocs       int64      `json:"Mallocs"`
	Frees         int64      `json:"Frees"`
	HeapAlloc     int64      `json:"HeapAlloc"`
	HeapSys       int64      `json:"HeapSys"`
	HeapIdle      int64      `json:"HeapIdle"`
	HeapInuse     int64      `json:"HeapInuse"`
	HeapReleased  int64      `json:"HeapReleased"`
	HeapObjects   int64      `json:"HeapObjects"`
	StackInuse    int64      `json:"StackInuse"`
	StackSys      int64      `json:"StackSys"`
	MSpanInuse    int64      `json:"MSpanInuse"`
	MSpanSys      int64      `json:"MSpanSys"`
	MCacheInuse   int64      `json:"MCacheInuse"`
	MCacheSys     int64      `json:"MCacheSys"`
	BuckHashSys   int64      `json:"BuckHashSys"`
	GCSys         int64      `json:"GCSys"`
	OtherSys      int64      `json:"OtherSys"`
	NextGC        int64      `json:"NextGC"`
	LastGC        int64      `json:"LastGC"`
	PauseTotalNs  int64      `json:"PauseTotalNs"`
	PauseNs       [256]int64 `json:"PauseNs"`
	NumGC         int64      `json:"NumGC"`
	GCCPUFraction float64    `json:"GCCPUFraction"`
}

type system struct {
	CurrentTime string `json:"currentTime"`
	Started     string `json:"started"`
	Uptime      uint64 `json:"uptime"`
}

type build struct {
	Branch    string `json:"Branch"`
	BuildTime string `json:"Build Time"`
	Commit    string `json:"Commit"`
	Version   string `json:"Version"`
}

type crypto struct {
	FIPS           bool   `json:"FIPS"`
	EnsureFIPS     bool   `json:"ensureFIPS"`
	Implementation string `json:"implementation"`
	PasswordHash   string `json:"passwordHash"`
}
