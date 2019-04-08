package kinesis_consumer

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"

	"github.com/golang/snappy"
)

func decompressGZip(recordSlug []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(recordSlug)
	decompressionReader, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, fmt.Errorf("Failed to create reader for gzip. Error: %s", err)
	}
	defer decompressionReader.Close()
	b, err := ioutil.ReadAll(decompressionReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from gzip reader. Error:s %s", err)
	}
	return b, nil
}

func decompressSnappy(recordSlug []byte) ([]byte, error) {
	b, err := snappy.Decode(nil, recordSlug)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode snappy data. Error: %s", err)
	}

	return b, nil
}
