package linux_battery

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const (
	BATTSTATUS   = "/sys/class/power_supply/battery/status"
	BATTVOLTAGE  = "/sys/class/power_supply/battery/voltage_now"
	BATTCURRENT  = "/sys/class/power_supply/battery/current_now"
	BATTCAPACITY = "/sys/class/power_supply/battery/capacity"
	BATTHEALTH   = "/sys/class/power_supply/battery/health"
)

// env variable names
const (
	BATT_STATUS   = "BATTSTATUS"
	BATT_VOLTAGE  = "BATTVOLTAGE"
	BATT_CURRENT  = "BATTCURRENT"
	BATT_CAPACITY = "BATTCAPACITY"
	BATT_HEALTH   = "BATTHEALTH"
)

var sampleConfig = `
  ## command  for reading. If empty default path will be used:
  ## This can also be overridden with env variable, see README.
  battstatus = "/sys/class/power_supply/battery/status
  battvoltage = "/sys/class/power_supply/battery/voltage_now"
  battcurrent = "/sys/class/power_supply/battery/current_now"
  battcapacity = "/sys/class/power_supply/battery/capacity"
  batthealth = "/sys/class/power_supply/battery/health"
`

type Batt struct {
	BATSTAT string `toml:"battstatus"`
	BATVOLT string `toml:"battvoltage"`
	BATCUR  string `toml:"battcurrent"`
	BATCAP  string `toml:"battcapacity"`
	BATHEAL string `toml:"batthealth"`
}

func (ns *Batt) Description() string {
	return "Collect Battery Health Stats"
}

func (ns *Batt) SampleConfig() string {
	return sampleConfig
}

func (ns *Batt) Gather(acc telegraf.Accumulator) error {
	tags := map[string]string{}
	metrics := map[string]interface{}{}
	mp := reflect.ValueOf(ns).Elem()
	for i := 0; i < mp.NumField(); i++ {
		f := mp.Field(i).String()
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		s := string(b[:len(b)])
		field := strings.Split(f, "/")
		stat := field[len(field)-1]
		if stat == "status" || stat == "health" {
			metrics[stat] = s
		} else {
			v, err := strconv.ParseFloat(strings.Trim(s, "\n"), 64)

			if err != nil {
				return err
			}
			if stat == "voltage_now" {
				v = (v / 10000) * 0.01
			}
			metrics[stat] = v
		}
	}
	//tags["sensor"] = "battery"
	acc.AddFields("battery", metrics, tags)
	return nil
}

// loadPath can be used to read path firstly from config
// if it is empty then try read from env variables
func (ns *Batt) loadPath() {

	if ns.BATSTAT == "" {
		ns.BATSTAT = proc(BATTSTATUS, "")
	}
	if ns.BATVOLT == "" {
		ns.BATVOLT = proc(BATTVOLTAGE, "")
	}
	if ns.BATCUR == "" {
		ns.BATCUR = proc(BATTCURRENT, "")
	}
	if ns.BATCAP == "" {
		ns.BATCAP = proc(BATTCAPACITY, "")
	}
	if ns.BATHEAL == "" {
		ns.BATHEAL = proc(BATTHEALTH, "")
	}
}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	return env
}

func init() {
	if runtime.GOOS != "linux" {
		return
	}
	inputs.Add("linux_battery", func() telegraf.Input {
		return &Batt{}
	})
}
