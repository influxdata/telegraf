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
	"github.com/influxdata/telegraf/internal"
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

	root         *os.Root
	metricGroups map[string]*metricGroup
}

func (*Parquet) SampleConfig() string {
	return sampleConfig
}

func (p *Parquet) Init() error {
	if p.Directory == "" {
		p.Directory = "."
	}

	if err := os.MkdirAll(p.Directory, 0750); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", p.Directory, err)
	}

	root, err := os.OpenRoot(p.Directory)
	if err != nil {
		return fmt.Errorf("failed to open directory %q: %w", p.Directory, err)
	}

	p.root = root
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

	if err := p.root.Close(); err != nil {
		p.Log.Errorf("failed to close directory %q: %v", p.Directory, err)
		errorOccurred = true
	}

	if errorOccurred {
		return errors.New("failed closing one or more parquet files")
	}

	return nil
}

func (p *Parquet) Write(metrics []telegraf.Metric) error {
	grouped := make(map[string][]int)
	for i, m := range metrics {
		grouped[m.Name()] = append(grouped[m.Name()], i)
	}

	var writeErr internal.PartialWriteError

	now := time.Now()
	for name, indices := range grouped {
		batch := make([]telegraf.Metric, len(indices))
		for j, i := range indices {
			batch[j] = metrics[i]
		}

		group, err := p.groupFor(name, batch, now)
		if err != nil {
			writeErr.MetricsReject = append(writeErr.MetricsReject, indices...)
			writeErr.MetricsRejectErrors = append(writeErr.MetricsRejectErrors, err)
			continue
		}

		if p.RotationInterval != 0 {
			if err := p.rotateIfNeeded(name); err != nil {
				return fmt.Errorf("failed to rotate file %q: %w", group.filename, err)
			}
		}

		record, err := p.createRecordBatch(batch, group.builder, group.schema)
		if err != nil {
			return fmt.Errorf("failed to create record for file %q: %w", group.filename, err)
		}
		err = group.writer.WriteBuffered(record)
		record.Release()
		if err != nil {
			return fmt.Errorf("failed to write to file %q: %w", group.filename, err)
		}
		writeErr.MetricsAccept = append(writeErr.MetricsAccept, indices...)
	}

	if len(writeErr.MetricsReject) == 0 {
		return nil
	}
	writeErr.Err = fmt.Errorf("rejected %d metric(s): %w", len(writeErr.MetricsReject), errors.Join(writeErr.MetricsRejectErrors...))

	return &writeErr
}

func (p *Parquet) groupFor(name string, metrics []telegraf.Metric, now time.Time) (*metricGroup, error) {
	if group, found := p.metricGroups[name]; found {
		return group, nil
	}

	schema, err := p.createSchema(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema for file %q: %w", name, err)
	}

	filename := parquetFilename(name, now)
	writer, err := p.createWriter(name, filename, schema)
	if err != nil {
		return nil, err
	}

	group := &metricGroup{
		builder:  array.NewRecordBuilder(memory.DefaultAllocator, schema),
		filename: filename,
		schema:   schema,
		writer:   writer,
	}
	p.metricGroups[name] = group

	return group, nil
}

func parquetFilename(name string, at time.Time) string {
	return fmt.Sprintf("%s-%s-%s.parquet", name, at.Format("2006-01-02"), strconv.FormatInt(at.Unix(), 10))
}

func (p *Parquet) rotateIfNeeded(name string) error {
	fileInfo, err := p.root.Stat(p.metricGroups[name].filename)
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

func (p *Parquet) createRecordBatch(metrics []telegraf.Metric, builder *array.RecordBuilder, schema *arrow.Schema) (arrow.RecordBatch, error) {
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

	record := builder.NewRecordBatch()
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
			Name:     key,
			Type:     value,
			Nullable: true,
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
	if _, err := p.root.Stat(filename); err == nil {
		if err := p.root.Rename(filename, parquetFilename(name, time.Now())); err != nil {
			return nil, fmt.Errorf("failed to rename file %q: %w", filename, err)
		}
	}

	file, err := p.root.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q: %w", filename, err)
	}

	writer, err := pqarrow.NewFileWriter(schema, file, parquet.NewWriterProperties(), pqarrow.DefaultWriterProps())
	if err != nil {
		file.Close()
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
