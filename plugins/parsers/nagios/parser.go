package nagios

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type NagiosParser struct {
	MetricName  string
	DefaultTags map[string]string
}

// Got from Alignak
// https://github.com/Alignak-monitoring/alignak/blob/develop/alignak/misc/perfdata.py
var perfSplitRegExp, _ = regexp.Compile(`([^=]+=\S+)`)
var nagiosRegExp, _ = regexp.Compile(`^([^=]+)=([\d\.\-\+eE]+)([\w\/%]*);?([\d\.\-\+eE:~@]+)?;?([\d\.\-\+eE:~@]+)?;?([\d\.\-\+eE]+)?;?([\d\.\-\+eE]+)?;?\s*`)

func (p *NagiosParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	return metrics[0], err
}

func (p *NagiosParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

//> rta,host=absol,unit=ms critical=6000,min=0,value=0.332,warning=4000 1456374625003628099
//> pl,host=absol,unit=% critical=90,min=0,value=0,warning=80 1456374625003693967

func (p *NagiosParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)
	// Convert to string
	out := string(buf)
	// Prepare output for splitting
	// Delete escaped pipes
	out = strings.Replace(out, `\|`, "___PROTECT_PIPE___", -1)
	// Split lines and get the first one
	lines := strings.Split(out, "\n")
	// Split output and perfdatas
	data_splitted := strings.Split(lines[0], "|")
	if len(data_splitted) <= 1 {
		// No pipe == no perf data
		return nil, nil
	}
	// Get perfdatas
	perfdatas := data_splitted[1]
	// Add escaped pipes
	perfdatas = strings.Replace(perfdatas, "___PROTECT_PIPE___", `\|`, -1)
	// Split perfs
	unParsedPerfs := perfSplitRegExp.FindAllSubmatch([]byte(perfdatas), -1)
	// Iterate on all perfs
	for _, unParsedPerfs := range unParsedPerfs {
		// Get metrics
		// Trim perf
		trimedPerf := strings.Trim(string(unParsedPerfs[0]), " ")
		// Parse perf
		perf := nagiosRegExp.FindAllSubmatch([]byte(trimedPerf), -1)
		// Bad string
		if len(perf) == 0 {
			continue
		}
		if len(perf[0]) <= 2 {
			continue
		}
		if perf[0][1] == nil || perf[0][2] == nil {
			continue
		}
		fieldName := string(perf[0][1])
		tags := make(map[string]string)
		if perf[0][3] != nil {
			str := string(perf[0][3])
			if str != "" {
				tags["unit"] = str
			}
		}
		fields := make(map[string]interface{})
		f, err := strconv.ParseFloat(string(perf[0][2]), 64)
		if err == nil {
			fields["value"] = f
		}
		// TODO should we set empty field
		// if metric if there is no data ?
		if perf[0][4] != nil {
			f, err := strconv.ParseFloat(string(perf[0][4]), 64)
			if err == nil {
				fields["warning"] = f
			}
		}
		if perf[0][5] != nil {
			f, err := strconv.ParseFloat(string(perf[0][5]), 64)
			if err == nil {
				fields["critical"] = f
			}
		}
		if perf[0][6] != nil {
			f, err := strconv.ParseFloat(string(perf[0][6]), 64)
			if err == nil {
				fields["min"] = f
			}
		}
		if perf[0][7] != nil {
			f, err := strconv.ParseFloat(string(perf[0][7]), 64)
			if err == nil {
				fields["max"] = f
			}
		}
		// Create metric
		metric, err := metric.New(fieldName, tags, fields, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		// Add Metric
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
