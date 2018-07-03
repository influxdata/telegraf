package rotatingfile

import (
	"errors"
	"fmt"
	"io"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type File struct {
	Root           string
	FilenamePrefix string
	MaxAge         string

	writer     io.WriteCloser
	serializer serializers.Serializer
}

var sampleConfig = `
  ## Path to write files into.
  root = "/tmp"
  filename_prefix = "metrics"
  max_age = "1m"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (f *File) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *File) Connect() error {
	if len(f.Root) == 0 {
		return errors.New("we need a root path")
	}

	var err error
	f.writer, err = NewRotatingWriter(f.Root, f.FilenamePrefix, f.MaxAge)
	if err != nil {
		return err
	}
	return nil
}

func (f *File) Close() error {
	return f.writer.Close()
}

func (f *File) SampleConfig() string {
	return sampleConfig
}

func (f *File) Description() string {
	return "Send telegraf metrics to a rotating file"
}
func (f *File) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		b, err := f.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}
		_, err = f.writer.Write(b)
		if err != nil {
			return fmt.Errorf("failed to write message: %s, %s", b, err)
		}
	}
	return nil
}

func init() {
	outputs.Add("rotating_file", func() telegraf.Output {
		return &File{}
	})
}
