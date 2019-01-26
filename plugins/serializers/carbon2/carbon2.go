package carbon2

import (
	"bytes"
	"fmt"
	"github.com/influxdata/telegraf"
	"strconv"
	"strings"
)

type serializer struct {
}

func NewSerializer() (*serializer, error) {
	s := &serializer{}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.createObject(metric), nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch bytes.Buffer
	for _, metric := range metrics {
		batch.Write(s.createObject(metric))
	}
	return batch.Bytes(), nil
}

func (s *serializer) createObject(metric telegraf.Metric) []byte {
	var m bytes.Buffer
	for fieldName, fieldValue := range metric.Fields() {
		if isNumeric(fieldValue) {
			m.WriteString("metric=")
			m.WriteString(strings.Replace(metric.Name(), " ", "_", -1))
			m.WriteString(" field=")
			m.WriteString(strings.Replace(fieldName, " ", "_", -1))
			m.WriteString(" ")
			for _, tag := range metric.TagList() {
				m.WriteString(strings.Replace(tag.Key, " ", "_", -1))
				m.WriteString("=")
				value := tag.Value
				if len(value) == 0 {
					value = "null"
				}
				m.WriteString(strings.Replace(value, " ", "_", -1))
				m.WriteString(" ")
			}
			m.WriteString(" ")
			m.WriteString(fmt.Sprintf("%v", fieldValue))
			m.WriteString(" ")
			m.WriteString(strconv.FormatInt(metric.Time().Unix(), 10))
			m.WriteString("\n")
		}
	}
	return m.Bytes()
}

func isNumeric(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	default:
		return true
	}
}
