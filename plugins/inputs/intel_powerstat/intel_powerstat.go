//go:generate ../../../tools/readme_config_includer/generator
//go:build linux && amd64

package intel_powerstat

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	ptel "github.com/intel/powertelemetry"
	cpuUtil "github.com/shirou/gopsutil/v3/cpu"
	"golang.org/x/exp/constraints"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// PowerStat plugin enables monitoring of platform metrics.
type PowerStat struct {
	CPUMetrics       []string        `toml:"cpu_metrics"`
	PackageMetrics   []string        `toml:"package_metrics"`
	IncludedCPUs     []string        `toml:"included_cpus"`
	ExcludedCPUs     []string        `toml:"excluded_cpus"`
	EventDefinitions string          `toml:"event_definitions"`
	MsrReadTimeout   config.Duration `toml:"msr_read_timeout"`
	Log              telegraf.Logger `toml:"-"`

	parsedIncludedCores []int
	parsedExcludedCores []int

	parsedCPUTimedMsrMetrics []string
	parsedCPUPerfMetrics     []string
	parsedPackageRaplMetrics []string
	parsedPackageMsrMetrics  []string

	option  OptGenerator
	fetcher MetricFetcher

	cpuMetricsMap    map[string]func(telegraf.Accumulator, int, int, int)
	packageMetricMap map[string]func(telegraf.Accumulator, int)

	needsCoreFreq       bool
	needsMsrCPU         bool
	needsPerf           bool
	needsTimeRelatedMsr bool

	needsRapl       bool
	needsMsrPackage bool

	logOnce map[string]struct{}
}

// SampleConfig returns a sample configuration (See sample.conf).
func (*PowerStat) SampleConfig() string {
	return sampleConfig
}

// Start initializes the MetricFetcher interface of the receiver to gather metrics.
func (p *PowerStat) Start(_ telegraf.Accumulator) error {
	if err := p.disableUnsupportedMetrics(); err != nil {
		return err
	}

	if err := p.parseConfig(); err != nil {
		return err
	}

	opts := p.option.Generate(OptConfig{
		cpuMetrics:     p.CPUMetrics,
		packageMetrics: p.PackageMetrics,
		includedCPUs:   p.parsedIncludedCores,
		excludedCPUs:   p.parsedExcludedCores,
		perfEventFile:  p.EventDefinitions,
		msrReadTimeout: time.Duration(p.MsrReadTimeout),
		log:            p.Log,
	})

	var err error
	var initErr *ptel.MultiError
	p.fetcher, err = ptel.New(opts...)
	if err != nil {
		if !errors.As(err, &initErr) {
			return fmt.Errorf("failed to initialize metric fetcher interface: %w", err)
		}
		p.Log.Warnf("Plugin initialized with errors: %v", err)
	}

	return nil
}

// Stop deactivates perf events if one or more of the requested metrics rely on perf.
func (p *PowerStat) Stop() {
	if !p.needsPerf {
		return
	}

	if err := p.fetcher.DeactivatePerfEvents(); err != nil {
		p.Log.Errorf("Failed to deactivate perf events: %v", err)
	}
}

// Gather collects the plugin's metrics.
func (p *PowerStat) Gather(acc telegraf.Accumulator) error {
	// gather CPU metrics relying on coreFreq and msr which share CPU IDs.
	if p.needsCoreFreq || p.needsMsrCPU {
		p.addCPUMetrics(acc)
	}

	// gather CPU metrics relying on perf.
	if p.needsPerf {
		p.addCPUPerfMetrics(acc)
	}

	// gather package metrics.
	if len(p.PackageMetrics) != 0 {
		p.addPackageMetrics(acc)
	}

	return nil
}

