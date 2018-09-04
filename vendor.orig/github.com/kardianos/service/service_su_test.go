// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// This needs to be run as root/admin hence the reason there is a build tag
// +build su

package service_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/kardianos/service"
)

const runAsServiceArg = "RunThisAsService"

func TestMain(m *testing.M) {
	reportDir := flag.String("su.reportDir", "", "")
	runAsService := flag.Bool("su.runAsService", false, "")
	flag.Parse()
	if !*runAsService {
		os.Exit(m.Run())
	}
	if len(*reportDir) == 0 {
		log.Fatal("missing su.reportDir argument")
	}
	writeReport(*reportDir, "call")
	runService()
	writeReport(*reportDir, "finished")
}

func TestInstallRunRestartStopRemove(t *testing.T) {
	p := &program{}
	reportDir := mustTempDir(t)
	defer os.RemoveAll(reportDir)

	s := mustNewRunAsService(t, p, reportDir)
	_ = s.Uninstall()

	if err := s.Install(); err != nil {
		t.Fatal("Install", err)
	}
	defer s.Uninstall()

	if err := s.Start(); err != nil {
		t.Fatal("Start", err)
	}
	defer s.Stop()
	checkReport(t, reportDir, "Start()", 1, 0)

	if err := s.Restart(); err != nil {
		t.Fatal("restart", err)
	}

	checkReport(t, reportDir, "Restart()", 2, 1)
	p.numStopped = 0
	if err := s.Stop(); err != nil {
		t.Fatal("stop", err)
	}
	checkReport(t, reportDir, "Stop()", 2, 2)

	if err := s.Uninstall(); err != nil {
		t.Fatal("uninstall", err)
	}
}

func runService() {
	p := &program{}
	sc := &service.Config{
		Name: "go_service_test",
	}
	s, err := service.New(p, sc)
	if err != nil {
		log.Fatal(err)
	}
	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}

func mustTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "servicetest")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeReport(reportDir string, action string) {
	b := []byte("go_test_service_report")
	timeStamp := time.Now().UnixNano()
	err := ioutil.WriteFile(
		filepath.Join(reportDir, fmt.Sprintf("%d-%s", timeStamp, action)),
		b,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
}

var matchActionRegexp = regexp.MustCompile("^(\\d+-)([a-z]*)$")

func getReport(
	t *testing.T,
	reportDir string,
) (numCalls int, numFinished int) {
	numCalls = 0
	numFinished = 0
	files, err := ioutil.ReadDir(reportDir)
	if err != nil {
		t.Fatalf("ReadDir(%s) err: %s", reportDir, err)
	}

	for _, file := range files {
		if matchActionRegexp.MatchString(file.Name()) {
			action := matchActionRegexp.ReplaceAllString(file.Name(), "$2")
			switch action {
			case "call":
				numCalls++
			case "finished":
				numFinished++
			default:
				t.Fatalf("getReport() found report with incorrect action: %s", action)
			}
		}
	}
	return
}

func checkReport(
	t *testing.T,
	reportDir string,
	msgPrefix string,
	wantNumCalled int,
	wantNumFinished int,
) {
	var numCalled int
	var numFinished int
	for i := 0; i < 25; i++ {
		numCalled, numFinished = getReport(t, reportDir)
		<-time.After(200 * time.Millisecond)
		if numCalled == wantNumCalled && numFinished == wantNumFinished {
			return
		}
	}
	if numCalled != wantNumCalled {
		t.Fatalf("%s - numCalled: %d, want %d",
			msgPrefix, numCalled, wantNumCalled)
	}
	if numFinished != wantNumFinished {
		t.Fatalf("%s - numFinished: %d, want %d",
			msgPrefix, numFinished, wantNumFinished)
	}
}

func mustNewRunAsService(
	t *testing.T,
	p *program,
	reportDir string,
) service.Service {
	sc := &service.Config{
		Name:      "go_service_test",
		Arguments: []string{"-test.v=true", "-su.runAsService", "-su.reportDir", reportDir},
	}
	s, err := service.New(p, sc)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
