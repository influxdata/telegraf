//go:generate ../../../tools/readme_config_includer/generator
package ipset

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/influxdata/telegraf"
)

type ipsetEntries struct {
	initizalized bool
	setName      string
	entries      int
	ips          int
}

func getCountInCidr(cidr string) (int, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// check if single IP
		if net.ParseIP(cidr) == nil {
			return 0, errors.New("invalid IP address format. Not CIDR format and not a single IP address")
		}
		return 1, nil // Single IP has only one address
	}

	ones, bits := ipNet.Mask.Size()
	if ones == 0 && bits == 0 {
		return 0, errors.New("invalid CIDR range")
	}
	numIps := 1 << (bits - ones)

	// exclude network and broadcast addresses if IPv4 and range > /31
	if bits == 32 && numIps > 2 {
		numIps -= 2
	}

	return numIps, nil
}

func (counter *ipsetEntries) reset() {
	counter.initSet("")
	counter.initizalized = false
}

func (counter *ipsetEntries) initSet(setName string) {
	counter.initizalized = true
	counter.setName = setName
	counter.entries = 0
	counter.ips = 0
}

func (counter *ipsetEntries) addLine(line string, acc telegraf.Accumulator) {
	data := strings.Fields(line)
	if len(data) < 3 {
		acc.AddError(fmt.Errorf("error parsing line (expected at least 3 fields): %s", line))
		return
	}

	operation := data[0]
	if operation == "create" {
		counter.commit(acc)
		counter.initSet(data[1])
	} else if operation == "add" {
		counter.entries++

		ip := data[2]
		count, err := getCountInCidr(ip)
		if err != nil {
			acc.AddError(err)
			return
		}
		counter.ips += count
	}
}

func (counter *ipsetEntries) commit(acc telegraf.Accumulator) {
	if !counter.initizalized {
		return
	}

	fields := make(map[string]interface{}, 3)
	fields["entries"] = counter.entries
	fields["ips"] = counter.ips

	tags := map[string]string{
		"set": counter.setName,
	}

	acc.AddGauge(measurement, fields, tags)

	// reset counter and prepare for next usage
	counter.reset()
}
