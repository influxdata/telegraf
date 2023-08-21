//go:generate ../../../tools/readme_config_includer/generator
//go:build linux && amd64

package intel_pmt

import (
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PaesslerAG/gval"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var hexToDecRegex = regexp.MustCompile(`0x[0-9a-fA-F]+`)

const (
	defaultPmtBasePath = "/sys/class/intel_pmt"
	pluginName         = "intel_pmt"
)

type pmtFileInfo []fileInfo

type fileInfo struct {
	path     string
	numaNode string
}

type IntelPMT struct {
	PmtSource       string          `toml:"pmt_source"`
	DatatypeMetrics []string        `toml:"datatype_metrics"`
	SampleMetrics   []string        `toml:"sample_metrics"`
	Log             telegraf.Logger `toml:"-"`

	pmtBasePath            string
	isFilePath             bool
	reader                 sourceReader
	pmtTelemetryFiles      map[string]pmtFileInfo
	pmtMetadata            *pmt
	pmtAggregator          map[string]aggregator
	pmtAggregatorInterface map[string]aggregatorInterface
	pmtTransformations     map[string]map[string]transformation
}

// SampleConfig returns a sample configuration (See sample.conf).
func (p *IntelPMT) SampleConfig() string {
	return sampleConfig
}

// Init performs one time setup of the plugin
func (p *IntelPMT) Init() error {
	err := p.definePmtSourceType()
	if err != nil {
		return err
	}

	return p.parseXMLs()
}

// Gather collects the plugin's metrics.
func (p *IntelPMT) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(p.pmtTelemetryFiles))
	for guid := range p.pmtTelemetryFiles {
		wg.Add(1)
		go func(guid string, fileInfo []fileInfo) {
			defer wg.Done()
			for _, info := range fileInfo {
				data, err := os.ReadFile(info.path)
				if err != nil {
					errorChan <- err
					return
				}

				err = p.aggregateSamples(guid, data, info.numaNode, acc)
				if err != nil {
					errorChan <- err
					return
				}
			}
		}(guid, p.pmtTelemetryFiles[guid])
	}
	wg.Wait()
	close(errorChan)

	var hasError bool
	for err := range errorChan {
		if err != nil {
			p.Log.Errorf("Error occurred while gathering metrics: %v", err)
			hasError = true
		}
	}
	if hasError {
		return errors.New("error(s) occurred while gathering metrics")
	}
	return nil
}

// definePmtSourceType defines if provided pmtSource is URL or filepath
//
// pmtSource is expected to be an absolute URL or absolute filepath.
// This function determines which one it is and creates a correct reader.
//
// Returns:
//
//	error - error if pmtSource is invalid or not absolute.
func (p *IntelPMT) definePmtSourceType() error {
	if p.PmtSource == "" {
		return fmt.Errorf("no source of XMLs provided")
	}

	parsedURL, err := url.Parse(p.PmtSource)
	// URL parse doesn't return an error if scheme is empty/invalid.
	// If scheme is empty or no Host in URL then check if provided source is a readable filePath.
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		if !isFileReadable(p.PmtSource) {
			return fmt.Errorf("provided pmt source is invalid %q", p.PmtSource)
		}
		p.isFilePath = true
	}

	lastSlash := strings.LastIndex(p.PmtSource, "/")
	// if pmtSource contains no "/"
	if lastSlash == -1 {
		return fmt.Errorf("provided pmt source is not an absolute path")
	}
	if p.isFilePath {
		p.pmtBasePath = p.PmtSource[:lastSlash]
		p.reader = fileReader{}
	} else {
		p.pmtBasePath = parsedURL.String()[:lastSlash]
		p.reader = httpReader{}
	}
	return nil
}

// parseXMLs reads and parses PMT XMLs.
//
// This method retrieves all metadata about known GUIDs from pmtSource.
// Then, it explores PMT sysfs to find all readable "telem" files and their GUIDs.
// It then matches found (readable) system GUIDs with GUIDs from metadata and
// reads corresponding sets of XMLs.
//
// Returns:
//
//	error - if PMT source is empty, if exploring PMT sysfs fails, or if reading XMLs fails.
func (p *IntelPMT) parseXMLs() error {
	err := parseXML(p.PmtSource, p.reader, &p.pmtMetadata)
	if err != nil {
		return err
	}
	if len(p.pmtMetadata.Mappings.Mapping) == 0 {
		return fmt.Errorf("pmt XML provided contains no mappings")
	}

	err = p.explorePmtInSysfs()
	if err != nil {
		return fmt.Errorf("error while exploring pmt sysfs: %w", err)
	}

	err = p.readXMLs()
	if err != nil {
		return err
	}

	p.pmtTransformations = make(map[string]map[string]transformation)
	for guid := range p.pmtTelemetryFiles {
		p.pmtTransformations[guid] = make(map[string]transformation)
		for _, transform := range p.pmtAggregatorInterface[guid].Transformations.Transformation {
			p.pmtTransformations[guid][transform.TransformID] = transform
		}
	}
	return nil
}

