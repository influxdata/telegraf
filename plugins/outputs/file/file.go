//go:generate ../../../tools/readme_config_includer/generator
package file

import (
	"bytes"
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
	gzip "github.com/klauspost/pgzip"
)

//go:embed sample.conf
var sampleConfig string

var ValidCompressionAlgorithmLevels = map[string][]int{
	"zstd": {1, 3, 7, 11},
	"gzip": {-3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
}

type File struct {
	Files                []string        `toml:"files"`
	RotationInterval     config.Duration `toml:"rotation_interval"`
	RotationMaxSize      config.Size     `toml:"rotation_max_size"`
	RotationMaxArchives  int             `toml:"rotation_max_archives"`
	UseBatchFormat       bool            `toml:"use_batch_format"`
	CompressionAlgorithm string          `toml:"compression_algorithm"`
	CompressionLevel     int             `toml:"compression_level"`
	encoder              interface{}
	Log                  telegraf.Logger `toml:"-"`

	writer     io.Writer
	closers    []io.Closer
	serializer serializers.Serializer
}

func validateCompressionAlgorithm(algorithm string) error {
	for validAlgorithm := range ValidCompressionAlgorithmLevels {
		if algorithm == validAlgorithm {
			return nil
		}
	}
	return fmt.Errorf("unknown or unsupported algorithm provided: %s", algorithm)
}

func validateCompressionLevel(algorithm string, level int) error {
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

func CompressGzip(data []byte, level int) ([]byte, error) {
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, level)

	if err != nil {
		return nil, err
	}

	_, err = gz.Write(data)
	if err != nil {
		return nil, err
	}

	if err = gz.Flush(); err != nil {
		return nil, err
	}

	if err = gz.Close(); err != nil {
		return nil, err
	}

	compressedData := b.Bytes()

	return compressedData, err
}

func closeZstdEncoder(encoder *zstd.Encoder) {
	encoder.Close()
}

func (*File) SampleConfig() string {
	return sampleConfig
}

func (f *File) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *File) Init() error {
	if f.CompressionAlgorithm == "" {
		return nil
	} else if f.CompressionAlgorithm != "" && f.CompressionLevel == 0 {
		if f.CompressionAlgorithm == "zstd" {
			f.CompressionLevel = 3
		} else if f.CompressionAlgorithm == "gzip" {
			f.CompressionLevel = -1
		}
	}
	err := validateCompressionAlgorithm(f.CompressionAlgorithm)
	if err != nil {
		return err
	}
	err = validateCompressionLevel(f.CompressionAlgorithm, f.CompressionLevel)
	if err != nil {
		return err
	}
	return nil
}

func (f *File) Connect() error {
	writers := []io.Writer{}

	if len(f.Files) == 0 {
		f.Files = []string{"stdout"}
	}

	if f.CompressionAlgorithm == "zstd" {
		f.encoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(f.CompressionLevel)))
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
	if f.CompressionAlgorithm == "zstd" {
		closeZstdEncoder(f.encoder.(*zstd.Encoder))
	}
	return err
}

func (f *File) Write(metrics []telegraf.Metric) error {
	var writeErr error

	if f.UseBatchFormat {
		octets, err := f.serializer.SerializeBatch(metrics)
		if f.CompressionAlgorithm == "zstd" {
			octets = CompressZstd(f.encoder.(*zstd.Encoder), octets)
		}
		if err != nil {
			f.Log.Errorf("Could not serialize metric: %v", err)
		}

		if f.CompressionAlgorithm == "gzip" {
			octets, err = CompressGzip(octets, f.CompressionLevel)
			if err != nil {
				f.Log.Errorf("Error writing to file: %v", err)
			}
		}
		_, err = f.writer.Write(octets)
		if err != nil {
			f.Log.Errorf("Error writing to file: %v", err)
		}
	} else {
		for _, metric := range metrics {
			b, err := f.serializer.Serialize(metric)
			if f.CompressionAlgorithm == "zstd" {
				b = CompressZstd(f.encoder.(*zstd.Encoder), b)
			}
			if err != nil {
				f.Log.Debugf("Could not serialize metric: %v", err)
			}

			if f.CompressionAlgorithm == "gzip" {
				b, err = CompressGzip(b, f.CompressionLevel)
				if err != nil {
					f.Log.Errorf("Error writing to file: %v", err)
				}
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
