//go:generate ../../../tools/readme_config_includer/generator
package parquet

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

var defaultTimestampFieldName = "timestamp"

type metricGroup struct {
	filename string
	builder  *array.RecordBuilder
	schema   *arrow.Schema
	writer   *pqarrow.FileWriter
}

type Parquet struct {
	Directory          string          `toml:"directory"`
	RotationInterval   config.Duration `toml:"rotation_interval"`
	TimestampFieldName string          `toml:"timestamp_field_name"`
	Log                telegraf.Logger `toml:"-"`

	metricGroups map[string]*metricGroup
}

func (*Parquet) SampleConfig() string {
	return sampleConfig
}

func (p *Parquet) Init() error {
	if p.Directory == "" {
		p.Directory = "."
	}

	stat, err := os.Stat(p.Directory)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(p.Directory, 0750); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", p.Directory, err)
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("provided directory %q is not a directory", p.Directory)
	}

	p.metricGroups = make(map[string]*metricGroup)

	return nil
}

func (*Parquet) Connect() error {
	return nil
}

func (p *Parquet) Close() error {
	var errorOccurred bool

	for _, metrics := range p.metricGroups {
		if err := metrics.writer.Close(); err != nil {
			p.Log.Errorf("failed to close file %q: %v", metrics.filename, err)
			errorOccurred = true
		}
	}

	if errorOccurred {
		return errors.New("failed closing one or more parquet files")
	}

	return nil
}

func (p *Parquet) Write(metrics []telegraf.Metric) error {
	groupedMetrics := make(map[string][]telegraf.Metric)
	for _, metric := range metrics {
		groupedMetrics[metric.Name()] = append(groupedMetrics[metric.Name()], metric)
	}

	now := time.Now()
	for name, metrics := range groupedMetrics {
		if _, ok := p.metricGroups[name]; !ok {
			filename := fmt.Sprintf("%s/%s-%s-%s.parquet", p.Directory, name, now.Format("2006-01-02"), strconv.FormatInt(now.Unix(), 10))
			schema, err := p.createSchema(metrics)
			if err != nil {
				return fmt.Errorf("failed to create schema for file %q: %w", name, err)
			}
			writer, err := p.createWriter(name, filename, schema)
			if err != nil {
				return fmt.Errorf("failed to create writer for file %q: %w", name, err)
			}
			p.metricGroups[name] = &metricGroup{
				builder:  array.NewRecordBuilder(memory.DefaultAllocator, schema),
				filename: filename,
				schema:   schema,
				writer:   writer,
			}
		}

		if p.RotationInterval != 0 {
			if err := p.rotateIfNeeded(name); err != nil {
				return fmt.Errorf("failed to rotate file %q: %w", p.metricGroups[name].filename, err)
			}
		}

		record, err := p.createRecord(metrics, p.metricGroups[name].builder, p.metricGroups[name].schema)
		if err != nil {
			return fmt.Errorf("failed to create record for file %q: %w", p.metricGroups[name].filename, err)
		}
		if err = p.metricGroups[name].writer.WriteBuffered(record); err != nil {
			return fmt.Errorf("failed to write to file %q: %w", p.metricGroups[name].filename, err)
		}
		record.Release()
	}

	return nil
}

func (p *Parquet) rotateIfNeeded(name string) error {
	fileInfo, err := os.Stat(p.metricGroups[name].filename)
	if err != nil {
		return fmt.Errorf("failed to stat file %q: %w", p.metricGroups[name].filename, err)
	}

	expireTime := fileInfo.ModTime().Add(time.Duration(p.RotationInterval))
	if time.Now().Before(expireTime) {
		return nil
	}

	if err := p.metricGroups[name].writer.Close(); err != nil {
		return fmt.Errorf("failed to close file for rotation %q: %w", p.metricGroups[name].filename, err)
	}

	writer, err := p.createWriter(name, p.metricGroups[name].filename, p.metricGroups[name].schema)
	if err != nil {
		return fmt.Errorf("failed to create new writer for file %q: %w", p.metricGroups[name].filename, err)
	}
	p.metricGroups[name].writer = writer

	return nil
}

