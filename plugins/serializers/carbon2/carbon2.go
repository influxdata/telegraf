package carbon2

import (
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
	return []byte(s.createObject(metric)), nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch strings.Builder
	for _, metric := range metrics {
		batch.WriteString(s.createObject(metric))
	}
	return []byte(batch.String()), nil
}

func (s *serializer) createObject(metric telegraf.Metric) string {
	var m strings.Builder
	for fieldName, fieldValue := range metric.Fields() {
		if isNumeric(fieldValue) {
			m.WriteString("metric=")
			m.WriteString(strings.Replace(metric.Name(), " ", "_", -1))
			m.WriteString(" field=")
			m.WriteString(strings.Replace(fieldName, " ", "_", -1))
			m.WriteString(" ")
			for k, v := range metric.Tags() {
				m.WriteString(strings.Replace(k, " ", "_", -1))
				m.WriteString("=")
				m.WriteString(strings.Replace(v, " ", "_", -1))
				m.WriteString(" ")
			}
			m.WriteString(" ")
			m.WriteString(fmt.Sprintf("%v", fieldValue))
			m.WriteString(" ")
			m.WriteString(strconv.FormatInt(metric.Time().Unix(), 10))
			m.WriteString("\n")
		}
	}
	return m.String()
}

func isNumeric(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	default:
		return true
	}
}
