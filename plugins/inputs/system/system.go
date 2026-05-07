//go:generate ../../../tools/readme_config_includer/generator
package system

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type System struct {
	Include          []string        `toml:"include"`
	OSCacheTTL       config.Duration `toml:"os_cache_ttl"`
	HardwareCacheTTL config.Duration `toml:"hardware_cache_ttl"`
	Log              telegraf.Logger `toml:"-"`

	osCache          map[string]interface{}
	osCachedAt       time.Time
	hardwareCache    map[string]interface{}
	hardwareCachedAt time.Time
}

func (*System) SampleConfig() string {
	return sampleConfig
}

func (s *System) Init() error {
	// Suppress deprecation warnings for default-only configs.
	userSupplied := len(s.Include) > 0
	if !userSupplied {
		s.Include = []string{"load", "users", "legacy_cpus", "legacy_uptime"}
	}

	enabled := make(map[string]bool, len(s.Include))
	deduped := make([]string, 0, len(s.Include))
	for _, incl := range s.Include {
		if enabled[incl] {
			continue
		}
		switch incl {
		case "load", "users", "cpus", "uptime", "os", "hardware":
		case "legacy_cpus":
			if userSupplied {
				config.PrintOptionValueDeprecationNotice(
					"inputs.system",
					"include",
					"legacy_cpus",
					telegraf.DeprecationInfo{
						Since:     "1.39.0",
						RemovalIn: "1.45.0",
						Notice:    "use 'cpus' instead",
					},
				)
			}
		case "legacy_uptime":
			if userSupplied {
				config.PrintOptionValueDeprecationNotice(
					"inputs.system",
					"include",
					"legacy_uptime",
					telegraf.DeprecationInfo{
						Since:     "1.39.0",
						RemovalIn: "1.45.0",
						Notice:    "use 'uptime' instead",
					},
				)
			}
		default:
			return fmt.Errorf("invalid 'include' option %q", incl)
		}
		enabled[incl] = true
		deduped = append(deduped, incl)
	}
	s.Include = deduped

	if enabled["cpus"] && enabled["legacy_cpus"] {
		return errors.New(`"cpus" and "legacy_cpus" are mutually exclusive`)
	}
	if enabled["uptime"] && enabled["legacy_uptime"] {
		return errors.New(`"uptime" and "legacy_uptime" are mutually exclusive`)
	}

	if enabled["hardware"] && !hardwareSupported {
		s.Log.Warn("'hardware' is not supported on this platform, ignoring")
	}

	return nil
}

func (s *System) Gather(acc telegraf.Accumulator) error {
	now := time.Now()
	fields := make(map[string]interface{}, 8)

	for _, incl := range s.Include {
		switch incl {
		case "os":
			if time.Since(s.osCachedAt) > time.Duration(s.OSCacheTTL) {
				osCache, err := gatherOS()
				if err != nil {
					acc.AddError(err)
				} else {
					s.osCache = osCache
					s.osCachedAt = now
				}
			}
			if len(s.osCache) > 0 {
				acc.AddFields("system_os", s.osCache, nil, now)
			}
		case "hardware":
			if time.Since(s.hardwareCachedAt) > time.Duration(s.HardwareCacheTTL) {
				hwCache, err := gatherHardware()
				if err != nil {
					acc.AddError(err)
				} else {
					s.hardwareCache = hwCache
					s.hardwareCachedAt = now
				}
			}
			if len(s.hardwareCache) > 0 {
				acc.AddFields("system_hardware", s.hardwareCache, nil, now)
			}
		case "load":
			loadavg, err := load.Avg()
			if err != nil {
				if !strings.Contains(err.Error(), "not implemented") {
					acc.AddError(fmt.Errorf("reading load averages: %w", err))
				}
				continue
			}
			fields["load1"] = loadavg.Load1
			fields["load5"] = loadavg.Load5
			fields["load15"] = loadavg.Load15
		case "users":
			users, err := host.Users()
			if err == nil {
				fields["n_users"] = len(users)
				fields["n_unique_users"] = findUniqueUsers(users)
			} else if os.IsNotExist(err) {
				s.Log.Debugf("Reading users: %s", err.Error())
			} else if os.IsPermission(err) {
				s.Log.Debug(err.Error())
			} else {
				s.Log.Warnf("Reading users: %s", err.Error())
			}
		case "cpus", "legacy_cpus":
			numLogicalCPUs, err := cpu.Counts(true)
			if err != nil {
				acc.AddError(fmt.Errorf("reading logical CPU count: %w", err))
				continue
			}
			numPhysicalCPUs, err := cpu.Counts(false)
			if err != nil {
				acc.AddError(fmt.Errorf("reading physical CPU count: %w", err))
				continue
			}
			if incl == "cpus" {
				fields["n_virtual_cpus"] = numLogicalCPUs
			} else {
				fields["n_cpus"] = numLogicalCPUs
			}
			fields["n_physical_cpus"] = numPhysicalCPUs
		case "uptime":
			uptime, err := host.Uptime()
			if err != nil {
				acc.AddError(fmt.Errorf("reading uptime: %w", err))
				continue
			}
			fields["uptime"] = uptime
		case "legacy_uptime":
			uptime, err := host.Uptime()
			if err != nil {
				acc.AddError(fmt.Errorf("reading uptime: %w", err))
				continue
			}
			acc.AddCounter("system", map[string]interface{}{
				"uptime": uptime,
			}, nil, now)
			acc.AddFields("system", map[string]interface{}{
				"uptime_format": formatUptime(uptime),
			}, nil, now)
		}
	}

	if len(fields) > 0 {
		acc.AddGauge("system", fields, nil, now)
	}

	return nil
}

