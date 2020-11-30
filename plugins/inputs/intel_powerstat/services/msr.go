// +build linux

package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/data"
	"golang.org/x/sync/errgroup"
)

const (
	systemCPUPath                      = "/sys/devices/system/cpu/"
	cpuCurrentFreqPartialPath          = "/sys/devices/system/cpu/cpu%s/cpufreq/scaling_cur_freq"
	msrPartialPath                     = "/dev/cpu/%s/msr"
	c3StateResidencyLocation           = "0x3FC"
	c6StateResidencyLocation           = "0x3FD"
	c7StateResidencyLocation           = "0x3FE"
	maximumFrequencyClockCountLocation = "0xE7"
	actualFrequencyClockCountLocation  = "0xE8"
	throttleTemperatureLocation        = "0x1A2"
	temperatureLocation                = "0x19C"
	// TimestampCounterLocation is an MSR offset.
	TimestampCounterLocation = "0x10"
)

// MsrService is responsible for interactions with MSR.
type MsrService interface {
	OpenAndReadMsr(core string) error
	RetrieveCPUFrequencyForCore(core string) (float64, int64, error)
	GetCpuIDs() []string
	GetThrottleTemperature(CPU string) uint64
	GetTemperature(CPU string) uint64
	GetTimestampDelta(CPU string) uint64
	GetMperfDelta(CPU string) uint64
	GetAperfDelta(CPU string) uint64
	GetC6Delta(CPU string) uint64
	GetC3Delta(CPU string) uint64
	GetC7Delta(CPU string) uint64
	GetReadDate(CPU string) int64
	SetReadDate(CPU string, date int64)
}

// MsrServiceImpl is implementation of MsrService.
type MsrServiceImpl struct {
	cpuCoresData map[string]*data.MsrData
	msrOffsets   map[string]int64
	fs           FileService
	log          telegraf.Logger
}

// GetCpuIDs returns array of logical CPUs.
func (m *MsrServiceImpl) GetCpuIDs() []string {
	IDs := make([]string, 0)

	for cpuID := range m.cpuCoresData {
		IDs = append(IDs, cpuID)
	}

	return IDs
}

// GetThrottleTemperature returns throttle temperature for CPU.
func (m *MsrServiceImpl) GetThrottleTemperature(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].ThrottleTemp
}

// GetTemperature returns temperature for CPU.
func (m *MsrServiceImpl) GetTemperature(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].Temp
}

// GetTimestampDelta returns timestamp delta for CPU.
func (m *MsrServiceImpl) GetTimestampDelta(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].TscDelta
}

// GetMperfDelta returns mperf delta for CPU.
func (m *MsrServiceImpl) GetMperfDelta(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].MperfDelta
}

// GetAperfDelta returns aperf delta for CPU.
func (m *MsrServiceImpl) GetAperfDelta(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].AperfDelta
}

// GetC6Delta returns C6 delta for CPU.
func (m *MsrServiceImpl) GetC6Delta(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].C6Delta
}

// GetC3Delta returns C3 delta for CPU.
func (m *MsrServiceImpl) GetC3Delta(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].C3Delta
}

// GetC7Delta returns C7 delta for CPU.
func (m *MsrServiceImpl) GetC7Delta(cpuID string) uint64 {
	return m.cpuCoresData[cpuID].C7Delta
}

// GetReadDate returns last read date.
func (m *MsrServiceImpl) GetReadDate(cpuID string) int64 {
	return m.cpuCoresData[cpuID].ReadDate
}

// SetReadDate sets last read date.
func (m *MsrServiceImpl) SetReadDate(cpuID string, date int64) {
	m.cpuCoresData[cpuID].ReadDate = date
}

// OpenAndReadMsr open and reads data from MSR.
func (m *MsrServiceImpl) OpenAndReadMsr(core string) error {
	path := fmt.Sprintf(msrPartialPath, core)
	msrFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening MSR file on path %s, err: %v", path, err)
	}
	defer msrFile.Close()

	err = m.readDataFromMsr(core, msrFile)
	if err != nil {
		return fmt.Errorf("error reading data from MSR for core %s, err: %v", core, err)
	}
	return nil
}

// RetrieveCPUFrequencyForCore retrieves CPU frequency for given core.
func (m *MsrServiceImpl) RetrieveCPUFrequencyForCore(core string) (float64, int64, error) {
	cpuFreqPath := fmt.Sprintf(cpuCurrentFreqPartialPath, core)
	cpuFreqFile, err := os.Open(cpuFreqPath)
	if err != nil {
		return 0, 0, fmt.Errorf("error opening scaling_cur_freq file on path %s, err: %v", cpuFreqPath, err)
	}
	defer cpuFreqFile.Close()
	return m.fs.ReadFileToFloat64(cpuFreqFile)
}

