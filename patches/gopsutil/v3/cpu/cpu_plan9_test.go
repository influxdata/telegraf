//go:build plan9
// +build plan9

package cpu

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var timesTests = []struct {
	mockedRootFS string
	stats        []TimesStat
}{
	{
		"2cores",
		[]TimesStat{
			{
				CPU:    "Core i7/Xeon",
				User:   2780.0 / 1000.0,
				System: 30020.0 / 1000.0,
				Idle:   (1412961713341830*2)/1000000000.0 - 2.78 - 30.02,
			},
		},
	},
}

func TestTimesPlan9(t *testing.T) {
	origRoot := os.Getenv("HOST_ROOT")
	t.Cleanup(func() {
		os.Setenv("HOST_ROOT", origRoot)
	})
	for _, tt := range timesTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			os.Setenv("HOST_ROOT", filepath.Join("testdata/plan9", tt.mockedRootFS))
			stats, err := Times(false)
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("error %v", err)
			}
			eps := cmpopts.EquateApprox(0, 0.00000001)
			if !cmp.Equal(stats, tt.stats, eps) {
				t.Errorf("got: %+v\nwant: %+v", stats, tt.stats)
			}
		})
	}
}
