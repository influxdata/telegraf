//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	ia "github.com/intel/iaevents"
)

// Linux availability: https://www.kernel.org/doc/Documentation/sysctl/fs.txt
const fileMaxPath = "/proc/sys/fs/file-max"

type fileInfoProvider interface {
	readFile(string) ([]byte, error)
	lstat(string) (os.FileInfo, error)
	fileLimit() (uint64, error)
}

type fileHelper struct{}

func (fileHelper) readFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (fileHelper) lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (fileHelper) fileLimit() (uint64, error) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	return rLimit.Cur, err
}

type sysInfoProvider interface {
	allCPUs() ([]int, error)
	allSockets() ([]int, error)
}

type iaSysInfo struct{}

func (iaSysInfo) allCPUs() ([]int, error) {
	return ia.AllCPUs()
}

func (iaSysInfo) allSockets() ([]int, error) {
	return ia.AllSockets()
}

// IntelPMU is the plugin type.
type IntelPMU struct {
	EventListPaths []string             `toml:"event_definitions"`
	CoreEntities   []*CoreEventEntity   `toml:"core_events"`
	UncoreEntities []*UncoreEventEntity `toml:"uncore_events"`

	Log telegraf.Logger `toml:"-"`

	fileInfo       fileInfoProvider
	entitiesReader entitiesValuesReader
}

// CoreEventEntity represents config section for core events.
type CoreEventEntity struct {
	Events    []string `toml:"events"`
	Cores     []string `toml:"cores"`
	EventsTag string   `toml:"events_tag"`
	PerfGroup bool     `toml:"perf_group"`

	parsedEvents []*eventWithQuals
	parsedCores  []int
	allEvents    bool

	activeEvents []*ia.ActiveEvent
}

// UncoreEventEntity represents config section for uncore events.
type UncoreEventEntity struct {
	Events    []string `toml:"events"`
	Sockets   []string `toml:"sockets"`
	Aggregate bool     `toml:"aggregate_uncore_units"`
	EventsTag string   `toml:"events_tag"`

	parsedEvents  []*eventWithQuals
	parsedSockets []int
	allEvents     bool

	activeMultiEvents []multiEvent
}

type multiEvent struct {
	activeEvents []*ia.ActiveEvent
	perfEvent    *ia.PerfEvent
	socket       int
}

type eventWithQuals struct {
	name       string
	qualifiers []string

	custom ia.CustomizableEvent
}

// Start is required for IntelPMU to implement the telegraf.ServiceInput interface.
// Necessary initialization and config checking are done in Init.
func (IntelPMU) Start(_ telegraf.Accumulator) error {
	return nil
}

func (i *IntelPMU) Init() error {
	err := checkFiles(i.EventListPaths, i.fileInfo)
	if err != nil {
		return fmt.Errorf("error during event definitions paths validation: %v", err)
	}

	reader, err := newReader(i.EventListPaths)
	if err != nil {
		return err
	}
	transformer := ia.NewPerfTransformer()
	resolver := &iaEntitiesResolver{reader: reader, transformer: transformer, log: i.Log}
	parser := &configParser{log: i.Log, sys: &iaSysInfo{}}
	activator := &iaEntitiesActivator{perfActivator: &iaEventsActivator{}, placementMaker: &iaPlacementMaker{}}

	i.entitiesReader = &iaEntitiesValuesReader{eventReader: &iaValuesReader{}, timer: &realClock{}}

	return i.initialization(parser, resolver, activator)
}

func (i *IntelPMU) initialization(parser entitiesParser, resolver entitiesResolver, activator entitiesActivator) error {
	if parser == nil || resolver == nil || activator == nil {
		return fmt.Errorf("entities parser and/or resolver and/or activator is nil")
	}

	err := parser.parseEntities(i.CoreEntities, i.UncoreEntities)
	if err != nil {
		return fmt.Errorf("error during parsing configuration sections: %v", err)
	}

	err = resolver.resolveEntities(i.CoreEntities, i.UncoreEntities)
	if err != nil {
		return fmt.Errorf("error during events resolving: %v", err)
	}

	err = i.checkFileDescriptors()
	if err != nil {
		return fmt.Errorf("error during file descriptors checking: %v", err)
	}

	err = activator.activateEntities(i.CoreEntities, i.UncoreEntities)
	if err != nil {
		return fmt.Errorf("error during events activation: %v", err)
	}
	return nil
}

