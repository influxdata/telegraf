//go:build linux && amd64

package rapl

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePowerZone(t *testing.T) {
	tests := []struct {
		name string
		zone *powerZone
		err  string
	}{
		{
			name: "intel-rapl:0",
			zone: &powerZone{ids: []int{0}},
			err:  "",
		},
		{
			name: "intel-rapl:12:34",
			zone: &powerZone{ids: []int{12, 34}},
			err:  "",
		},
		{
			name: "foo:0",
			zone: nil,
			err:  "invalid power zone",
		},
		{
			name: "intel-rapl:X:Y",
			zone: nil,
			err:  "invalid power zone",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			zone, err := parsePowerZone(tt.name)
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
			assert.Equal(t, tt.zone, zone)
		})
	}
}

func TestPowerZoneName(t *testing.T) {
	tests := []struct {
		zone *powerZone
		name string
	}{
		{
			zone: &powerZone{[]int{0}},
			name: "intel-rapl:0",
		},
		{
			zone: &powerZone{[]int{12, 34}},
			name: "intel-rapl:12:34",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.name, tt.zone.name())
		})
	}
}

func TestPowerZonePath(t *testing.T) {
	tests := []struct {
		zone *powerZone
		path string
	}{
		{
			zone: &powerZone{[]int{0}},
			path: "intel-rapl:0",
		},
		{
			zone: &powerZone{[]int{12, 34}},
			path: "intel-rapl:12/intel-rapl:12:34",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.path, tt.zone.path())
		})
	}
}

func TestReadName(t *testing.T) {
	tests := []struct {
		files map[string]string
		zone  *powerZone
		name  string
		err   string
	}{
		{
			files: map[string]string{
				"intel-rapl:0/name": "package-0",
			},
			zone: &powerZone{ids: []int{0}},
			name: "package-0",
			err:  "",
		},
		{
			files: map[string]string{
				"intel-rapl:0/intel-rapl:0:0/name": "core",
			},
			zone: &powerZone{ids: []int{0, 0}},
			name: "core",
			err:  "",
		},
		{
			files: map[string]string{},
			zone:  &powerZone{ids: []int{12, 23}},
			name:  "",
			err:   "could not read name",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir, err := createTempFiles(tt.files)
			require.NoError(t, err)
			defer os.RemoveAll(dir)
			rapl := intelRAPL{dir: dir}
			name, err := rapl.readName(tt.zone)
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
			assert.Equal(t, tt.name, name)
		})
	}
}

func TestReadEnergy(t *testing.T) {
	tests := []struct {
		files  map[string]string
		zone   *powerZone
		energy uint64
		err    string
	}{
		{
			files: map[string]string{
				"intel-rapl:0/energy_uj": "12345",
			},
			zone:   &powerZone{ids: []int{0}},
			energy: 12345,
			err:    "",
		},
		{
			files: map[string]string{
				"intel-rapl:0/intel-rapl:0:0/energy_uj": "54321",
			},
			zone:   &powerZone{ids: []int{0, 0}},
			energy: 54321,
			err:    "",
		},
		{
			files:  map[string]string{},
			zone:   &powerZone{ids: []int{12, 23}},
			energy: 0,
			err:    "could not read energy",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir, err := createTempFiles(tt.files)
			require.NoError(t, err)
			defer os.RemoveAll(dir)
			rapl := intelRAPL{dir: dir}
			energy, err := rapl.readEnergy(tt.zone)
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
			assert.Equal(t, tt.energy, energy)
		})
	}
}

func TestPowerZones(t *testing.T) {
	tests := []struct {
		files map[string]string
		zones []powerZone
		err   string
	}{
		{
			files: map[string]string{
				"intel-rapl:0/name":                "package-0",
				"intel-rapl:0/intel-rapl:0:0/name": "core",
			},
			zones: []powerZone{{ids: []int{0}}, {ids: []int{0, 0}}},
			err:   "",
		},
		{
			files: map[string]string{
				"intel-rapl:0/name":                  "package-0",
				"intel-rapl:0/intel-rapl:12:34/name": "core",
			},
			zones: nil,
			err:   "incorrectly nested power zone",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir, err := createTempFiles(tt.files)
			require.NoError(t, err)
			defer os.RemoveAll(dir)
			rapl := intelRAPL{dir: dir}
			zones, err := rapl.powerZones()
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
			assert.Equal(t, tt.zones, zones)
		})
	}
}

// Creates a temporary directory with the given files.
// We cannot rely on a testdata dir, because we require file names containing ":".
func createTempFiles(files map[string]string) (string, error) {
	dirPath, err := os.MkdirTemp("/tmp", "testdata")
	if err != nil {
		return "", err
	}
	for path, contents := range files {
		filePath := filepath.Join(dirPath, path)
		err := os.MkdirAll(filepath.Dir(filePath), 0750)
		if err != nil {
			return "", err
		}
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			return "", err
		}
		_, err = fmt.Fprintf(f, "%s\n", contents)
		if err != nil {
			return "", err
		}
		err = f.Close()
		if err != nil {
			return "", err
		}
	}
	return dirPath, nil
}
