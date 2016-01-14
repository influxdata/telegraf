package proCount

import (
	"fmt"
	"os/exec"
	"log"

	"github.com/influxdb/telegraf/plugins"
)

type proCount struct {
	count []byte	
} 

var procConfig = `
	# Set the processes count
	count = 200
	`

func (s *proCount) SampleConfig() string {
	return procConfig
}

func (s *proCount) Description() string {
	return "Counts all the processes running on a system"
}

func (s *proCount) Gather(acc plugins.Accumulator) error{ 
	holder, err := exec.Command("/bin/sh", "-c", "ps -A --no-headers | wc -l").Output()
	s.count = holder
	if err != nil {
		log.Fatal(err)
	}
	
	fieldData := string(holder)
	
	fields := make(map[string]interface{})
	fields["Processes"] = fieldData	
	
	tags := make(map[string]string)
	
	//Must subtract two from the processes running
	//because running the external commands fpr ps and wc
	//create two additional processes 

	fmt.Println("Processes running:   ", string(holder)-2)

	acc.AddFields("processes", fields, tags) 
	
	return nil
} 

func init(){
	plugins.Add("proCount", func() plugins.Plugin { return &proCount{} })
}
