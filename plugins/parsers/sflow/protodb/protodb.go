package protodb

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	// serviceNameMap is a map of ("%d:%d",protocol,port) -> string
	serviceNameMap map[string]string
)

//go:generate go run ../scripts/generate-embedded-data.go

func init() {
	serviceNameMap = make(map[string]string)
	switch len(os.Getenv("TELEGRAF_SFLOW_USE_ETC_SERVICES")) {
	case 0:
		populateServiceNameMapFromEmbedded(serviceNameMap)
	default:
		if e := populateServiceNameMapFromEtcServices(serviceNameMap); e != nil {
			populateServiceNameMapFromEmbedded(serviceNameMap)
		}
	}
}

func populateServiceNameMapFromEtcServices(snm map[string]string) error {
	data, err := ioutil.ReadFile("/etc/services")
	if err != nil {
		log.Println("W! [parser.sflow] /etc/services db not available")
		return err
	}
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
		snm[fmt.Sprintf("%s:%d", pnp[1], port)] = fields[0]
	}
	return nil
}

// GetServByPort answers the service name associaed with teh given protocol and port - if known
func GetServByPort(protocol string, port int) (string, bool) {
	value, ok := serviceNameMap[fmt.Sprintf("%s:%d", protocol, port)]
	return value, ok
}
