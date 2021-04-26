package postgresql

import (
	"context"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func BenchmarkPostgresql_sequential(b *testing.B) {
	gen := batchGenerator(ctx, b, 1000, 3, 8, 12, 100, 2)
	benchmarkPostgresql(b, gen, 1, true)
}
func BenchmarkPostgresql_concurrent(b *testing.B) {
	gen := batchGenerator(ctx, b, 1000, 3, 8, 12, 100, 2)
	benchmarkPostgresql(b, gen, 10, true)
}

func benchmarkPostgresql(b *testing.B, gen <-chan []telegraf.Metric, concurrency int, foreignTags bool) {
	p := newPostgresqlTest(b)
	p.Connection += fmt.Sprintf(" pool_max_conns=%d", concurrency)
	p.TagsAsForeignKeys = foreignTags
	p.LogLevel = ""
	if err := p.Connect(); err != nil {
		b.Fatalf("Error: %s", err)
	}

	metricCount := 0

	b.ResetTimer()
	tStart := time.Now()
	for i := 0; i < b.N; i++ {
		batch := <-gen
		if err := p.Write(batch); err != nil {
			b.Fatalf("Error: %s", err)
		}
		metricCount += len(batch)
	}
	p.Close()
	b.StopTimer()
	tStop := time.Now()
	b.ReportMetric(float64(metricCount) / tStop.Sub(tStart).Seconds(), "metrics/s")
}

// tagCardinality counts all the tag keys & values as one element. fieldCardinality counts all the field keys (not values) as one element.
func batchGenerator(ctx context.Context, b *testing.B, batchSize int, numTables int, numTags int, numFields int, tagCardinality int, fieldCardinality int) <-chan []telegraf.Metric {
	tagSets := make([]MSS, tagCardinality)
	for i := 0; i < tagCardinality; i++ {
		tags := MSS{}
		for j := 0; j < numTags; j++ {
			tags[fmt.Sprintf("tag_%d", j)] = fmt.Sprintf("%d", rand.Int())
		}
		tagSets[i] = tags
	}

	metricChan := make(chan []telegraf.Metric, 32)
	go func() {
		for {
			batch := make([]telegraf.Metric, batchSize)
			for i := 0; i < batchSize; i++ {
				tableName := b.Name() + "_" + strconv.Itoa(rand.Intn(numTables))

				tags := tagSets[rand.Intn(len(tagSets))]

				m := metric.New(tableName, tags, nil, time.Now())
				m.AddTag("tableName", tableName) // ensure the tag set is unique to this table. Just in case...

				// We do field cardinality by randomizing the name of the final field to an integer < cardinality.
				for j := 0; j < numFields-1; j++ { // use -1 to reserve the last field for cardinality
					m.AddField("f"+strconv.Itoa(j), rand.Int())
				}
				m.AddField("f"+strconv.Itoa(rand.Intn(fieldCardinality)), rand.Int())

				batch[i] = m
			}

			select {
			case metricChan <- batch:
			case <-ctx.Done():
				return
			}
		}
	}()

	return metricChan
}
