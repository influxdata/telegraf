package main

import (
	"fmt"
	"log"
	"os"
	"time"

	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func main() {
	// 1) Create a opentracing.Tracer that sends data to Zipkin
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Need 1 argument\n")
		os.Exit(0)
	}
	collector, err := zipkin.NewHTTPCollector(
		fmt.Sprintf("http://%s:9411/api/v1/spans", os.Args[1]))
	if err != nil {
		log.Fatalf("error: ", err)
	}
	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", "trivial"))

	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Println("Writing 1,000,000 spans to zipkin impl...")
	for i := 0; i < 1000000; i++ {
		parent := tracer.StartSpan("Parent")
		parent.LogEvent(fmt.Sprintf("Trace %d\n", i))
		parent.Finish()
		time.Sleep(10 * time.Microsecond)
	}
	log.Println("Done!")

}
