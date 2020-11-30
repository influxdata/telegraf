package config

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// API is the main interaction with this package
// This is set by telegraf when it loads up.
var API *api

// api is the general interface to interacting with Telegraf's current config
type api struct {
	config *Config
}

func newAPI(c *Config) *api {
	return &api{config: c}
}

// PluginConfig is a plugin name and details about the config fields.
type PluginConfig struct {
	Name   string
	Config map[string]FieldConfig
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

// PluginID is the random id assigned to the plugin so it can be referenced later
type PluginID string

// PluginState describes what the instantiated plugin is currently doing
type PluginState string

const (
	PluginStateCreated  PluginState = "created"
	PluginStateStarting PluginState = "starting"
	PluginStateRunning  PluginState = "running"
	PluginStateStopping PluginState = "stopping"
	PluginStateDead     PluginState = "" // or unknown
)

// Plugin is an instance of a plugin running with a specific configuration
type Plugin struct {
	ID   string
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

func (a *api) ListRunningPlugins() []Plugin {

	return []Plugin{}
}

func (a *api) CreatePlugin(config PluginConfig) PluginID {

	return PluginID("")
}

func (a *api) GetPluginStatus(ID PluginID) PluginState {

	return PluginState("")
}

func (a *api) DeletePlugin(ID PluginID) {

}

// func (a *API) PausePlugin(ID PluginID) {

// }

// func (a *API) ResumePlugin(ID PluginID) {

// }

var count = 0

// getFieldConfig expects a PluginDescriber and a map to populate.
// it calls itself recursively so p must be an interface{}
func getFieldConfig(p interface{}, cfg map[string]FieldConfig) {
	count++
	fmt.Println(count)
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
		fmt.Println(ft.Name)

		ftType := ft.Type
		if ftType.Kind() == reflect.Ptr {
			ftType = ftType.Elem()
			// f = f.Elem()
		}

		// check if it's not exported, and skip if so.
		// fmt.Println(ft.Name)
		if len(ft.Name) > 0 && strings.ToLower(string(ft.Name[0])) == string(ft.Name[0]) {
			fmt.Println("Skipped unexported field ", ft.Name)
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
		// fmt.Println("ftType.Kind() ", ftType.Kind().String())
		if ftType.Kind() == reflect.Struct && !isInternalStructFieldType(ft.Type) {
			// fmt.Println("plugin", structType.Name(), "name", ft.Name)
			if ft.Anonymous { // embedded
				t := getSubTypeType(ft)
				i := reflect.New(t)
				// fmt.Println("Anonymous ", t.Name())
				getFieldConfig(i.Interface(), cfg)
			} else {
				subCfg := map[string]FieldConfig{}
				t := getSubTypeType(ft)
				i := reflect.New(t)
				// fmt.Println("Named ", t.Name())
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

			// fmt.Println(t.Name())
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

// Invalid Kind = iota
// reflect.Uintptr
// reflect.Complex64
// reflect.Complex128
// reflect.Array
// reflect.Chan
// reflect.Func
// reflect.Interface
// reflect.Ptr
// reflect.Slice
// reflect.String
// reflect.Struct
// reflect.UnsafePointer