func (i *IntelPMU) checkFileDescriptors() error {
	coreFd, err := estimateCoresFd(i.CoreEntities)
	if err != nil {
		return fmt.Errorf("failed to estimate number of core events file descriptors: %v", err)
	}
	uncoreFd, err := estimateUncoreFd(i.UncoreEntities)
	if err != nil {
		return fmt.Errorf("failed to estimate nubmer of uncore events file descriptors: %v", err)
	}
	if coreFd > math.MaxUint64-uncoreFd {
		return fmt.Errorf("requested number of file descriptors exceeds uint64")
	}
	allFd := coreFd + uncoreFd

	// maximum file descriptors enforced on a kernel level
	maxFd, err := readMaxFD(i.fileInfo)
	if err != nil {
		i.Log.Warnf("cannot obtain number of available file descriptors: %v", err)
	} else if allFd > maxFd {
		return fmt.Errorf("required file descriptors number `%d` exceeds maximum number of available file descriptors `%d`"+
			": consider increasing the maximum number", allFd, maxFd)
	}

	// soft limit for current process
	limit, err := i.fileInfo.fileLimit()
	if err != nil {
		i.Log.Warnf("cannot obtain limit value of open files: %v", err)
	} else if allFd > limit {
		return fmt.Errorf("required file descriptors number `%d` exceeds soft limit of open files `%d`"+
			": consider increasing the limit", allFd, limit)
	}

	return nil
}

func (i *IntelPMU) Gather(acc telegraf.Accumulator) error {
	if i.entitiesReader == nil {
		return fmt.Errorf("entities reader is nil")
	}
	coreMetrics, uncoreMetrics, err := i.entitiesReader.readEntities(i.CoreEntities, i.UncoreEntities)
	if err != nil {
		return fmt.Errorf("failed to read entities events values: %v", err)
	}

	for id, m := range coreMetrics {
		scaled := ia.EventScaledValue(m.values)
		if !scaled.IsUint64() {
			return fmt.Errorf("cannot process `%s` scaled value `%s`: exceeds uint64", m.name, scaled.String())
		}
		coreMetrics[id].scaled = scaled.Uint64()
	}
	for id, m := range uncoreMetrics {
		scaled := ia.EventScaledValue(m.values)
		if !scaled.IsUint64() {
			return fmt.Errorf("cannot process `%s` scaled value `%s`: exceeds uint64", m.name, scaled.String())
		}
		uncoreMetrics[id].scaled = scaled.Uint64()
	}

	publishCoreMeasurements(coreMetrics, acc)
	publishUncoreMeasurements(uncoreMetrics, acc)

	return nil
}

func (i *IntelPMU) Stop() {
	for _, entity := range i.CoreEntities {
		if entity == nil {
			continue
		}
		for _, event := range entity.activeEvents {
			if event == nil {
				continue
			}
			err := event.Deactivate()
			if err != nil {
				i.Log.Warnf("failed to deactivate core event `%s`: %v", event, err)
			}
		}
	}
	for _, entity := range i.UncoreEntities {
		if entity == nil {
			continue
		}
		for _, multi := range entity.activeMultiEvents {
			for _, event := range multi.activeEvents {
				if event == nil {
					continue
				}
				err := event.Deactivate()
				if err != nil {
					i.Log.Warnf("failed to deactivate uncore event `%s`: %v", event, err)
				}
			}
		}
	}
}

func newReader(files []string) (*ia.JSONFilesReader, error) {
	reader := ia.NewFilesReader()
	for _, file := range files {
		err := reader.AddFiles(file)
		if err != nil {
			return nil, fmt.Errorf("failed to add files to reader: %v", err)
		}
	}
	return reader, nil
}

