package mssql

import "testing"

func TestBadOpen(t *testing.T) {
	drv := driverWithProcess(t)
	_, err := drv.open("port=bad")
	if err == nil {
		t.Fail()
	}
}
