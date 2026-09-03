//go:generate ../../../tools/readme_config_includer/generator
package parquet

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	root         *os.Root
}

func (*Parquet) SampleConfig() string {
	return sampleConfig
}

func (p *Parquet) Init() error {
	if p.Directory == "" {
		p.Directory = "."
	}

	// Get the absolute path of the directory
	d, err := filepath.Abs(p.Directory)
	if err != nil {
		return fmt.Errorf("getting absolute path for directory %q failed: %w", p.Directory, err)
	}
	p.Directory = d

	// Create the directory if it doesn't exist and check it actually is a directory
	stat, err := os.Stat(p.Directory)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to stat directory %q: %w", p.Directory, err)
		}
		if err := os.MkdirAll(p.Directory, 0750); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", p.Directory, err)
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("provided directory %q is not a directory", p.Directory)
	}

	// Make sure we cannot leave the given directory
	root, err := os.OpenRoot(p.Directory)
	if err != nil {
		return fmt.Errorf("opening directory %q as root failed: %w", p.Directory, err)
	}
	p.root = root

	p.metricGroups = make(map[string]*metricGroup)

	return nil
}

func (*Parquet) Connect() error {
	return nil
}

func (p *Parquet) Close() error {
	if p.root != nil {
		p.root.Close()
	}

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
	metricIndices := make(map[string][]int)
	for i, metric := range metrics {
		name := metric.Name()
		groupedMetrics[name] = append(groupedMetrics[name], metric)
		metricIndices[name] = append(metricIndices[name], i)
	}

	var perr internal.PartialWriteError
	now := time.Now()
	for name, metrics := range groupedMetrics {
		if _, ok := p.metricGroups[name]; !ok {
			filename := fmt.Sprintf("%s-%s-%s.parquet", name, now.Format("2006-01-02"), strconv.FormatInt(now.Unix(), 10))
			schema, err := p.createSchema(metrics)
			if err != nil {
				perr.MetricsReject = append(perr.MetricsReject, metricIndices[name]...)
				perr.MetricsRejectErrors = append(perr.MetricsRejectErrors, fmt.Errorf("failed to create schema for file %q: %w", name, err))
				perr.Err = fmt.Errorf("failed to create schema for file %q: %w", name, err)
				continue
			}
			writer, err := p.createWriter(name, filename, schema)
			if err != nil {
				perr.MetricsReject = append(perr.MetricsReject, metricIndices[name]...)
				perr.MetricsRejectErrors = append(perr.MetricsRejectErrors, fmt.Errorf("failed to create writer for file %q: %w", name, err))
				perr.Err = fmt.Errorf("failed to create writer for file %q: %w", name, err)
				continue
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
				perr.MetricsReject = append(perr.MetricsReject, metricIndices[name]...)
				perr.MetricsRejectErrors = append(perr.MetricsRejectErrors, fmt.Errorf("failed to rotate file %q: %w", p.metricGroups[name].filename, err))
				perr.Err = fmt.Errorf("failed to rotate file %q: %w", p.metricGroups[name].filename, err)
				continue
			}
		}

		record, err := p.createRecordBatch(metrics, p.metricGroups[name].builder, p.metricGroups[name].schema)
		accepted := metricIndices[name]
		if err != nil {
			var crbErr *internal.PartialWriteError
			if errors.As(err, &crbErr) {
				accepted = make([]int, 0, len(perr.MetricsAccept))
				for _, idx := range crbErr.MetricsAccept {
					accepted = append(accepted, metricIndices[name][idx])
				}
				for _, idx := range crbErr.MetricsReject {
					perr.MetricsReject = append(perr.MetricsReject, metricIndices[name][idx])
					perr.MetricsRejectErrors = append(perr.MetricsRejectErrors, crbErr.MetricsRejectErrors...)
				}
			} else {
				accepted = make([]int, 0)
				perr.MetricsReject = append(perr.MetricsReject, metricIndices[name]...)
				perr.MetricsRejectErrors = append(perr.MetricsRejectErrors, fmt.Errorf("failed to create record for file %q: %w", name, err))
			}
			perr.Err = fmt.Errorf("failed to create record for file %q: %w", p.metricGroups[name].filename, err)
		}
		if len(accepted) == 0 || record == nil {
			continue
		}
		if err := p.metricGroups[name].writer.WriteBuffered(record); err != nil {
			perr.Err = fmt.Errorf("failed to write to file %q: %w", p.metricGroups[name].filename, err)
			return &perr
		}
		perr.MetricsAccept = append(perr.MetricsAccept, accepted...)
		record.Release()
	}
	if perr.Err != nil {
		return &perr
	}

	return nil
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
	// Create a lookup table for the schema columns
	columnTypes := make(map[string]arrow.DataType, len(schema.Fields()))
	for _, col := range schema.Fields() {
		switch col.Type {
		case arrow.PrimitiveTypes.Int8,
			arrow.PrimitiveTypes.Int16,
			arrow.PrimitiveTypes.Int32,
			arrow.PrimitiveTypes.Int64,
			arrow.PrimitiveTypes.Uint8,
			arrow.PrimitiveTypes.Uint16,
			arrow.PrimitiveTypes.Uint32,
			arrow.PrimitiveTypes.Uint64,
			arrow.PrimitiveTypes.Float32,
			arrow.PrimitiveTypes.Float64,
			arrow.BinaryTypes.String,
			arrow.FixedWidthTypes.Boolean:
		default:
			// Return the error directly as all metrics will fail here
			return nil, fmt.Errorf("unsupported column type %q for column %q", col.Type.Name(), col.Name)
		}
		columnTypes[col.Name] = col.Type
	}

	var perr internal.PartialWriteError
	for i, m := range metrics {
		// Collect the row values in order of reverse precedence to add them to
		// the columns. We need to do this intermediate step to remove metric
		// rows with invalid types.
		row := make(map[string]any, len(columnTypes))
		for _, tag := range m.TagList() {
			if _, found := columnTypes[tag.Key]; !found {
				continue
			}
			row[tag.Key] = tag.Value
		}
		for _, field := range m.FieldList() {
			if _, found := columnTypes[field.Key]; !found {
				continue
			}
			row[field.Key] = field.Value
		}

		if p.TimestampFieldName != "" {
			row[p.TimestampFieldName] = m.Time().UnixNano()
		}

		// Check if the row values do match the corresponding schema column type
		var err error
		for name, value := range row {
			ctype := columnTypes[name]

			// Check if the value matches the schema type
			switch value.(type) {
			case int8:
				if ctype != arrow.PrimitiveTypes.Int8 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case int16:
				if ctype != arrow.PrimitiveTypes.Int16 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case int32:
				if ctype != arrow.PrimitiveTypes.Int32 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case int64:
				if ctype != arrow.PrimitiveTypes.Int64 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case uint8:
				if ctype != arrow.PrimitiveTypes.Uint8 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case uint16:
				if ctype != arrow.PrimitiveTypes.Uint16 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case uint32:
				if ctype != arrow.PrimitiveTypes.Uint32 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case uint64:
				if ctype != arrow.PrimitiveTypes.Uint64 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case float32:
				if ctype != arrow.PrimitiveTypes.Float32 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case float64:
				if ctype != arrow.PrimitiveTypes.Float64 {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case string:
				if ctype != arrow.BinaryTypes.String {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			case bool:
				if ctype != arrow.FixedWidthTypes.Boolean {
					err = fmt.Errorf("invalid value %v (%T) for column %q (%s)", value, value, name, ctype.Name())
				}
			default:
				err = fmt.Errorf("unknown field value type %T for column %q", value, name)
			}
			if err != nil {
				break
			}
		}
		if err != nil {
			perr.Err = err
			perr.MetricsRejectErrors = append(perr.MetricsRejectErrors, err)
			perr.MetricsReject = append(perr.MetricsReject, i)
			continue
		}

		for index, col := range schema.Fields() {
			value, found := row[col.Name]

			// If no value exists for the column we append a null value
			if !found {
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
			}
		}
		perr.MetricsAccept = append(perr.MetricsAccept, i)
	}

	record := builder.NewRecordBatch()
	if len(perr.MetricsReject) > 0 {
		return record, &perr
	}
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
	if _, err := p.root.Stat(filename); err == nil {
		now := time.Now()
		rotatedFilename := fmt.Sprintf("%s-%s-%s.parquet", name, now.Format("2006-01-02"), strconv.FormatInt(now.Unix(), 10))
		if err := p.root.Rename(filename, rotatedFilename); err != nil {
			return nil, fmt.Errorf("failed to rename file %q: %w", filename, err)
		}
	}
	file, err := p.root.Create(filename)
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
