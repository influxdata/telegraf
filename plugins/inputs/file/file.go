package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dimchansky/utfbom"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type File struct {
	Files             []string `toml:"files"`
	FileTag           string   `toml:"file_tag"`
	CharacterEncoding string   `toml:"character_encoding"`

	parserFunc telegraf.ParserFunc
	filenames  []string
	decoder    *encoding.Decoder
}

func (f *File) Init() error {
	var err error
	f.decoder, err = encoding.NewDecoder(f.CharacterEncoding)
	return err
}

func (f *File) Gather(acc telegraf.Accumulator) error {
	err := f.refreshFilePaths()
	if err != nil {
		return err
	}
	for _, k := range f.filenames {
		metrics, err := f.readMetric(k)
		if err != nil {
			return err
		}

		for _, m := range metrics {
			if f.FileTag != "" {
				m.AddTag(f.FileTag, filepath.Base(k))
			}
			acc.AddMetric(m)
		}
	}
	return nil
}

func (f *File) SetParserFunc(fn telegraf.ParserFunc) {
	f.parserFunc = fn
}

func (f *File) refreshFilePaths() error {
	var allFiles []string
	for _, file := range f.Files {
		g, err := globpath.Compile(file)
		if err != nil {
			return fmt.Errorf("could not compile glob %v: %v", file, err)
		}
		files := g.Match()
		if len(files) <= 0 {
			return fmt.Errorf("could not find file: %v", file)
		}
		allFiles = append(allFiles, files...)
	}

	f.filenames = allFiles
	return nil
}

func (f *File) readMetric(filename string) ([]telegraf.Metric, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r, _ := utfbom.Skip(f.decoder.Reader(file))
	fileContents, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read %q: %s", filename, err)
	}
	parser, err := f.parserFunc()
	if err != nil {
		return nil, fmt.Errorf("could not instantiate parser: %s", err)
	}
	return parser.Parse(fileContents)
}

func init() {
	inputs.Add("file", func() telegraf.Input {
		return &File{}
	})
}
