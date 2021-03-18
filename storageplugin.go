package telegraf

// StoragePlugin is the interface to implement if you're building a plugin that implements state storage
type StoragePlugin interface {
	Init() error

	Load(namespace string) map[string]interface{}
	Save(namespace string, values map[string]interface{})

	LoadKey(namespace, key string) interface{}
	SaveKey(namespace, key string, value interface{})

	Close() error
}