// explorePmtInSysfs finds necessary paths in pmt sysfs.
//
// This method finds "telem" files, used to retrieve telemetry values
// and saves them under their corresponding GUID.
// It also finds which NUMA node the samples belong to.
//
// Returns:
//
//	error - error if any of the operations failed.
func (p *IntelPMT) explorePmtInSysfs() error {
	pmtDirectories, err := os.ReadDir(defaultPmtBasePath)
	if err != nil {
		return fmt.Errorf("error reading pmt directory: %w", err)
	}
	p.pmtTelemetryFiles = make(map[string]pmtFileInfo)
	for _, dir := range pmtDirectories {
		if !strings.HasPrefix(dir.Name(), "telem") {
			continue
		}

		pmtGUIDPath := filepath.Join(defaultPmtBasePath, dir.Name(), "guid")
		if !isFileReadable(pmtGUIDPath) {
			p.Log.Warnf("GUID file is not readable %q", pmtGUIDPath)
			continue
		}

		rawGUID, err := os.ReadFile(pmtGUIDPath)
		if err != nil {
			return fmt.Errorf("cannot read GUID: %w", err)
		}
		// cut the newline char
		tID := strings.TrimRight(string(rawGUID), "\n")

		telemPath := filepath.Join(defaultPmtBasePath, dir.Name(), "telem")
		if !isFileReadable(telemPath) {
			p.Log.Warnf("telem file is not readable %q", telemPath)
			continue
		}

		numaNodePath := filepath.Join(defaultPmtBasePath, dir.Name(), "device", "numa_node")
		numaNodeSymlink, err := filepath.EvalSymlinks(numaNodePath)
		if err != nil {
			return fmt.Errorf("error while evaluating symlink %q: %w", numaNodePath, err)
		}

		numaNode, err := os.ReadFile(numaNodeSymlink)
		if err != nil {
			return fmt.Errorf("error while reading symlink %q: %w", numaNodeSymlink, err)
		}
		numaNodeString := strings.TrimRight(string(numaNode), "\n")
		if numaNodeString == "" {
			return fmt.Errorf("numa_node file %q is empty", numaNodeSymlink)
		}

		fi := fileInfo{
			path:     telemPath,
			numaNode: numaNodeString,
		}
		p.pmtTelemetryFiles[tID] = append(p.pmtTelemetryFiles[tID], fi)
	}
	if len(p.pmtTelemetryFiles) == 0 {
		return errors.New("no telemetry sources found - current platform doesn't support PMT or proper permissions needed to read them")
	}
	return nil
}

func isFileReadable(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	file, err := os.Open(path)
	if err != nil {
		return false
	}
	file.Close()

	return true
}

// readXMLs function reads all XMLs for found GUIDs.
//
// This method reads two required XMLs for each found GUID,
// checks if any of the provided filtering metrics were not found,
// and checks if there is at least one non-empty XML set.
//
// Returns:
//
//	error - error if reading operation failed or if all XMLs are empty.
func (p *IntelPMT) readXMLs() error {
	p.pmtAggregator = make(map[string]aggregator)
	p.pmtAggregatorInterface = make(map[string]aggregatorInterface)
	dtMetricsFound := make(map[string]bool)
	sampleMetricsFound := make(map[string]bool)
	for guid := range p.pmtTelemetryFiles {
		err := p.getAllXMLData(guid, dtMetricsFound, sampleMetricsFound)
		if err != nil {
			return fmt.Errorf("failed reading XMLs: %w", err)
		}
	}
	for _, dt := range p.DatatypeMetrics {
		if _, ok := dtMetricsFound[dt]; !ok {
			p.Log.Warnf("Configured datatype metric %q has not been found", dt)
		}
	}
	for _, sm := range p.SampleMetrics {
		if _, ok := sampleMetricsFound[sm]; !ok {
			p.Log.Warnf("Configured sample metric %q has not been found", sm)
		}
	}
	err := p.verifyNoEmpty()
	if err != nil {
		return fmt.Errorf("XMLs empty: %w", err)
	}
	return nil
}