// parseConfig is a helper method that parses configuration fields from the receiver such as included/excluded CPU IDs.
func (p *PowerStat) parseConfig() error {
	var err error

	if p.MsrReadTimeout < 0 {
		return errors.New("msr_read_timeout should be positive number or equal to 0 (to disable timeouts)")
	}

	p.PackageMetrics, err = parsePackageMetrics(p.PackageMetrics)
	if err != nil {
		return fmt.Errorf("failed to parse package metrics: %w", err)
	}

	p.CPUMetrics, err = parseCPUMetrics(p.CPUMetrics)
	if err != nil {
		return fmt.Errorf("failed to parse core metrics: %w", err)
	}

	if len(p.CPUMetrics) == 0 && len(p.PackageMetrics) == 0 {
		return fmt.Errorf("no metrics were found in the configuration file")
	}

	p.parsedCPUTimedMsrMetrics = parseCPUTimeRelatedMsrMetrics(p.CPUMetrics)
	p.parsedCPUPerfMetrics = parseCPUPerfMetrics(p.CPUMetrics)

	p.parsedPackageRaplMetrics = parsePackageRaplMetrics(p.PackageMetrics)
	p.parsedPackageMsrMetrics = parsePackageMsrMetrics(p.PackageMetrics)

	if len(p.ExcludedCPUs) != 0 && len(p.IncludedCPUs) != 0 {
		return errors.New("configuration error. Provide either one 'included_cpus'/'excluded_cpus' configuration option, or none")
	}

	if len(p.ExcludedCPUs) != 0 {
		p.parsedExcludedCores, err = parseCores(p.ExcludedCPUs)
		if err != nil {
			return fmt.Errorf("failed to parse excluded CPUs: %w", err)
		}
	}

	if len(p.IncludedCPUs) != 0 {
		p.parsedIncludedCores, err = parseCores(p.IncludedCPUs)
		if err != nil {
			return fmt.Errorf("failed to parse included CPUs: %w", err)
		}
	}

	p.needsCoreFreq = needsCoreFreq(p.CPUMetrics)
	p.needsMsrCPU = needsMsr(p.CPUMetrics)
	p.needsPerf = needsPerf(p.CPUMetrics)
	if p.needsPerf {
		if err = checkFile(p.EventDefinitions); err != nil {
			return fmt.Errorf("failed to parse event definitions path: %w", err)
		}
	}

	p.needsTimeRelatedMsr = needsTimeRelatedMsr(p.CPUMetrics)

	p.needsRapl = needsRapl(p.PackageMetrics)
	p.needsMsrPackage = needsMsr(p.PackageMetrics)

	p.cpuMetricsMap = map[string]func(telegraf.Accumulator, int, int, int){
		cpuC0StateResidency.String():        p.addCPUC0StateResidency,
		cpuC1StateResidency.String():        p.addCPUC1StateResidency,
		cpuC3StateResidency.String():        p.addCPUC3StateResidency,
		cpuC6StateResidency.String():        p.addCPUC6StateResidency,
		cpuC7StateResidency.String():        p.addCPUC7StateResidency,
		cpuBusyFrequency.String():           p.addCPUBusyFrequency,
		cpuBusyCycles.String():              p.addCPUBusyCycles,
		cpuC0SubstateC01Percent.String():    p.addCPUC0SubstateC01Percent,
		cpuC0SubstateC02Percent.String():    p.addCPUC0SubstateC02Percent,
		cpuC0SubstateC0WaitPercent.String(): p.addCPUC0SubstateC0WaitPercent,
	}

	p.packageMetricMap = map[string]func(telegraf.Accumulator, int){
		packageCurrentPowerConsumption.String():     p.addCurrentPackagePower,
		packageCurrentDramPowerConsumption.String(): p.addCurrentDramPower,
		packageThermalDesignPower.String():          p.addThermalDesignPower,
		packageCPUBaseFrequency.String():            p.addCPUBaseFrequency,
		packageTurboLimit.String():                  p.addMaxTurboFreqLimits,
	}

	return nil
}

// parsePackageMetrics takes a slice of package metrics. It ensures all metrics in the slice
// are supported package-specific, and there are no duplicates. If metrics slice is nil, then
// default package metrics are provided.
func parsePackageMetrics(metrics []string) ([]string, error) {
	if metrics == nil {
		return []string{
			packageCurrentPowerConsumption.String(),
			packageCurrentDramPowerConsumption.String(),
			packageThermalDesignPower.String(),
		}, nil
	}

	for _, m := range metrics {
		if !isValidPackageMetric(m) {
			return nil, fmt.Errorf("invalid package metric specified: %q", m)
		}
	}

	if hasDuplicate(metrics) {
		return nil, errors.New("package metrics contains duplicates")
	}
	return metrics, nil
}

// parseCPUMetrics takes a slice of CPU metrics. It ensures all metrics in the slice are
// supported CPU-specific, and there are no duplicates.
func parseCPUMetrics(metrics []string) ([]string, error) {
	for _, m := range metrics {
		if !isValidCoreMetric(m) {
			return nil, fmt.Errorf("invalid core metric specified: %q", m)
		}
	}

	if hasDuplicate(metrics) {
		return nil, errors.New("core metrics contains duplicates")
	}
	return metrics, nil
}

// parsedCPUTimedMsrMetrics takes a slice of unique CPU metrics, and returns only the metrics which
// depend on time-related MSR offset reads.
func parseCPUTimeRelatedMsrMetrics(metrics []string) []string {
	slice := []string{
		cpuC0StateResidency.String(),
		cpuC1StateResidency.String(),
		cpuC3StateResidency.String(),
		cpuC6StateResidency.String(),
		cpuC7StateResidency.String(),
		cpuBusyCycles.String(),
		cpuBusyFrequency.String(),
	}

	cpuTimeRelatedMsrMetrics := make([]string, 0, len(slice))
	for _, m := range metrics {
		if choice.Contains(m, slice) {
			cpuTimeRelatedMsrMetrics = append(cpuTimeRelatedMsrMetrics, m)
		}
	}

	return cpuTimeRelatedMsrMetrics
}

