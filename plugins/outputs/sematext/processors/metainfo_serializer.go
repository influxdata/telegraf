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
		_, _ = output.Write(serializeMetainfo(meta))
		_, _ = output.WriteString("\n")
	}

	return output.Bytes()
}

func serializeMetainfo(metainfo *MetricMetainfo) []byte {
	var output bytes.Buffer

	_, _ = output.WriteString((*metainfo).namespace)
	_, _ = output.WriteString(",label=")
	_, _ = output.WriteString((*metainfo).label)
	_, _ = output.WriteString(",numericType=")
	_, _ = output.WriteString(strings.ToLower((*metainfo).numericType.String()))
	_, _ = output.WriteString(",os.host=")
	_, _ = output.WriteString((*metainfo).host)
	_, _ = output.WriteString(",token=")
	_, _ = output.WriteString((*metainfo).token)
	_, _ = output.WriteString(",type=")
	_, _ = output.WriteString(strings.ToLower((*metainfo).semType.String()))
	_, _ = output.WriteString(" ")
	_, _ = output.WriteString((*metainfo).name)

	return output.Bytes()
}

func NewMetainfoSerializer(log telegraf.Logger) *MetainfoSerializer {
	return &MetainfoSerializer{
		log: log,
	}
}
