package file

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type File struct {
	Files []string

	writer  io.Writer
	closers []io.Closer

	serializer serializers.Serializer
}

var sampleConfig = `
	## Files to write to, "stdout" is a specially handled file.
	## For files which contain curly bracket tokens, these tokens will be interpretted as a date/time format
	## so file will be generated based on provided format and UTC time on creation.
	## This can be used to create dated directories or include time in name
	## for example to create a file called metrics.out in a dir within /tmp with todays date use /tmp/{020106}/metric.out
	## similarly if the filename was to also contain the current date and time on creation use /tmp/{020106}/metrics{020106.150406}.out
	## for more info on token time format notation see https://golang.org/pkg/time/#Time.Format
	files = ["stdout", "/tmp/metrics.out", "/tmp/{020106}/metrics{020106.150406}.out"]
	
	
	## Data format to output.
	## Each data format has it's own unique set of configuration options, read
	## more about them here:
	## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
	data_format = "influx"
`

func (f *File) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *File) Connect() error {
	writers := []io.Writer{}

	if len(f.Files) == 0 {
		f.Files = []string{"stdout"}
	}

	for _, file := range f.Files {
		if file == "stdout" {
			writers = append(writers, os.Stdout)
			f.closers = append(f.closers, os.Stdout)
		} else {
			var of *os.File
			var err error
			generatedFile := generateFileName(file)
			if _, err := os.Stat(generatedFile); os.IsNotExist(err) {
				// create directory f it doesn't exist
				lastSlash := strings.LastIndex(generatedFile, "/")
				if lastSlash != -1 {
					err = os.MkdirAll(generatedFile[:lastSlash], os.ModeDir)
				}
				// create file if it doesn't exist
				if err == nil {
					of, err = os.Create(generatedFile)
				}
			} else {
				of, err = os.OpenFile(generatedFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			}

			if err != nil {
				return err
			}
			writers = append(writers, of)
			f.closers = append(f.closers, of)
		}
	}
	f.writer = io.MultiWriter(writers...)
	return nil
}

func (f *File) Close() error {
	var errS string
	for _, c := range f.closers {
		if err := c.Close(); err != nil {
			errS += err.Error() + "\n"
		}
	}
	if errS != "" {
		return fmt.Errorf(errS)
	}
	return nil
}

func (f *File) SampleConfig() string {
	return sampleConfig
}

func (f *File) Description() string {
	return "Send telegraf metrics to file(s)"
}

func (f *File) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		values, err := f.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		for _, value := range values {
			_, err = f.writer.Write([]byte(value + "\n"))
			if err != nil {
				return fmt.Errorf("FAILED to write message: %s, %s", value, err)
			}
		}
	}
	return nil
}

// Generate filename, replace tokens enclosed by { and } with time format
func generateFileName(s string) string {
	t := time.Now().UTC()
	//split on opening brace
	tokens := strings.Split(s, "{")
	// first token has no left bracket
	outputString := tokens[0]
	// cycle through remaining tokens
	for j := 1; j < len(tokens); j++ {
		//find closing brace
		index := strings.Index(tokens[j], "}")
		// if -1 we have opening brace with no closing brace so we don't format
		if index == -1 {
			outputString += tokens[j]
		} else {
			// Extract enclosed token and format, concatenate anything after closing brace on output
			generatedString := t.Format(tokens[j][:index])
			outputString += generatedString + tokens[j][index+1:]
		}
	}
	return outputString
}

func init() {
	outputs.Add("file", func() telegraf.Output {
		return &File{}
	})
}
