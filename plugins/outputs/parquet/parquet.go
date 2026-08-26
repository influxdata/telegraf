//go:generate ../../../tools/readme_config_includer/generator
package parquet

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"
	"unicode"

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

const defaultMaxColumns = 1000

type metricGroup struct {
	name     string
	filename string
	created  time.Time
	columns  map[string]arrow.DataType
	limited  bool
	warned   map[string]bool
	builder  *array.RecordBuilder
	schema   *arrow.Schema
	writer   *pqarrow.FileWriter
}

type Parquet struct {
	Directory          string          `toml:"directory"`
	RotationInterval   config.Duration `toml:"rotation_interval"`
	TimestampFieldName string          `toml:"timestamp_field_name"`
	MaxColumns         int             `toml:"max_columns"`
	Log                telegraf.Logger `toml:"-"`

	metricGroups     map[string]*metricGroup
	warnedOnFilename bool
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

	for _, group := range p.metricGroups {
		if err := group.close(); err != nil {
			p.Log.Errorf("Failed to close file %q: %v", group.filename, err)
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
		name := p.metricToFile(metric.Name())
		groupedMetrics[name] = append(groupedMetrics[name], metric)
	}

	for name, metrics := range groupedMetrics {
		group, err := p.groupFor(name, metrics)
		if err != nil {
			return err
		}

		record := p.createRecordBatch(group, metrics)
		err = group.writer.WriteBuffered(record)
		record.Release()
		if err != nil {
			return fmt.Errorf("failed to write to file %q: %w", group.filename, err)
		}
	}

	return nil
}

func (p *Parquet) groupFor(name string, metrics []telegraf.Metric) (*metricGroup, error) {
	columns := p.collectColumns(metrics)

	group, found := p.metricGroups[name]
	if !found {
		group = &metricGroup{
			name:    name,
			columns: make(map[string]arrow.DataType, len(columns)),
			warned:  make(map[string]bool),
		}
	}

	added, dropped := group.addColumns(columns, p.MaxColumns)
	if len(dropped) > 0 && !group.limited {
		group.limited = true
		p.Log.Warnf(
			"Dropping column(s) %s from %q, which is at the %d column limit; raise 'max_columns' to keep them",
			strings.Join(dropped, ", "), name, p.MaxColumns,
		)
	}

	if !found {
		if err := p.openFile(group); err != nil {
			return nil, err
		}
		p.metricGroups[name] = group

		return group, nil
	}

	if len(added) == 0 {
		return group, p.rotateIfNeeded(group)
	}

	p.Log.Infof("Starting a new file for %q to add column(s) %s", name, strings.Join(added, ", "))
	if err := p.reopen(group); err != nil {
		return nil, err
	}

	return group, nil
}

func (g *metricGroup) addColumns(columns map[string]arrow.DataType, limit int) (added, dropped []string) {
	candidates := make([]string, 0, len(columns))
	for column := range columns {
		if _, known := g.columns[column]; !known {
			candidates = append(candidates, column)
		}
	}
	slices.Sort(candidates)

	room := len(candidates)
	if limit > 0 && len(g.columns)+room > limit {
		room = limit - len(g.columns)
		if room < 0 {
			room = 0
		}
	}

	for _, column := range candidates[:room] {
		g.columns[column] = columns[column]
	}

	return candidates[:room], candidates[room:]
}

const maxMeasurementLen = 255 - len("-2006-01-02-1234567890-999.parquet")

func (p *Parquet) metricToFile(name string) string {
	safe := strings.Map(func(r rune) rune {
		if reservedInFilename(r) {
			return '_'
		}
		return r
	}, name)

	if len(safe) > maxMeasurementLen {
		safe = strings.ToValidUTF8(safe[:maxMeasurementLen], "")
	}

	if safe != name && !p.warnedOnFilename {
		p.warnedOnFilename = true
		p.Log.Warnf("Metric %q is not usable as a file name, writing to %q instead; use the rename processor to choose the name", name, safe)
	}

	return safe
}

func reservedInFilename(r rune) bool {
	if r == '/' || r == '\\' || unicode.IsControl(r) {
		return true
	}

	return runtime.GOOS == "windows" && strings.ContainsRune(`<>:"|?*`, r)
}

func (p *Parquet) rotateIfNeeded(group *metricGroup) error {
	if p.RotationInterval == 0 || time.Since(group.created) < time.Duration(p.RotationInterval) {
		return nil
	}

	return p.reopen(group)
}

func (p *Parquet) reopen(group *metricGroup) error {
	if err := group.close(); err != nil {
		delete(p.metricGroups, group.name)
		return fmt.Errorf("failed to close file %q: %w", group.filename, err)
	}

	if err := p.openFile(group); err != nil {
		delete(p.metricGroups, group.name)
		return err
	}

	return nil
}

func (g *metricGroup) close() error {
	err := g.writer.Close()
	g.builder.Release()

	return err
}

func (p *Parquet) openFile(group *metricGroup) error {
	group.schema = p.createSchema(group.columns)
	group.builder = array.NewRecordBuilder(memory.DefaultAllocator, group.schema)
	group.filename = p.unusedFilename(group.name)
	group.created = time.Now()

	f, err := os.Create(group.filename)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", group.filename, err)
	}

	writer, err := pqarrow.NewFileWriter(group.schema, f, parquet.NewWriterProperties(), pqarrow.DefaultWriterProps())
	if err != nil {
		f.Close()
		return fmt.Errorf("failed to create parquet writer for file %q: %w", group.filename, err)
	}
	group.writer = writer

	return nil
}

