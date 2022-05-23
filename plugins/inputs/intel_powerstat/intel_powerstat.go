//go:build linux
// +build linux

package intel_powerstat

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	cpuFrequency                       = "cpu_frequency"
	cpuBusyFrequency                   = "cpu_busy_frequency"
	cpuTemperature                     = "cpu_temperature"
	cpuC0StateResidency                = "cpu_c0_state_residency"
	cpuC1StateResidency                = "cpu_c1_state_residency"
	cpuC6StateResidency                = "cpu_c6_state_residency"
	cpuBusyCycles                      = "cpu_busy_cycles"
	packageCurrentPowerConsumption     = "current_power_consumption"
	packageCurrentDramPowerConsumption = "current_dram_power_consumption"
	packageThermalDesignPower          = "thermal_design_power"
	packageTurboLimit                  = "max_turbo_frequency"
	percentageMultiplier               = 100
)

// PowerStat plugin enables monitoring of platform metrics (power, TDP) and Core metrics like temperature, power and utilization.
type PowerStat struct {
	CPUMetrics     []string        `toml:"cpu_metrics"`
	PackageMetrics []string        `toml:"package_metrics"`
	Log            telegraf.Logger `toml:"-"`

	fs   fileService
	rapl raplService
	msr  msrService

	cpuFrequency                       bool
	cpuBusyFrequency                   bool
	cpuTemperature                     bool
	cpuC0StateResidency                bool
	cpuC1StateResidency                bool
	cpuC6StateResidency                bool
	cpuBusyCycles                      bool
	packageTurboLimit                  bool
	packageCurrentPowerConsumption     bool
	packageCurrentDramPowerConsumption bool
	packageThermalDesignPower          bool
	cpuInfo                            map[string]*cpuInfo
	skipFirstIteration                 bool
	logOnce                            map[string]error
}

