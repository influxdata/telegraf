// Copyright 2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/gnatsd/server"
)

// waits until a calculated list of listeners is resolved or a timeout
func waitForFile(path string, dur time.Duration) ([]byte, error) {
	end := time.Now().Add(dur)
	for time.Now().Before(end) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			time.Sleep(25 * time.Millisecond)
			continue
		} else {
			return ioutil.ReadFile(path)
		}
	}
	return nil, errors.New("Timeout")
}

func portFile(dirname string) string {
	return path.Join(dirname, fmt.Sprintf("%s_%d.ports", path.Base(os.Args[0]), os.Getpid()))
}

func TestPortsFile(t *testing.T) {
	portFileDir := os.TempDir()

	opts := DefaultTestOptions
	opts.PortsFileDir = portFileDir
	opts.Port = -1
	opts.HTTPPort = -1
	opts.ProfPort = -1
	opts.Cluster.Port = -1

	s := RunServer(&opts)
	// this for test cleanup in case we fail - will be ignored if server already shutdown
	defer s.Shutdown()

	ports := s.PortsInfo(5 * time.Second)

	if ports == nil {
		t.Fatal("services failed to start in 5 seconds")
	}

	// the pid file should be
	portsFile := portFile(portFileDir)

	if portsFile == "" {
		t.Fatal("Expected a ports file")
	}

	// try to read a file here - the file should be a json
	buf, err := waitForFile(portsFile, 5*time.Second)
	if err != nil {
		t.Fatalf("Could not read ports file: %v", err)
	}
	if len(buf) <= 0 {
		t.Fatal("Expected a non-zero length ports file")
	}

	readPorts := server.Ports{}
	json.Unmarshal(buf, &readPorts)

	if len(readPorts.Nats) == 0 || !strings.HasPrefix(readPorts.Nats[0], "nats://") {
		t.Fatal("Expected at least one nats url")
	}

	if len(readPorts.Monitoring) == 0 || !strings.HasPrefix(readPorts.Monitoring[0], "http://") {
		t.Fatal("Expected at least one monitoring url")
	}

	if len(readPorts.Cluster) == 0 || !strings.HasPrefix(readPorts.Cluster[0], "nats://") {
		t.Fatal("Expected at least one cluster listen url")
	}

	if len(readPorts.Profile) == 0 || !strings.HasPrefix(readPorts.Profile[0], "http://") {
		t.Fatal("Expected at least one profile listen url")
	}

	// testing cleanup
	s.Shutdown()
	// if we called shutdown, the cleanup code should have kicked
	if _, err := os.Stat(portsFile); os.IsNotExist(err) {
		// good
	} else {
		t.Fatalf("the port file %s was not deleted", portsFile)
	}
}

// makes a temp directory with two directories 'A' and 'B'
// the location of the ports file is changed from dir A to dir B.
func TestPortsFileReload(t *testing.T) {
	// make a temp dir
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Error creating temp director (%s): %v", tempDir, err)
	}
	defer os.RemoveAll(tempDir)

	// make child temp dir A
	dirA := path.Join(tempDir, "A")
	os.MkdirAll(dirA, 0777)

	// write the config file with a reference to A
	config := fmt.Sprintf("ports_file_dir %s\nport -1", dirA)
	confPath := path.Join(tempDir, fmt.Sprintf("%d.conf", os.Getpid()))
	if err := ioutil.WriteFile(confPath, []byte(config), 0666); err != nil {
		t.Fatalf("Error writing ports file (%s): %v", confPath, err)
	}

	opts, err := server.ProcessConfigFile(confPath)
	if err != nil {
		t.Fatalf("Error processing the configuration: %v", err)
	}

	s := RunServer(opts)
	defer s.Shutdown()

	ports := s.PortsInfo(5 * time.Second)
	if ports == nil {
		t.Fatal("services failed to start in 5 seconds")
	}

	// get the ports file path name
	portsFileInA := portFile(dirA)
	// the file should be in dirA
	if !strings.HasPrefix(portsFileInA, dirA) {
		t.Fatalf("expected ports file to be in [%s] but was in [%s]", dirA, portsFileInA)
	}
	// wait for it
	buf, err := waitForFile(portsFileInA, 5*time.Second)
	if err != nil {
		t.Fatalf("Could not read ports file: %v", err)
	}
	if len(buf) <= 0 {
		t.Fatal("Expected a non-zero length ports file")
	}

	// change the configuration for the ports file to dirB
	dirB := path.Join(tempDir, "B")
	os.MkdirAll(dirB, 0777)

	config = fmt.Sprintf("ports_file_dir %s\nport -1", dirB)
	if err := ioutil.WriteFile(confPath, []byte(config), 0666); err != nil {
		t.Fatalf("Error writing ports file (%s): %v", confPath, err)
	}

	// reload the server
	if err := s.Reload(); err != nil {
		t.Fatalf("error reloading server: %v", err)
	}

	// wait for the new file to show up
	portsFileInB := portFile(dirB)
	buf, err = waitForFile(portsFileInB, 5*time.Second)
	if !strings.HasPrefix(portsFileInB, dirB) {
		t.Fatalf("expected ports file to be in [%s] but was in [%s]", dirB, portsFileInB)
	}
	if err != nil {
		t.Fatalf("Could not read ports file: %v", err)
	}
	if len(buf) <= 0 {
		t.Fatal("Expected a non-zero length ports file")
	}

	// the file in dirA should have deleted
	if _, err := os.Stat(portsFileInA); os.IsNotExist(err) {
		// good
	} else {
		t.Fatalf("the port file %s was not deleted", portsFileInA)
	}
}
