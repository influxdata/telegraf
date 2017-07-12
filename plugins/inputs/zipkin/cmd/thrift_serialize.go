package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
)

var (
	filename    *string
	outFileName *string
)

/*func init() {
	filename = flag.String("input", "", usage)
	outFileName = flag.String("output", "", usage)
}

const usage = `./json_serialize -input <input> -output output`*/

func main() {

	flag.Parse()
	b, err := ioutil.ReadFile("../testdata/json/distributed_trace_sample.json")
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	var spans []*zipkincore.Span
	span, err := serializeJSON(b)
	spans = append(spans, span)

	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	fmt.Println(spans)

	buf := thrift.NewTMemoryBuffer()
	transport := thrift.NewTBinaryProtocolTransport(buf)

	if err = transport.WriteListBegin(thrift.STRUCT, len(spans)); err != nil {
		panic(err)
	}

	for _, s := range spans {
		err = s.Write(transport)
		if err != nil {
			panic(err)
		}
	}

	if err = transport.WriteListEnd(); err != nil {
		panic(err)
	}

	log.Println(buf.Buffer.String())

	ioutil.WriteFile("../testdata/out.dat", buf.Buffer.Bytes(), 0644)

	b, err = ioutil.ReadFile("../testdata/out.dat")

	//log.Println("read bytes: ", b)
	if err != nil {
		log.Printf("%v\n", err)
	}
	deserializeThrift(b)
}

func serializeJSON(b []byte) (*zipkincore.Span, error) {
	var span *zipkincore.Span
	err := json.Unmarshal(b, &span)
	return span, err
}

func deserializeThrift(b []byte) {
	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(b); err != nil {
		log.Println("Error in buffer write: ", err)
		return
	}

	transport := thrift.NewTBinaryProtocolTransport(buffer)
	_, size, err := transport.ReadListBegin()
	if err != nil {
		log.Printf("Error in ReadListBegin: %s", err)
		return
	}

	var spans []*zipkincore.Span
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err = zs.Read(transport); err != nil {
			log.Printf("Error reading into zipkin struct: %s", err)
			return
		}
		spans = append(spans, zs)
	}

	err = transport.ReadListEnd()
	if err != nil {
		log.Printf("%s", err)
		return
	}

	//marshal json for debugging purposes
	out, _ := json.MarshalIndent(spans, "", "    ")
	log.Println(string(out))
}
