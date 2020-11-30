// +build linux

package intel_powerstat

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/data"
	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/services"
)

const (
	cpuC1                = "cpu_c1_state_residency"
	cpuC6                = "cpu_c6_state_residency"
	cpuBusyCycles        = "cpu_busy_cycles"
	cpuBusyFrequency     = "cpu_busy_frequency"
	cpuFrequency         = "cpu_frequency"
	cpuTemperature       = "cpu_temperature"
	percentageMultiplier = 100
)

var (
	allowedProcessorModelsForC1C6 = []string{"0x37", "0x4D", "0x5C", "0x5F", "0x7A", "0x4C", "0x86", "0x96", "0x9C",
		"0x1A", "0x1E", "0x1F", "0x2E", "0x25", "0x2C", "0x2F", "0x2A", "0x2D", "0x3A", "0x3E", "0x4E", "0x5E", "0x55", "0x8E",
		"0x9E", "0x6A", "0x6C", "0x7D", "0x7E", "0x9D", "0x3C", "0x3F", "0x45", "0x46", "0x3D", "0x47", "0x4F", "0x56",
		"0x66", "0x57", "0x85", "0xA5", "0xA6", "0x8F", "0x8C", "0x8D"}
)

// PowerStat plugin enables monitoring CPU utilization and power consumption on Intel platform.
type PowerStat struct {
	CPUMetrics []string        `toml:"cpu_metrics"`
	Log        telegraf.Logger `toml:"-"`

	fs   services.FileService
	rapl services.RaplService
	msr  services.MsrService

	perCoreMetrics         bool
	coreCPUFrequency       bool
	c1CoreStateResidency   bool
	c6CoreStateResidency   bool
	processorBusyCycles    bool
	coreCPUTemperature     bool
	processorBusyFrequency bool
	cpuInfo                map[string]*data.CPUInfo
	skipFirstIteration     bool
}

// Init performs one time setup of the plugin.
func (p *PowerStat) Init() error {
	p.ParseCPUMetricsConfig()
	err := p.verifyProcessor()
	if err != nil {
		return err
	}
	// Set perCoreMetrics flag only when there is at least one core metric enabled.
	if p.c1CoreStateResidency || p.c6CoreStateResidency || p.processorBusyCycles || p.coreCPUTemperature ||
		p.processorBusyFrequency || p.coreCPUFrequency {
		p.perCoreMetrics = true
		p.msr = services.NewMsrServiceWithFs(p.Log, p.fs)
	}
	p.rapl = services.NewRaplServiceWithFs(p.Log, p.fs)

	return nil
}

// SampleConfig returns the default configuration of the plugin.
func (p *PowerStat) SampleConfig() string {
	return `
  ## All global metrics are always collected by Intel PowerStat plugin.
  ## User can choose which per-CPU metrics are monitored by the plugin in cpu_metrics array.
  ## Empty array means no per-CPU specific metrics will be collected by the plugin - in this case only platform level
  ## telemetry will be exposed by Intel PowerStat plugin.
  ## Supported options:
  ## "cpu_frequency", "cpu_c1_state_residency", "cpu_c6_state_residency", "cpu_busy_cycles", "cpu_temperature", "cpu_busy_frequency"
  # cpu_metrics = []
`
}

// Description returns a one-sentence description on the plugin.
func (p *PowerStat) Description() string {
	return `Intel PowerStat plugin enables monitoring of platform metrics (power, TDP) and Core metrics like temperature, power and utilization.`
}

// Gather takes in an accumulator and adds the metrics that the Input gathers.
func (p *PowerStat) Gather(acc telegraf.Accumulator) error {
	p.addGlobalMetrics(acc)

	if p.perCoreMetrics && len(p.msr.GetCpuIDs()) > 0 {
		p.addPerCoreMetrics(acc)
	}

	// Gathering the first iteration of metrics was skipped for most of them because they are based on delta calculations.
	p.skipFirstIteration = false

	return nil
}

func (p *PowerStat) addPerCoreMetrics(acc telegraf.Accumulator) {
	var wg sync.WaitGroup
	cpuIDs := p.msr.GetCpuIDs()
	wg.Add(len(cpuIDs))

	for _, cpuID := range cpuIDs {
		go p.addMetricsForSingleCore(cpuID, acc, &wg)
	}

	wg.Wait()
}

