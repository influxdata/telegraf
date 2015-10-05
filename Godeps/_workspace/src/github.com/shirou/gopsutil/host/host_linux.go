// +build linux

package host

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	common "github.com/shirou/gopsutil/common"
)

type LSB struct {
	ID          string
	Release     string
	Codename    string
	Description string
}

func HostInfo() (*HostInfoStat, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	ret := &HostInfoStat{
		Hostname: hostname,
		OS:       runtime.GOOS,
	}

	platform, family, version, err := GetPlatformInformation()
	if err == nil {
		ret.Platform = platform
		ret.PlatformFamily = family
		ret.PlatformVersion = version
	}
	system, role, err := GetVirtualization()
	if err == nil {
		ret.VirtualizationSystem = system
		ret.VirtualizationRole = role
	}
	uptime, err := BootTime()
	if err == nil {
		ret.Uptime = uptime
	}

	return ret, nil
}

// BootTime returns the system boot time expressed in seconds since the epoch.
func BootTime() (uint64, error) {
	lines, err := common.ReadLines("/proc/stat")
	if err != nil {
		return 0, err
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "btime") {
			f := strings.Fields(line)
			if len(f) != 2 {
				return 0, fmt.Errorf("wrong btime format")
			}
			b, err := strconv.ParseInt(f[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return uint64(b), nil
		}
	}

	return 0, fmt.Errorf("could not find btime")
}

func Users() ([]UserStat, error) {
	utmpfile := "/var/run/utmp"

	file, err := os.Open(utmpfile)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	u := utmp{}
	entrySize := int(unsafe.Sizeof(u))
	count := len(buf) / entrySize

	ret := make([]UserStat, 0, count)

	for i := 0; i < count; i++ {
		b := buf[i*entrySize : i*entrySize+entrySize]

		var u utmp
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.LittleEndian, &u)
		if err != nil {
			continue
		}
		user := UserStat{
			User:     common.IntToString(u.User[:]),
			Terminal: common.IntToString(u.Line[:]),
			Host:     common.IntToString(u.Host[:]),
			Started:  int(u.Tv.TvSec),
		}
		ret = append(ret, user)
	}

	return ret, nil

}

func getLSB() (*LSB, error) {
	ret := &LSB{}
	if common.PathExists("/etc/lsb-release") {
		contents, err := common.ReadLines("/etc/lsb-release")
		if err != nil {
			return ret, err // return empty
		}
		for _, line := range contents {
			field := strings.Split(line, "=")
			if len(field) < 2 {
				continue
			}
			switch field[0] {
			case "DISTRIB_ID":
				ret.ID = field[1]
			case "DISTRIB_RELEASE":
				ret.Release = field[1]
			case "DISTRIB_CODENAME":
				ret.Codename = field[1]
			case "DISTRIB_DESCRIPTION":
				ret.Description = field[1]
			}
		}
	} else if common.PathExists("/usr/bin/lsb_release") {
		out, err := exec.Command("/usr/bin/lsb_release").Output()
		if err != nil {
			return ret, err
		}
		for _, line := range strings.Split(string(out), "\n") {
			field := strings.Split(line, ":")
			if len(field) < 2 {
				continue
			}
			switch field[0] {
			case "Distributor ID":
				ret.ID = field[1]
			case "Release":
				ret.Release = field[1]
			case "Codename":
				ret.Codename = field[1]
			case "Description":
				ret.Description = field[1]
			}
		}

	}

	return ret, nil
}

func GetPlatformInformation() (platform string, family string, version string, err error) {

	lsb, err := getLSB()
	if err != nil {
		lsb = &LSB{}
	}

	if common.PathExists("/etc/oracle-release") {
		platform = "oracle"
		contents, err := common.ReadLines("/etc/oracle-release")
		if err == nil {
			version = getRedhatishVersion(contents)
		}
	} else if common.PathExists("/etc/enterprise-release") {
		platform = "oracle"
		contents, err := common.ReadLines("/etc/enterprise-release")
		if err == nil {
			version = getRedhatishVersion(contents)
		}
	} else if common.PathExists("/etc/debian_version") {
		if lsb.ID == "Ubuntu" {
			platform = "ubuntu"
			version = lsb.Release
		} else if lsb.ID == "LinuxMint" {
			platform = "linuxmint"
			version = lsb.Release
		} else {
			if common.PathExists("/usr/bin/raspi-config") {
				platform = "raspbian"
			} else {
				platform = "debian"
			}
			contents, err := common.ReadLines("/etc/debian_version")
			if err == nil {
				version = contents[0]
			}
		}
	} else if common.PathExists("/etc/redhat-release") {
		contents, err := common.ReadLines("/etc/redhat-release")
		if err == nil {
			version = getRedhatishVersion(contents)
			platform = getRedhatishPlatform(contents)
		}
	} else if common.PathExists("/etc/system-release") {
		contents, err := common.ReadLines("/etc/system-release")
		if err == nil {
			version = getRedhatishVersion(contents)
			platform = getRedhatishPlatform(contents)
		}
	} else if common.PathExists("/etc/gentoo-release") {
		platform = "gentoo"
		contents, err := common.ReadLines("/etc/gentoo-release")
		if err == nil {
			version = getRedhatishVersion(contents)
		}
	} else if common.PathExists("/etc/SuSE-release") {
		contents, err := common.ReadLines("/etc/SuSE-release")
		if err == nil {
			version = getSuseVersion(contents)
			platform = getSusePlatform(contents)
		}
		// TODO: slackware detecion
	} else if common.PathExists("/etc/arch-release") {
		platform = "arch"
		// TODO: exherbo detection
	} else if lsb.ID == "RedHat" {
		platform = "redhat"
		version = lsb.Release
	} else if lsb.ID == "Amazon" {
		platform = "amazon"
		version = lsb.Release
	} else if lsb.ID == "ScientificSL" {
		platform = "scientific"
		version = lsb.Release
	} else if lsb.ID == "XenServer" {
		platform = "xenserver"
		version = lsb.Release
	} else if lsb.ID != "" {
		platform = strings.ToLower(lsb.ID)
		version = lsb.Release
	}

	switch platform {
	case "debian", "ubuntu", "linuxmint", "raspbian":
		family = "debian"
	case "fedora":
		family = "fedora"
	case "oracle", "centos", "redhat", "scientific", "enterpriseenterprise", "amazon", "xenserver", "cloudlinux", "ibm_powerkvm":
		family = "rhel"
	case "suse", "opensuse":
		family = "suse"
	case "gentoo":
		family = "gentoo"
	case "slackware":
		family = "slackware"
	case "arch":
		family = "arch"
	case "exherbo":
		family = "exherbo"
	}

	return platform, family, version, nil

}

