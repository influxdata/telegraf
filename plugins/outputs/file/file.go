package file

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type File struct {
	Files        []string
	RotateMaxAge string

	writer  io.Writer
	closers []io.Closer

	serializer serializers.Serializer
}

var sampleConfig = `
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## If this is defined, files will be rotated by the time.Duration specified
  # rotate_max_age = "1m"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
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
		} else {
			var of io.WriteCloser
			var err error
			if f.RotateMaxAge != "" {
				maxAge, err := time.ParseDuration(f.RotateMaxAge)
				if err != nil {
					return err
				}

				// Only rotate by file age for now, keep no archives.
				of, err = rotate.NewFileWriter(file, maxAge, 0, -1)
			} else {
				// Just open a normal file
				of, err = rotate.NewFileWriter(file, 0, 0, -1)
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
	var err error
	for _, c := range f.closers {
		errClose := c.Close()
		if errClose != nil {
			err = errClose
		}
	}
	return err
}

func (f *File) SampleConfig() string {
	return sampleConfig
}

func (f *File) Description() string {
	return "Send telegraf metrics to file(s)"
}

func (f *File) Write(metrics []telegraf.Metric) error {
	var writeErr error = nil

	for _, metric := range metrics {
		b, err := f.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}

		_, err = f.writer.Write(b)
		if err != nil {
			writeErr = fmt.Errorf("E! failed to write message: %s, %s", b, err)
		}
	}

	return writeErr
}

func init() {
	outputs.Add("file", func() telegraf.Output {
		return &File{}
	})
}
