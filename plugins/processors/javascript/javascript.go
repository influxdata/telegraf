package javascript

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/robertkrimen/otto"
)

var sampleConfig = `
  ## Code to run
  ##
  ## Path to JavaScript file or raw code string to run in the VM.
  ## If you put raw code, make sure to prepend "// javascript" as the first line of the code.
  code = "/path/to/file.js"

  ## Tags to set
  ##
  ## Set which tags from the metric to be set as variable
  ## in JavaScript VM.
  ## If tag is not found, it will be ignored and an error message will appear
  set_tags = [
	"mytag",
	"myothertag"
  ]

  ## Fields to set
  ##
  ## Set which fields from the metric to be set as variable
  ## in JavaScript VM.
  ## If field is not found, it will be ignored and an error message will appear.
  set_fields = [
	"myfield",
	"myotherfield"
  ]
	
  ## Tags to get
  ##
  ## Set to your target variable name, it must be a string
  ## and will replace original tags in the metric with the new values.
  ## If tag is not found, it will be ignored and an error message will appear.
  get_tags = [
	"mytag",
	"myothertag"
  ]

  ## Fields to get
  ##
  ## Consider JSON.stringify the variable to get if
  ## the value is not a boolean, string, float, or integer.
  ## It will be automatically unmarshalled into data type of your choice.
  ## Allowed data_type are:
  ## boolean, string, float, and integer.
  ## The value will then be resetted into the metric, replacing the original one.
  ## If field is not found, it will be ignored and an error message will appear.
  [[processors.javascript.get_fields]]
	name = "myfield",
	data_type = "string"
  [[processors.javascript.get_fields]]
	name = "myotherfield",
	data_type = "integer"
`

type JavaScript struct {
	Code      string      `toml:"code"`
	SetFields []string    `toml:"set_fields,omitempty"`
	SetTags   []string    `toml:"set_tags,omitempty"`
	GetFields []*Variable `toml:"get_fields,omitempty"`
	GetTags   []string    `toml:"get_tags,omitempty"`

	initialized bool
	vm          *otto.Otto
}

type Variable struct {
	Name     string `toml:"name"`
	DataType string `toml:"data_type"`
}

func (j *JavaScript) SampleConfig() string {
	return sampleConfig
}

func (j *JavaScript) Description() string {
	return "Process values by using JavaScript"
}

func (j *JavaScript) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	if !j.initialized {
		j.vm = otto.New()

		if !strings.HasPrefix(j.Code, "// javascript") {
			// load the code
			data, err := ioutil.ReadFile(j.Code)
			if err != nil {
				logPrintf("Error while reading code file %s: %v", j.Code, err)
				return metrics
			}
			j.Code = string(data)
		}
	}
	j.initialized = true

	var err error

	// set selected tags and fields
	for i, metric := range metrics {
		// tags first
		err = setTags(metric, j.SetTags, j.vm)
		if err != nil {
			logPrintf("%v", err)
			return metrics
		}

		// then fields
		err = setFields(metric, j.SetFields, j.vm)
		if err != nil {
			logPrintf("%v", err)
			return metrics
		}

		_, err = j.vm.Run(j.Code)
		if err != nil {
			logPrintf("error while running vm: %s", err)
			return metrics
		}

		// get selected tags and fields
		// and put it back, transformed
		metric, err = getTags(metric, j.GetTags, j.vm)
		if err != nil {
			logPrintf("%v", err)
			return metrics
		}

		metric, err = getFields(metric, j.GetFields, j.vm)
		if err != nil {
			logPrintf("%v", err)
			return metrics
		}

		metrics[i] = metric
	}

	return metrics
}

