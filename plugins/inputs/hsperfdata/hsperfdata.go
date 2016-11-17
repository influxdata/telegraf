package hsperfdata

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/tokuhirom/go-hsperfdata/hsperfdata"
)

type Hsperfdata struct {
	User string
	Tags []string
}

var sampleConfig = `
  ## Use the hsperfdata directory belonging to a different user.
  # user = "root"
  #
  ## Use the value for these keys in the hsperfdata as tags, not fields. By
  ## default everything is a field.
  # tags = ["sun.rt.jvmVersion"]
`

func (n *Hsperfdata) SampleConfig() string {
	return sampleConfig
}

func (n *Hsperfdata) Repo() (*hsperfdata.Repository, error) {
	if n.User == "" {
		return hsperfdata.New()
	} else {
		return hsperfdata.NewUser(n.User)
	}
}

func (n *Hsperfdata) Gather(acc telegraf.Accumulator) error {
	repo, err := n.Repo()
	if err != nil {
		return err
	}

	files, err := repo.GetFiles()
	if err != nil {
		// the directory doesn't exist - so there aren't any Java processes running
		return nil
	}

	for _, file := range files {
		result, err := file.Read()
		if err != nil {
			return err
		}

		tags := map[string]string{"pid": file.GetPid()}
		fields := result.GetMap()

		procname := result.GetProcName()
		if procname != "" {
			tags["procname"] = procname
		}

		for _, tag := range n.Tags {
			// don't tag metrics with "nil", just skip the tag if it's not there
			if value, ok := fields[tag]; ok {
				if valuestr, ok := value.(string); ok {
					tags[tag] = valuestr
				} else {
					tags[tag] = fmt.Sprintf("%#v", fields[tag])
				}
				delete(fields, tag)
			}
		}

		acc.AddFields("java", fields, tags)
	}

	return nil
}

func (n *Hsperfdata) Description() string {
	return "Read performance data from running hotspot JVMs from shared memory"
}

func init() {
	inputs.Add("hsperfdata", func() telegraf.Input {
		return &Hsperfdata{}
	})
}
