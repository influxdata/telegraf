package cpu

import (
	"encoding/json"
	"runtime"
	"strconv"
	"strings"
)

type CPUTimesStat struct {
	CPU       string  `json:"cpu"`
	User      float64 `json:"user"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guest_nice"`
	Stolen    float64 `json:"stolen"`
}

type CPUInfoStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendor_id"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physical_id"`
	CoreID     string   `json:"core_id"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"model_name"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cache_size"`
	Flags      []string `json:"flags"`
}

var lastCPUTimes []CPUTimesStat
var lastPerCPUTimes []CPUTimesStat

func CPUCounts(logical bool) (int, error) {
	return runtime.NumCPU(), nil
}

func (c CPUTimesStat) String() string {
	v := []string{
		`"cpu":"` + c.CPU + `"`,
		`"user":` + strconv.FormatFloat(c.User, 'f', 1, 64),
		`"system":` + strconv.FormatFloat(c.System, 'f', 1, 64),
		`"idle":` + strconv.FormatFloat(c.Idle, 'f', 1, 64),
		`"nice":` + strconv.FormatFloat(c.Nice, 'f', 1, 64),
		`"iowait":` + strconv.FormatFloat(c.Iowait, 'f', 1, 64),
		`"irq":` + strconv.FormatFloat(c.Irq, 'f', 1, 64),
		`"softirq":` + strconv.FormatFloat(c.Softirq, 'f', 1, 64),
		`"steal":` + strconv.FormatFloat(c.Steal, 'f', 1, 64),
		`"guest":` + strconv.FormatFloat(c.Guest, 'f', 1, 64),
		`"guest_nice":` + strconv.FormatFloat(c.GuestNice, 'f', 1, 64),
		`"stolen":` + strconv.FormatFloat(c.Stolen, 'f', 1, 64),
	}

	return `{` + strings.Join(v, ",") + `}`
}

func (c CPUInfoStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

func init() {
	lastCPUTimes, _ = CPUTimes(false)
	lastPerCPUTimes, _ = CPUTimes(true)
}
