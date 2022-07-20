package encoding

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
[[processors.encoding]]
  ## (required) Field specifies which string field to operate on
  # field = ""

  ## (required) Encoding is the algorithm used to encode the compressed binary
  ## data into a string Influx can process. Only base64 is supported for now.
  ## Because compression deals with binary data (not supported by Influx),
  ## encoding is required. However, encoding may be used with no compression if
  ## desired.
  # encoding = "base64"

  ## Destination field is the field where the encoding result will be stored.
  ## If not specified, field is used.
  # dest_field = ""

  ## Whether the original field should be removed if it doesn't match dest_field.
  # remove_original = false

  ## Operation determines whether to "encode" or "decode"
  # operation = "decode"

  ## Compression describes the compression algorithm used for the field. If
  ## empty, compression is skipped. Only gzip is supported for now. Compression
  ## requires that an encoding be set as Influx does not support binary data.
  # compression = "gzip"

  ## Compression field and tag allow the compression algorithm to be retrieved
  ## from a field or tag on a metric. Tag takes precedence over field. If
  ## neither is found, "compression" (if any) is used for the metric.
  # compression_field = ""
  # compression_tag = ""
`

type Encoding struct {
	Field            string `toml:"field"`
	RemoveOriginal   bool   `toml:"remove_original"`
	DestField        string `toml:"dest_field"`
	Operation        string `toml:"operation"`
	Compression      string `toml:"compression"`
	CompressionField string `toml:"compression_field"`
	CompressionTag   string `toml:"compression_tag"`
	Encoding         string `toml:"encoding"`

	Log       telegraf.Logger `toml:"-"`
	operation func(string, string, string) (string, error)
}

func (e *Encoding) SampleConfig() string {
	return sampleConfig
}

func (e *Encoding) Description() string {
	return "Encodes or decodes data that may also have been compressed"
}

func (e *Encoding) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		valueIntf, ok := point.GetField(e.Field)
		if !ok {
			continue
		}

		value, ok := valueIntf.(string)
		if !ok {
			e.Log.Warnf("skipping encoding because metric had non-string value for field %v: %v", e.Field, valueIntf)
			continue
		}

		var compression string
		var compressionFound bool

		if e.CompressionTag != "" {
			compression, compressionFound = point.GetTag(e.CompressionTag)
		}

		if !compressionFound && e.CompressionField != "" {
			if field, ok := point.GetField(e.CompressionField); ok {
				compression, compressionFound = field.(string)
			}
		}

		if !compressionFound {
			compression = e.Compression
		}

		newValue, err := e.operation(value, compression, e.Encoding)
		if err != nil {
			e.Log.Warnf("failed to %v a metric: %v", e.Operation, err)
			continue
		}

		point.AddField(e.DestField, newValue)

		if e.RemoveOriginal {
			point.RemoveField(e.Field)
		}
	}

	return in
}

func (e *Encoding) Init() error {
	switch e.Compression {
	case "gzip":
	case "":
	default:
		return fmt.Errorf("'%v' is not a supported compression algorithm. It must be 'gzip' or '' to skip.", e.Compression)
	}

	switch e.Encoding {
	case "base64":
	case "":
		return fmt.Errorf("'encoding' is required for the encoding processor.")
	default:
		return fmt.Errorf("'%v' is not a supported encoding. It must be 'base64'.", e.Encoding)
	}

	if e.DestField == "" {
		e.DestField = e.Field
	}

	if e.DestField == e.Field {
		e.RemoveOriginal = false
	}

	switch e.Operation {
	case "encode":
		e.operation = encode
	case "decode":
		e.operation = decode
	default:
		return fmt.Errorf("'%v' is not a supported operation. It must be one of 'encode' or 'decode'.", e.Operation)
	}

	return nil
}

func newEncoding() *Encoding {
	return &Encoding{
		Operation:   "decode",
		Compression: "gzip",
		Encoding:    "base64",
	}
}

func init() {
	processors.Add("encoding", func() telegraf.Processor {
		return newEncoding()
	})
}

func encode(value, compression, encoding string) (string, error) {
	switch compression {
	case "gzip":
		var buf bytes.Buffer
		zipper := gzip.NewWriter(&buf)
		_, err := zipper.Write([]byte(value))
		zipper.Close()
		if err != nil {
			return "", err
		}
		value = string(buf.Bytes())
	}

	switch encoding {
	case "base64":
		value = base64.StdEncoding.EncodeToString([]byte(value))
	}

	return value, nil
}

func decode(value, compression, encoding string) (string, error) {
	switch encoding {
	case "base64":
		var data []byte
		var err error
		if strings.HasSuffix(value, string(base64.StdPadding)) {
			data, err = base64.StdEncoding.DecodeString(value)
		} else {
			data, err = base64.RawStdEncoding.DecodeString(value)
		}
		if err != nil {
			return "", err
		}
		value = string(data)
	}

	switch compression {
	case "gzip":
		unzip, err := gzip.NewReader(strings.NewReader(value))
		if err != nil {
			return "", err
		}

		data, err := ioutil.ReadAll(unzip)
		if err != nil {
			return "", err
		}

		value = string(data)
	}

	return value, nil
}
