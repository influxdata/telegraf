package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/cmd/telegraf/commands"
)

//go:embed reference.md.tmpl
var reference string

type section struct {
	Heading, Path string
	Command       *cli.Command
}

func sections(commands []*cli.Command) []section {
	var result []section
	var walk func(string, string, []*cli.Command)
	walk = func(parent, heading string, items []*cli.Command) {
		for _, command := range items {
			path := parent + " " + command.Name
			result = append(result, section{heading, path, command})
			walk(path, heading+"#", command.VisibleCommands())
		}
	}
	walk("telegraf", "##", commands)
	return result
}
func writeCLIReference(w io.Writer) error {
	tmpl, err := template.New("reference").Funcs(template.FuncMap{"flags": formatFlags, "sections": sections}).Parse(reference)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, commands.AvailableCommands())
}
func generateCLIReference(path string) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".cli-flags-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	writeErr := writeCLIReference(file)
	modeErr := file.Chmod(0o644)
	closeErr := file.Close()
	if err := errors.Join(writeErr, modeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output-path>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	if err := generateCLIReference(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "generate CLI reference %q: %v\n", os.Args[1], err)
		os.Exit(1)
	}
}

func formatFlags(flags []cli.Flag) string {
	var doc strings.Builder
	if len(flags) == 0 {
		return ""
	}
	doc.WriteByte('\n')
	for _, flag := range flags {
		option := flag.(cli.DocGenerationFlag)
		names := make([]string, 0, len(flag.Names()))
		for _, name := range flag.Names() {
			prefix := "--"
			if len(name) == 1 {
				prefix = "-"
			}
			names = append(names, "`"+prefix+name+"`")
		}
		label := strings.Join(names, ", ")
		if option.TakesValue() {
			label += " `<value>`"
		}
		line := "- " + label + " —"
		for _, word := range strings.Fields(option.GetUsage()) {
			if len(line)+1+len(word) > 80 {
				doc.WriteString(line + "\n")
				line = "  " + word
			} else {
				line += " " + word
			}
		}
		doc.WriteString(line + "\n")
	}
	return doc.String()
}
