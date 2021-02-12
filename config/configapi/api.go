package configapi

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/alecthomas/units"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
)

// API is the main interaction with this package
// This is set by telegraf when it loads up.
var API *api

// api is the general interface to interacting with Telegraf's current config
type api struct {
	agent  *agent.Agent
	config *config.Config

	// api shutdown context
	ctx       context.Context
	outputCtx context.Context
}

func newAPI(ctx context.Context, agent *agent.Agent) (_ *api, outputCancel context.CancelFunc) {
	API = &api{
		config: agent.Config,
		agent:  agent,
		ctx:    ctx,
	}
	API.outputCtx, outputCancel = context.WithCancel(context.Background())
	return API, outputCancel
}

// PluginConfig is a plugin name and details about the config fields.
type PluginConfig struct {
	Name   string
	Config map[string]FieldConfig
}

type PluginConfigCreate struct {
	Name   string
	Config map[string]interface{} // map field name to field value
}

// FieldConfig describes a single field
type FieldConfig struct {
	Type      FieldType              `json:"type,omitempty"`       // see FieldType
	Default   interface{}            `json:"default,omitempty"`    // whatever the default value is
	Format    string                 `json:"format,omitempty"`     // type-specific format info. eg a url is a string, but has url-formatting rules.
	Required  bool                   `json:"required,omitempty"`   // this is sort of validation, which I'm not sure belongs here.
	SubType   FieldType              `json:"sub_type,omitempty"`   // The subtype. map[string]int subtype is int. []string subtype is string.
	SubFields map[string]FieldConfig `json:"sub_fields,omitempty"` // only for struct/object/FieldConfig types
}

// FieldType enumerable type. Describes config field type information to external applications
type FieldType string

// FieldTypes
const (
	FieldTypeUnknown     FieldType = ""
	FieldTypeString      FieldType = "string"
	FieldTypeInteger     FieldType = "integer"
	FieldTypeDuration    FieldType = "duration" // a special case of integer
	FieldTypeSize        FieldType = "size"     // a special case of integer
	FieldTypeFloat       FieldType = "float"
	FieldTypeBool        FieldType = "bool"
	FieldTypeInterface   FieldType = "any"
	FieldTypeSlice       FieldType = "array"  // array
	FieldTypeFieldConfig FieldType = "object" // a FieldConfig?
	FieldTypeMap         FieldType = "map"    // always map[string]FieldConfig ?
)

// Plugin is an instance of a plugin running with a specific configuration
type Plugin struct {
	ID   models.PluginID
	Name string
	// State()
	Config map[string]FieldConfig
}

func (a *api) ListPluginTypes() []PluginConfig {
	result := []PluginConfig{}
	inputNames := []string{}
	for name := range inputs.Inputs {
		inputNames = append(inputNames, name)
	}
	sort.Strings(inputNames)

	for _, name := range inputNames {
		creator := inputs.Inputs[name]
		cfg := PluginConfig{
			Name:   name,
			Config: map[string]FieldConfig{},
		}

		p := creator()
		getFieldConfig(p, cfg.Config)

		result = append(result, cfg)
	}
	return result
}

func (a *api) ListRunningPlugins() (runningPlugins []Plugin) {
	if a == nil {
		panic("api is nil")
	}
	for _, v := range a.agent.RunningInputs() {
		p := Plugin{
			ID:     idToString(v.ID),
			Name:   v.Config.Name,
			Config: map[string]FieldConfig{},
		}
		getFieldConfig(v.Config, p.Config)
		runningPlugins = append(runningPlugins, p)
	}
	for _, v := range a.agent.RunningProcessors() {
		p := Plugin{
			ID:     idToString(v.GetID()),
			Name:   v.LogName(),
			Config: map[string]FieldConfig{},
		}
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		pluginCfg := val.FieldByName("Config").Interface()
		getFieldConfig(pluginCfg, p.Config)
		runningPlugins = append(runningPlugins, p)
	}
	for _, v := range a.agent.RunningOutputs() {
		p := Plugin{
			ID:     idToString(v.ID),
			Name:   v.Config.Name,
			Config: map[string]FieldConfig{},
		}
		getFieldConfig(v.Config, p.Config)
		runningPlugins = append(runningPlugins, p)

	}
	return runningPlugins
}

