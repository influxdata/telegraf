package aurora

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Aurora struct {
	Timeout    int    `toml:"timeout"`
	Master     string `toml:"master"`
	HttpPrefix string `toml:"prefix"`
	Numeric    bool   `toml:"numeric"`
}

var sampleConfig = `
  ## Timeout, in ms.
  timeout = 100
  ## Aurora Master
  master = "localhost:8081"
  ## Http Prefix
  prefix = "http"
  ## Numeric values only
  numeric = true
`

// SampleConfig returns a sample configuration block
func (a *Aurora) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the Mesos plugin
func (a *Aurora) Description() string {
	return "Telegraf plugin for gathering metrics from N Apache Aurora Masters"
}

func (a *Aurora) SetDefaults() {
	if a.Timeout == 0 {
		log.Println("I! [aurora] Missing timeout value, setting default value (1000ms)")
		a.Timeout = 1000
	} else if a.HttpPrefix == "" {
		log.Println("I! [aurora] Missing http prefix value, setting default value (http)")
		a.HttpPrefix = "http"
	}
}


// Converts string values taken from aurora vars to numeric values for wavefront 
func convertToNumeric(value string) (interface{}, bool) {
	var err error
	var val interface{}
	if val, err = strconv.ParseFloat(value, 64); err == nil {
		return val, true
	}
	if val, err = strconv.ParseBool(value); err != nil {
		return val.(bool), false
	}
	return val, true
}

// Matches job keys like sla_role2/prod2/jobname2_job_uptime_50.00_sec
func isJobMetric(key string) bool {
	// Regex for matching job specific tasks
	re := regexp.MustCompile("^sla_(.*?)/(.*?)/.*")
	return re.MatchString(key)
}

// Checks if the job key starts with task_store indicating it's a task store metric
func isTaskStore(key string) bool {
	return strings.HasPrefix(key, "task_store_")
}

// Checks if the key is a framework_registered which is used to determine leader election
func isFramework(key string) bool {
	return strings.HasPrefix(key, "framework_registered")
}

// This function parses a job metric key like sla_role2/prod2/jobname2_job_uptime_50.00_sec
// It returns the fields and the tags associated with those fields
func parseJobSpecificMetric(key string, value interface{}) (map[string]interface{}, map[string]string) {
	// cut off the sla_
	key = key[4:]
	// We have previous checked if this is a job metric using isJobMetric so we know there will be 2 slashes
	slashSplit := strings.Split(key, "/")
	role := slashSplit[0]
	env := slashSplit[1]
	underscoreIdx := strings.Index(slashSplit[2], "_")
	job := slashSplit[2][:underscoreIdx]
	metric := slashSplit[2][underscoreIdx+1:]

	fields := make(map[string]interface{})
	fields[metric] = value

	tags := make(map[string]string)
	tags["role"] = role
	tags["env"] = env
	tags["job"] = job
	return fields, tags
}

// This function takes a metric like task_store_DRAINED and generates aurora.task.store.DRAINED
func parseTaskStore(key string, value interface{}) (map[string]interface{}, map[string]string) {
	metric := "task.store." + strings.Replace(key[len("task_store_"):], "_", ".", -1)
	fields := make(map[string]interface{})
	fields[metric] = value
	tags := make(map[string]string)
	return fields, tags
}

// This method parses the value out of the variable line which is always in the last place
// It returns the key, value, and error
func (a *Aurora) parseMetric(line string) (string, interface{}, error) {
	splitIdx := strings.Index(line, " ")
	if splitIdx == -1 {
		return "", nil, fmt.Errorf("Invalid metric line %s has no value", line)
	}
	key := line[:splitIdx]
	value := line[splitIdx+1:]
	// If numeric is true and the metric is not numeric then ignore
	numeric, isNumeric := convertToNumeric(value)
	if a.Numeric && !isNumeric {
		return "", nil, fmt.Errorf("Value is rejected due to being non-numeric")
	}
	return key, numeric, nil
}

// Gather() metrics from given list of Aurora Masters
func (a *Aurora) Gather(acc telegraf.Accumulator) error {
	a.SetDefaults()

	client := http.Client{
		Timeout: time.Duration(a.Timeout) * time.Second,
	}
	url := fmt.Sprintf("%s://%s/vars", a.HttpPrefix, a.Master)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if isJobMetric(line) {
			key, value, err := a.parseMetric(line)
			if err != nil {
				continue
			}
			fields, tags := parseJobSpecificMetric(key, value)
			// Per job there are different tags so need to add a field per line
			acc.AddFields("aurora", fields, tags)
		} else if isTaskStore(line) {
			key, value, err := a.parseMetric(line)
			if err != nil {
				continue
			}
			fields, tags := parseTaskStore(key, value)
			acc.AddFields("aurora", fields, tags)
		} else if isFramework(line) {
			key, value, err := a.parseMetric(line)
			if err != nil {
				continue
			}
			fields := map[string]interface{}{
				key: value,
			}
			tags := make(map[string]string)
			acc.AddFields("aurora", fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("aurora", func() telegraf.Input {
		return &Aurora{}
	})
}
