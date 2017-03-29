// +build linux

package dmcache

import (
	"strconv"
	"strings"

	"errors"

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

	total := make(map[string]interface{})

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
		}
		aggregateStats(total, fields)
	}

	acc.AddFields(metricName, total, map[string]string{"device": "all"})

	return nil
}

func parseDMSetupStatus(line string) (map[string]interface{}, error) {
	var err error
	status := make(map[string]interface{})
	values := strings.Fields(line)
	if len(values) < 15 {
		return nil, errors.New("dmsetup status data have invalid format")
	}

	status["device"] = values[0][:len(values[0])-1]
	status["length"], err = strconv.Atoi(values[2])
	if err != nil {
		return nil, err
	}
	status["target"] = values[3]
	status["metadata_blocksize"], err = strconv.Atoi(values[4])
	if err != nil {
		return nil, err
	}
	status["metadata_used"], err = strconv.Atoi(strings.Split(values[5], "/")[0])
	if err != nil {
		return nil, err
	}
	status["metadata_total"], err = strconv.Atoi(strings.Split(values[5], "/")[1])
	if err != nil {
		return nil, err
	}
	status["cache_blocksize"], err = strconv.Atoi(values[6])
	if err != nil {
		return nil, err
	}
	status["cache_used"], err = strconv.Atoi(strings.Split(values[7], "/")[0])
	if err != nil {
		return nil, err
	}
	status["cache_total"], err = strconv.Atoi(strings.Split(values[7], "/")[1])
	if err != nil {
		return nil, err
	}
	status["read_hits"], err = strconv.Atoi(values[8])
	if err != nil {
		return nil, err
	}
	status["read_misses"], err = strconv.Atoi(values[9])
	if err != nil {
		return nil, err
	}
	status["write_hits"], err = strconv.Atoi(values[10])
	if err != nil {
		return nil, err
	}
	status["write_misses"], err = strconv.Atoi(values[11])
	if err != nil {
		return nil, err
	}
	status["demotions"], err = strconv.Atoi(values[12])
	if err != nil {
		return nil, err
	}
	status["promotions"], err = strconv.Atoi(values[13])
	if err != nil {
		return nil, err
	}
	status["dirty"], err = strconv.Atoi(values[14])
	if err != nil {
		return nil, err
	}
	status["blocksize"] = 512

	return status, nil
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
