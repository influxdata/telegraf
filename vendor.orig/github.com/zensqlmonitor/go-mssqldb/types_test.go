package mssql

import (
	"testing"
	"reflect"
	"time"
)

func TestMakeGoLangScanType(t *testing.T) {
	if (reflect.TypeOf(int64(0)) != makeGoLangScanType(typeInfo{TypeId: typeInt8})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(float64(0)) != makeGoLangScanType(typeInfo{TypeId: typeFlt4})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(float64(0)) != makeGoLangScanType(typeInfo{TypeId: typeFlt8})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf("") != makeGoLangScanType(typeInfo{TypeId: typeVarChar})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(time.Time{}) != makeGoLangScanType(typeInfo{TypeId: typeDateTime})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(time.Time{}) != makeGoLangScanType(typeInfo{TypeId: typeDateTim4})) {
		t.Errorf("invalid type returned for typeDateTim4")
	}
}
