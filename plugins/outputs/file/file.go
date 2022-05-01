package file

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type File struct {
	Files               []string        `toml:"files"`
	RotationInterval    config.Duration `toml:"rotation_interval"`
	RotationMaxSize     config.Size     `toml:"rotation_max_size"`
	RotationMaxArchives int             `toml:"rotation_max_archives"`
	UseBatchFormat      bool            `toml:"use_batch_format"`
	Log                 telegraf.Logger `toml:"-"`

	writer     io.Writer
	closers    []io.Closer
	serializer serializers.Serializer
}

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
				file, time.Duration(f.RotationInterval), int64(f.RotationMaxSize), f.RotationMaxArchives)
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

func (f *File) Write(metrics []telegraf.Metric) error {
	var writeErr error

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
				writeErr = fmt.Errorf("failed to write message: %v", err)
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
