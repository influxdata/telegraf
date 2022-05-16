package modbus

import (
	"fmt"
	"math/big"
)

type value_offset struct {
	Val interface{}
}

func (v_o value_offset) check() value_offset {
	switch (v_o.Val).(type) {
	case int64, float64:
	default:
		fmt.Printf("value_offset (%v) is not a valid int64 or float64, it has been set to 0.0\n", v_o.Val)
		v_o.Val = 0.0
	}
	return v_o
}

func (v_o value_offset) asBigFloat() *big.Float {
	v_o.check()
	switch o := (v_o.Val).(type) {
	case int64:
		var i *big.Int = big.NewInt(o)
		return new(big.Float).SetInt(i)

	case float64:
		return big.NewFloat(o)
	default:
		return big.NewFloat(0.0)
	}
}

type value_gain struct {
	Val interface{}
}

func (v_g value_gain) check() value_gain {
	switch (v_g.Val).(type) {
	case int64, float64:
	default:
		fmt.Printf("scale (%v) is not a valid int64 or float64, it has been set to 1.0\n", v_g.Val)
		v_g.Val = 1.0
	}
	return v_g
}

func (v_g value_gain) asBigFloat() *big.Float {
	v_g.check()
	switch g := (v_g.Val).(type) {
	case int64:
		var i *big.Int = big.NewInt(g)
		return new(big.Float).SetInt(i)
	case float64:
		return big.NewFloat(g)
	default:
		return big.NewFloat(1.0)
	}
}
