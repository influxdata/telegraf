package parquet

import (
	"reflect"

	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/file"
)

func newColumnParser(reader file.ColumnChunkReader) *columnParser {
	batchSize := 128

	var valueBuffer interface{}
	switch reader.(type) {
	case *file.BooleanColumnChunkReader:
		valueBuffer = make([]bool, batchSize)
	case *file.Int32ColumnChunkReader:
		valueBuffer = make([]int32, batchSize)
	case *file.Int64ColumnChunkReader:
		valueBuffer = make([]int64, batchSize)
	case *file.Float32ColumnChunkReader:
		valueBuffer = make([]float32, batchSize)
	case *file.Float64ColumnChunkReader:
		valueBuffer = make([]float64, batchSize)
	case *file.ByteArrayColumnChunkReader:
		valueBuffer = make([]parquet.ByteArray, batchSize)
	case *file.FixedLenByteArrayColumnChunkReader:
		valueBuffer = make([]parquet.FixedLenByteArray, batchSize)
	}

	return &columnParser{
		name:        reader.Descriptor().Name(),
		reader:      reader,
		batchSize:   int64(batchSize),
		defLevels:   make([]int16, batchSize),
		repLevels:   make([]int16, batchSize),
		valueBuffer: valueBuffer,
	}
}

type columnParser struct {
	name           string
	reader         file.ColumnChunkReader
	batchSize      int64
	valueOffset    int
	valuesBuffered int

	levelOffset    int64
	levelsBuffered int64
	defLevels      []int16
	repLevels      []int16

	valueBuffer interface{}
}

func (c *columnParser) readNextBatch() error {
	var err error

	switch reader := c.reader.(type) {
	case *file.BooleanColumnChunkReader:
		values := c.valueBuffer.([]bool)
		c.levelsBuffered, c.valuesBuffered, err = reader.ReadBatch(c.batchSize, values, c.defLevels, c.repLevels)
	case *file.Int32ColumnChunkReader:
		values := c.valueBuffer.([]int32)
		c.levelsBuffered, c.valuesBuffered, err = reader.ReadBatch(c.batchSize, values, c.defLevels, c.repLevels)
	case *file.Int64ColumnChunkReader:
		values := c.valueBuffer.([]int64)
		c.levelsBuffered, c.valuesBuffered, err = reader.ReadBatch(c.batchSize, values, c.defLevels, c.repLevels)
	case *file.Float32ColumnChunkReader:
		values := c.valueBuffer.([]float32)
		c.levelsBuffered, c.valuesBuffered, err = reader.ReadBatch(c.batchSize, values, c.defLevels, c.repLevels)
	case *file.Float64ColumnChunkReader:
		values := c.valueBuffer.([]float64)
		c.levelsBuffered, c.valuesBuffered, err = reader.ReadBatch(c.batchSize, values, c.defLevels, c.repLevels)
	case *file.ByteArrayColumnChunkReader:
		values := c.valueBuffer.([]parquet.ByteArray)
		c.levelsBuffered, c.valuesBuffered, err = reader.ReadBatch(c.batchSize, values, c.defLevels, c.repLevels)
	case *file.FixedLenByteArrayColumnChunkReader:
		values := c.valueBuffer.([]parquet.FixedLenByteArray)
		c.levelsBuffered, c.valuesBuffered, err = reader.ReadBatch(c.batchSize, values, c.defLevels, c.repLevels)
	}

	c.valueOffset = 0
	c.levelOffset = 0

	return err
}

func (c *columnParser) HasNext() bool {
	return c.levelOffset < c.levelsBuffered || c.reader.HasNext()
}

func (c *columnParser) Next() (interface{}, bool) {
	if c.levelOffset == c.levelsBuffered {
		if !c.HasNext() {
			return nil, false
		}
		if err := c.readNextBatch(); err != nil {
			return nil, false
		}
		if c.levelsBuffered == 0 {
			return nil, false
		}
	}

	defLevel := c.defLevels[int(c.levelOffset)]
	c.levelOffset++

	if defLevel < c.reader.Descriptor().MaxDefinitionLevel() {
		return nil, true
	}

	vb := reflect.ValueOf(c.valueBuffer)
	val := vb.Index(c.valueOffset).Interface()
	c.valueOffset++

	// Convert byte arrays to strings
	switch v := val.(type) {
	case parquet.ByteArray:
		val = string(v)
	case parquet.FixedLenByteArray:
		val = string(v)
	}

	return val, true
}
