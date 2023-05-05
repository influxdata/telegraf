//go:generate ../../../tools/readme_config_includer/generator
package file

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/klauspost/compress/zstd"
)

//go:embed sample.conf
var sampleConfig string

var ValidCompressionAlgorithmLevels = map[string][]int{
	"zstd": {1, 3, 7, 11},
}

type File struct {
	Files               []string        `toml:"files"`
	RotationInterval    config.Duration `toml:"rotation_interval"`
	RotationMaxSize     config.Size     `toml:"rotation_max_size"`
	RotationMaxArchives int             `toml:"rotation_max_archives"`
	UseBatchFormat      bool            `toml:"use_batch_format"`
	Compression         Compression     `toml:"compression"`
	Log                 telegraf.Logger `toml:"-"`

	writer     io.Writer
	closers    []io.Closer
	serializer serializers.Serializer
}

type Compression struct {
	Enabled   bool   `toml:"enabled"`
	Algorithm string `toml:"algorithm"`
	Level     int    `toml:"level"`
}

func ValidateCompressionAlgorithm(algorithm string) error {
	for validAlgorithm := range ValidCompressionAlgorithmLevels {
		if algorithm == validAlgorithm {
			return nil
		}
	}
	return fmt.Errorf("unknown or unsupported algorithm provided: %s", algorithm)
}

func ValidateCompressionLevel(algorithm string, level int) error {
	for _, validAlgorithmLevel := range ValidCompressionAlgorithmLevels[algorithm] {
		if level == validAlgorithmLevel {
			return nil
		}
	}
	return fmt.Errorf("unsupported compression level provided: %d. only %v are supported", level, ValidCompressionAlgorithmLevels[algorithm])
}

func CompressZstd(encoder *zstd.Encoder, src []byte) []byte {
	return encoder.EncodeAll(src, make([]byte, 0, len(src)))
}

func (*File) SampleConfig() string {
	return sampleConfig
}

func (f *File) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *File) Init() error {
  if f.Compression.Enabled {
	  err := ValidateCompressionAlgorithm(f.Compression.Algorithm)
	  if err != nil {
	  	return err
	  }
	  err = ValidateCompressionLevel(f.Compression.Algorithm, f.Compression.Level)
	  if err != nil {
	  	return err
	  }
  }
	return nil
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
	var encoder interface{}

	if f.Compression.Enabled {
		if f.Compression.Algorithm == "zstd" {
			encoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(f.Compression.Level)))
		}
	}
	if f.UseBatchFormat {
		octets, err := f.serializer.SerializeBatch(metrics)
		if f.Compression.Enabled {
			if f.Compression.Algorithm == "zstd" {
				octets = CompressZstd(encoder.(*zstd.Encoder), octets)
			}
		}
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
			if f.Compression.Enabled {
				if f.Compression.Algorithm == "zstd" {
					b = CompressZstd(encoder.(*zstd.Encoder), b)
				}
			}
			if err != nil {
				f.Log.Debugf("Could not serialize metric: %v", err)
			}

			_, err = f.writer.Write(b)
			if err != nil {
				writeErr = fmt.Errorf("failed to write message: %w", err)
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
