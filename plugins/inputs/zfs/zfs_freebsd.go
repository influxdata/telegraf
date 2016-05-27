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

// $ zpool list -Hp
// freenas-boot    30601641984     2022177280      28579464704     -       -       6       1.00x   ONLINE  -
// red1    8933531975680   1126164848640   7807367127040   -       8%      12      1.83x   ONLINE  /mnt
// temp1   2989297238016   1626309320704   1362987917312   -       38%     54      1.28x   ONLINE  /mnt
// temp2   2989297238016   626958278656    2362338959360   -       12%     20      1.00x   ONLINE  /mnt

func gatherPoolStats(poolStats bool, acc telegraf.Accumulator) (string, error) {

	lines, err := run("zpool", []string{"list", "-Hp"}...)
	if err != nil {
		return "", err
	}

	pools := []string{}
	for _, line := range lines {
		col := strings.Split(line, "\t")

		pools = append(pools, col[0])
	}

	if poolStats {
		for _, line := range lines {
			col := strings.Split(line, "\t")
			tags := map[string]string{"pool": col[0], "health": col[8]}
			fields := map[string]interface{}{}

			size, err := strconv.ParseInt(col[1], 10, 64)
			if err != nil {
				return "", fmt.Errorf("Error parsing size: %s", err)
			}
			fields["size"] = size

			alloc, err := strconv.ParseInt(col[2], 10, 64)
			if err != nil {
				return "", fmt.Errorf("Error parsing alloc: %s", err)
			}
			fields["alloc"] = alloc

			free, err := strconv.ParseInt(col[3], 10, 64)
			if err != nil {
				return "", fmt.Errorf("Error parsing free: %s", err)
			}
			fields["free"] = free

			frag, err := strconv.ParseInt(strings.TrimSuffix(col[5], "%"), 10, 0)
			if err == nil {
				// This might be - for RO devs
				fields["frag"] = frag
			}

			capval, err := strconv.ParseInt(col[6], 10, 0)
			if err != nil {
				return "", fmt.Errorf("Error parsing cap: %s", err)
			}
			fields["cap"] = capval

			dedup, err := strconv.ParseFloat(strings.TrimSuffix(col[7], "x"), 32)
			if err != nil {
				return "", fmt.Errorf("Error parsing dedup: %s", err)
			}
			fields["dedup"] = dedup

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
	poolNames, err := gatherPoolStats(z.PoolMetrics, acc)
	if err != nil {
		return err
	}
	tags["pools"] = poolNames

	// kstat.zfs.misc.vdev_cache_stats
	// kstat.zfs.misc.arcstats
	// kstat.zfs.misc.zfetchstats
	fields := make(map[string]interface{})
	for _, metric := range kstatMetrics {
		stdout, err := run("sysctl", []string{"-q", fmt.Sprintf("kstat.zfs.misc.%s", metric)}...)
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
		return nil, fmt.Errorf("%s error: %s", cmd, stderr)
	}
	return strings.Split(stdout, "\n"), nil
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{}
	})
}
