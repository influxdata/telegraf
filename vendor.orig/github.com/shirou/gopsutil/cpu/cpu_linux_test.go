package cpu

import (
	"os"
	"testing"
)

func TestTimesEmpty(t *testing.T) {
	orig := os.Getenv("HOST_PROC")
	os.Setenv("HOST_PROC", "testdata/linux/times_empty")
	_, err := Times(true)
	if err != nil {
		t.Error("Times(true) failed")
	}
	_, err = Times(false)
	if err != nil {
		t.Error("Times(false) failed")
	}
	os.Setenv("HOST_PROC", orig)
}

func TestCPUparseStatLine_424(t *testing.T) {
	orig := os.Getenv("HOST_PROC")
	os.Setenv("HOST_PROC", "testdata/linux/424/proc")
	{
		l, err := Times(true)
		if err != nil || len(l) == 0 {
			t.Error("Times(true) failed")
		}
		t.Logf("Times(true): %#v", l)
	}
	{
		l, err := Times(false)
		if err != nil || len(l) == 0 {
			t.Error("Times(false) failed")
		}
		t.Logf("Times(false): %#v", l)
	}
	os.Setenv("HOST_PROC", orig)
}
