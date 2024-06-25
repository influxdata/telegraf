package metric

import "encoding/gob"

func Init() {
	gob.RegisterName("metric.metric", &metric{})
}
