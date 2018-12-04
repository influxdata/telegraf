package wmi

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type WmiMetricItem struct {
	measurement string
	fields      map[string]interface{}
	tags        map[string]string
	timestamp   time.Time
}

type WmiColumnDef struct {
	Name string
	As   string
	Type string
}

type WmiObject struct {
	Host      string
	Namespace string
	User      string
	Password  string
	Table     string
	As        string
	Something string
	Columns   []WmiColumnDef
	Where     string

	measurement string
	cmap        map[string]WmiColumnDef
	names       []string
	dst         reflect.Value
}

type WmiInput struct {
	Object map[string][]*WmiObject

	service      *wmi.SWbemServices
	configParsed bool
}

// SampleConfig returns the default configuration of the Input
func (w *WmiInput) SampleConfig() string {
	return ""
}

// Description returns a one-sentence description on the Input
func (w *WmiInput) Description() string {
	return "Gathers data from WMI/WQL Requests"
}

func (mi *WmiMetricItem) Process(cmap map[string]WmiColumnDef, dataIn reflect.Value) error {
	if dataIn.Kind() == reflect.Ptr {
		mi.Process(cmap, dataIn.Elem())
	} else if dataIn.Kind() == reflect.Slice {
		for i := 0; i < dataIn.Len(); i++ {
			mi.Process(cmap, dataIn.Index(i))
		}
	} else if dataIn.Kind() == reflect.Struct {
		for i := 0; i < dataIn.NumField(); i++ {
			var typ string
			field := dataIn.Field(i)
			name := dataIn.Type().Field(i).Name
			fmt.Println(field.Type())
			if _, ok := cmap[name]; ok {
				typ = cmap[name].Type
				if len(cmap[name].As) > 0 {
					name = cmap[name].As
				}
			}

			if typ == "tag" {
				mi.tags[name] = field.String()
			} else {
				mi.fields[name] = field.Interface()
			}
		}
	}

	return nil
}

func (w *WmiInput) Process(measurement string, acc telegraf.Accumulator) error {
	var mi WmiMetricItem
	mi.tags = make(map[string]string)
	mi.fields = make(map[string]interface{})

	for _, object := range w.Object[measurement] {
		mi.Process(object.cmap, object.dst)
	}

	acc.AddFields(measurement, mi.fields, mi.tags)

	return nil
}

func getType(s string) reflect.Type {
	var i interface{}
	switch s {
	case "string":
		i = "string"
		break
	case "bool":
		i = bool(true)
	case "int":
		i = int(1)
		break
	case "uint8":
		i = uint8(1)
		break
	case "uint16":
		i = uint16(1)
		break
	case "uint32":
		i = uint32(1)
		break
	case "uint64":
		i = uint64(1)
		break
	case "int8":
		i = int8(1)
		break
	case "int16":
		i = int16(1)
		break
	case "int32":
		i = int32(1)
		break
	case "int64":
		i = int64(1)
		break
	case "float32":
		i = float32(1.0)
		break
	case "float64":
		i = float64(1.0)
		break
	case "tag":
		i = "string"
		break
	case "[]string":
		i = [2]string{"string", "string"}
		break
	}

	return reflect.TypeOf(i)
}

func (w *WmiInput) ConnectWmi() error {
	var err error
	w.service, err = wmi.InitializeSWbemServices(&wmi.Client{})
	if err != nil {
		return err
	}

	return nil
}

func (w *WmiInput) ParseConfig() error {
	for name, objArr := range w.Object {
		for _, obj := range objArr {
			obj.measurement = name
			obj.cmap = make(map[string]WmiColumnDef)
			var fields []reflect.StructField
			for _, columnDef := range obj.Columns {
				fields = append(fields, reflect.StructField{
					Name: columnDef.Name,
					Type: getType(columnDef.Type),
				})
				obj.names = append(obj.names, columnDef.Name)
				obj.cmap[columnDef.Name] = columnDef
			}

			typ := reflect.SliceOf(reflect.StructOf(fields))
			obj.dst = reflect.New(typ)
		}
	}

	return nil
}

func (wo *WmiObject) Query(pwg *sync.WaitGroup, wmi *wmi.SWbemServices) {
	defer pwg.Done()
	query := "SELECT " + strings.Join(wo.names, ",") + " FROM " + wo.Table + " " + wo.Where
	fmt.Println("Query", query, wo.Host, wo.Namespace)
	err := wmi.Query(query, wo.dst.Interface(), wo.Host, wo.Namespace, wo.User, wo.Password)
	fmt.Println("Result", wo.Host, wo.Namespace, wo.dst)
	if err != nil {
		log.Println("ERROR [wmi.query]: ", err)
	}
}

func (w *WmiInput) Collect(acc telegraf.Accumulator, measurement string, wmiObjects []*WmiObject) error {
	var wg sync.WaitGroup
	for i, _ := range wmiObjects {
		wo := wmiObjects[i]
		wg.Add(1)
		go wo.Query(&wg, w.service)
	}
	wg.Wait()

	err := w.Process(measurement, acc)
	if err != nil {
		return err
	}

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (w *WmiInput) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	if !w.configParsed {
		err := w.ParseConfig()
		if err != nil {
			return err
		}

		err = w.ConnectWmi()
		if err != nil {
			return err
		}

		w.configParsed = true
	}

	for measurement, object := range w.Object {
		w.Collect(acc, measurement, object)
	}

	return nil
}

func init() {
	inputs.Add("wmi", func() telegraf.Input { return &WmiInput{configParsed: false} })
}
