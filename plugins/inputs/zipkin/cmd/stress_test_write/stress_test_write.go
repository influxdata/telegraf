/*
This is a development testing cli tool meant to stress the zipkin telegraf plugin.
It writes a specified number of zipkin spans to the plugin endpoint, with other
parameters which dictate batch size and flush timeout.

Usage as follows:

`./stress_test_write -batch_size=<batch_size> -max_backlog=<max_span_buffer_backlog> -batch_interval=<batch_interval_in_seconds> \
-span_count<number_of_spans_to_write> -zipkin_host=<zipkin_service_hostname>`

Or with a timer:

`time ./stress_test_write -batch_size=<batch_size> -max_backlog=<max_span_buffer_backlog> -batch_interval=<batch_interval_in_seconds> \
-span_count<number_of_spans_to_write> -zipkin_host=<zipkin_service_hostname>`

However, the flag defaults work just fine for a good write stress test (and are what
this tool has mainly been tested with), so there shouldn't be much need to
manually tweak the parameters.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

var (
	batchSize         int
	maxBackLog        int
	batchTimeInterval int
	spanCount         int
	zipkinServerHost  string
)

func init() {
	flag.IntVar(&batchSize, "batch_size", 10000, "")
	flag.IntVar(&maxBackLog, "max_backlog", 100000, "")
	flag.IntVar(&batchTimeInterval, "batch_interval", 1, "")
	flag.IntVar(&spanCount, "span_count", 100000, "")
	flag.StringVar(&zipkinServerHost, "zipkin_host", "localhost", "")
}

func main() {
	flag.Parse()
	var hostname = fmt.Sprintf("http://%s:9411/api/v1/spans", zipkinServerHost)
	reporter := zipkinhttp.NewReporter(
		hostname,
		zipkinhttp.BatchSize(batchSize),
		zipkinhttp.MaxBacklog(maxBackLog),
		zipkinhttp.BatchInterval(time.Duration(batchTimeInterval)*time.Second),
	)
	defer reporter.Close()

	endpoint, err := zipkin.NewEndpoint("Trivial", "127.0.0.1:0")
	if err != nil {
		log.Panicf("Error: %v\n", err)
	}

	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Panicf("Error: %v\n", err)
	}

	tracer := zipkinot.Wrap(nativeTracer)

	log.Printf("Writing %d spans to zipkin server at %s\n", spanCount, hostname)
	for i := 0; i < spanCount; i++ {
		parent := tracer.StartSpan("Parent")
		parent.LogFields(otlog.Message(fmt.Sprintf("Trace%d", i)))
		parent.Finish()
	}
	log.Println("Done. Flushing remaining spans...")
}
