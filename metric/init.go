package metric

import (
	"encoding/gob"
)

func init() {
	gob.RegisterName("metric.metric", &metric{})
}
