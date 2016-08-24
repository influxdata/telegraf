// +build linux

package dmcache

import (
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

const metricName = "dmcache"

var fieldNames = [...]string{
	"metadata_used",
	"metadata_free",
	"cache_used",
	"cache_free",
	"read_hits",
	"read_misses",
	"write_hits",
	"write_misses",
	"demotions",
	"promotions",
	"dirty",
}

func (c *DMCache) Gather(acc telegraf.Accumulator) error {
	outputLines, err := c.rawStatus()
	if err != nil {
		return err
	}

	var total map[string]interface{}
	if !c.PerDevice {
		total = make(map[string]interface{})
	}

	for _, s := range outputLines {
		fields := make(map[string]interface{})
		data, err := parseDMSetupStatus(s)
		if err != nil {
			return err
		}

		for _, f := range fieldNames {
			fields[f] = calculateSize(data, f)
		}

		if c.PerDevice {
			tags := map[string]string{"device": data["device"].(string)}
			acc.AddFields(metricName, fields, tags)
		} else {
			aggregateStats(total, fields)
		}
	}

	if !c.PerDevice {
		acc.AddFields(metricName, total, nil)
	}

	return nil
}

func parseDMSetupStatus(line string) (status map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			status = nil
			err = r.(error)
		}
	}()

	values := strings.Split(line, " ")
	status = make(map[string]interface{})

	status["device"] = values[0][:len(values[0])-1]
	status["length"] = toInt(values[2])
	status["target"] = values[3]
	status["metadata_blocksize"] = toInt(values[4])
	status["metadata_used"] = toInt(strings.Split(values[5], "/")[0])
	status["metadata_total"] = toInt(strings.Split(values[5], "/")[1])
	status["cache_blocksize"] = toInt(values[6])
	status["cache_used"] = toInt(strings.Split(values[7], "/")[0])
	status["cache_total"] = toInt(strings.Split(values[7], "/")[1])
	status["read_hits"] = toInt(values[8])
	status["read_misses"] = toInt(values[9])
	status["write_hits"] = toInt(values[10])
	status["write_misses"] = toInt(values[11])
	status["demotions"] = toInt(values[12])
	status["promotions"] = toInt(values[13])
	status["dirty"] = toInt(values[14])
	status["blocksize"] = 512

	return status, nil
}

func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func calculateSize(data map[string]interface{}, key string) (value int) {
	if key == "metadata_free" {
		value = data["metadata_total"].(int) - data["metadata_used"].(int)
	} else if key == "cache_free" {
		value = data["cache_total"].(int) - data["cache_used"].(int) - data["dirty"].(int)
	} else {
		value = data[key].(int)
	}

	if key == "metadata_free" || key == "metadata_used" {
		value = value * data["blocksize"].(int) * data["metadata_blocksize"].(int)
	} else {
		value = value * data["blocksize"].(int) * data["cache_blocksize"].(int)
	}

	return
}

func aggregateStats(total, fields map[string]interface{}) {
	for key, value := range fields {
		if _, ok := total[key]; ok {
			total[key] = total[key].(int) + value.(int)
		} else {
			total[key] = value.(int)
		}
	}
}
