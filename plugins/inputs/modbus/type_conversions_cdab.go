package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

func binaryLSWBEint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(binary.BigEndian.Uint16(b[2:]))<<16 | uint32(binary.BigEndian.Uint16(b[0:]))
}

func binaryLSWBEint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(binary.BigEndian.Uint16(b[6:]))<<48 | uint64(binary.BigEndian.Uint16(b[4:]))<<32 | uint64(binary.BigEndian.Uint16(b[2:]))<<16 | uint64(binary.BigEndian.Uint16(b[0:]))
}

// I32 - CDAB
func determineConverterCDABI32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binaryLSWBEint32(b))
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binaryLSWBEint32(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binaryLSWBEint32(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterCDABI32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binaryLSWBEint32(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binaryLSWBEint32(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binaryLSWBEint32(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U32 - CDAB
func determineConverterCDABU32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint32(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint32(b)
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint32(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterCDABU32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint32(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint32(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint32(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I64 - CDAB
func determineConverterCDABI64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binaryLSWBEint64(b))
			return in
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binaryLSWBEint64(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binaryLSWBEint64(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterCDABI64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binaryLSWBEint64(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binaryLSWBEint64(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binaryLSWBEint64(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U64 - CDAB
func determineConverterCDABU64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint64(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint64(b)
			return in
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint64(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterCDABU64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint64(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint64(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryLSWBEint64(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F32 - CDAB
func determineConverterCDABF32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryLSWBEint32(b)
			in := math.Float32frombits(raw)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterCDABF32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryLSWBEint32(b)
			in := math.Float32frombits(raw)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F64 - CDAB
func determineConverterCDABF64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryLSWBEint64(b)
			in := math.Float64frombits(raw)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterCDABF64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryLSWBEint64(b)
			in := math.Float64frombits(raw)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
