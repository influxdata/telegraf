package procfilter

//	"fmt"

// TODO add IO stats?
// TODO add percent for memory?

type tPid int32   // A PID, note that once a set of PIDs is packed as a single packStat it gets a negative pseudo ID
type tStamp uint8 // A stamp used to identify different sample times.

/* A real process or a group of (packed) processes will implement this interface to get to the underlying statistics. */
type stat interface {
	PID() tPid
	Users() ([]string, error)
	Groups() ([]string, error)
	UIDs() ([]int32, error)
	GIDs() ([]int32, error)
	RSS() (int64, error)
	VSZ() (int64, error)
	Swap() (int64, error)
	CPU() (float32, error)
	ProcessNumber() int64
	ThreadNumber() (int64, error)
	FDNumber() (int64, error)
	Path() (string, error)
	Exe() (string, error)
	Cmd() (string, error)
	CmdLine() (string, error)
	Args() ([]string, error)
	ChildrenPIDs(int) []tPid
}

type stats struct {
	stamp    tStamp // Stamp the pid2Stat to know to what sample it refers.
	pid2Stat map[tPid]stat
}

// A stamp to know when we change from one sample to the next
var stamp tStamp

// Fast lookup for UID -> user name
var uid2User = map[int32]string{}

// Sorting helpers
type sortCrit int
type statSlice []stat
type byRSS statSlice
type byVSZ statSlice
type bySwap statSlice
type byCPU statSlice
type byProcessNumber statSlice
type byThreadNumber statSlice
type byFDNumber statSlice

func (s byRSS) Len() int {
	return len(s)
}

func (s byRSS) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byRSS) Less(i, j int) bool {
	// use > (instead of <) to reverse the sort order and get the biggest first
	iv, _ := s[i].RSS()
	jv, _ := s[j].RSS()
	return iv > jv
}

func (s byVSZ) Len() int {
	return len(s)
}

func (s byVSZ) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byVSZ) Less(i, j int) bool {
	// use > (instead of <) to reverse the sort order and get the biggest first
	iv, _ := s[i].VSZ()
	jv, _ := s[j].VSZ()
	return iv > jv
}

func (s bySwap) Len() int {
	return len(s)
}

func (s bySwap) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s bySwap) Less(i, j int) bool {
	// use > (instead of <) to reverse the sort order and get the biggest first
	iv, _ := s[i].Swap()
	jv, _ := s[j].Swap()
	return iv > jv
}

func (s byProcessNumber) Len() int {
	return len(s)
}

func (s byProcessNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byProcessNumber) Less(i, j int) bool {
	// use > (instead of <) to reverse the sort order and get the biggest first
	iv := s[i].ProcessNumber()
	jv := s[j].ProcessNumber()
	return iv > jv
}

func (s byThreadNumber) Len() int {
	return len(s)
}

func (s byThreadNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byThreadNumber) Less(i, j int) bool {
	// use > (instead of <) to reverse the sort order and get the biggest first
	iv, _ := s[i].ThreadNumber()
	jv, _ := s[j].ThreadNumber()
	return iv > jv
}

func (s byFDNumber) Len() int {
	return len(s)
}

func (s byFDNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byFDNumber) Less(i, j int) bool {
	// use > (instead of <) to reverse the sort order and get the biggest first
	iv, _ := s[i].FDNumber()
	jv, _ := s[j].FDNumber()
	return iv > jv
}

func (s byCPU) Len() int {
	return len(s)
}

func (s byCPU) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byCPU) Less(i, j int) bool {
	// use > (instead of <) to reverse the sort order and get the biggest first
	iv, _ := s[i].CPU()
	jv, _ := s[j].CPU()
	return iv > jv
}

// resetGlobalStatSets update the (global) stats structures for the current sample. (ie: purge dead PIDs, and get new ones)
func resetGlobalStatSets() {
	resetAllProcStats()
	resetAllPackStats()
}

// reset checks if the stats are relevant to the current sample, if not it resets them. Returns true if we changed sample (reset occured)
func (s *stats) reset() bool {
	if s.pid2Stat == nil || s.stamp != stamp {
		s.pid2Stat = map[tPid]stat{}
		s.stamp = stamp
		return true
	}
	return false
}

// unpackAsSlice collect all procStats from all statss (unpack packStats).
func (s stats) unpackAsSlice(ss []*procStat) []*procStat {
	return unpackMapAsSlice(s.pid2Stat, ss)
}

// unpackAsMap collect all procStats from all statss (unpack packStats).
func (s stats) unpackAsMap(sm map[tPid]*procStat) map[tPid]*procStat {
	return unpackMapAsMap(s.pid2Stat, sm)
}