func (a *api) UpdatePlugin(ID models.PluginID, config PluginConfigCreate) error {
	// check config, call init.
	return nil
}

func (a *api) CreatePlugin(config PluginConfigCreate) (models.PluginID, error) {
	parts := strings.Split(config.Name, ".")
	pluginType, name := parts[0], parts[1]
	switch pluginType {
	case "inputs":
		// add an input
		input, ok := inputs.Inputs[name]
		if !ok {
			return models.PluginID(""), fmt.Errorf("Error finding plugin with name %s", name)
		}
		// create a copy
		i := input()
		// set the config
		if err := setFieldConfig(config.Config, i); err != nil {
			return models.PluginID(""), err
		}
		// start it and put it into the agent manager?
		pluginConfig := &models.InputConfig{Name: name}
		if err := setFieldConfig(config.Config, pluginConfig); err != nil {
			return models.PluginID(""), err
		}

		rp := models.NewRunningInput(i, pluginConfig)
		rp.SetDefaultTags(a.config.Tags)

		if err := rp.Init(); err != nil {
			return models.PluginID(""), fmt.Errorf("could not initialize plugin %w", err)
		}

		if err := a.agent.StartInput(rp); err != nil {
			return models.PluginID(""), fmt.Errorf("Could not start input: %w", err)
		}

		go func(rp *models.RunningInput) {
			a.agent.RunInput(rp, time.Now())
		}(rp)

		return idToString(rp.ID), nil
	case "outputs":
		// add an output
		output, ok := outputs.Outputs[name]
		if !ok {
			return models.PluginID(""), fmt.Errorf("Error finding plugin with name %s", name)
		}
		// create a copy
		o := output()
		// set the config
		if err := setFieldConfig(config.Config, o); err != nil {
			return models.PluginID(""), err
		}
		// start it and put it into the agent manager?
		pluginConfig := &models.OutputConfig{Name: name}
		if err := setFieldConfig(config.Config, pluginConfig); err != nil {
			return models.PluginID(""), err
		}

		ro := models.NewRunningOutput(o, pluginConfig, a.config.Agent.MetricBatchSize, a.config.Agent.MetricBufferLimit)

		if err := ro.Init(); err != nil {
			return models.PluginID(""), fmt.Errorf("could not initialize plugin %w", err)
		}

		if err := a.agent.StartOutput(ro); err != nil {
			return models.PluginID(""), fmt.Errorf("Could not start input: %w", err)
		}

		go func(ro *models.RunningOutput) {
			a.agent.RunOutput(a.outputCtx, ro)
		}(ro)

		return idToString(ro.ID), nil
	case "processors", "aggregators":
		processor, ok := processors.Processors[name]
		if !ok {
			return models.PluginID(""), fmt.Errorf("Error finding plugin with name %s", name)
		}
		// create a copy
		p := processor()
		// set the config
		if err := setFieldConfig(config.Config, p); err != nil {
			return models.PluginID(""), err
		}
		// start it and put it into the agent manager?
		pluginConfig := &models.ProcessorConfig{Name: name}
		if err := setFieldConfig(config.Config, pluginConfig); err != nil {
			return models.PluginID(""), err
		}

		rp := models.NewRunningProcessor(p, pluginConfig)

		if err := rp.Init(); err != nil {
			return models.PluginID(""), fmt.Errorf("could not initialize plugin %w", err)
		}

		if err := a.agent.StartProcessor(rp); err != nil {
			return models.PluginID(""), fmt.Errorf("Could not start input: %w", err)
		}

		go func(rp *models.RunningProcessor) {
			a.agent.RunProcessor(rp)
		}(rp)

		return idToString(rp.ID), nil
	default:
		return models.PluginID(""), errors.New("Unknown plugin type")
	}
}