func (p *PowerStat) addMetricsForSingleCore(cpuID string, acc telegraf.Accumulator, wg *sync.WaitGroup) {
	defer wg.Done()

	if p.coreCPUFrequency {
		p.addCPUFrequencyMetric(cpuID, acc)
	}

	// Read data from MSR only if required
	if p.c1CoreStateResidency || p.c6CoreStateResidency || p.processorBusyCycles || p.coreCPUTemperature ||
		p.processorBusyFrequency {
		err := p.msr.OpenAndReadMsr(cpuID)
		if err != nil {
			// In case of an error exit the function. All metrics past this point are dependant on MSR.
			p.Log.Debugf("error while reading msr: %v", err)
			return
		}
	}

	if p.coreCPUTemperature {
		p.addCoreCPUTemperatureMetric(cpuID, acc)
	}
	if !p.skipFirstIteration {
		if p.c6CoreStateResidency {
			p.addC6StateResidencyMetric(cpuID, acc)
		}

		if p.processorBusyCycles {
			p.addProcessorBusyCyclesMetric(cpuID, acc)
		}

		if p.c1CoreStateResidency {
			p.addC1StateResidencyMetric(cpuID, acc)
		}
	}
	// processorBusyFrequency metric does some calculations inside that are required in another plugin cycle.
	if p.processorBusyFrequency {
		p.addProcessorBusyFrequencyMetric(cpuID, acc)
	}
}