func (p *Parquet) createRecord(metrics []telegraf.Metric, builder *array.RecordBuilder, schema *arrow.Schema) (arrow.Record, error) {
	for index, col := range schema.Fields() {
		for _, m := range metrics {
			if p.TimestampFieldName != "" && col.Name == p.TimestampFieldName {
				builder.Field(index).(*array.Int64Builder).Append(m.Time().UnixNano())
				continue
			}

			// Try to get the value from a field first, then from a tag.
			var value any
			var ok bool
			value, ok = m.GetField(col.Name)
			if !ok {
				value, ok = m.GetTag(col.Name)
			}

			// if neither field nor tag exists, append a null value
			if !ok {
				switch col.Type {
				case arrow.PrimitiveTypes.Int8:
					builder.Field(index).(*array.Int8Builder).AppendNull()
				case arrow.PrimitiveTypes.Int16:
					builder.Field(index).(*array.Int16Builder).AppendNull()
				case arrow.PrimitiveTypes.Int32:
					builder.Field(index).(*array.Int32Builder).AppendNull()
				case arrow.PrimitiveTypes.Int64:
					builder.Field(index).(*array.Int64Builder).AppendNull()
				case arrow.PrimitiveTypes.Uint8:
					builder.Field(index).(*array.Uint8Builder).AppendNull()
				case arrow.PrimitiveTypes.Uint16:
					builder.Field(index).(*array.Uint16Builder).AppendNull()
				case arrow.PrimitiveTypes.Uint32:
					builder.Field(index).(*array.Uint32Builder).AppendNull()
				case arrow.PrimitiveTypes.Uint64:
					builder.Field(index).(*array.Uint64Builder).AppendNull()
				case arrow.PrimitiveTypes.Float32:
					builder.Field(index).(*array.Float32Builder).AppendNull()
				case arrow.PrimitiveTypes.Float64:
					builder.Field(index).(*array.Float64Builder).AppendNull()
				case arrow.BinaryTypes.String:
					builder.Field(index).(*array.StringBuilder).AppendNull()
				case arrow.FixedWidthTypes.Boolean:
					builder.Field(index).(*array.BooleanBuilder).AppendNull()
				default:
					return nil, fmt.Errorf("unsupported type: %T", value)
				}

				continue
			}

			switch col.Type {
			case arrow.PrimitiveTypes.Int8:
				builder.Field(index).(*array.Int8Builder).Append(value.(int8))
			case arrow.PrimitiveTypes.Int16:
				builder.Field(index).(*array.Int16Builder).Append(value.(int16))
			case arrow.PrimitiveTypes.Int32:
				builder.Field(index).(*array.Int32Builder).Append(value.(int32))
			case arrow.PrimitiveTypes.Int64:
				builder.Field(index).(*array.Int64Builder).Append(value.(int64))
			case arrow.PrimitiveTypes.Uint8:
				builder.Field(index).(*array.Uint8Builder).Append(value.(uint8))
			case arrow.PrimitiveTypes.Uint16:
				builder.Field(index).(*array.Uint16Builder).Append(value.(uint16))
			case arrow.PrimitiveTypes.Uint32:
				builder.Field(index).(*array.Uint32Builder).Append(value.(uint32))
			case arrow.PrimitiveTypes.Uint64:
				builder.Field(index).(*array.Uint64Builder).Append(value.(uint64))
			case arrow.PrimitiveTypes.Float32:
				builder.Field(index).(*array.Float32Builder).Append(value.(float32))
			case arrow.PrimitiveTypes.Float64:
				builder.Field(index).(*array.Float64Builder).Append(value.(float64))
			case arrow.BinaryTypes.String:
				builder.Field(index).(*array.StringBuilder).Append(value.(string))
			case arrow.FixedWidthTypes.Boolean:
				builder.Field(index).(*array.BooleanBuilder).Append(value.(bool))
			default:
				return nil, fmt.Errorf("unsupported type: %T", value)
			}
		}
	}

	record := builder.NewRecord()
	return record, nil
}

func (p *Parquet) createSchema(metrics []telegraf.Metric) (*arrow.Schema, error) {
	rawFields := make(map[string]arrow.DataType, 0)
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			if _, ok := rawFields[field.Key]; !ok {
				arrowType, err := goToArrowType(field.Value)
				if err != nil {
					return nil, fmt.Errorf("error converting '%s=%s' field to arrow type: %w", field.Key, field.Value, err)
				}
				rawFields[field.Key] = arrowType
			}
		}
		for _, tag := range metric.TagList() {
			if _, ok := rawFields[tag.Key]; !ok {
				rawFields[tag.Key] = arrow.BinaryTypes.String
			}
		}
	}

	fields := make([]arrow.Field, 0)
	for key, value := range rawFields {
		fields = append(fields, arrow.Field{
			Name: key,
			Type: value,
		})
	}

	if p.TimestampFieldName != "" {
		fields = append(fields, arrow.Field{
			Name: p.TimestampFieldName,
			Type: arrow.PrimitiveTypes.Int64,
		})
	}

	return arrow.NewSchema(fields, nil), nil
}

func (p *Parquet) createWriter(name, filename string, schema *arrow.Schema) (*pqarrow.FileWriter, error) {
	if _, err := os.Stat(filename); err == nil {
		now := time.Now()
		rotatedFilename := fmt.Sprintf("%s/%s-%s-%s.parquet", p.Directory, name, now.Format("2006-01-02"), strconv.FormatInt(now.Unix(), 10))
		if err := os.Rename(filename, rotatedFilename); err != nil {
			return nil, fmt.Errorf("failed to rename file %q: %w", filename, err)
		}
	}
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q: %w", filename, err)
	}

	writer, err := pqarrow.NewFileWriter(schema, file, parquet.NewWriterProperties(), pqarrow.DefaultWriterProps())
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet writer for file %q: %w", filename, err)
	}

	return writer, nil
}

func goToArrowType(value interface{}) (arrow.DataType, error) {
	switch value.(type) {
	case int8:
		return arrow.PrimitiveTypes.Int8, nil
	case int16:
		return arrow.PrimitiveTypes.Int16, nil
	case int32:
		return arrow.PrimitiveTypes.Int32, nil
	case int64, int:
		return arrow.PrimitiveTypes.Int64, nil
	case uint8:
		return arrow.PrimitiveTypes.Uint8, nil
	case uint16:
		return arrow.PrimitiveTypes.Uint16, nil
	case uint32:
		return arrow.PrimitiveTypes.Uint32, nil
	case uint64, uint:
		return arrow.PrimitiveTypes.Uint64, nil
	case float32:
		return arrow.PrimitiveTypes.Float32, nil
	case float64:
		return arrow.PrimitiveTypes.Float64, nil
	case string:
		return arrow.BinaryTypes.String, nil
	case bool:
		return arrow.FixedWidthTypes.Boolean, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

func init() {
	outputs.Add("parquet", func() telegraf.Output {
		return &Parquet{
			TimestampFieldName: defaultTimestampFieldName,
		}
	})
}