// parsedCPUTimedMsrMetrics takes a slice of unique CPU metrics, and returns only the metrics which
// depend on perf event reads.
func parseCPUPerfMetrics(metrics []string) []string {
	slice := []string{
		cpuC0SubstateC01Percent.String(),
		cpuC0SubstateC02Percent.String(),
		cpuC0SubstateC0WaitPercent.String(),
	}

	cpuPerfMetrics := make([]string, 0, len(slice))
	for _, m := range metrics {
		if choice.Contains(m, slice) {
			cpuPerfMetrics = append(cpuPerfMetrics, m)
		}
	}

	return cpuPerfMetrics
}

// parsePackageRaplMetrics takes a slice of unique package metrics, and returns only the metrics which depend on rapl.
func parsePackageRaplMetrics(metrics []string) []string {
	slice := []string{
		packageCurrentPowerConsumption.String(),
		packageCurrentDramPowerConsumption.String(),
		packageThermalDesignPower.String(),
	}

	packageRaplMetrics := make([]string, 0, len(slice))
	for _, m := range metrics {
		if choice.Contains(m, slice) {
			packageRaplMetrics = append(packageRaplMetrics, m)
		}
	}

	return packageRaplMetrics
}

// parsePackageMsrMetrics takes a slice of unique package metrics, and returns only the metrics which depend on msr.
func parsePackageMsrMetrics(metrics []string) []string {
	slice := []string{
		packageCPUBaseFrequency.String(),
		packageTurboLimit.String(),
	}

	packageMsrMetrics := make([]string, 0, len(slice))
	for _, m := range metrics {
		if choice.Contains(m, slice) {
			packageMsrMetrics = append(packageMsrMetrics, m)
		}
	}

	return packageMsrMetrics
}

// hasDuplicate takes a slice of a generic type, and returns true
// if the slice contains duplicates. Otherwise, it returns false.
func hasDuplicate[S ~[]E, E constraints.Ordered](s S) bool {
	m := make(map[E]struct{}, len(s))
	for _, v := range s {
		if _, ok := m[v]; ok {
			return true
		}
		m[v] = struct{}{}
	}
	return false
}

// parseCores takes a slice of strings where each string represents a group of
// one or more CPU IDs (e.g. ["0", "1-3", "4,5,6"] or ["1-3,4"]). It returns a slice
// of integers.
func parseCores(cores []string) ([]int, error) {
	parsedCores := make([]int, 0, len(cores))
	for _, elem := range cores {
		pCores, err := parseGroupCores(elem)
		if err != nil {
			return nil, fmt.Errorf("failed to parse core group: %w", err)
		}
		parsedCores = append(parsedCores, pCores...)
	}

	if hasDuplicate(parsedCores) {
		return nil, errors.New("core values cannot be duplicated")
	}
	return parsedCores, nil
}

// parseGroupCores takes a string which represents a group of one or more
// CPU IDs (e.g. "0", "1-3", or "4,5,6") and returns a slice of integers with
// all CPU IDs within the group.
func parseGroupCores(coreGroup string) ([]int, error) {
	coreElems := strings.Split(coreGroup, ",")
	cores := make([]int, 0, len(coreElems))

	for _, coreElem := range coreElems {
		if strings.Contains(coreElem, "-") {
			pCores, err := parseCoreRange(coreElem)
			if err != nil {
				return nil, fmt.Errorf("failed to parse core range %q: %w", coreElem, err)
			}
			cores = append(cores, pCores...)
		} else {
			singleCore, err := strconv.Atoi(coreElem)
			if err != nil {
				return nil, fmt.Errorf("failed to parse single core: %w", err)
			}
			cores = append(cores, singleCore)
		}
	}
	return cores, nil
}

// parseCoreRange takes a string representing a core range (e.g. "0-4"), and
// returns a slice of integers with all elements within this range.
func parseCoreRange(coreRange string) ([]int, error) {
	rangeVals := strings.Split(coreRange, "-")
	if len(rangeVals) != 2 {
		return nil, errors.New("invalid core range format")
	}

	low, err := strconv.Atoi(rangeVals[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse low bounds' core range: %w", err)
	}

	high, err := strconv.Atoi(rangeVals[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse high bounds' core range: %w", err)
	}

	if high < low {
		return nil, errors.New("high bound of core range cannot be less than low bound")
	}
	return makeCores(low, high), nil
}

// makeCores takes a low and high bounds of an interval of integers, and returns
// a slice will all integers within the interval.
func makeCores(low, high int) []int {
	cores := make([]int, high-low+1)
	for i := range cores {
		cores[i] = i + low
	}
	return cores
}

func checkFile(path string) error {
	if len(path) == 0 {
		return errors.New("file path is empty")
	}
	fInfo, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file %q does not exist", path)
		}
		return fmt.Errorf("could not get the info for file %q: %w", path, err)
	}
	if fMode := fInfo.Mode(); fMode&os.ModeSymlink != 0 {
		return fmt.Errorf("file %q is a symlink", path)
	}
	return nil
}