func (p *PowerStat) addCPUFrequencyMetric(cpuID string, acc telegraf.Accumulator) {
	frequency, _, err := p.msr.RetrieveCPUFrequencyForCore(cpuID)

	// In case of an error leave func
	if err != nil {
		p.Log.Debugf("error while reading file: %v", err)
		return
	}

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.PhysicalID,
		"core_id":    cpu.CoreID,
		"cpu_id":     cpu.CPUID,
	}

	fields := map[string]interface{}{
		"name":  "cpu_frequency",
		"unit":  "MHz",
		"value": services.RoundFloatToNearestTwoDecimalPlaces(services.ConvertKiloHertzToMegaHertz(frequency)),
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addCoreCPUTemperatureMetric(cpuID string, acc telegraf.Accumulator) {
	temp := p.msr.GetThrottleTemperature(cpuID) - p.msr.GetTemperature(cpuID)
	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.PhysicalID,
		"core_id":    cpu.CoreID,
		"cpu_id":     cpu.CPUID,
	}
	fields := map[string]interface{}{
		"name":  "cpu_temperature",
		"unit":  "celsius_degrees",
		"value": temp,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addProcessorBusyCyclesMetric(cpuID string, acc telegraf.Accumulator) {
	// Avoid division by 0
	if p.msr.GetTimestampDelta(cpuID) == 0 {
		p.Log.Errorf("timestamp counter on offset %s should not equal 0 on cpuID %s",
			services.TimestampCounterLocation, cpuID)
		return
	}
	busyCyclesValue := services.RoundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(p.msr.GetMperfDelta(cpuID)) / float64(p.msr.GetTimestampDelta(cpuID)))
	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.PhysicalID,
		"core_id":    cpu.CoreID,
		"cpu_id":     cpu.CPUID,
	}
	fields := map[string]interface{}{
		"name":  "cpu_busy_cycles",
		"unit":  "percentage",
		"value": busyCyclesValue,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addC6StateResidencyMetric(cpuID string, acc telegraf.Accumulator) {
	// Avoid division by 0
	if p.msr.GetTimestampDelta(cpuID) == 0 {
		p.Log.Errorf("timestamp counter on offset %s should not equal 0 on cpuID %s",
			services.TimestampCounterLocation, cpuID)
		return
	}
	c6Value := services.RoundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(p.msr.GetC6Delta(cpuID)) / float64(p.msr.GetTimestampDelta(cpuID)))

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.PhysicalID,
		"core_id":    cpu.CoreID,
		"cpu_id":     cpu.CPUID,
	}
	fields := map[string]interface{}{
		"name":  "cpu_c6_state_residency",
		"unit":  "percentage",
		"value": c6Value,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addC1StateResidencyMetric(cpuID string, acc telegraf.Accumulator) {
	c1Value := 0.0
	// Avoid division by 0
	timestampDelta := p.msr.GetTimestampDelta(cpuID)
	if timestampDelta < 1 {
		p.Log.Errorf("timestamp delta value %d should not be lower than 1", timestampDelta)
		return
	}
	// Since counter collection is not atomic it may happen that sum of C0, C1, C3, C6 and C7
	// is bigger value than TSC, in such case C1 residency shall be set to 0.
	if timestampDelta > (p.msr.GetMperfDelta(cpuID) + p.msr.GetC3Delta(cpuID) +
		p.msr.GetC6Delta(cpuID) + p.msr.GetC7Delta(cpuID)) {
		c1 := timestampDelta - p.msr.GetMperfDelta(cpuID) - p.msr.GetC3Delta(cpuID) -
			p.msr.GetC6Delta(cpuID) - p.msr.GetC7Delta(cpuID)

		c1Value = services.RoundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(c1) /
			float64(timestampDelta))
	}

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.PhysicalID,
		"core_id":    cpu.CoreID,
		"cpu_id":     cpu.CPUID,
	}
	fields := map[string]interface{}{
		"name":  "cpu_c1_state_residency",
		"unit":  "percentage",
		"value": c1Value,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addProcessorBusyFrequencyMetric(cpuID string, acc telegraf.Accumulator) {
	// Avoid division by 0
	if p.msr.GetMperfDelta(cpuID) == 0 {
		p.Log.Errorf("mperf delta should not equal 0 on core %s", cpuID)
		return
	}
	aperfMperf := float64(p.msr.GetAperfDelta(cpuID)) / float64(p.msr.GetMperfDelta(cpuID))
	tsc := services.ConvertProcessorCyclesToHertz(p.msr.GetTimestampDelta(cpuID))
	timeNow := time.Now().UnixNano()
	interval := services.ConvertNanoSecondsToSeconds(timeNow - p.msr.GetReadDate(cpuID))
	p.msr.SetReadDate(cpuID, timeNow)

	if p.skipFirstIteration {
		return
	}

	busyMhzValue := services.RoundFloatToNearestTwoDecimalPlaces(tsc * aperfMperf / interval)

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.PhysicalID,
		"core_id":    cpu.CoreID,
		"cpu_id":     cpu.CPUID,
	}
	fields := map[string]interface{}{
		"name":  "cpu_busy_frequency",
		"unit":  "Mhz",
		"value": busyMhzValue,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addGlobalMetrics(acc telegraf.Accumulator) {
	// Prepare RAPL data each gather because there is a possibility to disable rapl kernel module
	p.rapl.InitializeRaplData()

	for _, socketID := range p.rapl.GetSocketIDs() {
		err := p.rapl.RetrieveAndCalculateData(socketID)
		if err != nil {
			// In case of an error skip calculating metrics for this socket
			p.Log.Errorf("error fetching rapl data for socket %s, err: %v", socketID, err)
			continue
		}
		p.addThermalDesignPowerMetric(socketID, acc)
		if !p.skipFirstIteration {
			p.addCurrentSocketPowerConsumption(socketID, acc)
			p.addCurrentDramPowerConsumption(socketID, acc)
		}
	}
}

func (p *PowerStat) addThermalDesignPowerMetric(socketID string, acc telegraf.Accumulator) {
	maxPower, err := p.rapl.GetConstraintMaxPower(socketID)
	if err != nil {
		p.Log.Errorf("error while retrieving TDP of the socket %s, err: %v", socketID, err)
		return
	}

	tags := map[string]string{
		"package_id": socketID,
	}

	fields := map[string]interface{}{
		"name":  "thermal_design_power",
		"unit":  "Watt",
		"value": services.RoundFloatToNearestTwoDecimalPlaces(services.ConvertMicroWattToWatt(maxPower)),
	}

	acc.AddGauge("powerstat_package", fields, tags)
}

func (p *PowerStat) addCurrentSocketPowerConsumption(socketID string, acc telegraf.Accumulator) {
	tags := map[string]string{
		"package_id": socketID,
	}

	fields := map[string]interface{}{
		"name":  "current_power_consumption",
		"unit":  "Watt",
		"value": services.RoundFloatToNearestTwoDecimalPlaces(p.rapl.GetCurrentPackagePowerConsumption(socketID)),
	}

	acc.AddGauge("powerstat_package", fields, tags)
}

func (p *PowerStat) addCurrentDramPowerConsumption(socketID string, acc telegraf.Accumulator) {
	tags := map[string]string{
		"package_id": socketID,
	}

	fields := map[string]interface{}{
		"name":  "current_dram_power_consumption",
		"unit":  "Watt",
		"value": services.RoundFloatToNearestTwoDecimalPlaces(p.rapl.GetCurrentDramPowerConsumption(socketID)),
	}

	acc.AddGauge("powerstat_package", fields, tags)
}

func (p *PowerStat) verifyProcessor() error {
	stats, err := p.fs.GetCPUInfoStats()
	p.cpuInfo = stats

	if err != nil {
		return err
	}

	// First CPU is sufficient for verification.
	firstCPU := p.cpuInfo["0"]
	if firstCPU == nil {
		return fmt.Errorf("first core not found while parsing /proc/cpuinfo")
	}

	if firstCPU.VendorID != "GenuineIntel" || firstCPU.CPUFamily != "6" {
		return fmt.Errorf("Intel processor not found, vendorId: %s", firstCPU.VendorID)
	}

	// Disable c1CoreStateResidency and c6CoreStateResidency for specific processors.
	allowedProcessors, err := services.ConvertHexArrayToIntegerArray(allowedProcessorModelsForC1C6)
	if err != nil {
		return err
	}
	if !contains(services.ConvertIntegerArrayToStringArray(allowedProcessors), firstCPU.Model) {
		p.c1CoreStateResidency = false
		p.c6CoreStateResidency = false
	}

	if !strings.Contains(firstCPU.Flags, "msr") {
		p.coreCPUTemperature = false
		p.c6CoreStateResidency = false
		p.processorBusyCycles = false
		p.processorBusyFrequency = false
		p.c1CoreStateResidency = false
	}

	if !strings.Contains(firstCPU.Flags, "aperfmperf") {
		p.processorBusyFrequency = false
		p.processorBusyCycles = false
		p.c1CoreStateResidency = false
	}

	if !strings.Contains(firstCPU.Flags, "dts") {
		p.coreCPUTemperature = false
	}

	return nil
}

// ParseCPUMetricsConfig enables specific metrics based on config values.
func (p *PowerStat) ParseCPUMetricsConfig() {
	if len(p.CPUMetrics) == 0 {
		p.perCoreMetrics = false
		return
	}

	if contains(p.CPUMetrics, cpuFrequency) {
		p.coreCPUFrequency = true
	}

	if contains(p.CPUMetrics, cpuC1) {
		p.c1CoreStateResidency = true
	}

	if contains(p.CPUMetrics, cpuC6) {
		p.c6CoreStateResidency = true
	}

	if contains(p.CPUMetrics, cpuBusyCycles) {
		p.processorBusyCycles = true
	}

	if contains(p.CPUMetrics, cpuBusyFrequency) {
		p.processorBusyFrequency = true
	}

	if contains(p.CPUMetrics, cpuTemperature) {
		p.coreCPUTemperature = true
	}
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}

	return false
}

// NewPowerStat creates and returns PowerStat struct.
func NewPowerStat(fs services.FileService) *PowerStat {
	p := &PowerStat{
		perCoreMetrics:         false,
		coreCPUFrequency:       false,
		c1CoreStateResidency:   false,
		c6CoreStateResidency:   false,
		processorBusyCycles:    false,
		coreCPUTemperature:     false,
		processorBusyFrequency: false,
		skipFirstIteration:     true,
		fs:                     fs,
	}

	return p
}

func init() {
	inputs.Add("intel_powerstat", func() telegraf.Input {
		return NewPowerStat(services.NewFileService())
	})
}
