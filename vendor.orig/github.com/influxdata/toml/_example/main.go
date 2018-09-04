package main

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/influxdata/toml"
)

type tomlConfig struct {
	Title string
	Owner struct {
		Name string
		Org  string `toml:"organization"`
		Bio  string
		Dob  time.Time
	}
	Database struct {
		Server        string
		Ports         []int
		ConnectionMax uint
		Enabled       bool
	}
	Servers struct {
		Alpha Server
		Beta  Server
	}
	Clients struct {
		Data  [][]interface{}
		Hosts []string
	}
}

type Server struct {
	IP string
	DC string
}

func main() {
	f, err := os.Open("example.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var config tomlConfig
	if err := toml.Unmarshal(buf, &config); err != nil {
		panic(err)
	}
	// then to use the unmarshaled config...
}