// addCPUMetrics takes an accumulator, and adds to it enabled metrics which rely on
// coreFreq and msr.
func (p *PowerStat) addCPUMetrics(acc telegraf.Accumulator) {
	for _, cpuID := range p.fetcher.GetMsrCPUIDs() {
		coreID, packageID, err := getDataCPUID(p.fetcher, cpuID)
		if err != nil {
			acc.AddError(fmt.Errorf("failed to get coreFreq and/or msr metrics for CPU ID %v: %w", cpuID, err))
			continue
		}

		// Add requested metrics which rely on coreFreq.
		if p.needsCoreFreq {
			p.addCPUFrequency(acc, cpuID, coreID, packageID)
		}

		// Add requested metrics which rely on msr.
		if p.needsMsrCPU {
			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)
		}
	}
}

// addPerCPUMsrMetrics adds to the accumulator enabled metrics, which rely on msr,
// for a given CPU ID. MSR-related metrics comprise single-time MSR read and several
// time-related MSR offset reads.
func (p *PowerStat) addPerCPUMsrMetrics(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	// cpuTemperature metric is a single MSR offset read.
	if choice.Contains(cpuTemperature.String(), p.CPUMetrics) {
		p.addCPUTemperature(acc, cpuID, coreID, packageID)
	}

	if !p.needsTimeRelatedMsr {
		return
	}

	// Read several time-related MSR offsets.
	var moduleErr *ptel.ModuleNotInitializedError
	err := p.fetcher.UpdatePerCPUMetrics(cpuID)
	if err == nil {
		// Add time-related MSR offset metrics to the accumulator
		p.addCPUTimeRelatedMsrMetrics(acc, cpuID, coreID, packageID)
		return
	}

	// Always add to the accumulator errors not related to module not initialized.
	if !errors.As(err, &moduleErr) {
		acc.AddError(fmt.Errorf("failed to update MSR time-related metrics for CPU ID %v: %w", cpuID, err))
		return
	}

	// Add only once module not initialized error related to msr module and updating time-related msr metrics.
	logErrorOnce(
		acc,
		p.logOnce,
		"msr_time_related",
		fmt.Errorf("failed to update MSR time-related metrics: %w", moduleErr),
	)
}

// addCPUTimeRelatedMsrMetrics adds to the accumulator enabled time-related MSR metrics,
// for a given CPU ID. NOTE: Requires to run first fetcher.UpdatePerCPUMetrics method
// to update the values of MSR offsets read.
func (p *PowerStat) addCPUTimeRelatedMsrMetrics(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	for _, m := range p.parsedCPUTimedMsrMetrics {
		if fn, ok := p.cpuMetricsMap[m]; ok {
			fn(acc, cpuID, coreID, packageID)
		}
	}
}

// addCPUPerfMetrics takes an accumulator, and adds to it enabled metrics which rely on perf.
func (p *PowerStat) addCPUPerfMetrics(acc telegraf.Accumulator) {
	var moduleErr *ptel.ModuleNotInitializedError

	// Read events related to perf-related metrics.
	err := p.fetcher.ReadPerfEvents()
	if err != nil {
		// Always add to the accumulator errors not related to module not initialized.
		if !errors.As(err, &moduleErr) {
			acc.AddError(fmt.Errorf("failed to read perf events: %w", err))
			return
		}

		// Add only once module not initialized error related to perf module and reading perf-related metrics.
		logErrorOnce(
			acc,
			p.logOnce,
			"perf_read",
			fmt.Errorf("failed to read perf events: %w", moduleErr),
		)
		return
	}

	for _, cpuID := range p.fetcher.GetPerfCPUIDs() {
		coreID, packageID, err := getDataCPUID(p.fetcher, cpuID)
		if err != nil {
			acc.AddError(fmt.Errorf("failed to get perf metrics for CPU ID %v: %w", cpuID, err))
			continue
		}

		p.addPerCPUPerfMetrics(acc, cpuID, coreID, packageID)
	}
}

// addPerCPUPerfMetrics adds to the accumulator enabled metrics, which rely on perf, for a given CPU ID.
func (p *PowerStat) addPerCPUPerfMetrics(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	for _, m := range p.parsedCPUPerfMetrics {
		if fn, ok := p.cpuMetricsMap[m]; ok {
			fn(acc, cpuID, coreID, packageID)
		}
	}
}

