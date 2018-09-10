/*
* @Author: Jim Weber
* @Date:   2016-08-08 09:42:04
* @Last Modified by:   Jim Weber
* @Last Modified time: 2016-08-08 14:09:59
 */

package fleet

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, fleetResponseData)
}))

var fleetResponseData = `{
      "states": [
        {
          "hash": "67e33cb6ce9104fa451765128159eaccad11dee8",
          "machineID": "2d69b20e090a4859b2c9ec7d48b0188c",
          "name": "auth-api@34.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "67e33cb6ce9104fa451765128159eaccad11dee8",
          "machineID": "68d18238915842298dc6cd3b90824237",
          "name": "auth-api@35.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "1ebad1f4aa1b11af59c12e3d0fa58985807b54bb",
          "machineID": "635a42fc35b241ffa170c1dc1befa01c",
          "name": "ident@34.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "8a053bc6517baf8473d0ea7872acbd8c31dba0f8",
          "machineID": "39515ef8debc423c961543d45e382c63",
          "name": "help-api@55.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "30bf8d8bb392eb65497f2d0e4ea508401054949c",
          "machineID": "885814e701d94d67bd1264fb1b9c9958",
          "name": "fixer@50.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "30bf8d8bb392eb65497f2d0e4ea508401054949c",
          "machineID": "39515ef8debc423c961543d45e382c63",
          "name": "fixer@51.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "67cc24e573c05fba29de2bbc5cc4b522601ffcf4",
          "machineID": "2d69b20e090a4859b2c9ec7d48b0188c",
          "name": "logspout.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "67cc24e573c05fba29de2bbc5cc4b522601ffcf4",
          "machineID": "39515ef8debc423c961543d45e382c63",
          "name": "logspout.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "67cc24e573c05fba29de2bbc5cc4b522601ffcf4",
          "machineID": "635a42fc35b241ffa170c1dc1befa01c",
          "name": "logspout.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "67cc24e573c05fba29de2bbc5cc4b522601ffcf4",
          "machineID": "68d18238915842298dc6cd3b90824237",
          "name": "logspout.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "67cc24e573c05fba29de2bbc5cc4b522601ffcf4",
          "machineID": "885814e701d94d67bd1264fb1b9c9958",
          "name": "logspout.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "e6cd43573b54647d4508617f98ed6bae9db1be18",
          "machineID": "2d69b20e090a4859b2c9ec7d48b0188c",
          "name": "logstash@56.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "c40cddaed92a845a8ac93eccdc7a5a5517697816",
          "machineID": "635a42fc35b241ffa170c1dc1befa01c",
          "name": "logstash-serverlogs@10.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "f9a0c0c9f105bfac4133d5f23856146b27c48931",
          "machineID": "68d18238915842298dc6cd3b90824237",
          "name": "nginx@19.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "f9a0c0c9f105bfac4133d5f23856146b27c48931",
          "machineID": "39515ef8debc423c961543d45e382c63",
          "name": "nginx@20.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "failed"
        },
        {
          "hash": "7a2914683ef7bae3576bd1e48269839349f58752",
          "machineID": "885814e701d94d67bd1264fb1b9c9958",
          "name": "nginx@18.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "7a2914683ef7bae3576bd1e48269839349f58752",
          "machineID": "68d18238915842298dc6cd3b90824237",
          "name": "nginx@19.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "8a6b627f67b6ab113f083bef1d7c2e583a12eea5",
          "machineID": "39515ef8debc423c961543d45e382c63",
          "name": "weave.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "8a6b627f67b6ab113f083bef1d7c2e583a12eea5",
          "machineID": "635a42fc35b241ffa170c1dc1befa01c",
          "name": "weave.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        },
        {
          "hash": "8a6b627f67b6ab113f083bef1d7c2e583a12eea5",
          "machineID": "68d18238915842298dc6cd3b90824237",
          "name": "weave.service",
          "systemdActiveState": "active",
          "systemdLoadState": "loaded",
          "systemdSubState": "running"
        }
      ]
    }`

func TestGetInstanceStates(t *testing.T) {
	fleetStates := getInstanceStates(ts.URL, nil)
	if fleetStates.States[0].MachineID != "2d69b20e090a4859b2c9ec7d48b0188c" {
		t.Errorf("First machine id json response to be 2d69b20e090a4859b2c9ec7d48b0188c got %v instead", fleetStates.States[0].MachineID)
	}

}

func TestGetContainerCount(t *testing.T) {
	fleetStates := getInstanceStates(ts.URL, nil)
	containerCounts := getContainerCount(fleetStates)

	if containerCounts["auth-api"] != 2 {
		t.Errorf("Auth api count is incorrect got %d instead of 2", containerCounts["auth-api"])
	}

	if containerCounts["nginx"] != 3 {
		t.Errorf("nginx count is incorrect got %d instead of 4", containerCounts["nginx"])
	}
}