func (a *api) GetPluginStatus(ID models.PluginID) models.PluginState {
	for _, v := range a.agent.RunningInputs() {
		if v.ID == ID.Uint64() {
			return v.GetState()
		}
	}
	for _, v := range a.agent.RunningProcessors() {
		if v.GetID() == ID.Uint64() {
			return v.GetState()
		}
	}
	for _, v := range a.agent.RunningOutputs() {
		if v.ID == ID.Uint64() {
			return v.GetState()
		}
	}
	return models.PluginState(0)
}

func (a *api) getPluginByID(ID models.PluginID) {

}

func (a *api) DeletePlugin(ID models.PluginID) error {
	for _, v := range a.agent.RunningInputs() {
		if v.ID == ID.Uint64() {
			a.agent.StopInput(v)
			return nil
		}
	}
	for _, v := range a.agent.RunningProcessors() {
		if v.GetID() == ID.Uint64() {
			a.agent.StopProcessor(v)
			return nil
		}
	}
	for _, v := range a.agent.RunningOutputs() {
		if v.ID == ID.Uint64() {
			a.agent.StopOutput(v)
			return nil
		}
	}
	return nil
}

// func (a *API) PausePlugin(ID models.PluginID) {

// }

// func (a *API) ResumePlugin(ID models.PluginID) {

// }

// setFieldConfig takes a map of field names to field values and sets them on the plugin
func setFieldConfig(cfg map[string]interface{}, p interface{}) error {
	destStruct := reflect.ValueOf(p)
	if destStruct.Kind() == reflect.Ptr {
		destStruct = destStruct.Elem()
	}
	keys := make([]string, 0, len(cfg))
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := cfg[k]
		destField, destFieldType := getFieldByName(destStruct, k) // get by tag
		if !destField.IsValid() {
			continue
		}
		if !destField.CanSet() {
			destField.Addr()
			// TODO: error?
			fmt.Println("cannot set", k, destFieldType.Name())
			continue
		}
		val := reflect.ValueOf(v)
		if err := setObject(val, destField, destFieldType); err != nil {
			return fmt.Errorf("Could not set field %s: %w", k, err)
		}

	}
	return nil
}

// getFieldByName gets a reference to a struct field from it's name, considering the tag names
func getFieldByName(destStruct reflect.Value, fieldName string) (reflect.Value, reflect.Type) {
	if destStruct.Kind() == reflect.Ptr {
		if destStruct.IsNil() {
			return reflect.ValueOf(nil), reflect.TypeOf(nil)
		}
		destStruct = destStruct.Elem()
	}
	// may be an interface to a struct
	if destStruct.Kind() == reflect.Interface {
		destStruct = destStruct.Elem()
	}
	destStructType := reflect.TypeOf(destStruct.Interface())
	for i := 0; i < destStruct.NumField(); i++ {
		field := destStruct.Field(i)
		fieldType := destStructType.Field(i)
		if fieldType.Type.Kind() == reflect.Struct && fieldType.Anonymous {
			v, t := getFieldByName(field, fieldName)
			if t != reflect.TypeOf(nil) {
				return v, t
			}
		}
		if fieldType.Tag.Get("toml") == fieldName {
			return field, fieldType.Type
		}
		if strings.ToLower(fieldType.Name) == fieldName && unicode.IsUpper(rune(fieldType.Name[0])) {
			return field, fieldType.Type
		}
		// handle snake string case conversion
	}
	return reflect.ValueOf(nil), reflect.TypeOf(nil)
}