// getDataCPUID takes a TopologyFetcher and CPU ID, and returns the core ID and package ID corresponding to the CPU ID.
func getDataCPUID(t TopologyFetcher, cpuID int) (coreID int, packageID int, err error) {
	coreID, err = t.GetCPUCoreID(cpuID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get core ID from CPU ID %v: %w", cpuID, err)
	}

	packageID, err = t.GetCPUPackageID(cpuID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get package ID from CPU ID %v: %w", cpuID, err)
	}

	return coreID, packageID, nil
}

// addPackageMetrics takes an accumulator, and adds enabled package metrics to it.
func (p *PowerStat) addPackageMetrics(acc telegraf.Accumulator) {
	for _, packageID := range p.fetcher.GetPackageIDs() {
		// Add requested metrics which rely on rapl.
		if p.needsRapl {
			p.addPerPackageRaplMetrics(acc, packageID)
		}

		// Add requested metrics which rely on msr.
		if p.needsMsrPackage {
			p.addPerPackageMsrMetrics(acc, packageID)
		}

		// Add uncore frequency metric which relies on both uncoreFreq and msr.
		if choice.Contains(packageUncoreFrequency.String(), p.PackageMetrics) {
			p.addUncoreFrequency(acc, packageID)
		}
	}
}

// addPerPackageRaplMetrics adds to the accumulator enabled metrics, which rely on rapl, for a given package ID.
func (p *PowerStat) addPerPackageRaplMetrics(acc telegraf.Accumulator, packageID int) {
	for _, m := range p.parsedPackageRaplMetrics {
		if fn, ok := p.packageMetricMap[m]; ok {
			fn(acc, packageID)
		}
	}
}

// addPerPackageMsrMetrics adds to the accumulator enabled metrics, which rely on msr registers, for a given package ID.
func (p *PowerStat) addPerPackageMsrMetrics(acc telegraf.Accumulator, packageID int) {
	for _, m := range p.parsedPackageMsrMetrics {
		if fn, ok := p.packageMetricMap[m]; ok {
			fn(acc, packageID)
		}
	}
}

// addCPUFrequency fetches CPU frequency metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUFrequency(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuFrequency,
				units:  "mhz",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUFrequency,
		},
		p.logOnce,
	)
}

// addCPUFrequency fetches CPU temperature metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUTemperature(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[uint64]{
			metricCommon: metricCommon{
				metric: cpuTemperature,
				units:  "celsius",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUTemperature,
		},
		p.logOnce,
	)
}

// addCPUC0StateResidency fetches C0 state residency metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC0StateResidency(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC0StateResidency,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC0StateResidency,
		},
		p.logOnce,
	)
}

// addCPUC1StateResidency fetches C1 state residency metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC1StateResidency(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC1StateResidency,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC1StateResidency,
		},
		p.logOnce,
	)
}

// addCPUC3StateResidency fetches C3 state residency metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC3StateResidency(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC3StateResidency,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC3StateResidency,
		},
		p.logOnce,
	)
}

// addCPUC6StateResidency fetches C6 state residency metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC6StateResidency(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC6StateResidency,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC6StateResidency,
		},
		p.logOnce,
	)
}

// addCPUC7StateResidency fetches C7 state residency metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC7StateResidency(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC7StateResidency,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC7StateResidency,
		},
		p.logOnce,
	)
}

// addCPUBusyFrequency fetches CPU busy frequency metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUBusyFrequency(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuBusyFrequency,
				units:  "mhz",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUBusyFrequencyMhz,
		},
		p.logOnce,
	)
}

// addCPUBusyCycles fetches CPU busy cycles metric for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUBusyCycles(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuBusyCycles,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC0StateResidency,
		},
		p.logOnce,
	)
}

// addCPUC0SubstateC01Percent fetches a value indicating the percentage of time the processor spent in its C0.1 substate
// out of the total time in the C0 state for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC0SubstateC01Percent(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC0SubstateC01Percent,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC0SubstateC01Percent,
		},
		p.logOnce,
	)
}

// addCPUC0SubstateC02Percent fetches a value indicating the percentage of time the processor spent in its C0.2 substate
// out of the total time in the C0 state for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC0SubstateC02Percent(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC0SubstateC02Percent,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC0SubstateC02Percent,
		},
		p.logOnce,
	)
}

// addCPUC0SubstateC0WaitPercent fetches a value indicating the percentage of time the processor spent in its C0_Wait substate
// out of the total time in the C0 state for a given CPU ID, and adds it to the accumulator.
func (p *PowerStat) addCPUC0SubstateC0WaitPercent(acc telegraf.Accumulator, cpuID, coreID, packageID int) {
	addMetric(
		acc,
		&cpuMetric[float64]{
			metricCommon: metricCommon{
				metric: cpuC0SubstateC0WaitPercent,
				units:  "percent",
			},
			cpuID:     cpuID,
			coreID:    coreID,
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUC0SubstateC0WaitPercent,
		},
		p.logOnce,
	)
}

