//go:generate ../../../tools/readme_config_includer/generator
package ipset

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/influxdata/telegraf"
)

type IpsetEntries struct {
	acc          telegraf.Accumulator
	initizalized bool
	setName      string
	numEntries   int
	numIps       int
}

func NewIpsetEntries(acc telegraf.Accumulator) *IpsetEntries {
	return &IpsetEntries{acc: acc}
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

func (counter *IpsetEntries) reset(setName string) {
	counter.initizalized = true
	counter.setName = setName
	counter.numEntries = 0
	counter.numIps = 0
}

func (counter *IpsetEntries) addLine(line string) {
	data := strings.Fields(line)
	if strings.HasPrefix(line, "create ") {
		counter.commit()
		counter.reset(data[1])
	} else if strings.HasPrefix(line, "add ") {
		counter.numEntries++

		ip := data[2]
		count, err := getCountInCidr(ip)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		counter.numIps += count
	}
}

func (counter *IpsetEntries) commit() {
	if !counter.initizalized {
		return
	}

	fields := make(map[string]interface{}, 3)
	fields["num_entries"] = counter.numEntries
	fields["num_ips"] = counter.numIps

	tags := map[string]string{
		"set": counter.setName,
	}

	counter.acc.AddGauge(measurement, fields, tags)
}
