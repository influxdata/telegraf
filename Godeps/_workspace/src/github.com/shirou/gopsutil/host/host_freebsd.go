// +build freebsd

package host

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	common "github.com/shirou/gopsutil/common"
)

const (
	UTNameSize = 16 /* see MAXLOGNAME in <sys/param.h> */
	UTLineSize = 8
	UTHostSize = 16
)

func HostInfo() (*HostInfoStat, error) {
	ret := &HostInfoStat{
		OS:             runtime.GOOS,
		PlatformFamily: "freebsd",
	}

	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}
	ret.Hostname = hostname

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

	values, err := common.DoSysctrl("kern.boottime")
	if err == nil {
		// ex: { sec = 1392261637, usec = 627534 } Thu Feb 13 12:20:37 2014
		v := strings.Replace(values[2], ",", "", 1)
		t, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return ret, err
		}
		ret.Uptime = t
	}

	return ret, nil
}

func BootTime() (int64, error) {
	values, err := common.DoSysctrl("kern.boottime")
	if err != nil {
		return 0, err
	}
	// ex: { sec = 1392261637, usec = 627534 } Thu Feb 13 12:20:37 2014
	v := strings.Replace(values[2], ",", "", 1)

	boottime, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}

	return boottime, nil
}

func Users() ([]UserStat, error) {
	utmpfile := "/var/run/utx.active"
	if !common.PathExists(utmpfile) {
		utmpfile = "/var/run/utmp" // before 9.0
		return getUsersFromUtmp(utmpfile)
	}

	var ret []UserStat
	file, err := os.Open(utmpfile)
	if err != nil {
		return ret, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return ret, err
	}

	u := Utmpx{}
	entrySize := int(unsafe.Sizeof(u)) - 3
	entrySize = 197 // TODO: why should 197
	count := len(buf) / entrySize

	for i := 0; i < count; i++ {
		b := buf[i*entrySize : i*entrySize+entrySize]
		var u Utmpx
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.LittleEndian, &u)
		if err != nil || u.Type != 4 {
			continue
		}
		sec := (binary.LittleEndian.Uint32(u.Tv.Sec[:])) / 2 // TODO:
		user := UserStat{
			User:     common.IntToString(u.User[:]),
			Terminal: common.IntToString(u.Line[:]),
			Host:     common.IntToString(u.Host[:]),
			Started:  int(sec),
		}

		ret = append(ret, user)
	}

	return ret, nil

}

func GetPlatformInformation() (string, string, string, error) {
	platform := ""
	family := ""
	version := ""

	out, err := exec.Command("uname", "-s").Output()
	if err == nil {
		platform = strings.ToLower(strings.TrimSpace(string(out)))
	}

	out, err = exec.Command("uname", "-r").Output()
	if err == nil {
		version = strings.ToLower(strings.TrimSpace(string(out)))
	}

	return platform, family, version, nil
}

func GetVirtualization() (string, string, error) {
	system := ""
	role := ""

	return system, role, nil
}

// before 9.0
func getUsersFromUtmp(utmpfile string) ([]UserStat, error) {
	var ret []UserStat
	file, err := os.Open(utmpfile)
	if err != nil {
		return ret, err
	}
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return ret, err
	}

	u := Utmp{}
	entrySize := int(unsafe.Sizeof(u))
	count := len(buf) / entrySize

	for i := 0; i < count; i++ {
		b := buf[i*entrySize : i*entrySize+entrySize]
		var u Utmp
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.LittleEndian, &u)
		if err != nil || u.Time == 0 {
			continue
		}
		user := UserStat{
			User:     common.IntToString(u.Name[:]),
			Terminal: common.IntToString(u.Line[:]),
			Host:     common.IntToString(u.Host[:]),
			Started:  int(u.Time),
		}

		ret = append(ret, user)
	}

	return ret, nil
}