// addCurrentPackagePower fetches the current package power metric for a given package ID, and adds it to the accumulator.
func (p *PowerStat) addCurrentPackagePower(acc telegraf.Accumulator, packageID int) {
	addMetric(
		acc,
		&packageMetric[float64]{
			metricCommon: metricCommon{
				metric: packageCurrentPowerConsumption,
				units:  "watts",
			},
			packageID: packageID,
			fetchFn:   p.fetcher.GetCurrentPackagePowerConsumptionWatts,
		},
		p.logOnce,
	)
}

// addCurrentPackagePower fetches the current dram power metric for a given package ID, and adds it to the accumulator.
func (p *PowerStat) addCurrentDramPower(acc telegraf.Accumulator, packageID int) {
	addMetric(
		acc,
		&packageMetric[float64]{
			metricCommon: metricCommon{
				metric: packageCurrentDramPowerConsumption,
				units:  "watts",
			},
			packageID: packageID,
			fetchFn:   p.fetcher.GetCurrentDramPowerConsumptionWatts,
		},
		p.logOnce,
	)
}

// addCurrentPackagePower fetches the thermal design power metric for a given package ID, and adds it to the accumulator.
func (p *PowerStat) addThermalDesignPower(acc telegraf.Accumulator, packageID int) {
	addMetric(
		acc,
		&packageMetric[float64]{
			metricCommon: metricCommon{
				metric: packageThermalDesignPower,
				units:  "watts",
			},
			packageID: packageID,
			fetchFn:   p.fetcher.GetPackageThermalDesignPowerWatts,
		},
		p.logOnce,
	)
}

// addCPUBaseFrequency fetches the CPU base frequency metric for a given package ID, and adds it to the accumulator.
func (p *PowerStat) addCPUBaseFrequency(acc telegraf.Accumulator, packageID int) {
	addMetric(
		acc,
		&packageMetric[uint64]{
			metricCommon: metricCommon{
				metric: packageCPUBaseFrequency,
				units:  "mhz",
			},
			packageID: packageID,
			fetchFn:   p.fetcher.GetCPUBaseFrequency,
		},
		p.logOnce,
	)
}

// addUncoreFrequency fetches the uncore frequency metrics for a given package ID, and adds it to the accumulator.
func (p *PowerStat) addUncoreFrequency(acc telegraf.Accumulator, packageID int) {
	dieIDs, err := p.fetcher.GetPackageDieIDs(packageID)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to get die IDs for package ID %v: %w", packageID, err))
		return
	}

	for _, dieID := range dieIDs {
		// Add initial uncore frequency limits.
		p.addUncoreFrequencyInitialLimits(acc, packageID, dieID)

		// Add current uncore frequency limits and value.
		p.addUncoreFrequencyCurrentValues(acc, packageID, dieID)
	}
}

// addUncoreFrequencyInitialLimits fetches uncore frequency initial limits for a given pair of package and die ID,
// and adds it to the accumulator.
func (p *PowerStat) addUncoreFrequencyInitialLimits(acc telegraf.Accumulator, packageID, dieID int) {
	initMin, initMax, err := getUncoreFreqInitialLimits(p.fetcher, packageID, dieID)
	if err == nil {
		acc.AddGauge(
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": round(initMin),
				"uncore_frequency_limit_mhz_max": round(initMax),
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "initial",
				"die":        strconv.Itoa(dieID),
			},
		)
		return
	}

	// Always add to the accumulator errors not related to module not initialized.
	var moduleErr *ptel.ModuleNotInitializedError
	if !errors.As(err, &moduleErr) {
		acc.AddError(fmt.Errorf("failed to get initial uncore frequency limits for package ID %v and die ID %v: %w", packageID, dieID, err))
		return
	}

	// Add only once module not initialized error related to uncore_frequency module and uncore frequency initial limits.
	logErrorOnce(
		acc,
		p.logOnce,
		fmt.Sprintf("%s_%s_initial", moduleErr.Name, packageUncoreFrequency.String()),
		fmt.Errorf("failed to get %q initial limits: %w", packageUncoreFrequency, moduleErr),
	)
}

// addUncoreFrequencyCurrentValues fetches uncore frequency current limits and value for a given pair of package and die ID,
// and adds it to the accumulator.
func (p *PowerStat) addUncoreFrequencyCurrentValues(acc telegraf.Accumulator, packageID, dieID int) {
	val, err := getUncoreFreqCurrentValues(p.fetcher, packageID, dieID)
	if err == nil {
		acc.AddGauge(
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": round(val.currMin),
				"uncore_frequency_limit_mhz_max": round(val.currMax),
				"uncore_frequency_mhz_cur":       uint64(val.curr),
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "current",
				"die":        strconv.Itoa(dieID),
			},
		)
		return
	}

	// Always add to the accumulator errors not related to module not initialized.
	var moduleErr *ptel.ModuleNotInitializedError
	if !errors.As(err, &moduleErr) {
		acc.AddError(fmt.Errorf("failed to get current uncore frequency values for package ID %v and die ID %v: %w", packageID, dieID, err))
		return
	}

	// Add only once module not initialized error related to uncore_frequency module and uncore frequency current value and limits.
	logErrorOnce(
		acc,
		p.logOnce,
		fmt.Sprintf("%s_%s_current", moduleErr.Name, packageUncoreFrequency.String()),
		fmt.Errorf("failed to get %q current value and limits: %w", packageUncoreFrequency, moduleErr),
	)
}

