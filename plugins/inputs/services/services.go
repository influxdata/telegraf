package services

import (
    /*
	"strconv"
    */
	"strings"
	"bufio"
	"time"
	"fmt"
	"bytes"
	"os/exec"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)


// Services is a telegraf plugin to gather services status from systemd and windows services
type Services struct {
    Timeout     internal.Duration
    systemctl   systemctlCall
}

type systemctlCall func(Timeout internal.Duration) (*bytes.Buffer, error)

const measurement = "services"

var defaultTimeout = internal.Duration{Duration: time.Second}

// Description returns a short description of the plugin
func (services *Services) Description() string {
	return "Gather service status for systemd units and windows services"
}

// SampleConfig returns sample configuration options.
func (services *Services) SampleConfig() string {
	return `
  ## The default timeout of 1s for systemctl execution can be overridden here:
  # timeout = "1s"
`
}

func (services *Services) Gather(acc telegraf.Accumulator) error {
    out, e := services.systemctl(services.Timeout)
	if e != nil {
		acc.AddError(e)
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()

		data := strings.Fields(line)
		if len(data) < 4 {
			acc.AddError(fmt.Errorf("Error parsing line (expected at least 4 fields): %s", line))
			continue
		}
        tags := map[string]string{
            "name": data[0],
        }
        var status int
        switch active := data[2]; active {
        case "active":
            status = 0 // ok
        case "inactive":
            status = 0 // ok
        case "failed":
            status = 2 // error
        default:
            status = 3 // unknown
        }
        fields := map[string]interface{}{
            "status": status,
        }
        acc.AddCounter(measurement, fields, tags)
    }

    return nil
}

func setSystemctlCall(Timeout internal.Duration) (*bytes.Buffer, error) {
    // is systemctl available ?
	systemctlPath, err := exec.LookPath("systemctl")
	if err != nil {
		return nil, err
	}

    cmd := exec.Command(systemctlPath, "list-units", "--type=service", "--no-legend")

    var out bytes.Buffer
    cmd.Stdout = &out
	err = internal.RunTimeout(cmd, Timeout.Duration)
	if err != nil {
		return &out, fmt.Errorf("error running systemctl list-units --type=service --no-legend: %s", err)
	}

	return &out, nil
}

func init() {
	inputs.Add("services", func() telegraf.Input {
		return &Services{
            systemctl: setSystemctlCall,
			Timeout: defaultTimeout,
		}
	})
}
