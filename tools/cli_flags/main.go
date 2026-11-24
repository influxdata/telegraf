package main

import (
	_ "embed"
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
	Path    string
	Level   int
	Command *cli.Command
}

func sections(commands []*cli.Command) []section {
	var result []section
	var walk func(string, int, []*cli.Command)
	walk = func(parent string, level int, items []*cli.Command) {
		for _, command := range items {
			path := parent + " " + command.Name
			result = append(result, section{Path: path, Level: level, Command: command})
			walk(path, level+1, command.VisibleCommands())
		}
	}
	walk("telegraf", 2, commands)
	return result
}

func writeCLIReference(w io.Writer) error {
	tmpl, err := template.New("reference").Funcs(template.FuncMap{
		"flags":    formatFlags,
		"repeat":   strings.Repeat,
		"sections": sections,
	}).Parse(reference)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, commands.AvailableCommands())
}

func generateCLIReference(path string) error {
	var output strings.Builder
	if err := writeCLIReference(&output); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(output.String()), 0o644)
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
