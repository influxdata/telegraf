package glusterfs

// glusterfs.go

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"bufio"
	"os/exec"
	"regexp"
  "strconv"
)

var matchBrick = regexp.MustCompile("^Brick: (.*)$")
var matchRead = regexp.MustCompile("Data Read: ([0-9]+) bytes$")
var matchWrite = regexp.MustCompile("Data Written: ([0-9]+) bytes$")

type GlusterFS struct {
	Volumes []string
}

func (gfs *GlusterFS) Description() string {
	return "Plugin reading values from the GlusterFS profiler"
}

func (gfs *GlusterFS) SampleConfig() string {
	return "volumes = [\"volume-name\"]"
}

func (gfs *GlusterFS) Gather(acc telegraf.Accumulator) error {
	for _, volume := range gfs.Volumes {
		var cmdName = "sudo"
		var cmdArgs = []string{"gluster", "volume", "profile", volume, "info", "cumulative"}

		cmd := exec.Command(cmdName, cmdArgs...)
		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
      continue
    }

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			var tags map[string]string
			for scanner.Scan() {
				var txt = scanner.Text()
				if brick := matchBrick.FindStringSubmatch(txt); brick != nil {
					tags = map[string]string{"volume": volume, "brick": brick[1]}
				} else if gread := matchRead.FindStringSubmatch(txt); gread != nil {
          var val, _ = strconv.Atoi(gread[1])
					acc.AddFields("glusterfs", map[string]interface{}{"read": val}, tags)
				} else if gwrite := matchWrite.FindStringSubmatch(txt); gwrite != nil {
          var val, _ = strconv.Atoi(gwrite[1])
					acc.AddFields("glusterfs", map[string]interface{}{"write": val}, tags)
				}
			}
		}()

		err = cmd.Start()
		if err != nil {
      continue
		}

		cmd.Wait()
	}
	return nil
}

func init() {
	inputs.Add("glusterfs", func() telegraf.Input { return &GlusterFS{} })
}