// getFieldConfig expects a plugin, p, (of any type) and a map to populate.
// it calls itself recursively so p must be an interface{}
func getFieldConfig(p interface{}, cfg map[string]FieldConfig) {
	structVal := reflect.ValueOf(p)
	structType := structVal.Type()
	for structType.Kind() == reflect.Ptr {
		structVal = structVal.Elem()
		structType = structType.Elem()
	}

	// safety check.
	if structType.Kind() != reflect.Struct {
		// woah, what?
		panic(fmt.Sprintf("getFieldConfig expected a struct type, but got %v %v", p, structType.String()))
	}
	// structType.NumField()

	for i := 0; i < structType.NumField(); i++ {
		var f reflect.Value
		if !structVal.IsZero() {
			f = structVal.Field(i)
		}
		_ = f
		ft := structType.Field(i)

		ftType := ft.Type
		if ftType.Kind() == reflect.Ptr {
			ftType = ftType.Elem()
			// f = f.Elem()
		}

		// check if it's not exported, and skip if so.
		if len(ft.Name) > 0 && strings.ToLower(string(ft.Name[0])) == string(ft.Name[0]) {
			continue
		}
		tomlTag := ft.Tag.Get("toml")
		if tomlTag == "-" {
			continue
		}
		switch ftType.Kind() {
		case reflect.Func, reflect.Interface:
			continue
		}

		// if this field is a struct, get the structure of it.
		if ftType.Kind() == reflect.Struct && !isInternalStructFieldType(ft.Type) {
			if ft.Anonymous { // embedded
				t := getSubTypeType(ft)
				i := reflect.New(t)
				getFieldConfig(i.Interface(), cfg)
			} else {
				subCfg := map[string]FieldConfig{}
				t := getSubTypeType(ft)
				i := reflect.New(t)
				getFieldConfig(i.Interface(), subCfg)
				cfg[ft.Name] = FieldConfig{
					Type:      FieldTypeFieldConfig,
					SubFields: subCfg,
					SubType:   getFieldType(t),
				}
			}
			continue
		}

		// all other field types...
		fc := FieldConfig{
			Type:     getFieldTypeFromStructField(ft),
			Format:   ft.Tag.Get("format"),
			Required: ft.Tag.Get("required") == "true",
		}

		// set the default value for the field
		if f.IsValid() && !f.IsZero() {
			fc.Default = f.Interface()
			// special handling for internal struct types so the struct doesn't serialize to an object.
			if d, ok := fc.Default.(internal.Duration); ok {
				fc.Default = d.Duration
			}
			if s, ok := fc.Default.(internal.Size); ok {
				fc.Default = s.Size
			}
		}

		// if we found a slice of objects, get the structure of that object
		if hasSubType(ft.Type) {
			t := getSubTypeType(ft)
			n := t.Name()
			_ = n
			fc.SubType = getFieldType(t)

			if t.Kind() == reflect.Struct {
				i := reflect.New(t)
				subCfg := map[string]FieldConfig{}
				getFieldConfig(i.Interface(), subCfg)
				fc.SubFields = subCfg
			}
		}
		// if we found a map of objects, get the structure of that object

		cfg[ft.Name] = fc
	}
}

