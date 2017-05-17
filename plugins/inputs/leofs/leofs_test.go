package leofs

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
)

var fakeSNMP4Manager = `
package main

import "fmt"

const output = ` + "`" + `iso.3.6.1.4.1.35450.15.1.0 = STRING: "manager_888@127.0.0.1"
iso.3.6.1.4.1.35450.15.2.0 = Gauge32: 186
iso.3.6.1.4.1.35450.15.3.0 = Gauge32: 46235519
iso.3.6.1.4.1.35450.15.4.0 = Gauge32: 32168525
iso.3.6.1.4.1.35450.15.5.0 = Gauge32: 14066068
iso.3.6.1.4.1.35450.15.6.0 = Gauge32: 5512968
iso.3.6.1.4.1.35450.15.7.0 = Gauge32: 186
iso.3.6.1.4.1.35450.15.8.0 = Gauge32: 46269006
iso.3.6.1.4.1.35450.15.9.0 = Gauge32: 32202867
iso.3.6.1.4.1.35450.15.10.0 = Gauge32: 14064995
iso.3.6.1.4.1.35450.15.11.0 = Gauge32: 5492634
iso.3.6.1.4.1.35450.15.12.0 = Gauge32: 60
iso.3.6.1.4.1.35450.15.13.0 = Gauge32: 43515904
iso.3.6.1.4.1.35450.15.14.0 = Gauge32: 60
iso.3.6.1.4.1.35450.15.15.0 = Gauge32: 43533983` + "`" +
	`
func main() {
	fmt.Println(output)
}
`

var fakeSNMP4Storage = `
package main

import "fmt"

const output = ` + "`" + `iso.3.6.1.4.1.35450.34.1.0 = STRING: "storage_0@127.0.0.1"
iso.3.6.1.4.1.35450.34.2.0 = Gauge32: 512
iso.3.6.1.4.1.35450.34.3.0 = Gauge32: 38126307
iso.3.6.1.4.1.35450.34.4.0 = Gauge32: 22308716
iso.3.6.1.4.1.35450.34.5.0 = Gauge32: 15816448
iso.3.6.1.4.1.35450.34.6.0 = Gauge32: 5232008
iso.3.6.1.4.1.35450.34.7.0 = Gauge32: 512
iso.3.6.1.4.1.35450.34.8.0 = Gauge32: 38113176
iso.3.6.1.4.1.35450.34.9.0 = Gauge32: 22313398
iso.3.6.1.4.1.35450.34.10.0 = Gauge32: 15798779
iso.3.6.1.4.1.35450.34.11.0 = Gauge32: 5237315
iso.3.6.1.4.1.35450.34.12.0 = Gauge32: 191
iso.3.6.1.4.1.35450.34.13.0 = Gauge32: 824
iso.3.6.1.4.1.35450.34.14.0 = Gauge32: 0
iso.3.6.1.4.1.35450.34.15.0 = Gauge32: 50105
iso.3.6.1.4.1.35450.34.16.0 = Gauge32: 196654
iso.3.6.1.4.1.35450.34.17.0 = Gauge32: 0
iso.3.6.1.4.1.35450.34.18.0 = Gauge32: 2052
iso.3.6.1.4.1.35450.34.19.0 = Gauge32: 50296
iso.3.6.1.4.1.35450.34.20.0 = Gauge32: 35
iso.3.6.1.4.1.35450.34.21.0 = Gauge32: 898
iso.3.6.1.4.1.35450.34.22.0 = Gauge32: 0
iso.3.6.1.4.1.35450.34.23.0 = Gauge32: 0
iso.3.6.1.4.1.35450.34.24.0 = Gauge32: 0
iso.3.6.1.4.1.35450.34.31.0 = Gauge32: 51
iso.3.6.1.4.1.35450.34.32.0 = Gauge32: 53219328
iso.3.6.1.4.1.35450.34.33.0 = Gauge32: 51
iso.3.6.1.4.1.35450.34.34.0 = Gauge32: 53351083` + "`" +
	`
func main() {
	fmt.Println(output)
}
`