func estimateCoresFd(entities []*CoreEventEntity) (uint64, error) {
	var err error
	number := uint64(0)
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		events := uint64(len(entity.parsedEvents))
		cores := uint64(len(entity.parsedCores))
		number, err = multiplyAndAdd(events, cores, number)
		if err != nil {
			return 0, err
		}
	}
	return number, nil
}

func estimateUncoreFd(entities []*UncoreEventEntity) (uint64, error) {
	var err error
	number := uint64(0)
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		for _, e := range entity.parsedEvents {
			if e.custom.Event == nil {
				continue
			}
			pmus := uint64(len(e.custom.Event.PMUTypes))
			sockets := uint64(len(entity.parsedSockets))
			number, err = multiplyAndAdd(pmus, sockets, number)
			if err != nil {
				return 0, err
			}
		}
	}
	return number, nil
}

func multiplyAndAdd(factorA uint64, factorB uint64, sum uint64) (uint64, error) {
	bigA := new(big.Int).SetUint64(factorA)
	bigB := new(big.Int).SetUint64(factorB)
	activeEvents := new(big.Int).Mul(bigA, bigB)
	if !activeEvents.IsUint64() {
		return 0, fmt.Errorf("value `%s` cannot be represented as uint64", activeEvents.String())
	}
	if sum > math.MaxUint64-activeEvents.Uint64() {
		return 0, fmt.Errorf("value `%s` exceeds uint64", new(big.Int).Add(activeEvents, new(big.Int).SetUint64(sum)))
	}
	sum += activeEvents.Uint64()
	return sum, nil
}

func readMaxFD(reader fileInfoProvider) (uint64, error) {
	if reader == nil {
		return 0, fmt.Errorf("file reader is nil")
	}
	buf, err := reader.readFile(fileMaxPath)
	if err != nil {
		return 0, fmt.Errorf("cannot open `%s` file: %v", fileMaxPath, err)
	}
	max, err := strconv.ParseUint(strings.Trim(string(buf), "\n "), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse file content of `%s`: %v", fileMaxPath, err)
	}
	return max, nil
}

func checkFiles(paths []string, fileInfo fileInfoProvider) error {
	// No event definition JSON locations present
	if len(paths) == 0 {
		return fmt.Errorf("no paths were given")
	}
	if fileInfo == nil {
		return fmt.Errorf("file info provider is nil")
	}
	// Wrong files
	for _, path := range paths {
		lInfo, err := fileInfo.lstat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file `%s` doesn't exist", path)
			}
			return fmt.Errorf("cannot obtain file info of `%s`: %v", path, err)
		}
		mode := lInfo.Mode()
		if mode&os.ModeSymlink != 0 {
			return fmt.Errorf("file %s is a symlink", path)
		}
		if !mode.IsRegular() {
			return fmt.Errorf("file `%s` doesn't point to a reagular file", path)
		}
	}
	return nil
}

func publishCoreMeasurements(metrics []coreMetric, acc telegraf.Accumulator) {
	for _, m := range metrics {
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		fields["raw"] = m.values.Raw
		fields["enabled"] = m.values.Enabled
		fields["running"] = m.values.Running
		fields["scaled"] = m.scaled

		tags["event"] = m.name
		tags["cpu"] = strconv.Itoa(m.cpu)

		if len(m.tag) > 0 {
			tags["events_tag"] = m.tag
		}
		acc.AddFields("pmu_metric", fields, tags, m.time)
	}
}

func publishUncoreMeasurements(metrics []uncoreMetric, acc telegraf.Accumulator) {
	for _, m := range metrics {
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		fields["raw"] = m.values.Raw
		fields["enabled"] = m.values.Enabled
		fields["running"] = m.values.Running
		fields["scaled"] = m.scaled

		tags["event"] = m.name

		tags["socket"] = strconv.Itoa(m.socket)
		tags["unit_type"] = m.unitType
		if !m.agg {
			tags["unit"] = m.unit
		}
		if len(m.tag) > 0 {
			tags["events_tag"] = m.tag
		}
		acc.AddFields("pmu_metric", fields, tags, m.time)
	}
}

func init() {
	inputs.Add("intel_pmu", func() telegraf.Input {
		pmu := IntelPMU{
			fileInfo: &fileHelper{},
		}
		return &pmu
	})
}
