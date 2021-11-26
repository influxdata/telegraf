package influx_protobuf

import (
	"fmt"

	influx "github.com/influxdata/influxdb-pb-data-protocol/golang"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
)

// Serializer encodes metrics in the InfluxData protocol-buffer format
type Serializer struct {
	DatabaseName string
	IsIox        bool
}

// Serialize implements serializers.Serializer.Serialize
// github.com/influxdata/telegraf/plugins/serializers/Serializer
func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.SerializeBatch([]telegraf.Metric{metric})
}

// SerializeBatch implements serializers.Serializer.SerializeBatch
// github.com/influxdata/telegraf/plugins/serializers/Serializer
func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	// Collect the metrics into tables and columns in those tables
	collection := make(map[string]table)
	for _, m := range metrics {
		name := m.Name()

		// Create a new table if it doesn't exist yet
		if _, found := collection[name]; !found {
			collection[name] = newTable()
		}
		c := collection[name]
		if err := c.addMetric(m, s.IsIox); err != nil {
			return nil, fmt.Errorf("adding metric %q failed: %v", name, err)
		}
		collection[name] = c
	}

	// Convert the data into the protocol-buffer format
	tables := make([]*influx.TableBatch, 0, len(collection))
	for name, tbl := range collection {
		table := influx.TableBatch{
			TableName: name,
			Columns:   make([]*influx.Column, 0, len(tbl.Tags)+len(tbl.Fields)+1),
			RowCount:  tbl.rows,
		}
		table.Columns = append(table.Columns, tbl.Time)
		for _, col := range tbl.Tags {
			table.Columns = append(table.Columns, col)
		}
		for _, col := range tbl.Fields {
			table.Columns = append(table.Columns, col)
		}
		tables = append(tables, &table)
	}

	db := influx.DatabaseBatch{
		DatabaseName: s.DatabaseName,
		TableBatches: tables,
	}
	return proto.Marshal(&db)
}
