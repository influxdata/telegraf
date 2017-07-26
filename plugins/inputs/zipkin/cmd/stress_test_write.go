package main

import (
	"fmt"
	"log"
	"os"

	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func main() {
	// 1) Create a opentracing.Tracer that sends data to Zipkin
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Need 1 argument\n")
		os.Exit(0)
	}
	collector, err := zipkin.NewHTTPCollector(
		fmt.Sprintf("http://%s:9411/api/v1/spans", os.Args[1]), zipkin.HTTPBatchSize(1), zipkin.HTTPMaxBacklog(100000))
	defer collector.Close()
	if err != nil {
		log.Fatalf("error: ", err)
	}

	//zipkin.HTTPMaxBacklog(1000000)(collector.(*zipkin.HTTPCollector))
	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", "trivial"))

	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Println("Writing 1000000 spans to zipkin impl...")
	//var wg sync.WaitGroup
	for i := 0; i < 100000; i++ {

		log.Printf("Writing span %d\n", i)
		/*go func(i int) {
			wg.Add(1)
			defer wg.Done()
			log.Println("Writing span %d\n", i)
			parent := tracer.StartSpan("Parent")
			parent.LogEvent(fmt.Sprintf("Trace%d", i))
			parent.Finish()
			time.Sleep(2 * time.Second)
		}(i)*/
		parent := tracer.StartSpan("Parent")
		parent.LogEvent(fmt.Sprintf("Trace%d", i))
		parent.Finish()
		//	time.Sleep(2 * time.Second)
	}
	log.Println("Done!")
}
