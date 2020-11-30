// +build linux

package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/data"
)

// Paths for Intel RAPL.
const (
	IntelRaplPath                = "/sys/devices/virtual/powercap/intel-rapl"
	IntelRaplSocketPartialPath   = "%s/intel-rapl:%s"
	EnergyUjPartialPath          = "%s/energy_uj"
	MaxRangeUjPartialPath        = "%s/max_energy_range_uj"
	MaxPowerUwPartialPath        = "%s/constraint_0_max_power_uw"
	IntelRaplDramPartialPath     = "%s/intel-rapl:%s/%s"
	IntelRaplDramNamePartialPath = "%s/name"
)

// RaplService is responsible for interactions with RAPL.
type RaplService interface {
	InitializeRaplData()
	GetConstraintMaxPower(socketID string) (float64, error)
	RetrieveAndCalculateData(socketID string) error
	GetSocketIDs() []string
	GetCurrentPackagePowerConsumption(socketID string) float64
	GetCurrentDramPowerConsumption(socketID string) float64
	GetDramMaxEnergyRangeUj(socketID string) (float64, error)
	GetMaxEnergyRangeUj(socketID string) (float64, error)
}

// RaplServiceImpl is implementation of RaplService.
type RaplServiceImpl struct {
	log          telegraf.Logger
	isFirstError bool
	data         map[string]*data.RaplData
	dramFolders  map[string]string
	fs           FileService
}

// InitializeRaplData looks for RAPL folders and initializes data map with fetched information.
func (r *RaplServiceImpl) InitializeRaplData() {
	r.prepareData()
	r.findDramFolders()
}

// GetSocketIDs returns array of socketIDs.
func (r *RaplServiceImpl) GetSocketIDs() []string {
	sockets := make([]string, 0)

	for socketID := range r.data {
		sockets = append(sockets, socketID)
	}

	return sockets
}

// GetCurrentPackagePowerConsumption returns package power consumption for socket.
func (r *RaplServiceImpl) GetCurrentPackagePowerConsumption(socketID string) float64 {
	return r.data[socketID].SocketCurrentEnergy
}

// GetCurrentDramPowerConsumption returns dram package power consumption for socket.
func (r *RaplServiceImpl) GetCurrentDramPowerConsumption(socketID string) float64 {
	return r.data[socketID].DramCurrentEnergy
}

// RetrieveAndCalculateData reads data from RAPL and calculates energy consumption.
func (r *RaplServiceImpl) RetrieveAndCalculateData(socketID string) error {
	raplPath := fmt.Sprintf(IntelRaplSocketPartialPath, IntelRaplPath, socketID)
	energyUjPath := fmt.Sprintf(EnergyUjPartialPath, raplPath)
	energyUjFile, err := os.Open(energyUjPath)
	if err != nil {
		return fmt.Errorf("error opening energy_uj file on path %s, err: %v", energyUjPath, err)
	}
	defer energyUjFile.Close()

	dramPath := fmt.Sprintf(IntelRaplDramPartialPath, IntelRaplPath, socketID, r.dramFolders[socketID])
	dramEnergyPath := fmt.Sprintf(EnergyUjPartialPath, dramPath)
	dramEnergyFile, err := os.Open(dramEnergyPath)
	if err != nil {
		return fmt.Errorf("error opening dram energy_uj file on path %s, err: %v", dramEnergyPath, err)
	}
	defer dramEnergyFile.Close()

	return r.calculateData(socketID, energyUjFile, dramEnergyFile)
}

// GetConstraintMaxPower retrieves ConstraintMaxPower from constraint_0_max_power_uw file.
func (r *RaplServiceImpl) GetConstraintMaxPower(socketID string) (float64, error) {
	raplPath := fmt.Sprintf(IntelRaplSocketPartialPath, IntelRaplPath, socketID)
	maxPowerPath := fmt.Sprintf(MaxPowerUwPartialPath, raplPath)
	maxPowerFile, err := os.Open(maxPowerPath)
	if err != nil {
		return 0, fmt.Errorf("error opening constraint_0_max_power_uw file on path %s, err: %v", maxPowerPath, err)
	}
	defer maxPowerFile.Close()
	maxPower, _, err := r.fs.ReadFileToFloat64(maxPowerFile)

	return maxPower, err
}

// GetDramMaxEnergyRangeUj retrieves dram energy range value from max_energy_range_uj file.
func (r *RaplServiceImpl) GetDramMaxEnergyRangeUj(socketID string) (float64, error) {
	dramPath := fmt.Sprintf(IntelRaplDramPartialPath, IntelRaplPath, socketID, r.dramFolders[socketID])
	dramEnergyPath := fmt.Sprintf(MaxRangeUjPartialPath, dramPath)
	maxDramEnergyFile, err := os.Open(dramEnergyPath)
	if err != nil {
		return 0, fmt.Errorf("error opening max_energy_range_uj file on path %s, err: %v", dramEnergyPath, err)
	}
	defer maxDramEnergyFile.Close()
	maxDramEnergy, _, err := r.fs.ReadFileToFloat64(maxDramEnergyFile)

	return maxDramEnergy, err
}

// GetMaxEnergyRangeUj retrieves energy range value from max_energy_range_uj file.
func (r *RaplServiceImpl) GetMaxEnergyRangeUj(socketID string) (float64, error) {
	raplPath := fmt.Sprintf(IntelRaplSocketPartialPath, IntelRaplPath, socketID)
	maxEnergyPath := fmt.Sprintf(MaxRangeUjPartialPath, raplPath)
	maxEnergyFile, err := os.Open(maxEnergyPath)
	if err != nil {
		return 0, fmt.Errorf("error opening max_energy_range_uj file on path %s, err: %v", maxEnergyPath, err)
	}
	defer maxEnergyFile.Close()
	maxEnergy, _, err := r.fs.ReadFileToFloat64(maxEnergyFile)

	return maxEnergy, err
}

