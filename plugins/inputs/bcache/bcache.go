//go:build !windows
// +build !windows

// bcache doesn't aim for Windows

package bcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Bcache struct {
	BcachePath string
	BcacheDevs []string
}

func getTags(bdev string) map[string]string {
	backingDevFile, _ := os.Readlink(bdev)
	backingDevPath := strings.Split(backingDevFile, "/")
	backingDev := backingDevPath[len(backingDevPath)-2]

	bcacheDevFile, _ := os.Readlink(bdev + "/dev")
	bcacheDevPath := strings.Split(bcacheDevFile, "/")
	bcacheDev := bcacheDevPath[len(bcacheDevPath)-1]

	return map[string]string{"backing_dev": backingDev, "bcache_dev": bcacheDev}
}

func prettyToBytes(v string) uint64 {
	var factors = map[string]uint64{
		"k": 1 << 10,
		"M": 1 << 20,
		"G": 1 << 30,
		"T": 1 << 40,
		"P": 1 << 50,
		"E": 1 << 60,
	}
	var factor uint64
	factor = 1
	prefix := v[len(v)-1:]
	if factors[prefix] != 0 {
		v = v[:len(v)-1]
		factor = factors[prefix]
	}
	result, _ := strconv.ParseFloat(v, 32)
	result = result * float64(factor)

	return uint64(result)
}

func (b *Bcache) gatherBcache(bdev string, acc telegraf.Accumulator) error {
	tags := getTags(bdev)
	metrics, err := filepath.Glob(bdev + "/stats_total/*")
	if err != nil {
		return err
	}
	if len(metrics) == 0 {
		return errors.New("can't read any stats file")
	}
	file, err := os.ReadFile(bdev + "/dirty_data")
	if err != nil {
		return err
	}
	rawValue := strings.TrimSpace(string(file))
	value := prettyToBytes(rawValue)

	fields := make(map[string]interface{})
	fields["dirty_data"] = value

	for _, path := range metrics {
		key := filepath.Base(path)
		file, err := os.ReadFile(path)
		rawValue := strings.TrimSpace(string(file))
		if err != nil {
			return err
		}
		if key == "bypassed" {
			value := prettyToBytes(rawValue)
			fields[key] = value
		} else {
			value, _ := strconv.ParseUint(rawValue, 10, 64)
			fields[key] = value
		}
	}
	acc.AddFields("bcache", fields, tags)
	return nil
}

func (b *Bcache) Gather(acc telegraf.Accumulator) error {
	bcacheDevsChecked := make(map[string]bool)
	var restrictDevs bool
	if len(b.BcacheDevs) != 0 {
		restrictDevs = true
		for _, bcacheDev := range b.BcacheDevs {
			bcacheDevsChecked[bcacheDev] = true
		}
	}

	bcachePath := b.BcachePath
	if len(bcachePath) == 0 {
		bcachePath = "/sys/fs/bcache"
	}
	bdevs, _ := filepath.Glob(bcachePath + "/*/bdev*")
	if len(bdevs) < 1 {
		return errors.New("can't find any bcache device")
	}
	for _, bdev := range bdevs {
		if restrictDevs {
			bcacheDev := getTags(bdev)["bcache_dev"]
			if !bcacheDevsChecked[bcacheDev] {
				continue
			}
		}
		if err := b.gatherBcache(bdev, acc); err != nil {
			return fmt.Errorf("gathering bcache failed: %v", err)
		}
	}
	return nil
}

func init() {
	inputs.Add("bcache", func() telegraf.Input {
		return &Bcache{}
	})
}