var fakeSNMP4Gateway = `
package main

import "fmt"

const output = ` + "`" + `iso.3.6.1.4.1.35450.34.1.0 = STRING: "gateway_0@127.0.0.1"
iso.3.6.1.4.1.35450.34.2.0 = Gauge32: 465
iso.3.6.1.4.1.35450.34.3.0 = Gauge32: 61676335
iso.3.6.1.4.1.35450.34.4.0 = Gauge32: 46890415
iso.3.6.1.4.1.35450.34.5.0 = Gauge32: 14785011
iso.3.6.1.4.1.35450.34.6.0 = Gauge32: 5578855
iso.3.6.1.4.1.35450.34.7.0 = Gauge32: 465
iso.3.6.1.4.1.35450.34.8.0 = Gauge32: 61644426
iso.3.6.1.4.1.35450.34.9.0 = Gauge32: 46880358
iso.3.6.1.4.1.35450.34.10.0 = Gauge32: 14763002
iso.3.6.1.4.1.35450.34.11.0 = Gauge32: 5582125
iso.3.6.1.4.1.35450.34.12.0 = Gauge32: 191
iso.3.6.1.4.1.35450.34.13.0 = Gauge32: 827
iso.3.6.1.4.1.35450.34.14.0 = Gauge32: 0
iso.3.6.1.4.1.35450.34.15.0 = Gauge32: 50105
iso.3.6.1.4.1.35450.34.16.0 = Gauge32: 196650
iso.3.6.1.4.1.35450.34.17.0 = Gauge32: 0
iso.3.6.1.4.1.35450.34.18.0 = Gauge32: 30256
iso.3.6.1.4.1.35450.34.19.0 = Gauge32: 532158
iso.3.6.1.4.1.35450.34.20.0 = Gauge32: 34
iso.3.6.1.4.1.35450.34.21.0 = Gauge32: 1
iso.3.6.1.4.1.35450.34.31.0 = Gauge32: 53
iso.3.6.1.4.1.35450.34.32.0 = Gauge32: 55050240
iso.3.6.1.4.1.35450.34.33.0 = Gauge32: 53
iso.3.6.1.4.1.35450.34.34.0 = Gauge32: 55186538` + "`" +
	`
func main() {
	fmt.Println(output)
}
`

func makeFakeSNMPSrc(code string) string {
	path := os.TempDir() + "/test.go"
	err := ioutil.WriteFile(path, []byte(code), 0600)
	if err != nil {
		log.Fatalln(err)
	}
	return path
}

func buildFakeSNMPCmd(src string) {
	err := exec.Command("go", "build", "-o", "snmpwalk", src).Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func testMain(t *testing.T, code string, endpoint string, serverType ServerType) {
	// Build the fake snmpwalk for test
	src := makeFakeSNMPSrc(code)
	defer os.Remove(src)
	buildFakeSNMPCmd(src)
	defer os.Remove("./snmpwalk")
	envPathOrigin := os.Getenv("PATH")
	// Refer to the fake snmpwalk
	os.Setenv("PATH", ".")
	defer os.Setenv("PATH", envPathOrigin)

	l := &LeoFS{
		Servers: []string{endpoint},
	}

	var acc testutil.Accumulator
	acc.SetDebug(true)

	err := acc.GatherError(l.Gather)
	require.NoError(t, err)

	floatMetrics := KeyMapping[serverType]

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("leofs", metric), metric)
	}
}

func TestLeoFSManagerMasterMetrics(t *testing.T) {
	testMain(t, fakeSNMP4Manager, "localhost:4020", ServerTypeManagerMaster)
}

func TestLeoFSManagerSlaveMetrics(t *testing.T) {
	testMain(t, fakeSNMP4Manager, "localhost:4021", ServerTypeManagerSlave)
}

func TestLeoFSStorageMetrics(t *testing.T) {
	testMain(t, fakeSNMP4Storage, "localhost:4010", ServerTypeStorage)
}

func TestLeoFSGatewayMetrics(t *testing.T) {
	testMain(t, fakeSNMP4Gateway, "localhost:4000", ServerTypeGateway)
}