// getDramFolders returns Dram folder names from RAPL.
func (r *RaplServiceImpl) getDramFolders() map[string]string {
	return r.dramFolders
}

func (r *RaplServiceImpl) prepareData() {
	intelRaplPrefix := "intel-rapl:"
	intelRapl := fmt.Sprintf("%s%s", intelRaplPrefix, "[0-9]*")
	raplPaths, err := r.fs.GetStringsMatchingPatternOnPath(fmt.Sprintf("%s/%s", IntelRaplPath, intelRapl))
	if err != nil {
		if r.isFirstError {
			r.isFirstError = false
			r.log.Errorf("error while preparing RAPL data: %v", err)
		} else {
			r.log.Debugf("error while preparing RAPL data: %v", err)
		}

		r.data = make(map[string]*data.RaplData)
		return
	}

	r.isFirstError = true
	// If RAPL exists initialize data map (if it wasn't initialized before).
	if len(r.data) == 0 {
		for _, raplPath := range raplPaths {
			socketID := strings.TrimPrefix(filepath.Base(raplPath), intelRaplPrefix)
			r.data[socketID] = &data.RaplData{
				SocketCurrentEnergy: 0,
				DramCurrentEnergy:   0,
				SocketEnergy:        0,
				DramEnergy:          0,
				ReadDate:            0,
			}
		}
	}
}

func (r *RaplServiceImpl) findDramFolders() {
	intelRaplPrefix := "intel-rapl:"
	intelRaplDram := fmt.Sprintf("%s%s", intelRaplPrefix, "[0-9]*[0-9]*")
	// Clean existing map
	r.dramFolders = make(map[string]string)

	for socketID := range r.data {
		path := fmt.Sprintf(IntelRaplSocketPartialPath, IntelRaplPath, socketID)
		pathsToRaplFolders, err := r.fs.GetStringsMatchingPatternOnPath(fmt.Sprintf("%s/%s", path, intelRaplDram))
		if err != nil {
			r.log.Errorf("error during lookup for rapl dram: %v", err)
			continue
		}
		raplFolders := make([]string, 0)
		for _, folderPath := range pathsToRaplFolders {
			raplFolders = append(raplFolders, filepath.Base(folderPath))
		}

		r.findDramFolder(raplFolders, socketID)
	}
}

func (r *RaplServiceImpl) calculateData(socketID string, energyFile io.Reader, dramEnergyFile io.Reader) error {
	newSocketEnergy, _, err := r.readEnergyInJoules(energyFile)
	if err != nil {
		return err
	}

	newDramEnergy, readDate, err := r.readEnergyInJoules(dramEnergyFile)
	if err != nil {
		return err
	}

	interval := ConvertNanoSecondsToSeconds(readDate - r.data[socketID].ReadDate)
	r.data[socketID].ReadDate = readDate

	if newSocketEnergy > r.data[socketID].SocketEnergy {
		r.data[socketID].SocketCurrentEnergy = (newSocketEnergy - r.data[socketID].SocketEnergy) / interval
	} else {
		maxEnergy, err := r.GetMaxEnergyRangeUj(socketID)
		if err != nil {
			return err
		}
		// When energy_uj counter reaches maximum value defined in max_energy_range_uj file it
		// starts counting from 0.
		r.data[socketID].SocketCurrentEnergy = maxEnergy - r.data[socketID].SocketEnergy + newSocketEnergy
	}

	if newDramEnergy > r.data[socketID].DramEnergy {
		r.data[socketID].DramCurrentEnergy = (newDramEnergy - r.data[socketID].DramEnergy) / interval
	} else {
		dramMaxEnergy, err := r.GetDramMaxEnergyRangeUj(socketID)
		if err != nil {
			return err
		}
		// When dram energy_uj reaches maximum value defined in max_energy_range_uj file it
		// starts counting from 0.
		r.data[socketID].DramCurrentEnergy = dramMaxEnergy - r.data[socketID].DramEnergy + newDramEnergy
	}
	r.data[socketID].SocketEnergy = newSocketEnergy
	r.data[socketID].DramEnergy = newDramEnergy

	return nil
}

func (r *RaplServiceImpl) findDramFolder(raplFolders []string, socketID string) {
	for _, raplFolder := range raplFolders {
		potentialDramPath := fmt.Sprintf(IntelRaplDramPartialPath, IntelRaplPath, socketID, raplFolder)
		nameFilePath := fmt.Sprintf(IntelRaplDramNamePartialPath, potentialDramPath)
		read, err := r.fs.ReadFile(nameFilePath)
		if err != nil {
			r.log.Errorf("error reading file on path: %s, err: %v", nameFilePath, err)
			continue
		}

		// Remove new line character
		trimmedString := strings.TrimRight(string(read), "\n")
		if trimmedString == "dram" {
			// There should be only one DRAM folder per socket
			r.dramFolders[socketID] = raplFolder
			return
		}
	}
}

func (r *RaplServiceImpl) readEnergyInJoules(reader io.Reader) (float64, int64, error) {
	currentEnergy, readDate, err := r.fs.ReadFileToFloat64(reader)
	return ConvertMicroJoulesToJoules(currentEnergy), readDate, err
}

// NewRaplServiceWithFs returns new RaplServiceImpl struct with given FileService.
func NewRaplServiceWithFs(logger telegraf.Logger, fs FileService) *RaplServiceImpl {
	return &RaplServiceImpl{
		log:          logger,
		isFirstError: true,
		data:         make(map[string]*data.RaplData),
		dramFolders:  make(map[string]string),
		fs:           fs,
	}
}
