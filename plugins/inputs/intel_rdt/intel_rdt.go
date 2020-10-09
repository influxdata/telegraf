// +build !windows

package intel_rdt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	timestampFormat           = "2006-01-02 15:04:05"
	defaultSamplingInterval   = 10
	pqosInitOutputLinesNumber = 4
	numberOfMetrics           = 6
	secondsDenominator        = 10
)

var pqosMetricOrder = map[int]string{
	0: "IPC",        // Instructions Per Cycle
	1: "LLC_Misses", // Cache Misses
	2: "LLC",        // L3 Cache Occupancy
	3: "MBL",        // Memory Bandwidth on Local NUMA Node
	4: "MBR",        // Memory Bandwidth on Remote NUMA Node
	5: "MBT",        // Total Memory Bandwidth
}

type IntelRDT struct {
	PqosPath         string   `toml:"pqos_path"`
	Cores            []string `toml:"cores"`
	Processes        []string `toml:"processes"`
	SamplingInterval int32    `toml:"sampling_interval"`
	ShortenedMetrics bool     `toml:"shortened_metrics"`

	Log              telegraf.Logger  `toml:"-"`
	Publisher        Publisher        `toml:"-"`
	Processor        ProcessesHandler `toml:"-"`
	stopPQOSChan     chan bool
	quitChan         chan struct{}
	errorChan        chan error
	parsedCores      []string
	processesPIDsMap map[string]string
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

type processMeasurement struct {
	name        string
	measurement string
}

// All gathering is done in the Start function
func (r *IntelRDT) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (r *IntelRDT) Description() string {
	return "Intel Resource Director Technology plugin"
}

func (r *IntelRDT) SampleConfig() string {
	return `
	## Optionally set sampling interval to Nx100ms. 
	## This value is propagated to pqos tool. Interval format is defined by pqos itself.
	## If not provided or provided 0, will be set to 10 = 10x100ms = 1s.
	# sampling_interval = "10"
	
	## Optionally specify the path to pqos executable. 
	## If not provided, auto discovery will be performed.
	# pqos_path = "/usr/local/bin/pqos"

	## Optionally specify if IPC and LLC_Misses metrics shouldn't be propagated.
	## If not provided, default value is false.
	# shortened_metrics = false
	
	## Specify the list of groups of CPU core(s) to be provided as pqos input. 
	## Mandatory if processes aren't set and forbidden if processes are specified.
	## e.g. ["0-3", "4,5,6"] or ["1-3,4"]
	# cores = ["0-3"]
	
	## Specify the list of processes for which Metrics will be collected.
	## Mandatory if cores aren't set and forbidden if cores are specified.
	## e.g. ["qemu", "pmd"]
	# processes = ["process"]
`
}

func (r *IntelRDT) Start(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	r.Processor = NewProcessor()
	r.Publisher = NewPublisher(acc, r.Log, r.ShortenedMetrics)

	err := r.Initialize()
	if err != nil {
		return err
	}

	r.Publisher.publish(ctx)
	go r.errorHandler(ctx)
	go r.scheduler(ctx)

	return nil
}

func (r *IntelRDT) Initialize() error {
	r.stopPQOSChan = make(chan bool)
	r.quitChan = make(chan struct{})
	r.errorChan = make(chan error)

	err := validatePqosPath(r.PqosPath)
	if err != nil {
		return err
	}
	if len(r.Cores) != 0 && len(r.Processes) != 0 {
		return fmt.Errorf("monitoring start error, process and core tracking can not be done simultaneously")
	}
	if len(r.Cores) == 0 && len(r.Processes) == 0 {
		return fmt.Errorf("monitoring start error, at least one of cores or processes must be provided in config")
	}
	if r.SamplingInterval == 0 {
		r.SamplingInterval = defaultSamplingInterval
	}
	if err = validateInterval(r.SamplingInterval); err != nil {
		return err
	}
	r.parsedCores, err = parseCoresConfig(r.Cores)
	if err != nil {
		return err
	}
	r.processesPIDsMap, err = r.associateProcessesWithPIDs(r.Processes)
	if err != nil {
		return err
	}
	return nil
}

func (r *IntelRDT) errorHandler(ctx context.Context) {
	r.wg.Add(1)
	defer r.wg.Done()
	for {
		select {
		case err := <-r.errorChan:
			if err != nil {
				r.Log.Error(fmt.Sprintf("Error: %v", err))
				r.quitChan <- struct{}{}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *IntelRDT) scheduler(ctx context.Context) {
	r.wg.Add(1)
	defer r.wg.Done()
	interval := time.Duration(r.SamplingInterval)
	ticker := time.NewTicker(interval * time.Second / secondsDenominator)

	r.createArgsAndStartPQOS(ctx)

	for {
		select {
		case <-ticker.C:
			if len(r.Processes) != 0 {
				err := r.checkPIDsAssociation(ctx)
				if err != nil {
					r.errorChan <- err
				}
			}
		case <-r.quitChan:
			r.cancel()
			return
		case <-ctx.Done():
			return
		}
	}
}

func (r *IntelRDT) Stop() {
	r.cancel()
	r.wg.Wait()
}

func (r *IntelRDT) checkPIDsAssociation(ctx context.Context) error {
	newProcessesPIDsMap, err := r.associateProcessesWithPIDs(r.Processes)
	if err != nil {
		return err
	}
	// change in PIDs association appears
	if !cmp.Equal(newProcessesPIDsMap, r.processesPIDsMap) {
		r.Log.Warnf("PIDs association has changed. Refreshing...")
		if len(r.processesPIDsMap) != 0 {
			r.stopPQOSChan <- true
		}
		r.processesPIDsMap = newProcessesPIDsMap
		r.createArgsAndStartPQOS(ctx)
	}
	return nil
}

func (r *IntelRDT) associateProcessesWithPIDs(providedProcesses []string) (map[string]string, error) {
	mapProcessPIDs := map[string]string{}

	availableProcesses, err := r.Processor.getAllProcesses()
	if err != nil {
		return nil, fmt.Errorf("cannot gather information of all available processes")
	}
	for _, availableProcess := range availableProcesses {
		if choice.Contains(availableProcess.Name, providedProcesses) {
			PID := availableProcess.PID
			mapProcessPIDs[availableProcess.Name] = mapProcessPIDs[availableProcess.Name] + fmt.Sprintf("%d", PID) + ","
		}
	}
	for key := range mapProcessPIDs {
		mapProcessPIDs[key] = strings.TrimSuffix(mapProcessPIDs[key], ",")
	}
	return mapProcessPIDs, nil
}

func (r *IntelRDT) createArgsAndStartPQOS(ctx context.Context) {
	args := []string{"-r", "--iface-os", "--mon-file-type=csv", fmt.Sprintf("--mon-interval=%d", r.SamplingInterval)}

	if len(r.parsedCores) != 0 {
		coresArg := createArgCores(r.parsedCores)
		args = append(args, coresArg)
		go r.readData(args, nil, ctx)

	} else if len(r.processesPIDsMap) != 0 {
		processArg := createArgProcess(r.processesPIDsMap)
		args = append(args, processArg)
		go r.readData(args, r.processesPIDsMap, ctx)
	}
	return
}

func (r *IntelRDT) readData(args []string, processesPIDsAssociation map[string]string, ctx context.Context) {
	r.wg.Add(1)
	defer r.wg.Done()

	cmd := exec.Command(r.PqosPath, append(args)...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		r.errorChan <- err
	}
	go r.processOutput(cmdReader, processesPIDsAssociation)

	go func() {
		for {
			select {
			case <-r.stopPQOSChan:
				if err := shutDownPqos(cmd); err != nil {
					r.Log.Error(err)
				}
				return
			case <-ctx.Done():
				if err := shutDownPqos(cmd); err != nil {
					r.Log.Error(err)
				}
				return
			}
		}
	}()
	err = cmd.Start()
	if err != nil {
		r.errorChan <- fmt.Errorf("pqos: %v", err)
		return
	}
	err = cmd.Wait()
	if err != nil {
		r.errorChan <- fmt.Errorf("pqos: %v", err)
	}
}

func (r *IntelRDT) processOutput(cmdReader io.ReadCloser, processesPIDsAssociation map[string]string) {
	reader := bufio.NewScanner(cmdReader)
	/*
		Omit constant, first 4 lines :
		"NOTE:  Mixed use of MSR and kernel interfaces to manage
				CAT or CMT & MBM may lead to unexpected behavior.\n"
		CMT/MBM reset successful
		"Time,Core,IPC,LLC Misses,LLC[KB],MBL[MB/s],MBR[MB/s],MBT[MB/s]\n"
	*/
	toOmit := pqosInitOutputLinesNumber

	// omit first measurements which are zeroes
	if len(r.parsedCores) != 0 {
		toOmit = toOmit + len(r.parsedCores)
		// specify how many lines should pass before stopping
	} else if len(processesPIDsAssociation) != 0 {
		toOmit = toOmit + len(processesPIDsAssociation)
	}
	for omitCounter := 0; omitCounter < toOmit; omitCounter++ {
		reader.Scan()
	}
	for reader.Scan() {
		out := reader.Text()
		// to handle situation when monitored PID disappear and "err" is shown in output
		if strings.Contains(out, "err") {
			continue
		}
		if len(r.Processes) != 0 {
			newMetric := processMeasurement{}

			PIDs, err := findPIDsInMeasurement(out)
			if err != nil {
				r.errorChan <- err
				break
			}
			for processName, PIDsProcess := range processesPIDsAssociation {
				if PIDs == PIDsProcess {
					newMetric.name = processName
					newMetric.measurement = out
				}
			}
			r.Publisher.BufferChanProcess <- newMetric
		} else {
			r.Publisher.BufferChanCores <- out
		}
	}
}

func shutDownPqos(pqos *exec.Cmd) error {
	if pqos.Process != nil {
		err := pqos.Process.Signal(os.Interrupt)
		if err != nil {
			err = pqos.Process.Kill()
			if err != nil {
				return fmt.Errorf("failed to shut down pqos: %v", err)
			}
		}
	}
	return nil
}

func createArgCores(cores []string) string {
	allGroupsArg := "--mon-core="
	for _, coreGroup := range cores {
		argGroup := createArgsForGroups(strings.Split(coreGroup, ","))
		allGroupsArg = allGroupsArg + argGroup
	}
	return allGroupsArg
}

func createArgProcess(processPIDs map[string]string) string {
	allPIDsArg := "--mon-pid="
	for _, PIDs := range processPIDs {
		argPIDs := createArgsForGroups(strings.Split(PIDs, ","))
		allPIDsArg = allPIDsArg + argPIDs
	}
	return allPIDsArg
}

func createArgsForGroups(coresOrPIDs []string) string {
	template := "all:[%s];mbt:[%s];"
	group := ""

	for _, coreOrPID := range coresOrPIDs {
		group = group + coreOrPID + ","
	}
	if group != "" {
		group = strings.TrimSuffix(group, ",")
		return fmt.Sprintf(template, group, group)
	}
	return ""
}

func validatePqosPath(pqosPath string) error {
	if len(pqosPath) == 0 {
		return fmt.Errorf("monitoring start error, can not find pqos executable")
	}
	pathInfo, err := os.Stat(pqosPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("monitoring start error, provided pqos path not exist")
	}
	if mode := pathInfo.Mode(); !mode.IsRegular() {
		return fmt.Errorf("monitoring start error, provided pqos path does not point to a regular file")
	}
	return nil
}

func parseCoresConfig(cores []string) ([]string, error) {
	var parsedCores []string
	var allCores []int
	configError := fmt.Errorf("wrong cores input config data format")

	for _, singleCoreGroup := range cores {
		var actualGroupOfCores []int
		separatedCores := strings.Split(singleCoreGroup, ",")

		for _, coreStr := range separatedCores {
			actualCores, err := validateAndParseCores(coreStr)
			if err != nil {
				return nil, fmt.Errorf("%v: %v", configError, err)
			}
			if checkForDuplicates(allCores, actualCores) {
				return nil, fmt.Errorf("%v: %v", configError, "core value cannot be duplicated")
			}
			actualGroupOfCores = append(actualGroupOfCores, actualCores...)
			allCores = append(allCores, actualGroupOfCores...)
		}
		parsedCores = append(parsedCores, arrayToString(actualGroupOfCores))
	}
	return parsedCores, nil
}

func validateAndParseCores(coreStr string) ([]int, error) {
	var processedCores []int
	if strings.Contains(coreStr, "-") {
		rangeValues := strings.Split(coreStr, "-")

		if len(rangeValues) != 2 {
			return nil, fmt.Errorf("more than two values in range")
		}

		startValue, err := strconv.Atoi(rangeValues[0])
		if err != nil {
			return nil, err
		}
		stopValue, err := strconv.Atoi(rangeValues[1])
		if err != nil {
			return nil, err
		}

		if startValue > stopValue {
			return nil, fmt.Errorf("first value cannot be higher than second")
		}

		rangeOfCores := makeRange(startValue, stopValue)
		processedCores = append(processedCores, rangeOfCores...)
	} else {
		newCore, err := strconv.Atoi(coreStr)
		if err != nil {
			return nil, err
		}
		processedCores = append(processedCores, newCore)
	}
	return processedCores, nil
}

func findPIDsInMeasurement(measurements string) (string, error) {
	// to distinguish PIDs from Cores (PIDs should be in quotes)
	var insideQuoteRegex = regexp.MustCompile(`"(.*?)"`)
	PIDsMatch := insideQuoteRegex.FindStringSubmatch(measurements)
	if len(PIDsMatch) < 2 {
		return "", fmt.Errorf("cannot find PIDs in measurement line")
	}
	PIDs := PIDsMatch[1]
	return PIDs, nil
}

func splitCSVLineIntoValues(line string) (timeValue string, metricsValues, coreOrPIDsValues []string, err error) {
	values, err := splitMeasurementLine(line)
	if err != nil {
		return "", nil, nil, err
	}

	timeValue = values[0]
	// Because pqos csv format is broken when many cores are involved in PID or
	// group of PIDs, there is need to work around it. E.g.:
	// Time,PID,Core,IPC,LLC Misses,LLC[KB],MBL[MB/s],MBR[MB/s],MBT[MB/s]
	// 2020-08-12 13:34:36,"45417,29170,",37,44,0.00,0,0.0,0.0,0.0,0.0
	metricsValues = values[len(values)-numberOfMetrics:]
	coreOrPIDsValues = values[1 : len(values)-numberOfMetrics]

	return timeValue, metricsValues, coreOrPIDsValues, nil
}

func validateInterval(interval int32) error {
	if interval < 0 {
		return fmt.Errorf("interval cannot be lower than 0")
	}
	return nil
}

func splitMeasurementLine(line string) ([]string, error) {
	values := strings.Split(line, ",")
	if len(values) < 8 {
		return nil, fmt.Errorf(fmt.Sprintf("not valid line format from pqos: %s", values))
	}
	return values, nil
}

func parseTime(value string) (time.Time, error) {
	timestamp, err := time.Parse(timestampFormat, value)
	if err != nil {
		return time.Time{}, err
	}
	return timestamp, nil
}

func parseFloat(value string) (float64, error) {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return result, err
	}
	return result, nil
}

func arrayToString(array []int) string {
	result := ""
	for _, value := range array {
		result = fmt.Sprintf("%s%d,", result, value)
	}
	return strings.TrimSuffix(result, ",")
}

func checkForDuplicates(values []int, valuesToCheck []int) bool {
	for _, value := range values {
		for _, valueToCheck := range valuesToCheck {
			if value == valueToCheck {
				return true
			}
		}
	}
	return false
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func init() {
	inputs.Add("IntelRDT", func() telegraf.Input {
		rdt := IntelRDT{}
		pathPqos, _ := exec.LookPath("pqos")
		if len(pathPqos) > 0 {
			rdt.PqosPath = pathPqos
		}
		return &rdt
	})
}
