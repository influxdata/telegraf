/*
A small cli utility meant to convert json to zipkin thrift binary format, and
vice versa.

To convert from json to thrift,
the json is unmarshalled, converted to zipkincore.Span structures, and
marshalled into thrift binary protocol. The json must be in an array format (even if it only has one object),
because the tool automatically tries to unmarshal the json into an array of structs.

To convert from thrift to json,
the opposite process must happen. The thrift binary data must be read into an array of
zipkin span structures, and those spans must be marshalled into json.

Usage:

./thrift_serialize -input <input-file> -output <output-file> -deserialize<true|false>

If `deserialize` is set to true (false by default), the tool will interpret the input file as
thrift, and write it as json to the output file.
Otherwise, the input file will be interpreted as json, and the output will be encoded as thrift.
*/
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec/thrift/gen-go/zipkincore"
)

var (
	filename    string
	outFileName string
	inputType   string
)

const usage = `./json_serialize -input <input> -output output -input-type<json|thrift>`

func init() {
	flag.StringVar(&filename, "input", "", usage)
	flag.StringVar(&outFileName, "output", "", usage)
	flag.StringVar(&inputType, "input-type", "thrift", usage)
}

func main() {
	flag.Parse()
	contents, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file: %v\n", err)
	}

	switch inputType {
	case "json":
		raw, err := jsonToZipkinThrift(contents)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if err := os.WriteFile(outFileName, raw, 0640); err != nil {
			log.Fatalf("%v", err)
		}
	case "thrift":
		raw, err := thriftToJSONSpans(contents)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if err := os.WriteFile(outFileName, raw, 0640); err != nil {
			log.Fatalf("%v", err)
		}
	default:
		log.Fatalf("Unsupported input type")
	}
}

func jsonToZipkinThrift(jsonRaw []byte) ([]byte, error) {
	if len(jsonRaw) == 0 {
		return nil, errors.New("no data")
	}

	if string(jsonRaw)[0] != '[' {
		return nil, errors.New("cannot unmarshal non array type")
	}

	var spans []*zipkincore.Span
	err := json.Unmarshal(jsonRaw, &spans)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling: %w", err)
	}

	var zspans []*zipkincore.Span
	if err != nil {
		return nil, err
	}
	zspans = append(zspans, spans...)

	buf := thrift.NewTMemoryBuffer()
	transport := thrift.NewTBinaryProtocolConf(buf, nil)

	if err = transport.WriteListBegin(context.Background(), thrift.STRUCT, len(spans)); err != nil {
		return nil, fmt.Errorf("error in beginning thrift write: %w", err)
	}

	for _, span := range zspans {
		err = span.Write(context.Background(), transport)
		if err != nil {
			return nil, fmt.Errorf("error converting zipkin struct to thrift: %w", err)
		}
	}

	if err = transport.WriteListEnd(context.Background()); err != nil {
		return nil, fmt.Errorf("error finishing thrift write: %w", err)
	}

	return buf.Bytes(), nil
}

func thriftToJSONSpans(thriftData []byte) ([]byte, error) {
	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(thriftData); err != nil {
		return nil, fmt.Errorf("error in buffer write: %w", err)
	}

	transport := thrift.NewTBinaryProtocolConf(buffer, nil)
	_, size, err := transport.ReadListBegin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error in ReadListBegin: %w", err)
	}

	spans := make([]*zipkincore.Span, 0, size)
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err = zs.Read(context.Background(), transport); err != nil {
			return nil, fmt.Errorf("error reading into zipkin struct: %w", err)
		}
		spans = append(spans, zs)
	}

	err = transport.ReadListEnd(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error ending thrift read: %w", err)
	}

	out, _ := json.MarshalIndent(spans, "", "    ")
	return out, nil
}
