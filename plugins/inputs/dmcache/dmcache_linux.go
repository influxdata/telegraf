// +build linux

package dmcache

import (
	"os/exec"
	"strconv"
	"strings"

	"errors"

	"github.com/influxdata/telegraf"
)

const metricName = "dmcache"

type cacheStatus struct {
	device            string
	length            int
	target            string
	metadataBlocksize int
	metadataUsed      int
	metadataTotal     int
	cacheBlocksize    int
	cacheUsed         int
	cacheTotal        int
	readHits          int
	readMisses        int
	writeHits         int
	writeMisses       int
	demotions         int
	promotions        int
	dirty             int
}

func (c *DMCache) Gather(acc telegraf.Accumulator) error {
	outputLines, err := c.getCurrentStatus()
	if err != nil {
		return err
	}

	totalStatus := cacheStatus{}

	for _, s := range outputLines {
		status, err := parseDMSetupStatus(s)
		if err != nil {
			return err
		}

		if c.PerDevice {
			tags := map[string]string{"device": status.device}
			acc.AddFields(metricName, toFields(status), tags)
		}
		aggregateStats(&totalStatus, status)
	}

	acc.AddFields(metricName, toFields(totalStatus), map[string]string{"device": "all"})

	return nil
}

func parseDMSetupStatus(line string) (cacheStatus, error) {
	var err error
	parseError := errors.New("Output from dmsetup could not be parsed")
	status := cacheStatus{}
	values := strings.Fields(line)
	if len(values) < 15 {
		return cacheStatus{}, parseError
	}

	status.device = strings.TrimRight(values[0], ":")
	status.length, err = strconv.Atoi(values[2])
	if err != nil {
		return cacheStatus{}, err
	}
	status.target = values[3]
	status.metadataBlocksize, err = strconv.Atoi(values[4])
	if err != nil {
		return cacheStatus{}, err
	}
	metadata := strings.Split(values[5], "/")
	if len(metadata) != 2 {
		return cacheStatus{}, parseError
	}
	status.metadataUsed, err = strconv.Atoi(metadata[0])
	if err != nil {
		return cacheStatus{}, err
	}
	status.metadataTotal, err = strconv.Atoi(metadata[1])
	if err != nil {
		return cacheStatus{}, err
	}
	status.cacheBlocksize, err = strconv.Atoi(values[6])
	if err != nil {
		return cacheStatus{}, err
	}
	cache := strings.Split(values[7], "/")
	if len(cache) != 2 {
		return cacheStatus{}, parseError
	}
	status.cacheUsed, err = strconv.Atoi(cache[0])
	if err != nil {
		return cacheStatus{}, err
	}
	status.cacheTotal, err = strconv.Atoi(cache[1])
	if err != nil {
		return cacheStatus{}, err
	}
	status.readHits, err = strconv.Atoi(values[8])
	if err != nil {
		return cacheStatus{}, err
	}
	status.readMisses, err = strconv.Atoi(values[9])
	if err != nil {
		return cacheStatus{}, err
	}
	status.writeHits, err = strconv.Atoi(values[10])
	if err != nil {
		return cacheStatus{}, err
	}
	status.writeMisses, err = strconv.Atoi(values[11])
	if err != nil {
		return cacheStatus{}, err
	}
	status.demotions, err = strconv.Atoi(values[12])
	if err != nil {
		return cacheStatus{}, err
	}
	status.promotions, err = strconv.Atoi(values[13])
	if err != nil {
		return cacheStatus{}, err
	}
	status.dirty, err = strconv.Atoi(values[14])
	if err != nil {
		return cacheStatus{}, err
	}

	return status, nil
}

func aggregateStats(totalStatus *cacheStatus, status cacheStatus) {
	totalStatus.length += status.length
	totalStatus.metadataBlocksize += status.metadataBlocksize
	totalStatus.metadataUsed += status.metadataUsed
	totalStatus.metadataTotal += status.metadataTotal
	totalStatus.cacheBlocksize += status.cacheBlocksize
	totalStatus.cacheUsed += status.cacheUsed
	totalStatus.cacheTotal += status.cacheTotal
	totalStatus.readHits += status.readHits
	totalStatus.readMisses += status.readMisses
	totalStatus.writeHits += status.writeHits
	totalStatus.writeMisses += status.writeMisses
	totalStatus.demotions += status.demotions
	totalStatus.promotions += status.promotions
	totalStatus.dirty += status.dirty
}

func toFields(status cacheStatus) map[string]interface{} {
	fields := make(map[string]interface{})
	fields["length"] = status.length
	fields["metadata_blocksize"] = status.metadataBlocksize
	fields["metadata_used"] = status.metadataUsed
	fields["metadata_total"] = status.metadataTotal
	fields["cache_blocksize"] = status.cacheBlocksize
	fields["cache_used"] = status.cacheUsed
	fields["cache_total"] = status.cacheTotal
	fields["read_hits"] = status.readHits
	fields["read_misses"] = status.readMisses
	fields["write_hits"] = status.writeHits
	fields["write_misses"] = status.writeMisses
	fields["demotions"] = status.demotions
	fields["promotions"] = status.promotions
	fields["dirty"] = status.dirty
	return fields
}

func dmSetupStatus() ([]string, error) {
	out, err := exec.Command("/bin/sh", "-c", "sudo /sbin/dmsetup status --target cache").Output()
	if err != nil {
		return nil, err
	}
	if string(out) == "No devices found\n" {
		return []string{}, nil
	}

	outString := strings.TrimRight(string(out), "\n")
	status := strings.Split(outString, "\n")

	return status, nil
}