func getRedhatishVersion(contents []string) string {
	c := strings.ToLower(strings.Join(contents, ""))

	if strings.Contains(c, "rawhide") {
		return "rawhide"
	}
	if matches := regexp.MustCompile(`release (\d[\d.]*)`).FindStringSubmatch(c); matches != nil {
		return matches[1]
	}
	return ""
}

func getRedhatishPlatform(contents []string) string {
	c := strings.ToLower(strings.Join(contents, ""))

	if strings.Contains(c, "red hat") {
		return "redhat"
	}
	f := strings.Split(c, " ")

	return f[0]
}

func getSuseVersion(contents []string) string {
	version := ""
	for _, line := range contents {
		if matches := regexp.MustCompile(`VERSION = ([\d.]+)`).FindStringSubmatch(line); matches != nil {
			version = matches[1]
		} else if matches := regexp.MustCompile(`PATCHLEVEL = ([\d]+)`).FindStringSubmatch(line); matches != nil {
			version = version + "." + matches[1]
		}
	}
	return version
}

func getSusePlatform(contents []string) string {
	c := strings.ToLower(strings.Join(contents, ""))
	if strings.Contains(c, "opensuse") {
		return "opensuse"
	}
	return "suse"
}

func GetVirtualization() (string, string, error) {
	var system string
	var role string

	if common.PathExists("/proc/xen") {
		system = "xen"
		role = "guest" // assume guest

		if common.PathExists("/proc/xen/capabilities") {
			contents, err := common.ReadLines("/proc/xen/capabilities")
			if err == nil {
				if common.StringsHas(contents, "control_d") {
					role = "host"
				}
			}
		}
	}
	if common.PathExists("/proc/modules") {
		contents, err := common.ReadLines("/proc/modules")
		if err == nil {
			if common.StringsContains(contents, "kvm") {
				system = "kvm"
				role = "host"
			} else if common.StringsContains(contents, "vboxdrv") {
				system = "vbox"
				role = "host"
			} else if common.StringsContains(contents, "vboxguest") {
				system = "vbox"
				role = "guest"
			}
		}
	}

	if common.PathExists("/proc/cpuinfo") {
		contents, err := common.ReadLines("/proc/cpuinfo")
		if err == nil {
			if common.StringsHas(contents, "QEMU Virtual CPU") ||
				common.StringsHas(contents, "Common KVM processor") ||
				common.StringsHas(contents, "Common 32-bit KVM processor") {
				system = "kvm"
				role = "guest"
			}
		}
	}

	if common.PathExists("/proc/bc/0") {
		system = "openvz"
		role = "host"
	} else if common.PathExists("/proc/vz") {
		system = "openvz"
		role = "guest"
	}

	// not use dmidecode because it requires root

	if common.PathExists("/proc/self/status") {
		contents, err := common.ReadLines("/proc/self/status")
		if err == nil {

			if common.StringsHas(contents, "s_context:") ||
				common.StringsHas(contents, "VxID:") {
				system = "linux-vserver"
			}
			// TODO: guest or host
		}
	}

	if common.PathExists("/proc/self/cgroup") {
		contents, err := common.ReadLines("/proc/self/cgroup")
		if err == nil {
			if common.StringsHas(contents, "lxc") ||
				common.StringsHas(contents, "docker") {
				system = "lxc"
				role = "guest"
			} else if common.PathExists("/usr/bin/lxc-version") { // TODO: which
				system = "lxc"
				role = "host"
			}
		}
	}

	return system, role, nil
}
