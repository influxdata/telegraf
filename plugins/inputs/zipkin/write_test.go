package zipkin

import (
	"fmt"
	"log"
	"testing"
	"time"

	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func BenchmarkWrites(b *testing.B) {
	collector, err := zipkin.NewHTTPCollector(
		fmt.Sprintf("http://%s:9411/api/v1/spans", "localhost"))
	if err != nil {
		log.Fatalf("error: ", err)
	}
	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", "trivial"))

	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Println("Writing 1,000,000 spans to zipkin impl...")
	for i := 0; i < 20; i++ {
		log.Println("going")
		parent := tracer.StartSpan("Parent")
		parent.LogEvent(fmt.Sprintf("Trace %d\n", i))
		parent.Finish()
		time.Sleep(10 * time.Microsecond)
	}
	log.Println("Done!")
}