func setObject(from, to reflect.Value, destType reflect.Type) error {
	if from.Kind() == reflect.Interface {
		from = reflect.ValueOf(from.Interface())
	}
	// switch on source type
	switch from.Kind() {
	case reflect.Bool:
		if to.Kind() == reflect.Ptr {
			ptr := reflect.New(destType.Elem())
			to.Set(ptr)
			to = ptr.Elem()
		}
		if to.Kind() == reflect.Interface {
			to.Set(from)
		} else {
			to.SetBool(from.Bool())
		}
	case reflect.String:
		if to.Kind() == reflect.Ptr {
			ptr := reflect.New(destType.Elem())
			destType = destType.Elem()
			to.Set(ptr)
			to = ptr.Elem()
		}
		// consider duration and size types
		switch destType.String() {
		case "time.Duration", "config.Duration":
			d, err := time.ParseDuration(from.Interface().(string))
			if err != nil {
				return fmt.Errorf("Couldn't parse duration %q: %w", from.Interface().(string), err)
			}
			to.SetInt(int64(d))
		case "internal.Duration":
			d, err := time.ParseDuration(from.Interface().(string))
			if err != nil {
				return fmt.Errorf("Couldn't parse duration %q: %w", from.Interface().(string), err)
			}
			to.FieldByName("Duration").SetInt(int64(d))
		case "internal.Size":
			size, err := units.ParseStrictBytes(from.Interface().(string))
			if err != nil {
				return fmt.Errorf("Couldn't parse size %q: %w", from.Interface().(string), err)
			}
			to.FieldByName("Size").SetInt(size)
		case "config.Size":
			size, err := units.ParseStrictBytes(from.Interface().(string))
			if err != nil {
				return fmt.Errorf("Couldn't parse size %q: %w", from.Interface().(string), err)
			}
			to.SetInt(int64(size))
		// TODO: handle slice types?
		default:
			// to.SetString(from.Interface().(string))
			to.Set(from)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if to.Kind() == reflect.Ptr {
			ptr := reflect.New(destType.Elem())
			destType = destType.Elem()
			to.Set(ptr)
			to = ptr.Elem()
		}

		if destType.String() == "internal.Number" {
			n := internal.Number{Value: float64(from.Int())}
			to.Set(reflect.ValueOf(n))
			return nil
		}

		switch destType.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			to.SetUint(uint64(from.Int()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			to.SetInt(from.Int())
		case reflect.Float32, reflect.Float64:
			to.SetFloat(float64(from.Float()))
		case reflect.Interface:
			to.Set(from)
		default:
			return fmt.Errorf("cannot coerce int type into %s", destType.Kind().String())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if to.Kind() == reflect.Ptr {
			ptr := reflect.New(destType.Elem())
			// destType = destType.Elem()
			to.Set(ptr)
			to = ptr.Elem()
		}

		if destType.String() == "internal.Number" {
			n := internal.Number{Value: float64(from.Uint())}
			to.Set(reflect.ValueOf(n))
			return nil
		}

		if to.Kind() == reflect.Interface {
			to.Set(from)
		} else {
			to.SetUint(from.Uint())
		}
	case reflect.Float32, reflect.Float64:
		if to.Kind() == reflect.Ptr {
			ptr := reflect.New(destType.Elem())
			// destType = destType.Elem()
			to.Set(ptr)
			to = ptr.Elem()
		}
		if destType.String() == "internal.Number" {
			n := internal.Number{Value: from.Float()}
			to.Set(reflect.ValueOf(n))
			return nil
		}

		if to.Kind() == reflect.Interface {
			to.Set(from)
		} else {
			to.SetFloat(from.Float())
		}
	case reflect.Slice:
		if destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
			to = to.Elem()
		}
		if destType.Kind() != reflect.Slice {
			return fmt.Errorf("error setting slice field into %s", destType.Kind().String())
		}
		d := reflect.MakeSlice(destType, from.Len(), from.Len())
		for i := 0; i < from.Len(); i++ {
			if err := setObject(from.Index(i), d.Index(i), destType.Elem()); err != nil {
				return fmt.Errorf("couldn't set slice element: %w", err)
			}
		}
		to.Set(d)
	case reflect.Map:
		if destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
			ptr := reflect.New(destType)
			to.Set(ptr)
			to = to.Elem()
		}
		switch destType.Kind() {
		case reflect.Struct:
			structPtr := reflect.New(destType)
			err := setFieldConfig(from.Interface().(map[string]interface{}), structPtr.Interface())
			if err != nil {
				return err
			}
			to.Set(structPtr.Elem())
		case reflect.Map:
			//TODO: handle map[string]type
			if destType.Key().Kind() != reflect.String {
				panic("expecting string types for maps")
			}
			to.Set(reflect.MakeMap(destType))

			switch destType.Elem().Kind() {
			case reflect.Interface,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64,
				reflect.Bool:
				for _, k := range from.MapKeys() {
					t := from.MapIndex(k)
					if t.Kind() == reflect.Interface {
						t = reflect.ValueOf(t.Interface())
					}
					to.SetMapIndex(k, t)
				}
			case reflect.String:
				for _, k := range from.MapKeys() {
					t := from.MapIndex(k)
					if t.Kind() == reflect.Interface {
						t = reflect.ValueOf(t.Interface())
					}
					to.SetMapIndex(k, t)
				}
				// for _, k := range from.MapKeys() {
				// 	v := from.MapIndex(k)
				// 	s := v.Interface().(string)
				// 	to.SetMapIndex(k, reflect.ValueOf(s))
				// }
			case reflect.Slice:
				for _, k := range from.MapKeys() {
					// slice := reflect.MakeSlice(destType.Elem(), 0, 0)
					sliceptr := reflect.New(destType.Elem())
					// sliceptr.Elem().Set(slice)
					err := setObject(from.MapIndex(k), sliceptr, sliceptr.Type())
					if err != nil {
						return fmt.Errorf("could not set slice: %w", err)
					}
					to.SetMapIndex(k, sliceptr.Elem())
				}

			case reflect.Struct:
				for _, k := range from.MapKeys() {
					structPtr := reflect.New(destType.Elem())
					err := setFieldConfig(
						from.MapIndex(k).Interface().(map[string]interface{}),
						structPtr.Interface(),
					)
					// err := setObject(from.MapIndex(k), structPtr, structPtr.Type())
					if err != nil {
						return fmt.Errorf("could not set struct: %w", err)
					}
					to.SetMapIndex(k, structPtr.Elem())
				}

			default:
				return fmt.Errorf("can't write settings into map of type map[string]%s", destType.Elem().Kind().String())
			}
		default:
			return fmt.Errorf("Cannot load map into %s", destType.Kind().String())
			// panic("foo")
		}
		// to.Set(val)
	default:
		return fmt.Errorf("cannot convert unknown type %s to %s", from.Kind().String(), destType.String())
	}
	return nil
}

// hasSubType returns true when the field has a subtype (map,slice,struct)
func hasSubType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Map:
		return true
	case reflect.Struct:
		switch t.String() {
		case "internal.Duration", "config.Duration", "internal.Size", "config.Size":
			return false
		}
		return true
	default:
		return false
	}
}