// getUncoreFreqInitialLimits returns the initial uncore frequency limits of a given package ID and die ID.
func getUncoreFreqInitialLimits(fetcher MetricFetcher, packageID, dieID int) (initialMin float64, initialMax float64, err error) {
	initialMin, err = fetcher.GetInitialUncoreFrequencyMin(packageID, dieID)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("failed to get initial minimum uncore frequency limit: %w", err)
	}

	initialMax, err = fetcher.GetInitialUncoreFrequencyMax(packageID, dieID)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("failed to get initial maximum uncore frequency limit: %w", err)
	}

	return initialMin, initialMax, nil
}

type uncoreFreqValues struct {
	currMin float64
	currMax float64
	curr    float64
}

// getUncoreFreqCurrentValues returns the current uncore frequency value as well as current min and max uncore frequency limits of a given
// package ID and die ID.
func getUncoreFreqCurrentValues(fetcher MetricFetcher, packageID, dieID int) (uncoreFreqValues, error) {
	currMin, err := fetcher.GetCustomizedUncoreFrequencyMin(packageID, dieID)
	if err != nil {
		return uncoreFreqValues{}, fmt.Errorf("failed to get current minimum uncore frequency limit: %w", err)
	}

	currMax, err := fetcher.GetCustomizedUncoreFrequencyMax(packageID, dieID)
	if err != nil {
		return uncoreFreqValues{}, fmt.Errorf("failed to get current maximum uncore frequency limit: %w", err)
	}

	current, err := fetcher.GetCurrentUncoreFrequency(packageID, dieID)
	if err != nil {
		return uncoreFreqValues{}, fmt.Errorf("failed to get current uncore frequency: %w", err)
	}

	return uncoreFreqValues{
		currMin: currMin,
		currMax: currMax,
		curr:    current,
	}, nil
}

// addMaxTurboFreqLimits fetches the max turbo frequency limits metric for a given package ID, and adds it to the accumulator.
func (p *PowerStat) addMaxTurboFreqLimits(acc telegraf.Accumulator, packageID int) {
	var moduleErr *ptel.ModuleNotInitializedError

	turboFreqList, err := p.fetcher.GetMaxTurboFreqList(packageID)
	if err != nil {
		// Always add to the accumulator errors not related to module not initialized.
		if !errors.As(err, &moduleErr) {
			acc.AddError(fmt.Errorf("failed to get %q for package ID %v: %w", packageTurboLimit.String(), packageID, err))
			return
		}

		// Add only once module not initialized error related to msr module and max turbo frequency limits metric.
		logErrorOnce(
			acc,
			p.logOnce,
			fmt.Sprintf("%s_%s", moduleErr.Name, packageTurboLimit.String()),
			fmt.Errorf("failed to get %q: %w", packageTurboLimit, moduleErr),
		)
		return
	}

	isHybrid := isHybridCPU(turboFreqList)
	for _, v := range turboFreqList {
		tags := map[string]string{
			"package_id":   strconv.Itoa(packageID),
			"active_cores": strconv.Itoa(int(v.ActiveCores)),
		}

		if isHybrid {
			var hybridTag string
			if v.Secondary {
				hybridTag = "secondary"
			} else {
				hybridTag = "primary"
			}
			tags["hybrid"] = hybridTag
		}

		acc.AddGauge(
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"max_turbo_frequency_mhz": v.Value,
			},
			// tags
			tags,
		)
	}
}

// isHybridCPU is a helper function that takes a slice of MaxTurboFreq structs and returns true if the CPU where these values belong to,
// is a hybrid CPU. Otherwise, returns false.
func isHybridCPU(turboFreqList []ptel.MaxTurboFreq) bool {
	for _, v := range turboFreqList {
		if v.Secondary {
			return true
		}
	}
	return false
}

