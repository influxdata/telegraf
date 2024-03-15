package parquet

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	MeasurementColumn string   `toml:"measurement_column"`
	TagColumns        []string `toml:"tag_columns"`
	TimestampColumn   string   `toml:"timestamp_column"`
	TimestampFormat   string   `toml:"timestamp_format"`
	TimestampTimezone string   `toml:"timestamp_timezone"`

	defaultTags map[string]string
	location    *time.Location
	metricName  string
}

func (p *Parser) Init() error {
	if p.TimestampFormat == "" {
		p.TimestampFormat = "unix"
	}
	if p.TimestampTimezone == "" {
		p.location = time.UTC
	} else {
		loc, err := time.LoadLocation(p.TimestampTimezone)
		if err != nil {
			return fmt.Errorf("invalid location %s: %w", p.TimestampTimezone, err)
		}
		p.location = loc
	}

	return nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	reader := bytes.NewReader(buf)
	parquetReader, err := file.NewParquetReader(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to create parquet reader: %w", err)
	}
	metadata := parquetReader.MetaData()
	selectedColumns := make([]int, 0)
	for i := 0; i < metadata.Schema.NumColumns(); i++ {
		selectedColumns = append(selectedColumns, i)
	}

	columns := make([]string, len(selectedColumns))
	data := make(map[int][]any, metadata.Schema.NumColumns())
	for i := 0; i < parquetReader.NumRowGroups(); i++ {
		rowGroup := parquetReader.RowGroup(i)
		scanners := make([]*ColumnParser, len(selectedColumns))
		for idx, j := range selectedColumns {
			col, err := rowGroup.Column(j)
			if err != nil {
				return nil, fmt.Errorf("unable to fetch column %q: %w", j, err)
			}

			scanners[idx] = newColumnParser(col)
			columns[idx] = col.Descriptor().Name()
		}

		for i, s := range scanners {
			for s.HasNext() {
				if val, ok := s.Next(); ok {
					data[i] = append(data[i], val)
				}
			}
		}
	}

	metrics := make([]telegraf.Metric, len(data[0]))
	for colIndex, col := range data {
		for i, val := range col {
			if val == nil {
				continue
			}

			if metrics[i] == nil {
				metrics[i] = metric.New(p.metricName, p.defaultTags, nil, time.Now())
			}

			if p.MeasurementColumn != "" && columns[colIndex] == p.MeasurementColumn {
				metrics[i].SetName(fmt.Sprintf("%v", val))
			} else if p.TagColumns != nil && slices.Contains(p.TagColumns, columns[colIndex]) {
				metrics[i].AddTag(columns[colIndex], fmt.Sprintf("%v", val))
			} else if p.TimestampColumn != "" && columns[colIndex] == p.TimestampColumn {
				rawTime := fmt.Sprintf("%v", val)
				timestamp, err := internal.ParseTimestamp(p.TimestampFormat, rawTime, p.location)
				if err != nil {
					return nil, fmt.Errorf("could not parse '%s' to '%s'", rawTime, p.TimestampFormat)
				}
				metrics[i].SetTime(timestamp)
			} else {
				metrics[i].AddField(columns[colIndex], val)
			}
		}
	}

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) != 1 {
		return nil, errors.New("line contains multiple metrics")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.defaultTags = tags
}

func init() {
	parsers.Add("parquet",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{metricName: defaultMetricName}
		},
	)
}
