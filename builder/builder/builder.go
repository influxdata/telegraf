package builder

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Build struct {
	GOOS    string
	GOARCH  string
	Plugins []string
}

// Compile compiles a custom build of telegraf
// GOARCH="amd64"
// GOOS="darwin"
func (b *Build) Compile() error {
	os.MkdirAll("tmp", os.ModeDir|0447)
	f, err := os.Create("cmd/telegraf/customplugins.go")
	if err != nil {
		return err
	}

	if _, err = f.WriteString("package main\n\n"); err != nil {
		return err
	}

	for _, p := range b.Plugins {
		if p == "inputs.internal" {
			continue
		}
		path := "github.com/influxdata/telegraf/plugins/" + strings.ReplaceAll(p, ".", "/")
		if _, err = f.WriteString(`import _ "` + path + `"` + "\n"); err != nil {
			return err
		}
	}
	if err = f.Close(); err != nil {
		return err
	}

	outputFile := "tmp/telegraf"
	cmd := exec.Command("go", "build", "-o", outputFile, "./cmd/telegraf/telegraf.go", "./cmd/telegraf/telegraf_posix.go", "./cmd/telegraf/customplugins.go")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Env, []string{
		"GOOS=" + b.GOOS,
		"GOARCH=" + b.GOARCH,
		// `GOCACHE=/Users/stevensoroka/Library/Caches/go-build`,
		// "GOPATH=/Users/stevensoroka/go",
	}...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}
	return nil
}
