package config

import (
	"log" //nolint:revive  // This is a subpart of config which is allowed to have log imported.
	"reflect"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf/internal/choice"
)

const secretStoreConfig = `
# Store secrets like credentials using a service external to telegraf
# [[secretstore]]
  ## Name of the secret-store used to reference the secrets later via @{name:secret_key} (mandatory)
  name = secretstore

  ## Define the service for storing the credentials, can be one of
  ##     file://<path>
  ##       Encrypted file at the given "path" (mandatory) for storing the secrets.
  ##     kwallet://[[application]/folder]   (default: "kwallet://telegraf")
  ##       kWallet with the given "application" ID and an optional subfolder.
  ##     os://[collection]                  (default: "os://telegraf")
  ##       OS's native secret store with "collection" being the keychain/keyring name or Windows' credential prefix
  ##     secret-service://[collection]      (default: "secret-service://telegraf")
  ##       Freedesktop secret-service implementation.
  # service = "os://telegraf"
`

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
	storename := parts[0] // Ignore the storename for now. This is in preparation for using multiple stores
	keyname := parts[1]

	log.Printf("I! [secretstore] Replacing secret %q in %q of %q...", keyname, field.Name, pluginType)
	store, found := c.SecretStore[storename]
	if !found {
		log.Printf("E! [secretstore] Unknown store %q for secret %q of %q", storename, matches[1], pluginType)
		return
	}
	secret, err := store.Get(keyname)
	if err != nil {
		log.Printf("E! [secretstore] Retrieving secret %q in %q of %q failed: %v", matches[1], field.Name, pluginType, err)
		return
	}
	value.SetString(secret)
}

func walkPluginStruct(value reflect.Value, fn func(f reflect.StructField, fv reflect.Value)) {
	v := reflect.Indirect(value)
	t := v.Type()

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