func setTags(metric telegraf.Metric, setTagKeys []string, jsvm *otto.Otto) (err error) {
	for _, selectedTag := range setTagKeys {
		tag, ok := metric.GetTag(selectedTag)
		if !ok {
			logPrintf("tag %s not found in the metric", selectedTag)
			continue
		}

		// and set to vm
		err = jsvm.Set(selectedTag, tag)
		if err != nil {
			err = fmt.Errorf("error while setting %s tag to VM: %v", selectedTag, err)
			continue
		}
	}

	return
}

func setFields(metric telegraf.Metric, setFieldKeys []string, jsvm *otto.Otto) (err error) {
	for _, selectedField := range setFieldKeys {
		field, ok := metric.GetField(selectedField)
		if !ok {
			logPrintf("field %s not found in the metric", selectedField)
			continue
		}

		// and set to vm
		err = jsvm.Set(selectedField, field)
		if err != nil {
			err = fmt.Errorf("error while setting %s field to VM: %v", selectedField, err)
			continue
		}
	}

	return
}

func getTags(metric telegraf.Metric, getTagKeys []string, jsvm *otto.Otto) (newMetric telegraf.Metric, err error) {
	newMetric = metric

	var val otto.Value
	for _, selectedTag := range getTagKeys {
		val, err = jsvm.Get(selectedTag)
		if err != nil {
			logPrintf("error while getting %s tag: %v", selectedTag, err)
			continue
		}
		if !val.IsDefined() {
			logPrintf("tag %s is undefined", selectedTag)
			continue
		}

		var tag string
		tag, err = val.ToString()
		if err != nil {
			err = fmt.Errorf("error while cast %s tag back to string: %v", selectedTag, err)
			continue
		}

		newMetric.RemoveTag(selectedTag)
		newMetric.AddTag(selectedTag, tag)
	}

	return
}

func getFields(metric telegraf.Metric, getFieldKeyVals []*Variable, jsvm *otto.Otto) (newMetric telegraf.Metric, err error) {
	newMetric = metric

	for _, selectedField := range getFieldKeyVals {
		var val otto.Value
		val, err = jsvm.Get(selectedField.Name)
		if err != nil {
			logPrintf("error while getting %s field: %v", selectedField.Name, err)
			continue
		}
		if !val.IsDefined() {
			logPrintf("field %s is undefined", selectedField.Name)
			continue
		}

		switch selectedField.DataType {
		case "boolean":
			var fieldData bool
			fieldData, err = val.ToBoolean()
			if err != nil {
				err = fmt.Errorf("error while cast %s field back to boolean: %v", selectedField.Name, err)
				continue
			}

			newMetric.RemoveField(selectedField.Name)
			newMetric.AddField(selectedField.Name, fieldData)
		case "string":
			var fieldData string
			fieldData, err = val.ToString()
			if err != nil {
				err = fmt.Errorf("error while cast %s field back to string: %v", selectedField.Name, err)
				continue
			}

			newMetric.RemoveField(selectedField.Name)
			newMetric.AddField(selectedField.Name, fieldData)
		case "float":
			var fieldData float64
			fieldData, err = val.ToFloat()
			if err != nil {
				err = fmt.Errorf("error while cast %s field back to float: %v", selectedField.Name, err)
				continue
			}

			newMetric.RemoveField(selectedField.Name)
			newMetric.AddField(selectedField.Name, fieldData)
		case "integer":
			var fieldData int64
			fieldData, err = val.ToInteger()
			if err != nil {
				err = fmt.Errorf("error while cast %s field back to integer: %v", selectedField.Name, err)
				continue
			}

			newMetric.RemoveField(selectedField.Name)
			newMetric.AddField(selectedField.Name, fieldData)
		default:
			err = fmt.Errorf("invalid get_fields %s data type %s", selectedField.Name, selectedField.DataType)
			continue
		}
	}

	return
}

func logPrintf(format string, v ...interface{}) {
	log.Printf("E! [processors.javascript] "+format, v...)
}

func init() {
	processors.Add("javascript", func() telegraf.Processor {
		return &JavaScript{}
	})
}
