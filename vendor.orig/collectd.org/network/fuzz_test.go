// +build gofuzz

package network // import "collectd.org/network"

import (
	"io/ioutil"
	"testing"
)

func TestFuzz(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/packet2.bin")
	if err != nil {
		panic(err)
	}

	got := Fuzz(data)
	if got != 1 {
		t.Errorf("Failed to fuzz a sample packet. Wanted [%v] Got [%v]\n", 1, got)
	}

}
