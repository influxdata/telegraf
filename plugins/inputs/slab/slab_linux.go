//go:build linux
// +build linux

package slab

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

func (ss *SlabStats) Init() error {
	return nil
}

func (ss *SlabStats) Gather(acc telegraf.Accumulator) error {
	fields, err := ss.getSlabStats()
	if err != nil {
		return err
	}

	acc.AddGauge("slab", fields, nil)
	return nil
}

func (ss *SlabStats) getSlabStats() (map[string]interface{}, error) {
	fields := map[string]interface{}{}

	file, err := os.Open(ss.statFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read header rows
	scanner.Scan() // for "slabinfo - version: 2.1"
	scanner.Scan() // for "# name <active_objs> <num_objs> <objsize> ..."

	// Read data rows
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Fields(line)

		if len(cols) < 4 {
			return nil, errors.New("the content of /proc/slabinfo is invalid")
		}

		var numObj, sizObj int

		numObj, err = strconv.Atoi(cols[2])
		if err != nil {
			return nil, err
		}

		sizObj, err = strconv.Atoi(cols[3])
		if err != nil {
			return nil, err
		}

		fields[cols[0]+"_size_in_bytes"] = numObj * sizObj
	}
	return fields, nil
}