// getAllXMLData retrieves two XMLs for given GUID.
//
// This method reads where to find the Aggregator and Aggregator interface XMLs
// from pmt metadata and reads found XMLs.
// This method also filters read XMLs before saving them
// and extracts additional tags from the data.
//
// Parameters:
//
//	guid - GUID saying which XMLs should be read.
//	dtMetricsFound - a map of found datatype metrics for all GUIDs.
//	smFound - a map of found sample names for all GUIDs.
//
// Returns:
//
//	error - if reading XML has failed.
func (p *IntelPMT) getAllXMLData(guid string, dtMetricsFound map[string]bool, smFound map[string]bool) error {
	for _, mapping := range p.pmtMetadata.Mappings.Mapping {
		if mapping.GUID == guid {
			basedir := mapping.XMLSet.Basedir
			guid := mapping.GUID
			var aggSource, aggInterfaceSource string

			if !p.isFilePath {
				pmtBaseURL, err := url.Parse(p.pmtBasePath)
				if err != nil {
					return err
				}
				aggSource = pmtBaseURL.JoinPath(basedir, mapping.XMLSet.Aggregator).String()
				aggInterfaceSource = pmtBaseURL.JoinPath(basedir, mapping.XMLSet.AggregatorInterface).String()
			} else {
				aggSource = filepath.Join(p.pmtBasePath, basedir, mapping.XMLSet.Aggregator)
				aggInterfaceSource = filepath.Join(p.pmtBasePath, basedir, mapping.XMLSet.AggregatorInterface)
			}

			tAgg := aggregator{}
			tAggInterface := aggregatorInterface{}

			err := parseXML(aggSource, p.reader, &tAgg)
			if err != nil {
				return fmt.Errorf("failed reading aggregator XML: %w", err)
			}
			err = parseXML(aggInterfaceSource, p.reader, &tAggInterface)
			if err != nil {
				return fmt.Errorf("failed reading aggregator interface XML: %w", err)
			}
			if len(p.DatatypeMetrics) > 0 {
				tAgg.filterAggregatorByDatatype(p.DatatypeMetrics)
				tAggInterface.filterAggInterfaceByDatatype(p.DatatypeMetrics, dtMetricsFound)
			}
			if len(p.SampleMetrics) > 0 {
				tAgg.filterAggregatorBySampleName(p.SampleMetrics)
				tAggInterface.filterAggInterfaceBySampleName(p.SampleMetrics, smFound)
			}
			p.pmtAggregator[guid] = tAgg
			tAggInterface.extractTagsFromSample()
			p.pmtAggregatorInterface[guid] = tAggInterface
		}
	}
	return nil
}

// getSampleValues reads all sample values for all sample groups.
//
// This method reads all telemetry samples for given GUID from given data
// and saves it in results map.
//
// Parameters:
//
//	guid - GUID saying which Aggregator XML will be read.
//	data - data read from "telem" file.
//
// Returns:
//
//	map[string]uint64 - results map with read data.
//	error - error if getting any of the values failed.
func (p *IntelPMT) getSampleValues(guid string, data []byte) (map[string]uint64, error) {
	results := make(map[string]uint64)
	for _, group := range p.pmtAggregator[guid].SampleGroup {
		// Determine starting position of the Sample Group.
		// Each Sample Group occupies 8 bytes.
		offset := 8 * group.SampleID
		for _, sample := range group.Sample {
			var err error
			results[sample.SampleID], err = getTelemSample(sample, data, offset)
			if err != nil {
				return nil, err
			}
		}
	}
	return results, nil
}

