package disk

import (
	"bufio"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const defaultFstabPath = "/etc/fstab"

type fstabEntry struct {
	device     string
	mountpoint string
	fstype     string
	opts       []string
}

// fstab(5) octal escape sequences for whitespace and backslashes in fields.
var fstabUnescaper = strings.NewReplacer(
	`\040`, " ",
	`\011`, "\t",
	`\012`, "\n",
	`\134`, `\`,
)

// Kernel and virtual pseudo-filesystems that are not storage mounts.
var pseudoFS = map[string]struct{}{
	"autofs":      {},
	"binfmt_misc": {},
	"bpf":         {},
	"cgroup":      {},
	"cgroup2":     {},
	"configfs":    {},
	"debugfs":     {},
	"devpts":      {},
	"efivarfs":    {},
	"fusectl":     {},
	"hugetlbfs":   {},
	"mqueue":      {},
	"nfsd":        {},
	"proc":        {},
	"procfs":      {},
	"pstore":      {},
	"rpc_pipefs":  {},
	"securityfs":  {},
	"sysfs":       {},
	"tracefs":     {},
}

func isPseudoFS(fstype string) bool {
	_, ok := pseudoFS[fstype]
	return ok
}

func normalizeMountPath(p string) string {
	return path.Clean(filepath.ToSlash(p))
}

func readFstabFile(fstabPath string) ([]fstabEntry, error) {
	f, err := os.Open(fstabPath) //nolint:gosec // Path is configured by administrator or defaults to /etc/fstab
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseFstab(f)
}

func parseFstab(r io.Reader) ([]fstabEntry, error) {
	var entries []fstabEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		device := fstabUnescaper.Replace(fields[0])
		mountpoint := fstabUnescaper.Replace(fields[1])
		fstype := fields[2]

		if fstype == "swap" || fstype == "ignore" || mountpoint == "none" || mountpoint == "swap" {
			continue
		}
		if isPseudoFS(fstype) {
			continue
		}

		var opts []string
		if len(fields) >= 4 {
			opts = strings.Split(fields[3], ",")
		}
		if slices.Contains(opts, "noauto") {
			continue
		}

		entries = append(entries, fstabEntry{
			device:     device,
			mountpoint: normalizeMountPath(mountpoint),
			fstype:     fstype,
			opts:       opts,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
