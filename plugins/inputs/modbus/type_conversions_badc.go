package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

func binaryMSWLEUint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(binary.LittleEndian.Uint16(b[0:]))<<16 | uint32(binary.LittleEndian.Uint16(b[2:]))
}

func binaryMSWLEUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(binary.LittleEndian.Uint16(b[0:]))<<48 | uint64(binary.LittleEndian.Uint16(b[2:]))<<32 | uint64(binary.LittleEndian.Uint16(b[4:]))<<16 | uint64(binary.LittleEndian.Uint16(b[6:]))
}

// I32 - BADC
func determineConverterBADCI32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binaryMSWLEUint32(b))
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binaryMSWLEUint32(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binaryMSWLEUint32(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterBADCI32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int32(binaryMSWLEUint32(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int32(binaryMSWLEUint32(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int32(binaryMSWLEUint32(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U32 - BADC
func determineConverterBADCU32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint32(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint32(b)
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint32(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterBADCU32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint32(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint32(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint32(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// I64 - BADC
func determineConverterBADCI64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binaryMSWLEUint64(b))
			return in
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binaryMSWLEUint64(b))
			return uint64(in)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binaryMSWLEUint64(b))
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterBADCI64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := int64(binaryMSWLEUint64(b))
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int64(binaryMSWLEUint64(b))
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int64(binaryMSWLEUint64(b))
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U64 - BADC
func determineConverterBADCU64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint64(b)
			return int64(in)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint64(b)
			return in
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint64(b)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterBADCU64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "INT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint64(b)
			return int64(float64(in) * scale)
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint64(b)
			return uint64(float64(in) * scale)
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := binaryMSWLEUint64(b)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F32 - BADC
func determineConverterBADCF32NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryMSWLEUint32(b)
			in := math.Float32frombits(raw)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterBADCF32Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryMSWLEUint32(b)
			in := math.Float32frombits(raw)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// F64 - BADC
func determineConverterBADCF64NoScale(outType string) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryMSWLEUint64(b)
			in := math.Float64frombits(raw)
			return float64(in)
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

func determineConverterBADCF64Scale(outType string, scale float64) (fieldConverterFunc, error) {
	switch outType {
	case "FLOAT64":
		return func(b []byte) interface{} {
			raw := binaryMSWLEUint64(b)
			in := math.Float64frombits(raw)
			return float64(in) * scale
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
