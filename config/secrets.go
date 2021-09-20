package config

import (
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf/internal/choice"
)

// secretPattern is a regex to extract references to secrets stored in a secret-store.
var secretPattern = regexp.MustCompile(`@\{(\w+:\w+)\}`)

func (c *Config) replaceSecrets(pluginType string, plugin interface{}) {
	walkPluginStruct(reflect.ValueOf(plugin), func(f reflect.StructField, fv reflect.Value) {
		c.replaceFieldSecret(pluginType, f, fv)
	})
}

func (c *Config) replaceFieldSecret(pluginType string, field reflect.StructField, value reflect.Value) {
	tags := strings.Split(field.Tag.Get("telegraf"), ",")
	if !choice.Contains("secret", tags) {
		return
	}

	// We only support string replacement
	if value.Kind() != reflect.String {
		log.Printf("W! [secretstore] unsupported type %q for field %q of %q", field.Type.Kind().String(), field.Name, pluginType)
		return
	}

	// Secret references are in the form @{<store name>:<keyname>}
	matches := secretPattern.FindStringSubmatch(value.String())
	if len(matches) < 2 {
		return
	}

	// There should _ALWAYS_ be two parts due to the regular expression match
	parts := strings.SplitN(matches[1], ":", 2)
	_ = parts[0] // Ignore the storename for now. This is in preparation for using multiple stores
	keyname := parts[1]

	log.Printf("I! [secretstore] Replacing secret %q in %q of %q...", keyname, field.Name, pluginType)
	secret, err := c.SecretStore.Get(keyname)
	if err != nil {
		log.Printf("E! [secretstore] Retrieving secret for %q of %q failed: %v", field.Name, pluginType, err)
		return
	}
	value.SetString(secret)
}

func walkPluginStruct(value reflect.Value, fn func(f reflect.StructField, fv reflect.Value)) {
	v := reflect.Indirect(value)
	t := v.Type()
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			if field.PkgPath != "" {
				continue
			}
			switch field.Type.Kind() {
			case reflect.Struct:
				walkPluginStruct(fieldValue, fn)

			case reflect.Array, reflect.Slice:
				for j := 0; j < fieldValue.Len(); j++ {
					fn(field, fieldValue.Index(j))
				}
			case reflect.Map:
				iter := fieldValue.MapRange()
				for iter.Next() {
					fn(field, iter.Value())
				}
			default:
				fn(field, fieldValue)
			}
		}
	}
}
