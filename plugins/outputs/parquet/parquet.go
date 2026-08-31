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
	grouped := make(map[string][]int)
	for i, m := range metrics {
		grouped[m.Name()] = append(grouped[m.Name()], i)
	}

	var writeErr internal.PartialWriteError

	now := time.Now()
	for name, indices := range grouped {
		group, found := p.metricGroups[name]
		if !found {
			batch := make([]telegraf.Metric, 0, len(indices))
			for _, i := range indices {
				batch = append(batch, metrics[i])
			}

			filename := fmt.Sprintf("%s/%s-%s-%s.parquet", p.Directory, name, now.Format("2006-01-02"), strconv.FormatInt(now.Unix(), 10))
			schema, err := p.createSchema(batch)
			if err != nil {
				return fmt.Errorf("failed to create schema for file %q: %w", name, err)
			}
			writer, err := p.createWriter(name, filename, schema)
			if err != nil {
				return fmt.Errorf("failed to create writer for file %q: %w", name, err)
			}

			group = &metricGroup{
				builder:  array.NewRecordBuilder(memory.DefaultAllocator, schema),
				filename: filename,
				schema:   schema,
				writer:   writer,
			}
			p.metricGroups[name] = group
		}

		if p.RotationInterval != 0 {
			if err := p.rotateIfNeeded(name); err != nil {
				return fmt.Errorf("failed to rotate file %q: %w", group.filename, err)
			}
		}

		accepted := make([]telegraf.Metric, 0, len(indices))
		for _, i := range indices {
			if err := p.checkSchema(group, metrics[i]); err != nil {
				writeErr.MetricsReject = append(writeErr.MetricsReject, i)
				writeErr.MetricsRejectErrors = append(writeErr.MetricsRejectErrors, err)
				continue
			}
			accepted = append(accepted, metrics[i])
			writeErr.MetricsAccept = append(writeErr.MetricsAccept, i)
		}

		record := p.createRecordBatch(group, accepted)
		err := group.writer.WriteBuffered(record)
		record.Release()
		if err != nil {
			return fmt.Errorf("failed to write to file %q: %w", group.filename, err)
		}
	}

	if len(writeErr.MetricsReject) == 0 {
		return nil
	}
	writeErr.Err = fmt.Errorf("rejected %d metric(s): %w", len(writeErr.MetricsReject), errors.Join(writeErr.MetricsRejectErrors...))

	return &writeErr
}

func (p *Parquet) checkSchema(group *metricGroup, m telegraf.Metric) error {
	for _, column := range group.schema.Fields() {
		value := p.valueFor(m, column.Name)
		if value == nil {
			continue
		}

		datatype, err := goToArrowType(value)
		if err != nil {
			return fmt.Errorf("column %q of file %q: %w", column.Name, group.filename, err)
		}
		if !arrow.TypeEqual(datatype, column.Type) {
			return fmt.Errorf("column %q of file %q holds %s but the metric has a %s value", column.Name, group.filename, column.Type, datatype)
		}
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

func (p *Parquet) createRecordBatch(group *metricGroup, metrics []telegraf.Metric) arrow.RecordBatch {
	for index, column := range group.schema.Fields() {
		builder := group.builder.Field(index)

		for _, m := range metrics {
			appendValue(builder, p.valueFor(m, column.Name))
		}
	}

	return group.builder.NewRecordBatch()
}

func (p *Parquet) valueFor(m telegraf.Metric, column string) interface{} {
	if p.TimestampFieldName != "" && column == p.TimestampFieldName {
		return m.Time().UnixNano()
	}
	if value, found := m.GetField(column); found {
		return value
	}
	if value, found := m.GetTag(column); found {
		return value
	}

	return nil
}

func appendValue(builder array.Builder, value interface{}) {
	switch v := value.(type) {
	case nil:
		builder.AppendNull()
	case int8:
		appendTyped(builder, v)
	case int16:
		appendTyped(builder, v)
	case int32:
		appendTyped(builder, v)
	case int64:
		appendTyped(builder, v)
	case int:
		appendTyped(builder, int64(v))
	case uint8:
		appendTyped(builder, v)
	case uint16:
		appendTyped(builder, v)
	case uint32:
		appendTyped(builder, v)
	case uint64:
		appendTyped(builder, v)
	case uint:
		appendTyped(builder, uint64(v))
	case float32:
		appendTyped(builder, v)
	case float64:
		appendTyped(builder, v)
	case string:
		appendTyped(builder, v)
	case bool:
		appendTyped(builder, v)
	default:
		builder.AppendNull()
	}
}

func appendTyped[T any](builder array.Builder, value T) {
	column, ok := builder.(interface{ Append(T) })
	if !ok {
		builder.AppendNull()
		return
	}
	column.Append(value)
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
		if p.TimestampFieldName != "" && key == p.TimestampFieldName {
			p.Log.Warnf("Ignoring the %q field or tag as that column holds the metric time; "+
				"set 'timestamp_field_name' to another name to keep it", key)
			continue
		}
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
