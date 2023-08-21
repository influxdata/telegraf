//go:build linux && amd64

package intel_pmt

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

type httpReader struct{}

func (httpReader) getReadCloser(source string) (io.ReadCloser, error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("error during url parsing: %w", err)
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reading %q, expected status code 200, got %d", source, resp.StatusCode)
	}
	return resp.Body, nil
}

func parseXML(source string, sr sourceReader, v interface{}) error {
	if sr == nil {
		return fmt.Errorf("XML reader failed to initialize")
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
