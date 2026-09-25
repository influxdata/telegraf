//go:generate ../../../tools/readme_config_includer/generator
package puppetagent

import (
	_ "embed"
	"fmt"
	"os"
	"reflect"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type PuppetAgent struct {
	Location            string   `toml:"location"`
	AdditionalResources []string `toml:"additional_resources"`
}

type state struct {
	Events    event
	Resources resource
	Changes   change
	Time      time
	Version   version
}

type additionalState struct {
	Time map[string]any `yaml:"time"`
}

type event struct {
	Failure int64 `yaml:"failure"`
	Noop    int64 `yaml:"noop"`
	Total   int64 `yaml:"total"`
	Success int64 `yaml:"success"`
}

type resource struct {
	Changed          int64 `yaml:"changed"`
	CorrectiveChange int64 `yaml:"corrective_change"`
	Failed           int64 `yaml:"failed"`
	FailedToRestart  int64 `yaml:"failed_to_restart"`
	OutOfSync        int64 `yaml:"out_of_sync"`
	Restarted        int64 `yaml:"restarted"`
	Scheduled        int64 `yaml:"scheduled"`
	Skipped          int64 `yaml:"skipped"`
	Total            int64 `yaml:"total"`
}

type change struct {
	Total int64 `yaml:"total"`
}

type time struct {
	Anchor                float64 `yaml:"anchor"`
	CataLogApplication    float64 `yaml:"catalog_application"`
	ConfigRetrieval       float64 `yaml:"config_retrieval"`
	ConvertCatalog        float64 `yaml:"convert_catalog"`
	Cron                  float64 `yaml:"cron"`
	Exec                  float64 `yaml:"exec"`
	FactGeneration        float64 `yaml:"fact_generation"`
	File                  float64 `yaml:"file"`
	FileBucket            float64 `yaml:"filebucket"`
	Group                 float64 `yaml:"group"`
	LastRun               int64   `yaml:"last_run"`
	NodeRetrieval         float64 `yaml:"node_retrieval"`
	Notify                float64 `yaml:"notify"`
	Package               float64 `yaml:"package"`
	PluginSync            float64 `yaml:"plugin_sync"`
	Schedule              float64 `yaml:"schedule"`
	Service               float64 `yaml:"service"`
	SSHAuthorizedKey      float64 `yaml:"ssh_authorized_key"`
	Total                 float64 `yaml:"total"`
	TransactionEvaluation float64 `yaml:"transaction_evaluation"`
	User                  float64 `yaml:"user"`
}

type version struct {
	ConfigString string `yaml:"config"`
	Puppet       string `yaml:"puppet"`
}

func (*PuppetAgent) SampleConfig() string {
	return sampleConfig
}

func (pa *PuppetAgent) Gather(acc telegraf.Accumulator) error {
	if len(pa.Location) == 0 {
		pa.Location = "/var/lib/puppet/state/last_run_summary.yaml"
	}

	if _, err := os.Stat(pa.Location); err != nil {
		return err
	}

	contents, err := os.ReadFile(pa.Location)
	if err != nil {
		return err
	}

	var puppetState state

	err = yaml.Unmarshal(contents, &puppetState)
	if err != nil {
		return err
	}

	fields := structFields(&puppetState)

	if len(pa.AdditionalResources) > 0 {
		var additionalState additionalState
		if err := yaml.Unmarshal(contents, &additionalState); err != nil {
			return err
		}
		addAdditionalResources(additionalState.Time, pa.AdditionalResources, fields)
	}

	tags := map[string]string{"location": pa.Location}
	acc.AddFields("puppetagent", fields, tags)

	return nil
}

func structFields(s *state) map[string]any {
	e := reflect.ValueOf(s).Elem()

	fields := make(map[string]any)
	for tLevelFNum := 0; tLevelFNum < e.NumField(); tLevelFNum++ {
		name := e.Type().Field(tLevelFNum).Name
		nameNumField := e.FieldByName(name).NumField()

		for sLevelFNum := range nameNumField {
			sName := e.FieldByName(name).Type().Field(sLevelFNum).Name
			sValue := e.FieldByName(name).Field(sLevelFNum).Interface()

			lname := strings.ToLower(name)
			lsName := strings.ToLower(sName)
			fields[fmt.Sprintf("%s_%s", lname, lsName)] = sValue
		}
	}
	return fields
}

func addAdditionalResources(resources map[string]any, additional []string, fields map[string]any) {
	for _, resource := range additional {
		fieldName := "time_" + strings.ToLower(resource)
		if _, exists := fields[fieldName]; exists {
			continue
		}

		if value, exists := resources[resource]; exists {
			fields[fieldName] = value
		}
	}
}

func init() {
	inputs.Add("puppetagent", func() telegraf.Input {
		return &PuppetAgent{}
	})
}
