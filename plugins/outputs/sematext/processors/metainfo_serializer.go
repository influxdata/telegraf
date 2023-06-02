package processors

import (
	"bytes"
	"github.com/influxdata/telegraf"
	"strings"
)

type MetainfoSerializer struct {
	log telegraf.Logger
}

func (m *MetainfoSerializer) Write(mList []*MetricMetainfo) []byte {
	var output bytes.Buffer

	for _, meta := range mList {
		output.Write(serializeMetainfo(meta))
		output.WriteString("\n")
	}

	return output.Bytes()
}

func serializeMetainfo(metainfo *MetricMetainfo) []byte {
	var output bytes.Buffer

	output.WriteString((*metainfo).namespace)
	output.WriteString(",label=")
	output.WriteString((*metainfo).label)
	output.WriteString(",numericType=")
	output.WriteString(strings.ToLower((*metainfo).numericType.String()))
	output.WriteString(",os.host=")
	output.WriteString((*metainfo).host)
	output.WriteString(",token=")
	output.WriteString((*metainfo).token)
	output.WriteString(",type=")
	output.WriteString(strings.ToLower((*metainfo).semType.String()))
	output.WriteString(" ")
	output.WriteString((*metainfo).name)

	return output.Bytes()
}

func NewMetainfoSerializer(log telegraf.Logger) *MetainfoSerializer {
	return &MetainfoSerializer{
		log: log,
	}
}
