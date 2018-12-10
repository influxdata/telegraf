package vsphere

import (
	"errors"
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
)

func HandlePanicWithAcc(acc telegraf.Accumulator) {
	if p := recover(); p != nil {
		switch p.(type) {
		case string:
			acc.AddError(errors.New(p.(string)))
		case error:
			acc.AddError(p.(error))
		default:
			acc.AddError(fmt.Errorf("Unknown panic: %s", p))
		}
	}
}

func HandlePanic() {
	if p := recover(); p != nil {
		log.Printf("E! [input.vsphere] PANIC (recovered): %s", p)
	}
}
