package file

import (
	"fmt"
	"io"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type File struct {
	Files               []string          `toml:"files"`
	RotationInterval    internal.Duration `toml:"rotation_interval"`
	RotationMaxSize     internal.Size     `toml:"rotation_max_size"`
	RotationMaxArchives int               `toml:"rotation_max_archives"`
	UseBatchFormat      bool              `toml:"use_batch_format"`
	Log                 telegraf.Logger   `toml:"-"`

	writer     io.Writer
	closers    []io.Closer
	serializer serializers.Serializer
}

var sampleConfig = `
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Use batch serialization format instead of line based delimiting.  The
  ## batch format allows for the production of non line based output formats and
  ## may more efficiently encode metric groups.
  # use_batch_format = false

  ## The file will be rotated after the time interval specified.  When set
  ## to 0 no time based rotation is performed.
  # rotation_interval = "0d"

  ## The logfile will be rotated when it becomes larger than the specified
  ## size.  When set to 0 no size based rotation is performed.
  # rotation_max_size = "0MB"

  ## Maximum number of rotated archives to keep, any older logs are deleted.
  ## If set to -1, no archives are removed.
  # rotation_max_archives = 5

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
			of, err := rotate.NewFileWriter(
				file, f.RotationInterval.Duration, f.RotationMaxSize.Size, f.RotationMaxArchives)
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

	if f.UseBatchFormat {
		octets, err := f.serializer.SerializeBatch(metrics)
		if err != nil {
			f.Log.Errorf("Could not serialize metric: %v", err)
		}

		_, err = f.writer.Write(octets)
		if err != nil {
			f.Log.Errorf("Error writing to file: %v", err)
		}
	} else {
		for _, metric := range metrics {
			b, err := f.serializer.Serialize(metric)
			if err != nil {
				f.Log.Debugf("Could not serialize metric: %v", err)
			}

			_, err = f.writer.Write(b)
			if err != nil {
				writeErr = fmt.Errorf("E! [outputs.file] failed to write message: %v", err)
			}
		}
	}

	return writeErr
}

func init() {
	outputs.Add("file", func() telegraf.Output {
		return &File{}
	})
}
