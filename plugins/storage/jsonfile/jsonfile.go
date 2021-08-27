package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/storage"
)

type JSONFileStorage struct {
	Filename string `toml:"file"`
}

func (s *JSONFileStorage) Init() error {
	if len(s.Filename) == 0 {
		return fmt.Errorf("Storage service requires filename")
	}
	return nil
}

func (s *JSONFileStorage) Close() error {
	return nil
}

func (s *JSONFileStorage) Load(namespace, key string, obj interface{}) error {
	f, err := os.Open(s.Filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	dec := json.NewDecoder(f)
	m := map[string]interface{}{}
	err = dec.Decode(&m)
	if err != nil {
		return err
	}
	if v, ok := m[namespace]; ok {
		m = v.(map[string]interface{})
	}
	if v, ok := m[key]; ok {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, obj)
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}

func (s *JSONFileStorage) Save(namespace, key string, value interface{}) error {
	m := map[string]interface{}{}
	m[namespace] = map[string]interface{}{
		key: value,
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Filename, data, 0600)
}

func (s *JSONFileStorage) GetName() string {
	return "jsonfile"
}

func init() {
	storage.Add("jsonfile", func() config.StoragePlugin {
		return &JSONFileStorage{}
	})
}