func (m *MsrServiceImpl) readDataFromMsr(core string, reader io.ReaderAt) error {
	g, ctx := errgroup.WithContext(context.Background())

	// Create and populate a map that contains msr offsets along with their respective channels
	msrOffsetsWithChannels := make(map[string]chan uint64)
	for offset := range m.msrOffsets {
		msrOffsetsWithChannels[offset] = make(chan uint64)
	}

	// Start a goroutine for each msr offset
	for offset, channel := range msrOffsetsWithChannels {
		// Wrap around function to avoid race on loop counter
		func(off string, ch chan uint64) {
			g.Go(func() error {
				defer close(ch)

				err := m.readValueFromFileAtOffset(ctx, ch, reader, m.msrOffsets[off])
				if err != nil {
					return fmt.Errorf("error reading MSR file, err: %v", err)
				}

				return nil
			})
		}(offset, channel)
	}

	newC3 := <-msrOffsetsWithChannels[c3StateResidencyLocation]
	newC6 := <-msrOffsetsWithChannels[c6StateResidencyLocation]
	newC7 := <-msrOffsetsWithChannels[c7StateResidencyLocation]
	newMperf := <-msrOffsetsWithChannels[maximumFrequencyClockCountLocation]
	newAperf := <-msrOffsetsWithChannels[actualFrequencyClockCountLocation]
	newTsc := <-msrOffsetsWithChannels[TimestampCounterLocation]
	newThrottleTemp := <-msrOffsetsWithChannels[throttleTemperatureLocation]
	newTemp := <-msrOffsetsWithChannels[temperatureLocation]

	if err := g.Wait(); err != nil {
		return fmt.Errorf("received error during reading MSR values in goroutines: %v", err)
	}

	m.cpuCoresData[core].C3Delta = newC3 - m.cpuCoresData[core].C3
	m.cpuCoresData[core].C6Delta = newC6 - m.cpuCoresData[core].C6
	m.cpuCoresData[core].C7Delta = newC7 - m.cpuCoresData[core].C7
	m.cpuCoresData[core].MperfDelta = newMperf - m.cpuCoresData[core].Mperf
	m.cpuCoresData[core].AperfDelta = newAperf - m.cpuCoresData[core].Aperf
	m.cpuCoresData[core].TscDelta = newTsc - m.cpuCoresData[core].Tsc

	m.cpuCoresData[core].C3 = newC3
	m.cpuCoresData[core].C6 = newC6
	m.cpuCoresData[core].C7 = newC7
	m.cpuCoresData[core].Mperf = newMperf
	m.cpuCoresData[core].Aperf = newAperf
	m.cpuCoresData[core].Tsc = newTsc
	// MSR (1A2h) IA32_TEMPERATURE_TARGET bits 23:16.
	m.cpuCoresData[core].ThrottleTemp = (newThrottleTemp >> 16) & 0xFF
	// MSR (19Ch) IA32_THERM_STATUS bits 22:16.
	m.cpuCoresData[core].Temp = (newTemp >> 16) & 0x7F

	return nil
}

func (m *MsrServiceImpl) readValueFromFileAtOffset(ctx context.Context, ch chan uint64, reader io.ReaderAt, offset int64) error {
	value, err := m.fs.ReadFileAtOffsetToUint64(reader, offset)
	if err != nil {
		return err
	}

	// Detect context cancellation and return an error if other goroutine fails
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- value:
	}

	return nil
}

// setCPUCores initialize cpuCoresData map.
func (m *MsrServiceImpl) setCPUCores() error {
	cpuPrefix := "cpu"
	cpuCore := fmt.Sprintf("%s%s", cpuPrefix, "[0-9]*")
	m.cpuCoresData = make(map[string]*data.MsrData)
	cpuPaths, err := m.fs.GetStringsMatchingPatternOnPath(fmt.Sprintf("%s/%s", systemCPUPath, cpuCore))
	if err != nil {
		return err
	}

	for _, cpuPath := range cpuPaths {
		core := strings.TrimPrefix(filepath.Base(cpuPath), cpuPrefix)
		m.cpuCoresData[core] = &data.MsrData{
			Mperf:        0,
			Aperf:        0,
			Tsc:          0,
			C3:           0,
			C6:           0,
			C7:           0,
			ThrottleTemp: 0,
			Temp:         0,
			MperfDelta:   0,
			AperfDelta:   0,
			TscDelta:     0,
			C3Delta:      0,
			C6Delta:      0,
			C7Delta:      0,
		}
	}

	return nil
}

func (m *MsrServiceImpl) getMsrOffsets() map[string]int64 {
	return m.msrOffsets
}

func (m *MsrServiceImpl) calculateMsrOffsets() {
	m.msrOffsets = map[string]int64{
		c3StateResidencyLocation:           m.parseStringHexToInt(c3StateResidencyLocation),
		c6StateResidencyLocation:           m.parseStringHexToInt(c6StateResidencyLocation),
		c7StateResidencyLocation:           m.parseStringHexToInt(c7StateResidencyLocation),
		maximumFrequencyClockCountLocation: m.parseStringHexToInt(maximumFrequencyClockCountLocation),
		actualFrequencyClockCountLocation:  m.parseStringHexToInt(actualFrequencyClockCountLocation),
		TimestampCounterLocation:           m.parseStringHexToInt(TimestampCounterLocation),
		throttleTemperatureLocation:        m.parseStringHexToInt(throttleTemperatureLocation),
		temperatureLocation:                m.parseStringHexToInt(temperatureLocation),
	}
}

func (m *MsrServiceImpl) parseStringHexToInt(s string) int64 {
	parsedInt, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		m.log.Errorf("error on parsing offset %s, err: %v", s, err)
		return 0
	}

	return parsedInt
}

// NewMsrServiceWithFs returns new RaplServiceImpl struct with given FileService.
func NewMsrServiceWithFs(logger telegraf.Logger, fs FileService) *MsrServiceImpl {
	msrService := &MsrServiceImpl{
		fs:  fs,
		log: logger,
	}
	err := msrService.setCPUCores()
	if err != nil {
		// This error does not prevent plugin from working thus it is not returned.
		msrService.log.Error(err)
	}
	msrService.calculateMsrOffsets()

	return msrService
}
