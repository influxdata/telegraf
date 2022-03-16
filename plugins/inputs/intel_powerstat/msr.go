//go:build linux
// +build linux

package intel_powerstat

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/influxdata/telegraf"
)

const (
	systemCPUPath                      = "/sys/devices/system/cpu/"
	cpuCurrentFreqPartialPath          = "/sys/devices/system/cpu/cpu%s/cpufreq/scaling_cur_freq"
	msrPartialPath                     = "/dev/cpu/%s/msr"
	c3StateResidencyLocation           = 0x3FC
	c6StateResidencyLocation           = 0x3FD
	c7StateResidencyLocation           = 0x3FE
	maximumFrequencyClockCountLocation = 0xE7
	actualFrequencyClockCountLocation  = 0xE8
	throttleTemperatureLocation        = 0x1A2
	temperatureLocation                = 0x19C
	timestampCounterLocation           = 0x10
)

// msrService is responsible for interactions with MSR.
type msrService interface {
	getCPUCoresData() map[string]*msrData
	retrieveCPUFrequencyForCore(core string) (float64, error)
	openAndReadMsr(core string) error
}

type msrServiceImpl struct {
	cpuCoresData map[string]*msrData
	msrOffsets   []int64
	fs           fileService
	log          telegraf.Logger
}

func (m *msrServiceImpl) getCPUCoresData() map[string]*msrData {
	return m.cpuCoresData
}

func (m *msrServiceImpl) retrieveCPUFrequencyForCore(core string) (float64, error) {
	cpuFreqPath := fmt.Sprintf(cpuCurrentFreqPartialPath, core)
	cpuFreqFile, err := os.Open(cpuFreqPath)
	if err != nil {
		return 0, fmt.Errorf("error opening scaling_cur_freq file on path %s, err: %v", cpuFreqPath, err)
	}
	defer cpuFreqFile.Close()

	cpuFreq, _, err := m.fs.readFileToFloat64(cpuFreqFile)
	return convertKiloHertzToMegaHertz(cpuFreq), err
}

func (m *msrServiceImpl) openAndReadMsr(core string) error {
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

func (m *msrServiceImpl) readDataFromMsr(core string, reader io.ReaderAt) error {
	g, ctx := errgroup.WithContext(context.Background())

	// Create and populate a map that contains msr offsets along with their respective channels
	msrOffsetsWithChannels := make(map[int64]chan uint64)
	for _, offset := range m.msrOffsets {
		msrOffsetsWithChannels[offset] = make(chan uint64)
	}

	// Start a goroutine for each msr offset
	for offset, channel := range msrOffsetsWithChannels {
		// Wrap around function to avoid race on loop counter
		func(off int64, ch chan uint64) {
			g.Go(func() error {
				defer close(ch)

				err := m.readValueFromFileAtOffset(ctx, ch, reader, off)
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
	newTsc := <-msrOffsetsWithChannels[timestampCounterLocation]
	newThrottleTemp := <-msrOffsetsWithChannels[throttleTemperatureLocation]
	newTemp := <-msrOffsetsWithChannels[temperatureLocation]

	if err := g.Wait(); err != nil {
		return fmt.Errorf("received error during reading MSR values in goroutines: %v", err)
	}

	m.cpuCoresData[core].c3Delta = newC3 - m.cpuCoresData[core].c3
	m.cpuCoresData[core].c6Delta = newC6 - m.cpuCoresData[core].c6
	m.cpuCoresData[core].c7Delta = newC7 - m.cpuCoresData[core].c7
	m.cpuCoresData[core].mperfDelta = newMperf - m.cpuCoresData[core].mperf
	m.cpuCoresData[core].aperfDelta = newAperf - m.cpuCoresData[core].aperf
	m.cpuCoresData[core].timeStampCounterDelta = newTsc - m.cpuCoresData[core].timeStampCounter

	m.cpuCoresData[core].c3 = newC3
	m.cpuCoresData[core].c6 = newC6
	m.cpuCoresData[core].c7 = newC7
	m.cpuCoresData[core].mperf = newMperf
	m.cpuCoresData[core].aperf = newAperf
	m.cpuCoresData[core].timeStampCounter = newTsc
	// MSR (1A2h) IA32_TEMPERATURE_TARGET bits 23:16.
	m.cpuCoresData[core].throttleTemp = (newThrottleTemp >> 16) & 0xFF
	// MSR (19Ch) IA32_THERM_STATUS bits 22:16.
	m.cpuCoresData[core].temp = (newTemp >> 16) & 0x7F

	return nil
}

func (m *msrServiceImpl) readValueFromFileAtOffset(ctx context.Context, ch chan uint64, reader io.ReaderAt, offset int64) error {
	value, err := m.fs.readFileAtOffsetToUint64(reader, offset)
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
func (m *msrServiceImpl) setCPUCores() error {
	m.cpuCoresData = make(map[string]*msrData)
	cpuPrefix := "cpu"
	cpuCore := fmt.Sprintf("%s%s", cpuPrefix, "[0-9]*")
	cpuCorePattern := fmt.Sprintf("%s/%s", systemCPUPath, cpuCore)
	cpuPaths, err := m.fs.getStringsMatchingPatternOnPath(cpuCorePattern)
	if err != nil {
		return err
	}
	if len(cpuPaths) == 0 {
		m.log.Debugf("CPU core data wasn't found using pattern: %s", cpuCorePattern)
		return nil
	}

	for _, cpuPath := range cpuPaths {
		core := strings.TrimPrefix(filepath.Base(cpuPath), cpuPrefix)
		m.cpuCoresData[core] = &msrData{
			mperf:                 0,
			aperf:                 0,
			timeStampCounter:      0,
			c3:                    0,
			c6:                    0,
			c7:                    0,
			throttleTemp:          0,
			temp:                  0,
			mperfDelta:            0,
			aperfDelta:            0,
			timeStampCounterDelta: 0,
			c3Delta:               0,
			c6Delta:               0,
			c7Delta:               0,
		}
	}

	return nil
}

func newMsrServiceWithFs(logger telegraf.Logger, fs fileService) *msrServiceImpl {
	msrService := &msrServiceImpl{
		fs:  fs,
		log: logger,
	}
	err := msrService.setCPUCores()
	if err != nil {
		// This error does not prevent plugin from working thus it is not returned.
		msrService.log.Error(err)
	}

	msrService.msrOffsets = []int64{c3StateResidencyLocation, c6StateResidencyLocation, c7StateResidencyLocation,
		maximumFrequencyClockCountLocation, actualFrequencyClockCountLocation, timestampCounterLocation,
		throttleTemperatureLocation, temperatureLocation}

	return msrService
}
