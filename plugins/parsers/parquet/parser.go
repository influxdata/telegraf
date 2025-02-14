package parquet

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/apache/arrow-go/v18/parquet/file"

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

	now := time.Now()
	metrics := make([]telegraf.Metric, 0, metadata.NumRows)
	for i := 0; i < parquetReader.NumRowGroups(); i++ {
		rowGroup := parquetReader.RowGroup(i)
		scanners := make([]*columnParser, metadata.Schema.NumColumns())
		for colIndex := range metadata.Schema.NumColumns() {
			col, err := rowGroup.Column(colIndex)
			if err != nil {
				return nil, fmt.Errorf("unable to fetch column %q: %w", colIndex, err)
			}

			scanners[colIndex] = newColumnParser(col)
		}

		rowIndex := 0
		rowGroupMetrics := make([]telegraf.Metric, rowGroup.NumRows())
		for _, s := range scanners {
			for s.HasNext() {
				if rowIndex%int(rowGroup.NumRows()) == 0 {
					rowIndex = 0
				}

				val, ok := s.Next()
				if !ok || val == nil {
					rowIndex++
					continue
				}

				if rowGroupMetrics[rowIndex] == nil {
					rowGroupMetrics[rowIndex] = metric.New(p.metricName, p.defaultTags, nil, now)
				}

				if p.MeasurementColumn != "" && s.name == p.MeasurementColumn {
					valStr, err := internal.ToString(val)
					if err != nil {
						return nil, fmt.Errorf("could not convert value to string: %w", err)
					}
					rowGroupMetrics[rowIndex].SetName(valStr)
				} else if p.TagColumns != nil && slices.Contains(p.TagColumns, s.name) {
					valStr, err := internal.ToString(val)
					if err != nil {
						return nil, fmt.Errorf("could not convert value to string: %w", err)
					}
					rowGroupMetrics[rowIndex].AddTag(s.name, valStr)
				} else if p.TimestampColumn != "" && s.name == p.TimestampColumn {
					valStr, err := internal.ToString(val)
					if err != nil {
						return nil, fmt.Errorf("could not convert value to string: %w", err)
					}
					timestamp, err := internal.ParseTimestamp(p.TimestampFormat, valStr, p.location)
					if err != nil {
						return nil, fmt.Errorf("could not parse '%s' to '%s'", valStr, p.TimestampFormat)
					}
					rowGroupMetrics[rowIndex].SetTime(timestamp)
				} else {
					rowGroupMetrics[rowIndex].AddField(s.name, val)
				}

				rowIndex++
			}
		}

		metrics = append(metrics, rowGroupMetrics...)
	}

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, nil
	}
	if len(metrics) > 1 {
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
