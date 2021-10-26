//go:build linux
// +build linux

package intel_powerstat

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
)

const (
	intelRaplPath                = "/sys/devices/virtual/powercap/intel-rapl"
	intelRaplSocketPartialPath   = "%s/intel-rapl:%s"
	energyUjPartialPath          = "%s/energy_uj"
	maxEnergyRangeUjPartialPath  = "%s/max_energy_range_uj"
	maxPowerUwPartialPath        = "%s/constraint_0_max_power_uw"
	intelRaplDramPartialPath     = "%s/intel-rapl:%s/%s"
	intelRaplDramNamePartialPath = "%s/name"
)

// raplService is responsible for interactions with RAPL.
type raplService interface {
	initializeRaplData()
	getRaplData() map[string]*raplData
	retrieveAndCalculateData(socketID string) error
	getConstraintMaxPowerWatts(socketID string) (float64, error)
}

type raplServiceImpl struct {
	log         telegraf.Logger
	data        map[string]*raplData
	dramFolders map[string]string
	fs          fileService
}

// initializeRaplData looks for RAPL folders and initializes data map with fetched information.
func (r *raplServiceImpl) initializeRaplData() {
	r.prepareData()
	r.findDramFolders()
}

func (r *raplServiceImpl) getRaplData() map[string]*raplData {
	return r.data
}

func (r *raplServiceImpl) retrieveAndCalculateData(socketID string) error {
	socketRaplPath := fmt.Sprintf(intelRaplSocketPartialPath, intelRaplPath, socketID)
	socketEnergyUjPath := fmt.Sprintf(energyUjPartialPath, socketRaplPath)
	socketEnergyUjFile, err := os.Open(socketEnergyUjPath)
	if err != nil {
		return fmt.Errorf("error opening socket energy_uj file on path %s, err: %v", socketEnergyUjPath, err)
	}
	defer socketEnergyUjFile.Close()

	dramRaplPath := fmt.Sprintf(intelRaplDramPartialPath, intelRaplPath, socketID, r.dramFolders[socketID])
	dramEnergyUjPath := fmt.Sprintf(energyUjPartialPath, dramRaplPath)
	dramEnergyUjFile, err := os.Open(dramEnergyUjPath)
	if err != nil {
		return fmt.Errorf("error opening dram energy_uj file on path %s, err: %v", dramEnergyUjPath, err)
	}
	defer dramEnergyUjFile.Close()

	socketMaxEnergyUjPath := fmt.Sprintf(maxEnergyRangeUjPartialPath, socketRaplPath)
	socketMaxEnergyUjFile, err := os.Open(socketMaxEnergyUjPath)
	if err != nil {
		return fmt.Errorf("error opening socket max_energy_range_uj file on path %s, err: %v", socketMaxEnergyUjPath, err)
	}
	defer socketMaxEnergyUjFile.Close()

	dramMaxEnergyUjPath := fmt.Sprintf(maxEnergyRangeUjPartialPath, dramRaplPath)
	dramMaxEnergyUjFile, err := os.Open(dramMaxEnergyUjPath)
	if err != nil {
		return fmt.Errorf("error opening dram max_energy_range_uj file on path %s, err: %v", dramMaxEnergyUjPath, err)
	}
	defer dramMaxEnergyUjFile.Close()

	return r.calculateData(socketID, socketEnergyUjFile, dramEnergyUjFile, socketMaxEnergyUjFile, dramMaxEnergyUjFile)
}

func (r *raplServiceImpl) getConstraintMaxPowerWatts(socketID string) (float64, error) {
	socketRaplPath := fmt.Sprintf(intelRaplSocketPartialPath, intelRaplPath, socketID)
	socketMaxPowerPath := fmt.Sprintf(maxPowerUwPartialPath, socketRaplPath)
	socketMaxPowerFile, err := os.Open(socketMaxPowerPath)
	if err != nil {
		return 0, fmt.Errorf("error opening constraint_0_max_power_uw file on path %s, err: %v", socketMaxPowerPath, err)
	}
	defer socketMaxPowerFile.Close()

	socketMaxPower, _, err := r.fs.readFileToFloat64(socketMaxPowerFile)
	return convertMicroWattToWatt(socketMaxPower), err
}

func (r *raplServiceImpl) prepareData() {
	intelRaplPrefix := "intel-rapl:"
	intelRapl := fmt.Sprintf("%s%s", intelRaplPrefix, "[0-9]*")
	raplPattern := fmt.Sprintf("%s/%s", intelRaplPath, intelRapl)

	raplPaths, err := r.fs.getStringsMatchingPatternOnPath(raplPattern)
	if err != nil {
		r.log.Errorf("error while preparing RAPL data: %v", err)
		r.data = make(map[string]*raplData)
		return
	}
	if len(raplPaths) == 0 {
		r.log.Debugf("RAPL data wasn't found using pattern: %s", raplPattern)
		r.data = make(map[string]*raplData)
		return
	}

	// If RAPL exists initialize data map (if it wasn't initialized before).
	if len(r.data) == 0 {
		for _, raplPath := range raplPaths {
			socketID := strings.TrimPrefix(filepath.Base(raplPath), intelRaplPrefix)
			r.data[socketID] = &raplData{
				socketCurrentEnergy: 0,
				dramCurrentEnergy:   0,
				socketEnergy:        0,
				dramEnergy:          0,
				readDate:            0,
			}
		}
	}
}