// disableUnsupportedMetrics checks whether the processor is capable of gathering specific metrics.
// In case it is not, disableUnsupportedMetrics will disable the option to gather those metrics.
// Error is returned if there is an issue with retrieving processor information.
func (p *PowerStat) disableUnsupportedMetrics() error {
	cpus, err := cpuUtil.Info()
	if err != nil {
		return fmt.Errorf("error occurred while parsing CPU information: %w", err)
	}
	if len(cpus) == 0 {
		return errors.New("no CPUs were found")
	}

	// First CPU is sufficient for verification
	firstCPU := cpus[0]
	cpuModel, err := strconv.Atoi(firstCPU.Model)
	if err != nil {
		return fmt.Errorf("error occurred while parsing CPU model: %w", err)
	}

	if err := ptel.CheckIfCPUC1StateResidencySupported(cpuModel); err != nil {
		p.disableCPUMetric(cpuC1StateResidency)
	}

	if err := ptel.CheckIfCPUC3StateResidencySupported(cpuModel); err != nil {
		p.disableCPUMetric(cpuC3StateResidency)
	}

	if err := ptel.CheckIfCPUC6StateResidencySupported(cpuModel); err != nil {
		p.disableCPUMetric(cpuC6StateResidency)
	}

	if err := ptel.CheckIfCPUC7StateResidencySupported(cpuModel); err != nil {
		p.disableCPUMetric(cpuC7StateResidency)
	}

	if err := ptel.CheckIfCPUTemperatureSupported(cpuModel); err != nil {
		p.disableCPUMetric(cpuTemperature)
	}

	if err := ptel.CheckIfCPUBaseFrequencySupported(cpuModel); err != nil {
		p.disablePackageMetric(packageCPUBaseFrequency)
	}

	allowedModelsForPerfRelated := []int{
		0x8F, // INTEL_FAM6_SAPPHIRERAPIDS_X
		0xCF, // INTEL_FAM6_EMERALDRAPIDS_X
	}
	if !slices.Contains(allowedModelsForPerfRelated, cpuModel) {
		p.disableCPUMetric(cpuC0SubstateC01Percent)
		p.disableCPUMetric(cpuC0SubstateC02Percent)
		p.disableCPUMetric(cpuC0SubstateC0WaitPercent)
	}

	if !slices.Contains(firstCPU.Flags, "msr") {
		p.disableCPUMetric(cpuC0StateResidency)
		p.disableCPUMetric(cpuC1StateResidency)
		p.disableCPUMetric(cpuC3StateResidency)
		p.disableCPUMetric(cpuC6StateResidency)
		p.disableCPUMetric(cpuC7StateResidency)
		p.disableCPUMetric(cpuBusyCycles)
		p.disableCPUMetric(cpuBusyFrequency)
		p.disableCPUMetric(cpuTemperature)
		p.disablePackageMetric(packageCPUBaseFrequency)
		p.disablePackageMetric(packageTurboLimit)
	}

	if !slices.Contains(firstCPU.Flags, "aperfmperf") {
		p.disableCPUMetric(cpuC0StateResidency)
		p.disableCPUMetric(cpuC1StateResidency)
		p.disableCPUMetric(cpuBusyCycles)
		p.disableCPUMetric(cpuBusyFrequency)
	}

	if !slices.Contains(firstCPU.Flags, "dts") {
		p.disableCPUMetric(cpuTemperature)
	}

	return nil
}

// disableCPUMetric removes given cpu metric from cpu_metrics.
func (p *PowerStat) disableCPUMetric(metricToDisable cpuMetricType) {
	startLen := len(p.CPUMetrics)
	p.CPUMetrics = slices.DeleteFunc(p.CPUMetrics, func(cpuMetric string) bool {
		return cpuMetric == metricToDisable.String()
	})

	if len(p.CPUMetrics) < startLen {
		p.Log.Warnf("%q is not supported by CPU, metric will not be gathered.", metricToDisable)
	}
}

// disablePackageMetric removes given package metric from package_metrics.
func (p *PowerStat) disablePackageMetric(metricToDisable packageMetricType) {
	startLen := len(p.PackageMetrics)
	p.PackageMetrics = slices.DeleteFunc(p.PackageMetrics, func(packageMetric string) bool {
		return packageMetric == metricToDisable.String()
	})

	if len(p.PackageMetrics) < startLen {
		p.Log.Warnf("%q is not supported by CPU, metric will not be gathered.", metricToDisable)
	}
}

// logErrorOnce takes an accumulator, a key string value error map, a key string and an error. It adds the error to the accumulator only if the
// key is not in the logOnceMap. Additionally, if the key is not in logOnceMap map, adds the key to it. This is to prevent excessive error messages
// from flooding the accumulator.
func logErrorOnce(acc telegraf.Accumulator, logOnceMap map[string]struct{}, key string, err error) {
	if _, ok := logOnceMap[key]; !ok {
		acc.AddError(err)
		logOnceMap[key] = struct{}{}
	}
}

func init() {
	inputs.Add("intel_powerstat", func() telegraf.Input {
		return &PowerStat{
			MsrReadTimeout: 0,

			option:  &optGenerator{},
			logOnce: make(map[string]struct{}),
		}
	})
}
