package modbus

import (
	"fmt"
	"math/big"
)

type valueOffset struct {
	Val interface{}
}

func (in valueOffset) check() valueOffset {
	out := in
	switch (in.Val).(type) {
	case int64, float64:
	default:
		fmt.Printf("valueOffset (%v) is not a valid int64 or float64, it has been set to 0.0\n", in.Val)
		out.Val = 0.0
	}
	return out
}

func (in valueOffset) asBigFloat() *big.Float {
	switch o := (in.Val).(type) {
	case int64:
		i := big.NewInt(o)
		return new(big.Float).SetInt(i)

	case float64:
		return big.NewFloat(o)
	default:
		return big.NewFloat(0.0)
	}
}

type valueGain struct {
	Val interface{}
}

func (in valueGain) check() valueGain {
	out := in
	switch (in.Val).(type) {
	case int64, float64:
	default:
		fmt.Printf("scale (%v) is not a valid int64 or float64, it has been set to 1.0\n", in.Val)
		out.Val = 1.0
	}
	return out
}

func (in valueGain) asBigFloat() *big.Float {
	switch g := (in.Val).(type) {
	case int64:
		i := big.NewInt(g)
		return new(big.Float).SetInt(i)
	case float64:
		return big.NewFloat(g)
	default:
		return big.NewFloat(1.0)
	}
}