// Init performs one time setup of the plugin
func (p *PowerStat) Init() error {
	p.parsePackageMetricsConfig()
	p.parseCPUMetricsConfig()
	err := p.verifyProcessor()
	if err != nil {
		return err
	}
	// Initialize MSR service only when there is at least one metric enabled
	if p.cpuFrequency || p.cpuBusyFrequency || p.cpuTemperature || p.cpuC0StateResidency || p.cpuC1StateResidency ||
		p.cpuC6StateResidency || p.cpuBusyCycles || p.packageTurboLimit {
		p.msr = newMsrServiceWithFs(p.Log, p.fs)
	}
	if p.packageCurrentPowerConsumption || p.packageCurrentDramPowerConsumption || p.packageThermalDesignPower || p.packageTurboLimit {
		p.rapl = newRaplServiceWithFs(p.Log, p.fs)
	}

	if !p.areCoreMetricsEnabled() && !p.areGlobalMetricsEnabled() {
		return fmt.Errorf("all configuration options are empty or invalid. Did not find anything to gather")
	}

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input gathers
func (p *PowerStat) Gather(acc telegraf.Accumulator) error {
	if p.areGlobalMetricsEnabled() {
		p.addGlobalMetrics(acc)
	}

	if p.areCoreMetricsEnabled() {
		p.addPerCoreMetrics(acc)
	}

	// Gathering the first iteration of metrics was skipped for most of them because they are based on delta calculations
	p.skipFirstIteration = false

	return nil
}

func (p *PowerStat) addGlobalMetrics(acc telegraf.Accumulator) {
	// Prepare RAPL data each gather because there is a possibility to disable rapl kernel module
	p.rapl.initializeRaplData()

	for socketID := range p.rapl.getRaplData() {
		if p.packageTurboLimit {
			p.addTurboRatioLimit(socketID, acc)
		}

		err := p.rapl.retrieveAndCalculateData(socketID)
		if err != nil {
			// In case of an error skip calculating metrics for this socket
			if val := p.logOnce[socketID]; val == nil || val.Error() != err.Error() {
				p.Log.Errorf("error fetching rapl data for socket %s, err: %v", socketID, err)
				// Remember that specific error occurs for socketID to omit logging next time
				p.logOnce[socketID] = err
			}
			continue
		}

		// If error stops occurring, clear logOnce indicator
		p.logOnce[socketID] = nil
		if p.packageThermalDesignPower {
			p.addThermalDesignPowerMetric(socketID, acc)
		}

		if p.skipFirstIteration {
			continue
		}
		if p.packageCurrentPowerConsumption {
			p.addCurrentSocketPowerConsumption(socketID, acc)
		}
		if p.packageCurrentDramPowerConsumption {
			p.addCurrentDramPowerConsumption(socketID, acc)
		}
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
	if p.cpuC0StateResidency || p.cpuC1StateResidency || p.cpuC6StateResidency || p.cpuBusyCycles || p.cpuTemperature || p.cpuBusyFrequency {
		err := p.msr.openAndReadMsr(cpuID)
		if err != nil {
			// In case of an error exit the function. All metrics past this point are dependent on MSR
			p.Log.Debugf("error while reading msr: %v", err)
			return
		}
	}

	if p.cpuTemperature {
		p.addCPUTemperatureMetric(cpuID, acc)
	}

	// cpuBusyFrequency metric does some calculations inside that are required in another plugin cycle
	if p.cpuBusyFrequency {
		p.addCPUBusyFrequencyMetric(cpuID, acc)
	}

	if !p.skipFirstIteration {
		if p.cpuC0StateResidency || p.cpuBusyCycles {
			p.addCPUC0StateResidencyMetric(cpuID, acc)
		}

		if p.cpuC1StateResidency {
			p.addCPUC1StateResidencyMetric(cpuID, acc)
		}

		if p.cpuC6StateResidency {
			p.addCPUC6StateResidencyMetric(cpuID, acc)
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

func calculateTurboRatioGroup(coreCounts uint64, msr uint64, group map[int]uint64) {
	from := coreCounts & 0xFF // value of number of active cores of bucket 1 is written in the first 8 bits. The next buckets values are saved on the following 8-bit sides
	for i := 0; i < 8; i++ {
		to := (coreCounts >> (i * 8)) & 0xFF
		if to == 0 {
			break
		}
		value := (msr >> (i * 8)) & 0xFF
		// value of freq ratio is stored in 8-bit blocks, and their real value is obtained after multiplication by 100
		if value != 0 && to != 0 {
			for ; from <= to; from++ {
				group[int(from)] = value * 100
			}
		}
		from = to + 1
	}
}

func (p *PowerStat) addTurboRatioLimit(socketID string, acc telegraf.Accumulator) {
	var err error
	turboRatioLimitGroups := make(map[int]uint64)

	var cpuID = ""
	var model = ""
	for _, v := range p.cpuInfo {
		if v.physicalID == socketID {
			cpuID = v.cpuID
			model = v.model
		}
	}
	if cpuID == "" || model == "" {
		p.Log.Debugf("error while reading socket ID")
		return
	}
	// dump_hsw_turbo_ratio_limit
	if model == strconv.FormatInt(0x3F, 10) { // INTEL_FAM6_HASWELL_X
		coreCounts := uint64(0x1211) // counting the number of active cores 17 and 18
		msrTurboRatioLimit2, err := p.msr.readSingleMsr(cpuID, "MSR_TURBO_RATIO_LIMIT2")
		if err != nil {
			p.Log.Debugf("error while reading MSR_TURBO_RATIO_LIMIT2: %v", err)
			return
		}

		calculateTurboRatioGroup(coreCounts, msrTurboRatioLimit2, turboRatioLimitGroups)
	}

	// dump_ivt_turbo_ratio_limit
	if (model == strconv.FormatInt(0x3E, 10)) || // INTEL_FAM6_IVYBRIDGE_X
		(model == strconv.FormatInt(0x3F, 10)) { // INTEL_FAM6_HASWELL_X
		coreCounts := uint64(0x100F0E0D0C0B0A09) // counting the number of active cores 9 to 16
		msrTurboRatioLimit1, err := p.msr.readSingleMsr(cpuID, "MSR_TURBO_RATIO_LIMIT1")
		if err != nil {
			p.Log.Debugf("error while reading MSR_TURBO_RATIO_LIMIT1: %v", err)
			return
		}
		calculateTurboRatioGroup(coreCounts, msrTurboRatioLimit1, turboRatioLimitGroups)
	}

	if (model != strconv.FormatInt(0x37, 10)) && // INTEL_FAM6_ATOM_SILVERMONT
		(model != strconv.FormatInt(0x4A, 10)) && // INTEL_FAM6_ATOM_SILVERMONT_MID:
		(model != strconv.FormatInt(0x5A, 10)) && // INTEL_FAM6_ATOM_AIRMONT_MID:
		(model != strconv.FormatInt(0x2E, 10)) && // INTEL_FAM6_NEHALEM_EX
		(model != strconv.FormatInt(0x2F, 10)) && // INTEL_FAM6_WESTMERE_EX
		(model != strconv.FormatInt(0x57, 10)) && // INTEL_FAM6_XEON_PHI_KNL
		(model != strconv.FormatInt(0x85, 10)) { // INTEL_FAM6_XEON_PHI_KNM
		coreCounts := uint64(0x0807060504030201)     // default value (counting the number of active cores 1 to 8). May be changed in "if" segment below
		if (model == strconv.FormatInt(0x5C, 10)) || // INTEL_FAM6_ATOM_GOLDMONT
			(model == strconv.FormatInt(0x55, 10)) || // INTEL_FAM6_SKYLAKE_X
			(model == strconv.FormatInt(0x6C, 10) || model == strconv.FormatInt(0x8F, 10) || model == strconv.FormatInt(0x6A, 10)) || // INTEL_FAM6_ICELAKE_X
			(model == strconv.FormatInt(0x5F, 10)) || // INTEL_FAM6_ATOM_GOLDMONT_D
			(model == strconv.FormatInt(0x86, 10)) { // INTEL_FAM6_ATOM_TREMONT_D
			coreCounts, err = p.msr.readSingleMsr(cpuID, "MSR_TURBO_RATIO_LIMIT1")

			if err != nil {
				p.Log.Debugf("error while reading MSR_TURBO_RATIO_LIMIT1: %v", err)
				return
			}
		}

		msrTurboRatioLimit, err := p.msr.readSingleMsr(cpuID, "MSR_TURBO_RATIO_LIMIT")
		if err != nil {
			p.Log.Debugf("error while reading MSR_TURBO_RATIO_LIMIT: %v", err)
			return
		}
		calculateTurboRatioGroup(coreCounts, msrTurboRatioLimit, turboRatioLimitGroups)
	}
	// dump_atom_turbo_ratio_limits
	if model == strconv.FormatInt(0x37, 10) || // INTEL_FAM6_ATOM_SILVERMONT
		model == strconv.FormatInt(0x4A, 10) || // INTEL_FAM6_ATOM_SILVERMONT_MID:
		model == strconv.FormatInt(0x5A, 10) { // INTEL_FAM6_ATOM_AIRMONT_MID
		coreCounts := uint64(0x04030201) // counting the number of active cores 1 to 4
		msrTurboRatioLimit, err := p.msr.readSingleMsr(cpuID, "MSR_ATOM_CORE_TURBO_RATIOS")

		if err != nil {
			p.Log.Debugf("error while reading MSR_ATOM_CORE_TURBO_RATIOS: %v", err)
			return
		}
		value := uint64(0)
		newValue := uint64(0)

		for i := 0; i < 4; i++ { // value "4" is specific for this group of processors
			newValue = (msrTurboRatioLimit >> (8 * (i))) & 0x3F // value of freq ratio is stored in 6-bit blocks, saved every 8 bits
			value = value + (newValue << ((i - 1) * 8))         // now value of freq ratio is stored in 8-bit blocks, saved every 8 bits
		}

		calculateTurboRatioGroup(coreCounts, value, turboRatioLimitGroups)
	}
	// dump_knl_turbo_ratio_limits
	if model == strconv.FormatInt(0x57, 10) { // INTEL_FAM6_XEON_PHI_KNL
		msrTurboRatioLimit, err := p.msr.readSingleMsr(cpuID, "MSR_TURBO_RATIO_LIMIT")
		if err != nil {
			p.Log.Debugf("error while reading MSR_TURBO_RATIO_LIMIT: %v", err)
			return
		}

		// value of freq ratio of bucket 1 is saved in bits 15 to 8.
		// each next value is calculated as the previous value - delta. Delta is stored in 3-bit blocks every 8 bits (start at 21 (2*8+5))
		value := (msrTurboRatioLimit >> 8) & 0xFF
		newValue := value
		for i := 2; i < 8; i++ {
			newValue = newValue - (msrTurboRatioLimit>>(8*i+5))&0x7
			value = value + (newValue << ((i - 1) * 8))
		}

		// value of number of active cores of bucket 1 is saved in bits 1 to 7.
		// each next value is calculated as the previous value + delta. Delta is stored in 5-bit blocks every 8 bits (start at 16 (2*8))
		coreCounts := (msrTurboRatioLimit & 0xFF) >> 1
		newBucket := coreCounts
		for i := 2; i < 8; i++ {
			newBucket = newBucket + (msrTurboRatioLimit>>(8*i))&0x1F
			coreCounts = coreCounts + (newBucket << ((i - 1) * 8))
		}
		calculateTurboRatioGroup(coreCounts, value, turboRatioLimitGroups)
	}

	for key, val := range turboRatioLimitGroups {
		tags := map[string]string{
			"package_id":   socketID,
			"active_cores": strconv.Itoa(key),
		}
		fields := map[string]interface{}{
			"max_turbo_frequency_mhz": val,
		}
		acc.AddGauge("powerstat_package", fields, tags)
	}
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

func (p *PowerStat) addCPUC0StateResidencyMetric(cpuID string, acc telegraf.Accumulator) {
	coresData := p.msr.getCPUCoresData()
	// Avoid division by 0
	if coresData[cpuID].timeStampCounterDelta == 0 {
		p.Log.Errorf("timestamp counter on offset %s should not equal 0 on cpuID %s",
			timestampCounterLocation, cpuID)
		return
	}
	c0Value := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(coresData[cpuID].mperfDelta) / float64(coresData[cpuID].timeStampCounterDelta))
	cpu := p.cpuInfo[cpuID]
	tags := map[string]string{
		"package_id": cpu.physicalID,
		"core_id":    cpu.coreID,
		"cpu_id":     cpu.cpuID,
	}
	if p.cpuC0StateResidency {
		fields := map[string]interface{}{
			"cpu_c0_state_residency_percent": c0Value,
		}
		acc.AddGauge("powerstat_core", fields, tags)
	}
	if p.cpuBusyCycles {
		deprecatedFields := map[string]interface{}{
			"cpu_busy_cycles_percent": c0Value,
		}
		acc.AddGauge("powerstat_core", deprecatedFields, tags)
	}
}

func (p *PowerStat) parsePackageMetricsConfig() {
	if p.PackageMetrics == nil {
		// if Package Metric config is empty, use the default settings.
		p.packageCurrentPowerConsumption = true
		p.packageCurrentDramPowerConsumption = true
		p.packageThermalDesignPower = true
		return
	}

	if contains(p.PackageMetrics, packageTurboLimit) {
		p.packageTurboLimit = true
	}
	if contains(p.PackageMetrics, packageCurrentPowerConsumption) {
		p.packageCurrentPowerConsumption = true
	}

	if contains(p.PackageMetrics, packageCurrentDramPowerConsumption) {
		p.packageCurrentDramPowerConsumption = true
	}
	if contains(p.PackageMetrics, packageThermalDesignPower) {
		p.packageThermalDesignPower = true
	}
}

func (p *PowerStat) parseCPUMetricsConfig() {
	if len(p.CPUMetrics) == 0 {
		return
	}

	if contains(p.CPUMetrics, cpuFrequency) {
		p.cpuFrequency = true
	}

	if contains(p.CPUMetrics, cpuC0StateResidency) {
		p.cpuC0StateResidency = true
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

	// First CPU is sufficient for verification
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
		p.cpuC0StateResidency = false
		p.cpuBusyCycles = false
		p.cpuBusyFrequency = false
		p.cpuC1StateResidency = false
	}

	if !strings.Contains(firstCPU.flags, "aperfmperf") {
		p.cpuBusyCycles = false
		p.cpuBusyFrequency = false
		p.cpuC0StateResidency = false
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

func (p *PowerStat) areGlobalMetricsEnabled() bool {
	return p.rapl != nil
}

// newPowerStat creates and returns PowerStat struct
func newPowerStat(fs fileService) *PowerStat {
	p := &PowerStat{
		cpuFrequency:                       false,
		cpuC0StateResidency:                false,
		cpuC1StateResidency:                false,
		cpuC6StateResidency:                false,
		cpuBusyCycles:                      false,
		cpuTemperature:                     false,
		cpuBusyFrequency:                   false,
		packageTurboLimit:                  false,
		packageCurrentPowerConsumption:     false,
		packageCurrentDramPowerConsumption: false,
		packageThermalDesignPower:          false,
		skipFirstIteration:                 true,
		fs:                                 fs,
		logOnce:                            make(map[string]error),
	}

	return p
}

func init() {
	inputs.Add("intel_powerstat", func() telegraf.Input {
		return newPowerStat(newFileService())
	})
}
