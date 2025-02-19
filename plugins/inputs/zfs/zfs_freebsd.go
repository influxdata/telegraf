//go:build freebsd

package zfs

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (z *Zfs) Init() error {
	// Determine the kernel version to adapt parsing
	release, err := z.uname()
	if err != nil {
		return fmt.Errorf("determining uname failed: %w", err)
	}
	parts := strings.SplitN(release, ".", 2)
	z.version, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("determining version from %q failed: %w", release, err)
	}

	// Setup default metrics if they are not specified.
	// Please note that starting from FreeBSD 14 the 'vdev_cache_stats' are
	// no longer available.
	if len(z.KstatMetrics) == 0 {
		if z.version < 14 {
			z.KstatMetrics = []string{"arcstats", "zfetchstats", "vdev_cache_stats"}
		} else {
			z.KstatMetrics = []string{"arcstats", "zfetchstats"}
		}
	}

	return nil
}

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	tags := map[string]string{}

	poolNames, err := z.gatherPoolStats(acc)
	if err != nil {
		return err
	}
	if poolNames != "" {
		tags["pools"] = poolNames
	}

	datasetNames, err := z.gatherDatasetStats(acc)
	if err != nil {
		return err
	}
	if datasetNames != "" {
		tags["datasets"] = datasetNames
	}

	// Gather information form the kernel using sysctl
	fields := make(map[string]interface{})
	var removeIndices []int
	for i, metric := range z.KstatMetrics {
		stdout, err := z.sysctl(metric)
		if err != nil {
			z.Log.Warnf("sysctl for 'kstat.zfs.misc.%s' failed: %v; removing metric", metric, err)
			removeIndices = append(removeIndices, i)
			continue
		}
		for _, line := range stdout {
			rawData := strings.Split(line, ": ")
			key := metric + "_" + strings.Split(rawData[0], ".")[4]
			value, _ := strconv.ParseInt(rawData[1], 10, 64)
			fields[key] = value
		}
	}
	acc.AddFields("zfs", fields, tags)

	// Remove the invalid kstat metrics
	if len(removeIndices) > 0 {
		for i := len(removeIndices) - 1; i >= 0; i-- {
			idx := removeIndices[i]
			z.KstatMetrics = append(z.KstatMetrics[:idx], z.KstatMetrics[idx+1:]...)
		}
	}

	return nil
}

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

	if !z.PoolMetrics {
		return strings.Join(pools, "::"), nil
	}

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

	return strings.Join(pools, "::"), nil
}

func (z *Zfs) gatherDatasetStats(acc telegraf.Accumulator) (string, error) {
	properties := []string{"name", "avail", "used", "usedsnap", "usedds"}

	lines, err := z.zdataset(properties)
	if err != nil {
		return "", err
	}

	datasets := []string{}
	for _, line := range lines {
		col := strings.Split(line, "\t")
		datasets = append(datasets, col[0])
	}

	if !z.DatasetMetrics {
		return strings.Join(datasets, "::"), nil
	}

	for _, line := range lines {
		col := strings.Split(line, "\t")
		if len(col) != len(properties) {
			z.Log.Warnf("Invalid number of columns for line: %s", line)
			continue
		}

		tags := map[string]string{"dataset": col[0]}
		fields := map[string]interface{}{}

		for i, key := range properties[1:] {
			// Treat '-' entries as zero
			if col[i+1] == "-" {
				fields[key] = int64(0)
				continue
			}
			value, err := strconv.ParseInt(col[i+1], 10, 64)
			if err != nil {
				return "", fmt.Errorf("Error parsing %s %q: %s", key, col[i+1], err)
			}
			fields[key] = value
		}

		acc.AddFields("zfs_dataset", fields, tags)
	}

	return strings.Join(datasets, "::"), nil
}

func run(command string, args ...string) ([]string, error) {
	cmd := exec.Command(command, args...)
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	stdout := strings.TrimSpace(outbuf.String())
	stderr := strings.TrimSpace(errbuf.String())

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s error: %s", command, stderr)
		}
		return nil, fmt.Errorf("%s error: %s", command, err)
	}
	return strings.Split(stdout, "\n"), nil
}

func zpool() ([]string, error) {
	return run("zpool", []string{"list", "-Hp", "-o", "name,health,size,alloc,free,fragmentation,capacity,dedupratio"}...)
}

func zdataset(properties []string) ([]string, error) {
	return run("zfs", []string{"list", "-Hp", "-t", "filesystem,volume", "-o", strings.Join(properties, ",")}...)
}

func sysctl(metric string) ([]string, error) {
	return run("sysctl", []string{"-q", fmt.Sprintf("kstat.zfs.misc.%s", metric)}...)
}

func uname() (string, error) {
	var info unix.Utsname
	if err := unix.Uname(&info); err != nil {
		return "", err
	}
	release := unix.ByteSliceToString(info.Release[:])
	return release, nil
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{
			sysctl:   sysctl,
			zpool:    zpool,
			zdataset: zdataset,
			uname:    uname,
		}
	})
}
