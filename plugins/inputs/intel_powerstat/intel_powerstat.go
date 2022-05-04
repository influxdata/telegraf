//go:build linux
// +build linux

package intel_powerstat

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	cpuFrequency         = "cpu_frequency"
	cpuBusyFrequency     = "cpu_busy_frequency"
	cpuTemperature       = "cpu_temperature"
	cpuC1StateResidency  = "cpu_c1_state_residency"
	cpuC6StateResidency  = "cpu_c6_state_residency"
	cpuBusyCycles        = "cpu_busy_cycles"
	percentageMultiplier = 100
)

// PowerStat plugin enables monitoring of platform metrics (power, TDP) and Core metrics like temperature, power and utilization.
type PowerStat struct {
	CPUMetrics []string        `toml:"cpu_metrics"`
	Log        telegraf.Logger `toml:"-"`

	fs   fileService
	rapl raplService
	msr  msrService

	cpuFrequency        bool
	cpuBusyFrequency    bool
	cpuTemperature      bool
	cpuC1StateResidency bool
	cpuC6StateResidency bool
	cpuBusyCycles       bool
	cpuInfo             map[string]*cpuInfo
	skipFirstIteration  bool
}

// Init performs one time setup of the plugin.
func (p *PowerStat) Init() error {
	p.parseCPUMetricsConfig()
	err := p.verifyProcessor()
	if err != nil {
		return err
	}
	// Initialize MSR service only when there is at least one core metric enabled.
	if p.cpuFrequency || p.cpuBusyFrequency || p.cpuTemperature || p.cpuC1StateResidency ||
		p.cpuC6StateResidency || p.cpuBusyCycles {
		p.msr = newMsrServiceWithFs(p.Log, p.fs)
	}
	p.rapl = newRaplServiceWithFs(p.Log, p.fs)

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input gathers.
func (p *PowerStat) Gather(acc telegraf.Accumulator) error {
	p.addGlobalMetrics(acc)

	if p.areCoreMetricsEnabled() {
		p.addPerCoreMetrics(acc)
	}

	// Gathering the first iteration of metrics was skipped for most of them because they are based on delta calculations.
	p.skipFirstIteration = false

	return nil
}

func (p *PowerStat) addGlobalMetrics(acc telegraf.Accumulator) {
	// Prepare RAPL data each gather because there is a possibility to disable rapl kernel module
	p.rapl.initializeRaplData()

	for socketID := range p.rapl.getRaplData() {
		err := p.rapl.retrieveAndCalculateData(socketID)
		if err != nil {
			// In case of an error skip calculating metrics for this socket
			p.Log.Errorf("error fetching rapl data for socket %s, err: %v", socketID, err)
			continue
		}
		p.addThermalDesignPowerMetric(socketID, acc)
		if p.skipFirstIteration {
			continue
		}
		p.addCurrentSocketPowerConsumption(socketID, acc)
		p.addCurrentDramPowerConsumption(socketID, acc)
	}
}

func (p *PowerStat) addThermalDesignPowerMetric(socketID string, acc telegraf.Accumulator) {
	maxPower, err := p.rapl.getConstraintMaxPowerWatts(socketID)
	if err != nil {
		p.Log.Errorf("error while retrieving TDP of the socket %s, err: %v", socketID, err)
		return
	}

	tags := map[string]string{
		"package_id": socketID,
	}

	fields := map[string]interface{}{
		"thermal_design_power_watts": roundFloatToNearestTwoDecimalPlaces(maxPower),
	}

	acc.AddGauge("powerstat_package", fields, tags)
}

func (p *PowerStat) addCurrentSocketPowerConsumption(socketID string, acc telegraf.Accumulator) {
	tags := map[string]string{
		"package_id": socketID,
	}

	fields := map[string]interface{}{
		"current_power_consumption_watts": roundFloatToNearestTwoDecimalPlaces(p.rapl.getRaplData()[socketID].socketCurrentEnergy),
	}

	acc.AddGauge("powerstat_package", fields, tags)
}

func (p *PowerStat) addCurrentDramPowerConsumption(socketID string, acc telegraf.Accumulator) {
	tags := map[string]string{
		"package_id": socketID,
	}

	fields := map[string]interface{}{
		"current_dram_power_consumption_watts": roundFloatToNearestTwoDecimalPlaces(p.rapl.getRaplData()[socketID].dramCurrentEnergy),
	}

	acc.AddGauge("powerstat_package", fields, tags)
}

func (p *PowerStat) addPerCoreMetrics(acc telegraf.Accumulator) {
	var wg sync.WaitGroup
	wg.Add(len(p.msr.getCPUCoresData()))

	for cpuID := range p.msr.getCPUCoresData() {
		go p.addMetricsForSingleCore(cpuID, acc, &wg)
	}

	wg.Wait()
}

func (p *PowerStat) addMetricsForSingleCore(cpuID string, acc telegraf.Accumulator, wg *sync.WaitGroup) {
	defer wg.Done()

	if p.cpuFrequency {
		p.addCPUFrequencyMetric(cpuID, acc)
	}

	// Read data from MSR only if required
	if p.cpuC1StateResidency || p.cpuC6StateResidency || p.cpuBusyCycles || p.cpuTemperature ||
		p.cpuBusyFrequency {
		err := p.msr.openAndReadMsr(cpuID)
		if err != nil {
			// In case of an error exit the function. All metrics past this point are dependant on MSR.
			p.Log.Debugf("error while reading msr: %v", err)
			return
		}
	}

	if p.cpuTemperature {
		p.addCPUTemperatureMetric(cpuID, acc)
	}

	// cpuBusyFrequency metric does some calculations inside that are required in another plugin cycle.
	if p.cpuBusyFrequency {
		p.addCPUBusyFrequencyMetric(cpuID, acc)
	}

	if !p.skipFirstIteration {
		if p.cpuC1StateResidency {
			p.addCPUC1StateResidencyMetric(cpuID, acc)
		}

		if p.cpuC6StateResidency {
			p.addCPUC6StateResidencyMetric(cpuID, acc)
		}

		if p.cpuBusyCycles {
			p.addCPUBusyCyclesMetric(cpuID, acc)
		}
	}
}

func (p *PowerStat) addCPUFrequencyMetric(cpuID string, acc telegraf.Accumulator) {
	frequency, err := p.msr.retrieveCPUFrequencyForCore(cpuID)

	// In case of an error leave func
	if err != nil {
		p.Log.Debugf("error while reading file: %v", err)
		return
	}

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.physicalID,
		"core_id":    cpu.coreID,
		"cpu_id":     cpu.cpuID,
	}

	fields := map[string]interface{}{
		"cpu_frequency_mhz": roundFloatToNearestTwoDecimalPlaces(frequency),
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addCPUTemperatureMetric(cpuID string, acc telegraf.Accumulator) {
	coresData := p.msr.getCPUCoresData()
	temp := coresData[cpuID].throttleTemp - coresData[cpuID].temp

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.physicalID,
		"core_id":    cpu.coreID,
		"cpu_id":     cpu.cpuID,
	}
	fields := map[string]interface{}{
		"cpu_temperature_celsius": temp,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addCPUBusyFrequencyMetric(cpuID string, acc telegraf.Accumulator) {
	coresData := p.msr.getCPUCoresData()
	mperfDelta := coresData[cpuID].mperfDelta
	// Avoid division by 0
	if mperfDelta == 0 {
		p.Log.Errorf("mperf delta should not equal 0 on core %s", cpuID)
		return
	}
	aperfMperf := float64(coresData[cpuID].aperfDelta) / float64(mperfDelta)
	tsc := convertProcessorCyclesToHertz(coresData[cpuID].timeStampCounterDelta)
	timeNow := time.Now().UnixNano()
	interval := convertNanoSecondsToSeconds(timeNow - coresData[cpuID].readDate)
	coresData[cpuID].readDate = timeNow

	if p.skipFirstIteration {
		return
	}

	if interval == 0 {
		p.Log.Errorf("interval between last two Telegraf cycles is 0")
		return
	}

	busyMhzValue := roundFloatToNearestTwoDecimalPlaces(tsc * aperfMperf / interval)

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.physicalID,
		"core_id":    cpu.coreID,
		"cpu_id":     cpu.cpuID,
	}
	fields := map[string]interface{}{
		"cpu_busy_frequency_mhz": busyMhzValue,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addCPUC1StateResidencyMetric(cpuID string, acc telegraf.Accumulator) {
	coresData := p.msr.getCPUCoresData()
	timestampDeltaBig := new(big.Int).SetUint64(coresData[cpuID].timeStampCounterDelta)
	// Avoid division by 0
	if timestampDeltaBig.Sign() < 1 {
		p.Log.Errorf("timestamp delta value %v should not be lower than 1", timestampDeltaBig)
		return
	}

	// Since counter collection is not atomic it may happen that sum of C0, C1, C3, C6 and C7
	// is bigger value than TSC, in such case C1 residency shall be set to 0.
	// Operating on big.Int to avoid overflow
	mperfDeltaBig := new(big.Int).SetUint64(coresData[cpuID].mperfDelta)
	c3DeltaBig := new(big.Int).SetUint64(coresData[cpuID].c3Delta)
	c6DeltaBig := new(big.Int).SetUint64(coresData[cpuID].c6Delta)
	c7DeltaBig := new(big.Int).SetUint64(coresData[cpuID].c7Delta)

	c1Big := new(big.Int).Sub(timestampDeltaBig, mperfDeltaBig)
	c1Big.Sub(c1Big, c3DeltaBig)
	c1Big.Sub(c1Big, c6DeltaBig)
	c1Big.Sub(c1Big, c7DeltaBig)

	if c1Big.Sign() < 0 {
		c1Big = c1Big.SetInt64(0)
	}
	c1Value := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(c1Big.Uint64()) / float64(timestampDeltaBig.Uint64()))

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.physicalID,
		"core_id":    cpu.coreID,
		"cpu_id":     cpu.cpuID,
	}
	fields := map[string]interface{}{
		"cpu_c1_state_residency_percent": c1Value,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addCPUC6StateResidencyMetric(cpuID string, acc telegraf.Accumulator) {
	coresData := p.msr.getCPUCoresData()
	// Avoid division by 0
	if coresData[cpuID].timeStampCounterDelta == 0 {
		p.Log.Errorf("timestamp counter on offset %s should not equal 0 on cpuID %s",
			timestampCounterLocation, cpuID)
		return
	}
	c6Value := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(coresData[cpuID].c6Delta) / float64(coresData[cpuID].timeStampCounterDelta))

	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.physicalID,
		"core_id":    cpu.coreID,
		"cpu_id":     cpu.cpuID,
	}
	fields := map[string]interface{}{
		"cpu_c6_state_residency_percent": c6Value,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) addCPUBusyCyclesMetric(cpuID string, acc telegraf.Accumulator) {
	coresData := p.msr.getCPUCoresData()
	// Avoid division by 0
	if coresData[cpuID].timeStampCounterDelta == 0 {
		p.Log.Errorf("timestamp counter on offset %s should not equal 0 on cpuID %s",
			timestampCounterLocation, cpuID)
		return
	}
	busyCyclesValue := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(coresData[cpuID].mperfDelta) / float64(coresData[cpuID].timeStampCounterDelta))
	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.physicalID,
		"core_id":    cpu.coreID,
		"cpu_id":     cpu.cpuID,
	}
	fields := map[string]interface{}{
		"cpu_busy_cycles_percent": busyCyclesValue,
	}

	acc.AddGauge("powerstat_core", fields, tags)
}

func (p *PowerStat) parseCPUMetricsConfig() {
	if len(p.CPUMetrics) == 0 {
		return
	}

	if contains(p.CPUMetrics, cpuFrequency) {
		p.cpuFrequency = true
	}

	if contains(p.CPUMetrics, cpuC1StateResidency) {
		p.cpuC1StateResidency = true
	}

	if contains(p.CPUMetrics, cpuC6StateResidency) {
		p.cpuC6StateResidency = true
	}

	if contains(p.CPUMetrics, cpuBusyCycles) {
		p.cpuBusyCycles = true
	}

	if contains(p.CPUMetrics, cpuBusyFrequency) {
		p.cpuBusyFrequency = true
	}

	if contains(p.CPUMetrics, cpuTemperature) {
		p.cpuTemperature = true
	}
}

func (p *PowerStat) verifyProcessor() error {
	allowedProcessorModelsForC1C6 := []int64{0x37, 0x4D, 0x5C, 0x5F, 0x7A, 0x4C, 0x86, 0x96, 0x9C,
		0x1A, 0x1E, 0x1F, 0x2E, 0x25, 0x2C, 0x2F, 0x2A, 0x2D, 0x3A, 0x3E, 0x4E, 0x5E, 0x55, 0x8E,
		0x9E, 0x6A, 0x6C, 0x7D, 0x7E, 0x9D, 0x3C, 0x3F, 0x45, 0x46, 0x3D, 0x47, 0x4F, 0x56,
		0x66, 0x57, 0x85, 0xA5, 0xA6, 0x8F, 0x8C, 0x8D}
	stats, err := p.fs.getCPUInfoStats()
	if err != nil {
		return err
	}

	p.cpuInfo = stats

	// First CPU is sufficient for verification.
	firstCPU := p.cpuInfo["0"]
	if firstCPU == nil {
		return fmt.Errorf("first core not found while parsing /proc/cpuinfo")
	}

	if firstCPU.vendorID != "GenuineIntel" || firstCPU.cpuFamily != "6" {
		return fmt.Errorf("Intel processor not found, vendorId: %s", firstCPU.vendorID)
	}

	if !contains(convertIntegerArrayToStringArray(allowedProcessorModelsForC1C6), firstCPU.model) {
		p.cpuC1StateResidency = false
		p.cpuC6StateResidency = false
	}

	if !strings.Contains(firstCPU.flags, "msr") {
		p.cpuTemperature = false
		p.cpuC6StateResidency = false
		p.cpuBusyCycles = false
		p.cpuBusyFrequency = false
		p.cpuC1StateResidency = false
	}

	if !strings.Contains(firstCPU.flags, "aperfmperf") {
		p.cpuBusyFrequency = false
		p.cpuBusyCycles = false
		p.cpuC1StateResidency = false
	}

	if !strings.Contains(firstCPU.flags, "dts") {
		p.cpuTemperature = false
	}

	return nil
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}

	return false
}

func (p *PowerStat) areCoreMetricsEnabled() bool {
	return p.msr != nil && len(p.msr.getCPUCoresData()) > 0
}

// newPowerStat creates and returns PowerStat struct.
func newPowerStat(fs fileService) *PowerStat {
	p := &PowerStat{
		cpuFrequency:        false,
		cpuC1StateResidency: false,
		cpuC6StateResidency: false,
		cpuBusyCycles:       false,
		cpuTemperature:      false,
		cpuBusyFrequency:    false,
		skipFirstIteration:  true,
		fs:                  fs,
	}

	return p
}

func init() {
	inputs.Add("intel_powerstat", func() telegraf.Input {
		return newPowerStat(newFileService())
	})
}
