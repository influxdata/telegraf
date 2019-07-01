package kinesis

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/golang/snappy"
)

var (
	// gzipCompressionLevel sets the compression level. Tests indicate that 7 gives the best trade off
	// between speed and compression.
	gzipCompressionLevel = 7
)

func gzipMetrics(metrics []byte) ([]byte, error) {
	var buffer bytes.Buffer

	gzw, err := gzip.NewWriterLevel(&buffer, gzipCompressionLevel)
	if err != nil {
		return []byte{}, fmt.Errorf("Compression level is incorrect for gzip")
	}
	_, err = gzw.Write(metrics)
	if err != nil {
		return []byte{}, fmt.Errorf("There was an error in writing to the gzip writer")
	}
	if err := gzw.Close(); err != nil {
		return []byte{}, fmt.Errorf("There was an error in closing the gzip writer")
	}

	return buffer.Bytes(), nil
}

func snappyMetrics(metrics []byte) ([]byte, error) {
	return snappy.Encode(nil, metrics), nil
}