// gatherOS reads OS release and uname information via gopsutil, skipping
// host.Info() to avoid the unrelated virtualization, boot-time and
// process-count probes.
func gatherOS() (map[string]interface{}, error) {
	platform, family, version, err := host.PlatformInformation()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return nil, fmt.Errorf("reading platform information: %w", err)
	}
	kernelVersion, err := host.KernelVersion()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return nil, fmt.Errorf("reading kernel version: %w", err)
	}
	arch, err := host.KernelArch()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return nil, fmt.Errorf("reading kernel architecture: %w", err)
	}
	if arch == "" {
		arch = runtime.GOARCH
	}

	return map[string]interface{}{
		"os":               runtime.GOOS,
		"arch":             arch,
		"platform":         platform,
		"platform_family":  family,
		"platform_version": version,
		"kernel_version":   kernelVersion,
	}, nil
}

// gatherHardware reads BIOS, baseboard, chassis and product DMI/SMBIOS
// information. Fields that cannot be read are omitted.
func gatherHardware() (map[string]interface{}, error) {
	// Disable ghw warnings; honor GHW_CHROOT and other GHW_* env variables.
	ctx := ghw.WithDisableWarnings()(ghw.ContextFromEnv())

	fields := make(map[string]interface{})

	bios, err := ghw.BIOS(ctx)
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return nil, fmt.Errorf("reading BIOS information: %w", err)
	}
	if bios != nil {
		addNonEmpty(fields, "bios_vendor", bios.Vendor)
		addNonEmpty(fields, "bios_version", bios.Version)
		addNonEmpty(fields, "bios_date", bios.Date)
	}

	bb, err := ghw.Baseboard(ctx)
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return nil, fmt.Errorf("reading baseboard information: %w", err)
	}
	if bb != nil {
		addNonEmpty(fields, "board_vendor", bb.Vendor)
		addNonEmpty(fields, "board_product", bb.Product)
		addNonEmpty(fields, "board_version", bb.Version)
	}

	ch, err := ghw.Chassis(ctx)
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return nil, fmt.Errorf("reading chassis information: %w", err)
	}
	if ch != nil {
		addNonEmpty(fields, "chassis_vendor", ch.Vendor)
		addNonEmpty(fields, "chassis_type", ch.Type)
		addNonEmpty(fields, "chassis_type_description", ch.TypeDescription)
		addNonEmpty(fields, "chassis_version", ch.Version)
	}

	prod, err := ghw.Product(ctx)
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return nil, fmt.Errorf("reading product information: %w", err)
	}
	if prod != nil {
		addNonEmpty(fields, "product_vendor", prod.Vendor)
		addNonEmpty(fields, "product_name", prod.Name)
		addNonEmpty(fields, "product_family", prod.Family)
	}

	return fields, nil
}

// addNonEmpty adds value to fields under key, dropping empty strings and the
// ghw "unknown" sentinel.
func addNonEmpty(fields map[string]interface{}, key, value string) {
	value = strings.TrimSpace(value)
	if value == "" || value == "unknown" {
		return
	}
	fields[key] = value
}

func findUniqueUsers(userStats []host.UserStat) int {
	uniqueUsers := make(map[string]bool)
	for _, userstat := range userStats {
		if _, ok := uniqueUsers[userstat.User]; !ok {
			uniqueUsers[userstat.User] = true
		}
	}
	return len(uniqueUsers)
}

func formatUptime(uptime uint64) string {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	days := uptime / (60 * 60 * 24)
	if days != 0 {
		s := ""
		if days > 1 {
			s = "s"
		}
		fmt.Fprintf(w, "%d day%s, ", days, s)
	}

	minutes := uptime / 60
	hours := minutes / 60
	hours %= 24
	minutes %= 60

	fmt.Fprintf(w, "%2d:%02d", hours, minutes)
	w.Flush()
	return buf.String()
}

func init() {
	inputs.Add("system", func() telegraf.Input {
		return &System{
			OSCacheTTL:       config.Duration(8 * time.Hour),
			HardwareCacheTTL: config.Duration(8 * time.Hour),
		}
	})
}
