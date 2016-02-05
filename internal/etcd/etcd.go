package etcd

import (
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/etcd/client"
	influxconfig "github.com/influxdata/config"
	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
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
	splittedUrls := strings.Split(urls, ",")
	// Create a new etcd client
	cfg := client.Config{
		Endpoints: splittedUrls,
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

func (e *EtcdClient) DeleteConfig(path string) error {
	// removeWrite main config file in etcd
	key := e.Folder + "/" + path
	_, err := e.Kapi.Delete(context.Background(), key, &client.DeleteOptions{Recursive: true})
	return err
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
	// Read file
	tbl, err := influxconfig.ParseFile(path)
	if err != nil {
		return err
	}
	// Get toml
	toml_data := tbl.Source()
	// Write it
	key := e.Folder + "/" + relative_key
	resp, _ := e.Kapi.Get(context.Background(), key, nil)
	if resp == nil {
		_, err = e.Kapi.Set(context.Background(), key, string(toml_data), nil)
	} else {
		_, err = e.Kapi.Update(context.Background(), key, string(toml_data))
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
		tbl, err := toml2table(resp)
		if err != nil {
			log.Printf("WARNING: [etcd] %s", err)
		}
		c.LoadConfigFromTable(tbl)
		if err != nil {
			log.Print(err, "")
		}
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
			tbl, err := toml2table(resp)
			if err != nil {
				log.Print(err)
			}
			c.LoadConfigFromTable(tbl)
			if err != nil {
				log.Print(err, "")
			}
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
		tbl, err := toml2table(resp)
		if err != nil {
			log.Print(err)
			continue
		}
		// Load config
		err = c.LoadConfigFromTable(tbl)
		if err != nil {
			log.Print(err, "")
		}
	}

	return c, nil
}

func toml2table(resp *client.Response) (*ast.Table, error) {
	// Convert json to toml
	return toml.Parse([]byte(resp.Node.Value))
}
