// +build freebsd

package zfs

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (z *Zfs) gatherPoolStats(acc telegraf.Accumulator) (string, error) {

	lines, err := z.zpool()
	if err != nil {
		return "", err
	}

	pools := []string{}
	for _, line := range lines {
		col := strings.Split(line, "\t")

		pools = append(pools, col[0])
	}

	if z.PoolMetrics {
		for _, line := range lines {
			col := strings.Split(line, "\t")
			if len(col) != 8 {
				continue
			}

			tags := map[string]string{"pool": col[0], "health": col[1]}
			fields := map[string]interface{}{}

			if tags["health"] == "UNAVAIL" {

				fields["size"] = int64(0)

			} else {

				size, err := strconv.ParseInt(col[2], 10, 64)
				if err != nil {
					return "", fmt.Errorf("Error parsing size: %s", err)
				}
				fields["size"] = size

				alloc, err := strconv.ParseInt(col[3], 10, 64)
				if err != nil {
					return "", fmt.Errorf("Error parsing allocation: %s", err)
				}
				fields["allocated"] = alloc

				free, err := strconv.ParseInt(col[4], 10, 64)
				if err != nil {
					return "", fmt.Errorf("Error parsing free: %s", err)
				}
				fields["free"] = free

				frag, err := strconv.ParseInt(strings.TrimSuffix(col[5], "%"), 10, 0)
				if err != nil { // This might be - for RO devs
					frag = 0
				}
				fields["fragmentation"] = frag

				capval, err := strconv.ParseInt(col[6], 10, 0)
				if err != nil {
					return "", fmt.Errorf("Error parsing capacity: %s", err)
				}
				fields["capacity"] = capval

				dedup, err := strconv.ParseFloat(strings.TrimSuffix(col[7], "x"), 32)
				if err != nil {
					return "", fmt.Errorf("Error parsing dedupratio: %s", err)
				}
				fields["dedupratio"] = dedup
			}

			acc.AddFields("zfs_pool", fields, tags)
		}
	}

	return strings.Join(pools, "::"), nil
}

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	kstatMetrics := z.KstatMetrics
	if len(kstatMetrics) == 0 {
		kstatMetrics = []string{"arcstats", "zfetchstats", "vdev_cache_stats"}
	}

	tags := map[string]string{}
	poolNames, err := z.gatherPoolStats(acc)
	if err != nil {
		return err
	}
	tags["pools"] = poolNames

	fields := make(map[string]interface{})
	for _, metric := range kstatMetrics {
		stdout, err := z.sysctl(metric)
		if err != nil {
			return err
		}
		for _, line := range stdout {
			rawData := strings.Split(line, ": ")
			key := metric + "_" + strings.Split(rawData[0], ".")[4]
			value, _ := strconv.ParseInt(rawData[1], 10, 64)
			fields[key] = value
		}
	}
	acc.AddFields("zfs", fields, tags)
	return nil
}

func run(command string, args ...string) ([]string, error) {
	cmd := exec.Command(command, args...)
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	stdout := strings.TrimSpace(outbuf.String())
	stderr := strings.TrimSpace(errbuf.String())

	if _, ok := err.(*exec.ExitError); ok {
		return nil, fmt.Errorf("%s error: %s", command, stderr)
	}
	return strings.Split(stdout, "\n"), nil
}

func zpool() ([]string, error) {
	return run("zpool", []string{"list", "-Hp", "-o", "name,health,size,alloc,free,fragmentation,capacity,dedupratio"}...)
}

func sysctl(metric string) ([]string, error) {
	return run("sysctl", []string{"-q", fmt.Sprintf("kstat.zfs.misc.%s", metric)}...)
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{
			sysctl: sysctl,
			zpool:  zpool,
		}
	})
}
