package protodb

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

var (
	// serviceNameMap is a map of ("%d:%d",protocol,port) -> string
	serviceNameMap map[string]string
)

func init() {
	serviceNameMap = make(map[string]string)
	data, err := ioutil.ReadFile("/etc/services")
	if err != nil {
		log.Println("W! [parser.sflow] /etc/services db not available")
	} else {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			split := strings.SplitN(line, "#", 2)
			fields := strings.Fields(split[0])
			if len(fields) < 2 {
				continue
			}

			pnp := strings.SplitN(fields[1], "/", 2)
			port, err := strconv.ParseInt(pnp[0], 10, 32)
			if err != nil {
				log.Printf("W! [parser.sflow] /etc/services unable to parse %s as port number from line %s", pnp[0], line)
				continue
			}

			serviceNameMap[fmt.Sprintf("%s:%d", pnp[1], port)] = fields[0]
		}
	}
}

// GetServByPort answers the service name associaed with teh given protocol and port - if known
func GetServByPort(protocol string, port int) (string, bool) {
	value, ok := serviceNameMap[fmt.Sprintf("%s:%d", protocol, port)]
	return value, ok
}
