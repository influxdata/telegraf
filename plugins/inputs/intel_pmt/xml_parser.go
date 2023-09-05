//go:build linux && amd64

package intel_pmt

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type pmt struct {
	XMLName  xml.Name `xml:"pmt"`
	Mappings mappings `xml:"mappings"`
}
type mappings struct {
	XMLName xml.Name  `xml:"mappings"`
	Mapping []mapping `xml:"mapping"`
}

type mapping struct {
	XMLName xml.Name `xml:"mapping"`
	GUID    string   `xml:"guid,attr"`
	XMLSet  xmlset   `xml:"xmlset"`
}

type xmlset struct {
	XMLName             xml.Name `xml:"xmlset"`
	Basedir             string   `xml:"basedir"`
	Aggregator          string   `xml:"aggregator"`
	AggregatorInterface string   `xml:"aggregatorinterface"`
}

type aggregator struct {
	XMLName     xml.Name      `xml:"Aggregator"`
	Name        string        `xml:"name"`
	SampleGroup []sampleGroup `xml:"SampleGroup"`
}

type sampleGroup struct {
	XMLName  xml.Name `xml:"SampleGroup"`
	SampleID uint64   `xml:"sampleID,attr"`
	Sample   []sample `xml:"sample"`
}

type sample struct {
	XMLName       xml.Name `xml:"sample"`
	SampleName    string   `xml:"name,attr"`
	DatatypeIDRef string   `xml:"datatypeIDREF,attr"`
	SampleID      string   `xml:"sampleID,attr"`
	Lsb           uint64   `xml:"lsb"`
	Msb           uint64   `xml:"msb"`

	mask uint64
}

type aggregatorInterface struct {
	XMLName           xml.Name          `xml:"AggregatorInterface"`
	Transformations   transformations   `xml:"TransFormations"`
	AggregatorSamples aggregatorSamples `xml:"AggregatorSamples"`
}

type transformations struct {
	XMLName        xml.Name         `xml:"TransFormations"`
	Transformation []transformation `xml:"TransFormation"`
}

type transformation struct {
	XMLName     xml.Name `xml:"TransFormation"`
	Name        string   `xml:"name,attr"`
	TransformID string   `xml:"transformID,attr"`
	Transform   string   `xml:"transform"`
}

type aggregatorSamples struct {
	XMLName          xml.Name           `xml:"AggregatorSamples"`
	AggregatorSample []aggregatorSample `xml:"T_AggregatorSample"`
}

type aggregatorSample struct {
	XMLName         xml.Name        `xml:"T_AggregatorSample"`
	SampleName      string          `xml:"sampleName,attr"`
	SampleGroup     string          `xml:"sampleGroup,attr"`
	DatatypeIDRef   string          `xml:"datatypeIDREF,attr"`
	TransformInputs transformInputs `xml:"TransFormInputs"`
	TransformREF    string          `xml:"transformREF"`

	core string
	cha  string
}

type transformInputs struct {
	XMLName        xml.Name         `xml:"TransFormInputs"`
	TransformInput []transformInput `xml:"TransFormInput"`
}

type transformInput struct {
	XMLName     xml.Name `xml:"TransFormInput"`
	VarName     string   `xml:"varName,attr"`
	SampleIDREF string   `xml:"sampleIDREF"`
}

type sourceReader interface {
	getReadCloser(source string) (io.ReadCloser, error)
}

type fileReader struct{}

func (fileReader) getReadCloser(source string) (io.ReadCloser, error) {
	return os.Open(source)
}

// parseXMLs reads and parses PMT XMLs.
//
// This method retrieves all metadata about known GUIDs from PmtSpec.
// Then, it explores PMT sysfs to find all readable "telem" files and their GUIDs.
// It then matches found (readable) system GUIDs with GUIDs from metadata and
// reads corresponding sets of XMLs.
//
// Returns:
//
//	error - if PMT spec is empty, if exploring PMT sysfs fails, or if reading XMLs fails.
func (p *IntelPMT) parseXMLs() error {
	err := parseXML(p.PmtSpec, p.reader, &p.pmtMetadata)
	if err != nil {
		return err
	}
	if len(p.pmtMetadata.Mappings.Mapping) == 0 {
		return errors.New("pmt XML provided contains no mappings")
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
	sampleFilterFound := make(map[string]bool)
	for guid := range p.pmtTelemetryFiles {
		err := p.getAllXMLData(guid, dtMetricsFound, sampleFilterFound)
		if err != nil {
			return fmt.Errorf("failed reading XMLs: %w", err)
		}
	}
	for _, dt := range p.DatatypeFilter {
		if _, ok := dtMetricsFound[dt]; !ok {
			p.Log.Warnf("Configured datatype metric %q has not been found", dt)
		}
	}
	for _, sm := range p.SampleFilter {
		if _, ok := sampleFilterFound[sm]; !ok {
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

			aggSource = filepath.Join(p.pmtBasePath, basedir, mapping.XMLSet.Aggregator)
			aggInterfaceSource = filepath.Join(p.pmtBasePath, basedir, mapping.XMLSet.AggregatorInterface)

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
			if len(p.DatatypeFilter) > 0 {
				tAgg.filterAggregatorByDatatype(p.DatatypeFilter)
				tAggInterface.filterAggInterfaceByDatatype(p.DatatypeFilter, dtMetricsFound)
			}
			if len(p.SampleFilter) > 0 {
				tAgg.filterAggregatorBySampleName(p.SampleFilter)
				tAggInterface.filterAggInterfaceBySampleName(p.SampleFilter, smFound)
			}
			tAgg.calculateMasks()
			p.pmtAggregator[guid] = tAgg
			tAggInterface.extractTagsFromSample()
			p.pmtAggregatorInterface[guid] = tAggInterface
		}
	}
	return nil
}

func (a *aggregator) calculateMasks() {
	for i := range a.SampleGroup {
		for j, sample := range a.SampleGroup[i].Sample {
			mask := computeMask(sample.Msb, sample.Lsb)
			a.SampleGroup[i].Sample[j].mask = mask
		}
	}
}

func computeMask(msb uint64, lsb uint64) uint64 {
	msbMask := uint64(0xffffffffffffffff) & ((1 << (msb + 1)) - 1)
	lsbMask := uint64(0xffffffffffffffff) & (1<<lsb - 1)
	return msbMask & (^lsbMask)
}

func parseXML(source string, sr sourceReader, v interface{}) error {
	if sr == nil {
		return errors.New("XML reader failed to initialize")
	}
	reader, err := sr.getReadCloser(source)
	if err != nil {
		return fmt.Errorf("error reading source %q: %w", source, err)
	}
	defer reader.Close()

	parser := xml.NewDecoder(reader)
	parser.AutoClose = xml.HTMLAutoClose
	parser.Entity = xml.HTMLEntity
	// There are "&" in XMLs in entity references.
	// Parser sees it as not allowed characters.
	// Strict mode disabled to handle that.
	parser.Strict = false
	err = parser.Decode(v)
	if err != nil {
		return fmt.Errorf("error decoding an XML %q: %w", source, err)
	}
	return nil
}