func (p *Parquet) unusedFilename(name string) string {
	now := time.Now()
	prefix := filepath.Join(p.Directory, fmt.Sprintf("%s-%s-%d", name, now.Format("2006-01-02"), now.Unix()))

	filename := prefix + ".parquet"
	for suffix := 1; ; suffix++ {
		if _, err := os.Stat(filename); err != nil {
			return filename
		}
		filename = fmt.Sprintf("%s-%d.parquet", prefix, suffix)
	}
}

func (p *Parquet) createRecordBatch(group *metricGroup, metrics []telegraf.Metric) arrow.RecordBatch {
	for index, column := range group.schema.Fields() {
		builder := group.builder.Field(index)

		for _, m := range metrics {
			value := p.valueFor(m, column.Name)
			if !appendValue(builder, value) {
				p.warnOncef(
					group, column.Name,
					"Writing null for column %q of file %q as a %T value does not fit its %s column",
					column.Name, group.filename, value, column.Type,
				)
			}
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

func (p *Parquet) warnOncef(group *metricGroup, column, format string, args ...interface{}) {
	if group.warned[column] {
		return
	}
	group.warned[column] = true
	p.Log.Warnf(format, args...)
}

func appendValue(builder array.Builder, value interface{}) bool {
	switch v := value.(type) {
	case nil:
		builder.AppendNull()
		return true
	case int8:
		return appendTyped(builder, v)
	case int16:
		return appendTyped(builder, v)
	case int32:
		return appendTyped(builder, v)
	case int64:
		return appendTyped(builder, v)
	case int:
		return appendTyped(builder, int64(v))
	case uint8:
		return appendTyped(builder, v)
	case uint16:
		return appendTyped(builder, v)
	case uint32:
		return appendTyped(builder, v)
	case uint64:
		return appendTyped(builder, v)
	case uint:
		return appendTyped(builder, uint64(v))
	case float32:
		return appendTyped(builder, v)
	case float64:
		return appendTyped(builder, v)
	case string:
		return appendTyped(builder, v)
	case bool:
		return appendTyped(builder, v)
	default:
		builder.AppendNull()
		return false
	}
}

func appendTyped[T any](builder array.Builder, value T) bool {
	column, ok := builder.(interface{ Append(T) })
	if !ok {
		builder.AppendNull()
		return false
	}
	column.Append(value)

	return true
}

func (p *Parquet) collectColumns(metrics []telegraf.Metric) map[string]arrow.DataType {
	columns := make(map[string]arrow.DataType)
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			if _, known := columns[field.Key]; known {
				continue
			}
			datatype, err := goToArrowType(field.Value)
			if err != nil {
				p.Log.Warnf("Skipping field %q of metric %q: %v", field.Key, metric.Name(), err)
				continue
			}
			columns[field.Key] = datatype
		}
	}

	for _, metric := range metrics {
		for _, tag := range metric.TagList() {
			if _, known := columns[tag.Key]; !known {
				columns[tag.Key] = arrow.BinaryTypes.String
			}
		}
	}

	if p.TimestampFieldName != "" {
		if _, taken := columns[p.TimestampFieldName]; taken {
			delete(columns, p.TimestampFieldName)
			p.Log.Warnf(
				"Ignoring the %q field or tag as that column holds the metric time; "+
					"set 'timestamp_field_name' to another name to keep it",
				p.TimestampFieldName,
			)
		}
	}

	return columns
}

func (p *Parquet) createSchema(columns map[string]arrow.DataType) *arrow.Schema {
	names := make([]string, 0, len(columns))
	for name := range columns {
		names = append(names, name)
	}
	slices.Sort(names)

	fields := make([]arrow.Field, 0, len(names)+1)
	for _, name := range names {
		fields = append(fields, arrow.Field{
			Name:     name,
			Type:     columns[name],
			Nullable: true,
		})
	}

	if p.TimestampFieldName != "" {
		fields = append(fields, arrow.Field{
			Name: p.TimestampFieldName,
			Type: arrow.PrimitiveTypes.Int64,
		})
	}

	return arrow.NewSchema(fields, nil)
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
			MaxColumns:         defaultMaxColumns,
		}
	})
}
