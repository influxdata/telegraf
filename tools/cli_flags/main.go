package main

import (
	"bytes"
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

//go:embed COMMANDS_AND_FLAGS.tmpl
var reference string

type section struct {
	Path      string
	Level     int
	Usage     string
	ArgsUsage string
	Flags     string
}

func sections(parent string, level int, commands []*cli.Command) []section {
	var result []section
	for _, command := range commands {
		path := parent + " " + command.Name
		result = append(result, section{
			Path:      path,
			Level:     level,
			Usage:     command.Usage,
			ArgsUsage: command.ArgsUsage,
			Flags:     formatFlags(command.VisibleFlags()),
		})
		result = append(result, sections(path, level+1, command.VisibleCommands())...)
	}
	return result
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

func writeCLIReference(w io.Writer) error {
	root := commands.AvailableCommands()
	data := struct {
		GlobalFlags string
		Sections    []section
	}{
		GlobalFlags: formatFlags(root.VisibleFlags()),
		Sections:    sections(root.Name, 2, root.VisibleCommands()),
	}
	tmpl, err := template.New("reference").Funcs(template.FuncMap{"repeat": strings.Repeat}).Parse(reference)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output-path>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	var output bytes.Buffer
	if err := writeCLIReference(&output); err != nil {
		fmt.Fprintf(os.Stderr, "generate CLI reference: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(os.Args[1], output.Bytes(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
