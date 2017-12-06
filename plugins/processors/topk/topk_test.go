package topk

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf"
)

// NewTestTopk creates new test topk processor with specified config
func NewTestTopk() telegraf.Processor {
	topk := &TopK{}

	return topk
}

func TestTopkAvg(t *testing.T) {
	fmt.Println("sucess")
}
