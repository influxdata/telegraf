package modbus

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

type convert16 func([]byte) uint16

func endianessConverter16(byteOrder string) (convert16, error) {
	switch byteOrder {
	case "ABCD", "CDAB": // Big endian (Motorola)
		return binary.BigEndian.Uint16, nil
	case "DCBA", "BADC": // Little endian (Intel)
		return binary.LittleEndian.Uint16, nil
	}
	return nil, fmt.Errorf("invalid byte-order: %s", byteOrder)
}
func rescaleI16AsBigFloat(in int16, valueGain valueGain, valueOffset valueOffset) *big.Float {
	t := big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), valueGain.asBigFloat())
	t.Add(t, valueOffset.asBigFloat())
	return t
}
func rescaleU16AsBigFloat(in uint16, valueGain valueGain, valueOffset valueOffset) *big.Float {
	t := big.NewFloat(0.0)
	t.Mul(big.NewFloat(float64(in)), valueGain.asBigFloat())
	t.Add(t, valueOffset.asBigFloat())
	return t
}

// I16
func determineConverterI16(outType, byteOrder string, valueGain valueGain, valueOffset valueOffset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}
	switch outType {
	case "native":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return int16(forceIntToMinMax(rescaleI16AsBigFloat(in, valueGain, valueOffset), "INT16").Int64())
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return forceIntToMinMax(rescaleI16AsBigFloat(in, valueGain, valueOffset), "INT64").Int64()
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			return forceIntToMinMax(rescaleI16AsBigFloat(in, valueGain, valueOffset), "UINT64").Uint64()
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := int16(tohost(b))
			fmt.Printf("In: %v\n", in)
			return forceFloat64ToMinMax(rescaleI16AsBigFloat(in, valueGain, valueOffset))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}

// U16
func determineConverterU16(outType, byteOrder string, valueGain valueGain, valueOffset valueOffset) (fieldConverterFunc, error) {
	tohost, err := endianessConverter16(byteOrder)
	if err != nil {
		return nil, err
	}
	switch outType {
	case "native": //U16
		return func(b []byte) interface{} {
			in := tohost(b)
			return uint16(forceIntToMinMax(rescaleU16AsBigFloat(in, valueGain, valueOffset), "UINT16").Uint64())
		}, nil
	case "INT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceIntToMinMax(rescaleU16AsBigFloat(in, valueGain, valueOffset), "INT64").Int64()
		}, nil
	case "UINT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceIntToMinMax(rescaleU16AsBigFloat(in, valueGain, valueOffset), "UINT64").Uint64()
		}, nil
	case "FLOAT64":
		return func(b []byte) interface{} {
			in := tohost(b)
			return forceFloat64ToMinMax(rescaleU16AsBigFloat(in, valueGain, valueOffset))
		}, nil
	}
	return nil, fmt.Errorf("invalid output data-type: %s", outType)
}
