package grok

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrokParse(t *testing.T) {
	parser := Parser{
		Measurement: "t_met",
		Patterns:    []string{"%{COMMON_LOG_FORMAT}"},
	}
	parser.Compile()
	metrics, err := parser.Parse([]byte(`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`))
	log.Printf("metric_tags: %v, metric_fields: %v", metrics[0].Tags(), metrics[0].Fields())
	assert.NoError(t, err)
}
