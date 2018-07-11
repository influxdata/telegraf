package linux_sysctl_fs

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"

	"path"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// https://www.kernel.org/doc/Documentation/sysctl/fs.txt
type SysctlFS struct {
	path string
}

var sysctlFSDescription = `Provides Linux sysctl fs metrics`
var sysctlFSSampleConfig = ``

func (_ SysctlFS) Description() string {
	return sysctlFSDescription
}
func (_ SysctlFS) SampleConfig() string {
	return sysctlFSSampleConfig
}

func (sfs *SysctlFS) gatherList(file string, fields map[string]interface{}, fieldNames ...string) error {
	bs, err := ioutil.ReadFile(sfs.path + "/" + file)
	if err != nil {
		return err
	}

	bsplit := bytes.Split(bytes.TrimRight(bs, "\n"), []byte{'\t'})
	for i, name := range fieldNames {
		if i >= len(bsplit) {
			break
		}
		if name == "" {
			continue
		}

		v, err := strconv.ParseUint(string(bsplit[i]), 10, 64)
		if err != nil {
			return err
		}
		fields[name] = v
	}

	return nil
}

func (sfs *SysctlFS) gatherOne(name string, fields map[string]interface{}) error {
	bs, err := ioutil.ReadFile(sfs.path + "/" + name)
	if err != nil {
		return err
	}

	v, err := strconv.ParseUint(string(bytes.TrimRight(bs, "\n")), 10, 64)
	if err != nil {
		return err
	}

	fields[name] = v
	return nil
}

func (sfs *SysctlFS) Gather(acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}

	for _, n := range []string{"aio-nr", "aio-max-nr", "dquot-nr", "dquot-max", "super-nr", "super-max"} {
		sfs.gatherOne(n, fields)
	}

	sfs.gatherList("inode-state", fields, "inode-nr", "inode-free-nr", "inode-preshrink-nr")
	sfs.gatherList("dentry-state", fields, "dentry-nr", "dentry-unused-nr", "dentry-age-limit", "dentry-want-pages")
	sfs.gatherList("file-nr", fields, "file-nr", "", "file-max")

	acc.AddFields("linux_sysctl_fs", fields, nil)
	return nil
}

func GetHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

func init() {

	inputs.Add("linux_sysctl_fs", func() telegraf.Input {
		return &SysctlFS{
			path: path.Join(GetHostProc(), "/sys/fs"),
		}
	})
}
