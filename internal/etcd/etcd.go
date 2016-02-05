package etcd

import (
	"bytes"
	"encoding/json"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/coreos/etcd/client"

	"github.com/influxdata/telegraf/internal/config"
)

type EtcdClient struct {
	Kapi   client.KeysAPI
	Folder string
}

func (e *EtcdClient) LaunchWatcher(shutdown chan struct{}, signals chan os.Signal) {
	// TODO: All telegraf client will reload for each changes...
	// Maybe we want to reload on those we need to ???
	// So we need to create a watcher by labels ??

	for {
		watcherOpts := client.WatcherOptions{AfterIndex: 0, Recursive: true}
		w := e.Kapi.Watcher(e.Folder, &watcherOpts)
		r, err := w.Next(context.Background())
		if err != nil {
			// TODO What we have to do here ????
			log.Fatal("Error occurred", err)
		}
		if r.Action == "set" || r.Action == "update" {
			// do something with Response r here
			log.Printf("Changes detected in etcd (%s action detected)\n", r.Action)
			log.Print("Reloading Telegraf config\n")
			signals <- syscall.SIGHUP
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}

func NewEtcdClient(urls string, folder string) *EtcdClient {
	// Create a new etcd client
	cfg := client.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: client.DefaultTransport,
	}

	e := &EtcdClient{}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	e.Kapi = kapi
	e.Folder = folder

	return e
}

func (e *EtcdClient) Connect() error {
	//c, err := eclient.New(cfg)
	return nil
}

func (e *EtcdClient) WriteConfigDir(configdir string) error {
	directoryEntries, err := ioutil.ReadDir(configdir)
	if err != nil {
		return err
	}
	for _, entry := range directoryEntries {
		name := entry.Name()
		if entry.IsDir() {
			if name == "labels" {
				// Handle labels
				directoryEntries, err := ioutil.ReadDir(path.Join(configdir, name))
				if err != nil {
					return err
				}
				for _, entry := range directoryEntries {
					filename := entry.Name()
					if len(filename) < 6 || filename[len(filename)-5:] != ".conf" {
						continue
					}
					label := filename[:len(filename)-5]
					err = e.WriteLabelConfig(label, path.Join(configdir, name, filename))
					if err != nil {
						return err
					}
				}
			} else if name == "hosts" {
				// Handle hosts specific config
				directoryEntries, err := ioutil.ReadDir(path.Join(configdir, name))
				if err != nil {
					return err
				}

				for _, entry := range directoryEntries {
					filename := entry.Name()
					if len(filename) < 6 || filename[len(filename)-5:] != ".conf" {
						continue
					}
					hostname := filename[:len(filename)-5]
					err = e.WriteHostConfig(hostname, path.Join(configdir, name, filename))
					if err != nil {
						return err
					}
				}
			}
			continue
		}
		if name == "main.conf" {
			// Handle main config
			err := e.WriteMainConfig(path.Join(configdir, name))
			if err != nil {
				return err
			}
		} else {
			continue
		}
	}

	return nil
}

func (e *EtcdClient) WriteMainConfig(path string) error {
	// Write main config file in etcd
	key := "main"
	err := e.WriteConfig(key, path)
	return err
}

func (e *EtcdClient) WriteLabelConfig(label string, path string) error {
	// Write label config file in etcd
	key := "labels/" + label
	err := e.WriteConfig(key, path)
	return err
}

func (e *EtcdClient) WriteHostConfig(host string, path string) error {
	// Write host config file in etcd
	key := "hosts/" + host
	err := e.WriteConfig(key, path)
	return err
}

func (e *EtcdClient) WriteConfig(relative_key string, path string) error {
	// Read config file, get conf in tomlformat, convert to json
	// Then write it to etcd
	// TODO: Maybe we just want to store toml in etcd ? Is json really needed ????
	// Read file
	raw_data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// Get toml
	var data interface{}
	_, err = toml.Decode(string(raw_data), &data)
	if err != nil {
		return err
	}
	// Get json
	json_data, _ := json.Marshal(&data)
	// Write it
	key := e.Folder + "/" + relative_key
	resp, _ := e.Kapi.Get(context.Background(), key, nil)
	if resp == nil {
		_, err = e.Kapi.Set(context.Background(), key, string(json_data), nil)
	} else {
		_, err = e.Kapi.Update(context.Background(), key, string(json_data))
	}
	if err != nil {
		log.Fatal(err)
		return err
	} else {
		log.Printf("Config written with key %s\n", key)
	}
	return nil
}

//func (e *EtcdClient) ReadConfig(labels []string) (*config.Config, error) {
func (e *EtcdClient) ReadConfig(c *config.Config, labels string) (*config.Config, error) {
	// Get default config in etcd
	// key = /telegraf/default
	key := e.Folder + "/main"
	resp, err := e.Kapi.Get(context.Background(), key, nil)
	if err != nil {
		log.Printf("WARNING: [etcd] %s", err)
	} else {
		// Put it in toml
		data, err := json2toml(resp)
		if err != nil {
			log.Printf("WARNING: [etcd] %s", err)
		}
		c.LoadConfigFromText(data)
	}

	// Get specific host config
	// key = /telegraf/hosts/HOSTNAME
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("WARNING: [etcd] %s", err)
	} else if hostname != "" {
		key = e.Folder + "/hosts/" + hostname
		resp, err := e.Kapi.Get(context.Background(), key, nil)
		if err != nil {
			log.Printf("WARNING: [etcd] %s", err)
		} else {
			// Put it in toml
			data, err := json2toml(resp)
			if err != nil {
				log.Print(err)
			}
			c.LoadConfigFromText(data)
		}
	}

	// Concat labels from etcd and labels from command line
	labels_list := c.Agent.Labels
	if labels != "" {
		labels_list = append(labels_list, strings.Split(labels, ",")...)
	}

	// Iterate on all labels
	// TODO check label order of importance ?
	for _, label := range labels_list {
		// Read from etcd
		// key = /telegraf/labels/LABEL
		key := e.Folder + "/labels/" + label
		resp, err := e.Kapi.Get(context.Background(), key, nil)
		if err != nil {
			log.Print(err)
			continue
		}
		// Put it in toml
		data, err := json2toml(resp)
		if err != nil {
			log.Print(err)
			continue
		}
		// Load config
		err = c.LoadConfigFromText(data)
		if err != nil {
			log.Print(err)
		}
	}

	return c, nil
}

func json2toml(resp *client.Response) ([]byte, error) {
	// Convert json to toml
	var json_data interface{}
	var data []byte
	json.Unmarshal([]byte(resp.Node.Value), &json_data)
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(json_data)
	data = buf.Bytes()
	return data, err
}
