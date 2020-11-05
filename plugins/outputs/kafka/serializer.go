package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"strings"
	"time"
)


type format string

type Serializer struct {
	metricsFormat format
}
func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.createObject(metric), nil
}

//func (s *Serializer) getMetricsFormat() format {
//	return s.metricsFormat
//}


func serializeMetricFieldSeparate(name, fieldName string) string {
	return fmt.Sprintf("metric=%s field=%s ",
		strings.Replace(name, " ", "_", -1),
		strings.Replace(fieldName, " ", "_", -1),
	)
}

func serializeMetricIncludeField(name, fieldName string) string {
	return fmt.Sprintf("metric=%s_%s ",
		strings.Replace(name, " ", "_", -1),
		strings.Replace(fieldName, " ", "_", -1),
	)
}

// Serialize writes the telegraf.Metric to a byte slice.  May produce multiple
// lines of output if longer than maximum line length.  Lines are terminated
// with a newline (LF) char.
func (s *Serializer) createObject(metric telegraf.Metric) []byte  {

	//var zeroTime = time.Unix(0, 0)

	res :=  make(map[string]interface{})

	for fieldName, fieldValue := range metric.Fields() {
		res[fieldName] = fieldValue
		for _, tag := range metric.TagList() {
			res[tag.Key] = tag.Value
		}
	}
	//add timestamp
	res["timestamp"]  = time.Now().UnixNano() / 1000000
	res["name"] = metric.Name()

	data, err := json.Marshal(&res)
	if err != nil{
		fmt.Printf("S err=%v\n",err)
		return nil
	}
	return data
}
