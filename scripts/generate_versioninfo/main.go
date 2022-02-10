// Generate the versioninfo.json with the current build version from the makefile
// The file versioninfo.json is used by the goversioninfo package to add version info into a windows binary
package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log" //nolint:revive
	"os/exec"
	"strings"
)

type VersionInfo struct {
	StringFileInfo StringFileInfo
}

type StringFileInfo struct {
	ProductName    string
	ProductVersion string
}

func main() {
	e := exec.Command("make", "version")
	var out bytes.Buffer
	e.Stdout = &out
	if err := e.Run(); err != nil {
		log.Fatalf("Failed to get version from makefile: %v", err)
	}
	version := strings.TrimSuffix(out.String(), "\n")

	v := VersionInfo{
		StringFileInfo: StringFileInfo{
			ProductName:    "Telegraf",
			ProductVersion: version,
		},
	}

	file, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		log.Fatalf("Failed to marshal json: %v", err)
	}
	if err := ioutil.WriteFile("cmd/telegraf/versioninfo.json", file, 0644); err != nil {
		log.Fatalf("Failed to write versioninfo.json: %v", err)
	}
}