func (r *raplServiceImpl) findDramFolders() {
	intelRaplPrefix := "intel-rapl:"
	intelRaplDram := fmt.Sprintf("%s%s", intelRaplPrefix, "[0-9]*[0-9]*")
	// Clean existing map
	r.dramFolders = make(map[string]string)

	for socketID := range r.data {
		path := fmt.Sprintf(intelRaplSocketPartialPath, intelRaplPath, socketID)
		raplFoldersPattern := fmt.Sprintf("%s/%s", path, intelRaplDram)
		pathsToRaplFolders, err := r.fs.getStringsMatchingPatternOnPath(raplFoldersPattern)
		if err != nil {
			r.log.Errorf("error during lookup for rapl dram: %v", err)
			continue
		}
		if len(pathsToRaplFolders) == 0 {
			r.log.Debugf("RAPL folders weren't found using pattern: %s", raplFoldersPattern)
			continue
		}

		raplFolders := make([]string, 0)
		for _, folderPath := range pathsToRaplFolders {
			raplFolders = append(raplFolders, filepath.Base(folderPath))
		}

		r.findDramFolder(raplFolders, socketID)
	}
}

func (r *raplServiceImpl) findDramFolder(raplFolders []string, socketID string) {
	for _, raplFolder := range raplFolders {
		potentialDramPath := fmt.Sprintf(intelRaplDramPartialPath, intelRaplPath, socketID, raplFolder)
		nameFilePath := fmt.Sprintf(intelRaplDramNamePartialPath, potentialDramPath)
		read, err := r.fs.readFile(nameFilePath)
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

func (r *raplServiceImpl) calculateData(socketID string, socketEnergyUjFile io.Reader, dramEnergyUjFile io.Reader,
	socketMaxEnergyUjFile io.Reader, dramMaxEnergyUjFile io.Reader,
) error {
	newSocketEnergy, _, err := r.readEnergyInJoules(socketEnergyUjFile)
	if err != nil {
		return err
	}

	newDramEnergy, readDate, err := r.readEnergyInJoules(dramEnergyUjFile)
	if err != nil {
		return err
	}

	interval := convertNanoSecondsToSeconds(readDate - r.data[socketID].readDate)
	r.data[socketID].readDate = readDate
	if interval == 0 {
		return fmt.Errorf("interval between last two Telegraf cycles is 0")
	}

	if newSocketEnergy > r.data[socketID].socketEnergy {
		r.data[socketID].socketCurrentEnergy = (newSocketEnergy - r.data[socketID].socketEnergy) / interval
	} else {
		socketMaxEnergy, _, err := r.readEnergyInJoules(socketMaxEnergyUjFile)
		if err != nil {
			return err
		}
		// When socket energy_uj counter reaches maximum value defined in max_energy_range_uj file it
		// starts counting from 0.
		r.data[socketID].socketCurrentEnergy = (socketMaxEnergy - r.data[socketID].socketEnergy + newSocketEnergy) / interval
	}

	if newDramEnergy > r.data[socketID].dramEnergy {
		r.data[socketID].dramCurrentEnergy = (newDramEnergy - r.data[socketID].dramEnergy) / interval
	} else {
		dramMaxEnergy, _, err := r.readEnergyInJoules(dramMaxEnergyUjFile)
		if err != nil {
			return err
		}
		// When dram energy_uj counter reaches maximum value defined in max_energy_range_uj file it
		// starts counting from 0.
		r.data[socketID].dramCurrentEnergy = (dramMaxEnergy - r.data[socketID].dramEnergy + newDramEnergy) / interval
	}
	r.data[socketID].socketEnergy = newSocketEnergy
	r.data[socketID].dramEnergy = newDramEnergy

	return nil
}

func (r *raplServiceImpl) readEnergyInJoules(reader io.Reader) (float64, int64, error) {
	currentEnergy, readDate, err := r.fs.readFileToFloat64(reader)
	return convertMicroJoulesToJoules(currentEnergy), readDate, err
}

func newRaplServiceWithFs(logger telegraf.Logger, fs fileService) *raplServiceImpl {
	return &raplServiceImpl{
		log:         logger,
		data:        make(map[string]*raplData),
		dramFolders: make(map[string]string),
		fs:          fs,
	}
}
