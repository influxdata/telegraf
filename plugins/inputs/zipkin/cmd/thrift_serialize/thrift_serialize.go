/*
A small cli utility meant to convert json to zipkin thrift binary format, and
vice versa.

To convert from json to thrift,
the json is unmarshalled, converted to zipkincore.Span structures, and
marshalled into thrift binary protocol. The json must be in an array format (even if it only has one object),
because the tool automatically tries to unmarshall the json into an array of structs.

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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
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
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file: %v\n", err)
	}

	switch inputType {
	case "json":
		raw, err := jsonToZipkinThrift(contents)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if err := ioutil.WriteFile(outFileName, raw, 0644); err != nil {
			log.Fatalf("%v", err)
		}
	case "thrift":
		raw, err := thriftToJSONSpans(contents)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if err := ioutil.WriteFile(outFileName, raw, 0644); err != nil {
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
		return nil, fmt.Errorf("error unmarshalling: %v", err)
	}

	var zspans []*zipkincore.Span
	if err != nil {
		return nil, err
	}
	zspans = append(zspans, spans...)

	fmt.Println(spans)

	buf := thrift.NewTMemoryBuffer()
	transport := thrift.NewTBinaryProtocolTransport(buf)

	if err = transport.WriteListBegin(thrift.STRUCT, len(spans)); err != nil {
		return nil, fmt.Errorf("error in beginning thrift write: %v", err)
	}

	for _, span := range zspans {
		err = span.Write(transport)
		if err != nil {
			return nil, fmt.Errorf("error converting zipkin struct to thrift: %v", err)
		}
	}

	if err = transport.WriteListEnd(); err != nil {
		return nil, fmt.Errorf("error finishing thrift write: %v", err)
	}

	return buf.Bytes(), nil
}

func thriftToJSONSpans(thriftData []byte) ([]byte, error) {
	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(thriftData); err != nil {
		err = fmt.Errorf("error in buffer write: %v", err)
		return nil, err
	}

	transport := thrift.NewTBinaryProtocolTransport(buffer)
	_, size, err := transport.ReadListBegin()
	if err != nil {
		err = fmt.Errorf("error in ReadListBegin: %v", err)
		return nil, err
	}

	var spans []*zipkincore.Span
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err = zs.Read(transport); err != nil {
			err = fmt.Errorf("Error reading into zipkin struct: %v", err)
			return nil, err
		}
		spans = append(spans, zs)
	}

	err = transport.ReadListEnd()
	if err != nil {
		err = fmt.Errorf("error ending thrift read: %v", err)
		return nil, err
	}

	out, _ := json.MarshalIndent(spans, "", "    ")
	return out, nil
}
