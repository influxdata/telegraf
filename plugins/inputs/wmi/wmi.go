package wmi

import (
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
	Type string
}

type WmiObject struct {
	Host      string
	Namespace string
	Table     string
	Something string
	Columns   []WmiColumnDef
	Where     string

	names []string
	typ   reflect.Type
}

type WmiInput struct {
	Object []WmiObject

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

func (mi *WmiMetricItem) Process(measurement string, dataIn reflect.Value, acc telegraf.Accumulator) error {
	if dataIn.Kind() == reflect.Ptr {
		mi.Process(measurement, dataIn.Elem(), acc)
	} else if dataIn.Kind() == reflect.Slice {
		for i := 0; i < dataIn.Len(); i++ {
			mi.Process(measurement, dataIn.Index(i), acc)
		}
	} else if dataIn.Kind() == reflect.Struct {
		mi.tags = make(map[string]string)
		mi.fields = make(map[string]interface{})
		for i := 0; i < dataIn.NumField(); i++ {
			field := dataIn.Field(i)
			name := dataIn.Type().Field(i).Name
			mi.fields[name] = field.Interface()
			acc.AddFields(measurement, mi.fields, mi.tags)
		}
	}

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
	}

	return reflect.TypeOf(i)
}

func (w *WmiInput) ParseConfig() error {
	for i, object := range w.Object {
		var fields []reflect.StructField
		for _, columnDef := range object.Columns {
			fields = append(fields, reflect.StructField{
				Name: columnDef.Name,
				Type: getType(columnDef.Type),
			})
			w.Object[i].names = append(w.Object[i].names, columnDef.Name)
		}

		w.Object[i].typ = reflect.SliceOf(reflect.StructOf(fields))
	}

	return nil
}

func (w *WmiInput) Collect(acc telegraf.Accumulator, object WmiObject, pwg *sync.WaitGroup) error {
	defer pwg.Done()
	dst := reflect.New(object.typ)
	query := "SELECT " + strings.Join(object.names, ",") + " FROM " + object.Table + " " + object.Where
	err := wmi.Query(query, dst.Interface(), object.Host, object.Namespace)
	if err != nil {
		log.Println("ERROR [wmi.query]: ", err)
		return err
	}

	wmiItem := WmiMetricItem{}
	wmiItem.Process(object.Table, dst, acc)

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (w *WmiInput) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	if !w.configParsed {
		err := w.ParseConfig()
		w.configParsed = true
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	for _, object := range w.Object {
		wg.Add(1)
		go w.Collect(acc, object, &wg)
	}
	wg.Wait()

	return nil
}

func init() {
	inputs.Add("wmi", func() telegraf.Input { return &WmiInput{configParsed: false} })
}