// getSubTypeType gets the underlying subtype's reflect.Type
// examples:
//   []string => string
//   map[string]int => int
//   User => User
func getSubTypeType(structField reflect.StructField) reflect.Type {
	ft := structField.Type
	if ft.Kind() == reflect.Ptr {
		ft = ft.Elem()
	}
	switch ft.Kind() {
	case reflect.Slice:
		t := ft.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t
	case reflect.Map:
		t := ft.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t
	case reflect.Struct:
		return ft
	}
	panic(ft.String() + " is not a type that has subtype information (map, slice, struct)")
}

// getFieldType translates reflect.Types to our API field types.
func getFieldType(t reflect.Type) FieldType {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return FieldTypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return FieldTypeInteger
	case reflect.Float32, reflect.Float64:
		return FieldTypeFloat
	case reflect.Bool:
		return FieldTypeBool
	case reflect.Slice:
		return FieldTypeSlice
	case reflect.Map:
		return FieldTypeMap
	case reflect.Struct:
		switch t.String() {
		case "internal.Duration", "config.Duration":
			return FieldTypeDuration
		case "internal.Size", "config.Size":
			return FieldTypeSize
		}
		return FieldTypeFieldConfig
	}
	return FieldTypeUnknown
}

func getFieldTypeFromStructField(structField reflect.StructField) FieldType {
	fieldName := structField.Name
	ft := structField.Type
	result := getFieldType(ft)
	if result == FieldTypeUnknown {
		panic(fmt.Sprintf("unknown type, name: %q, string: %q", fieldName, ft.String()))
	}
	return result
}

func isInternalStructFieldType(t reflect.Type) bool {
	switch t.String() {
	case "internal.Duration", "config.Duration":
		return true
	case "internal.Size", "config.Size":
		return true
	default:
		return false
	}
}

func idToString(id uint64) models.PluginID {
	return models.PluginID(fmt.Sprintf("%016x", id))
}

var _ models.RunningPlugin = &models.RunningProcessor{}
var _ models.RunningPlugin = &models.RunningAggregator{}
var _ models.RunningPlugin = &models.RunningInput{}
var _ models.RunningPlugin = &models.RunningOutput{}