// getTelemSample extracts a telemetry sample from a given buffer.
//
// This function uses offset as a starting position.
// Then it uses LSB and MSB from sample to determine which bits
// to read from the given buffer.
//
// Parameters:
//
//	s - sample from Aggregator XML containing LSB and MSB info.
//	buf - the byte buffer containing the telemetry data.
//	offset - the starting position (in bytes) in the buffer.
//
// Returns:
//
//	uint64 - the extracted sample as a 64-bit unsigned integer.
//	error - error if offset+8 exceeds the size of the buffer.
func getTelemSample(s sample, buf []byte, offset uint64) (uint64, error) {
	if len(buf) < int(offset+8) {
		return 0, fmt.Errorf("error reading telemetry sample: insufficient bytes from offset %d in buffer of size %d", offset, len(buf))
	}
	data := binary.LittleEndian.Uint64(buf[offset : offset+8])
	msbMask := uint64(0xffffffffffffffff) & ((1 << (s.Msb + 1)) - 1)
	lsbMask := uint64(0xffffffffffffffff) & (1<<s.Lsb - 1)
	mask := msbMask & (^lsbMask)

	// Apply mask and shift right
	value := (data & mask) >> s.Lsb
	return value, nil
}

// aggregateSamples outputs transformed metrics to Telegraf.
//
// This method transforms low level samples
// into high-level samples with appropriate transformation equation.
// Then it creates fields and tags and adds them to Telegraf Accumulator.
//
// Parameters:
//
//	guid - GUID saying which Aggregator Interface will be read.
//	data - contents of the "telem" file.
//	numaNode - which NUMA node this sample belongs to.
//	acc - Telegraf Accumulator.
//
// Returns:
//
//	error - error if getting values has failed, if sample IDref is missing or if equation evaluation has failed.
func (p *IntelPMT) aggregateSamples(guid string, data []byte, numaNode string, acc telegraf.Accumulator) error {
	results, err := p.getSampleValues(guid, data)
	if err != nil {
		return err
	}
	for _, sample := range p.pmtAggregatorInterface[guid].AggregatorSamples.AggregatorSample {
		parameters := make(map[string]interface{})
		for _, input := range sample.TransformInputs.TransformInput {
			if _, ok := results[input.SampleIDREF]; !ok {
				return fmt.Errorf("sample with IDREF %q has not been found", input.SampleIDREF)
			}
			parameters[input.VarName] = results[input.SampleIDREF]
		}
		eq := transformEquation(p.pmtTransformations[guid][sample.TransformREF].Transform)
		res, err := eval(eq, parameters)
		if err != nil {
			return fmt.Errorf("error during eval of sample %q: %w", sample.SampleName, err)
		}
		fields := map[string]interface{}{
			"value": res,
		}
		tags := map[string]string{
			"guid":           guid,
			"numa_node":      numaNode,
			"sample_name":    sample.SampleName,
			"sample_group":   sample.SampleGroup,
			"datatype_idref": sample.DatatypeIDRef,
		}
		if sample.core != "" {
			tags["core"] = sample.core
		}
		if sample.cha != "" {
			tags["cha"] = sample.cha
		}

		acc.AddFields(pluginName, fields, tags)
	}
	return nil
}

// transformEquation changes the equation string to be ready for eval.
//
// This function removes "$" signs, which prefixes every parameter in equations.
// Then escapes special characters from XML
// like "&lt;" into "<", "&amp;" into "&" and "&gt;" into ">"
// so they can be used in evaluation.
//
// Parameters:
//
//	eq - string which should be transformed.
//
// Returns:
//
//	string - transformed string.
func transformEquation(eq string) string {
	withoutDollar := strings.ReplaceAll(eq, "$", "")
	decoded := html.UnescapeString(withoutDollar)
	return decoded
}

// eval calculates the value of given equation for given parameters.
//
// This function evaluates arbitrary equations with parameters.
// It substitutes the parameters in the equation with their values
// and calculates its value.
// Example: equation "a + b", with params: a: 2, b: 3.
// a and b will be substituted with their values so the equation becomes "2 + 3".
// If any of the parameters are missing then the equation is invalid and returns an error.
// Parameters:
//
//	eq - equation which should be calculated.
//	params - parameters to substitute in the equation.
//
// Returns:
//
//	interface - the value of calculation.
//	error - error if the equation is empty, if hex to dec conversion failed or if the equation is invalid.
func eval(eq string, params map[string]interface{}) (interface{}, error) {
	if eq == "" {
		return nil, fmt.Errorf("no transformation equation found")
	}
	// gval doesn't support hexadecimals
	eq = hexToDecRegex.ReplaceAllStringFunc(eq, hexToDec)
	if eq == "" {
		return nil, fmt.Errorf("error during hex to decimal conversion")
	}
	result, err := gval.Evaluate(eq, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func hexToDec(hexStr string) string {
	dec, err := strconv.ParseInt(hexStr, 0, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatInt(dec, 10)
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		return new(IntelPMT)
	})
}
