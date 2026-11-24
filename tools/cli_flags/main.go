package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
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
func main() {
	if err := writeCLIReference(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
