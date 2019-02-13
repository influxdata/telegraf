package fileinfo

import (
	"os"
	"regexp"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type FileInfoParser struct {
	DefaultTags map[string]string

	relativeDir   string
	fileRegexp    map[string]*regexp.Regexp
	fileTagRegexp map[string]*regexp.Regexp
}

func NewFileInfoParser(fileRegex map[string]string, fileTagRegex map[string]string) (*FileInfoParser, error) {
	r := make(map[string]*regexp.Regexp)
	tr := make(map[string]*regexp.Regexp)
	for key, value := range fileRegex {
		r[key] = regexp.MustCompile(value)
	}

	for key, value := range fileTagRegex {
		tr[key] = regexp.MustCompile(value)
	}

	return &FileInfoParser{
		fileRegexp:    r,
		fileTagRegexp: tr,
	}, nil
}

// Provided so that you can accurately calcuate the relative path against
// A specific source directory
func (p *FileInfoParser) SetRelativeDir(dir string) {
	p.relativeDir = dir
}

func (p *FileInfoParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	line := string(buf[:len(buf)])
	var metrics []telegraf.Metric
	metric, err := p.ParseLine(line)
	if metric == nil && err == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	metrics = append(metrics, metric)
	return metrics, nil
}

func (p *FileInfoParser) ParseLine(line string) (telegraf.Metric, error) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	for name, regex := range p.fileRegexp {
		match := regex.FindStringSubmatch(line)
		if len(match) > 1 {
			fields[name] = match[1]
		}
	}

	for name, regex := range p.fileTagRegexp {
		match := regex.FindStringSubmatch(line)
		if len(match) > 1 {
			tags[name] = match[1]
		}
	}

	osfi, err := os.Stat(line)
	if err != nil {
		return nil, err
	}

	fields["filesize"] = osfi.Size()

	m, err := metric.New("fileinfo", tags, fields, time.Now())

	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p *FileInfoParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
