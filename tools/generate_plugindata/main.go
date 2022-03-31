// generate_plugindata is a tool used to inject the sample configuration into all the plugins
// It extracts the sample configuration from the plugins README.md
// Then the plugin's main source file is used as a template, and {{ .SampleConfig }} is replaced with the configuration
// This tool is then also used to revert these changes with the `--clean` flag
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log" //nolint:revive
	"os"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func createSourceName(packageName string) string {
	return fmt.Sprintf("%s_sample_config.go", packageName)
}

// extractPluginData reads the README.md to get the sample configuration
func extractPluginData() (string, error) {
	readMe, err := os.ReadFile("README.md")
	if err != nil {
		return "", err
	}
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	mdParser := parser.NewWithExtensions(extensions)
	md := markdown.Parse(readMe, mdParser)

	var currentSection string
	for _, t := range md.GetChildren() {
		switch tok := t.(type) {
		case *ast.Heading:
			currentSection = tok.HeadingID
		case *ast.CodeBlock:
			if currentSection == "configuration" && string(tok.Info) == "toml" {
				return string(tok.Literal), nil
			}
		}
	}

	fmt.Printf("No configuration found for plugin: %s\n", os.Getenv("GOPACKAGE"))

	return "", nil
}

// generatePluginData parses the main source file of the plugin as a template and updates it with the sample configuration
// The original source file is saved so that these changes can be reverted
func generatePluginData(packageName string, sampleConfig string) error {
	sourceName := createSourceName(packageName)

	plugin, err := os.ReadFile(sourceName)
	if err != nil {
		return err
	}

	generatedTemplate := template.Must(template.New("").Parse(string(plugin)))

	f, err := os.Create(sourceName)
	if err != nil {
		return err
	}
	defer f.Close()

	err = generatedTemplate.Execute(f, struct {
		SampleConfig string
	}{
		SampleConfig: sampleConfig,
	})
	if err != nil {
		return err
	}

	return nil
}

var newSampleConfigFunc = `	return ` + "`{{ .SampleConfig }}`\n"

// cleanGeneratedFiles will revert the changes made by generatePluginData
func cleanGeneratedFiles(packageName string) error {
	sourceName := createSourceName(packageName)
	sourcefile, err := os.Open(sourceName)
	if err != nil {
		return err
	}
	defer sourcefile.Close()

	var c []byte
	buf := bytes.NewBuffer(c)

	scanner := bufio.NewScanner(sourcefile)

	var sampleconfigSection bool
	for scanner.Scan() {
		if sampleconfigSection && strings.TrimSpace(scanner.Text()) == "}" {
			sampleconfigSection = false
			if _, err := buf.Write([]byte(newSampleConfigFunc)); err != nil {
				return err
			}
		}

		if !sampleconfigSection {
			if _, err := buf.Write(scanner.Bytes()); err != nil {
				return err
			}
			if _, err = buf.WriteString("\n"); err != nil {
				return err
			}
		}
		if !sampleconfigSection && strings.Contains(scanner.Text(), "SampleConfig() string") {
			sampleconfigSection = true
		}
	}

	err = os.WriteFile(sourceName, buf.Bytes(), 0664)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	clean := flag.Bool("clean", false, "Remove generated files")
	flag.Parse()

	goPackage := os.Getenv("GOPACKAGE")

	if *clean {
		err := cleanGeneratedFiles(goPackage)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		s, err := extractPluginData()
		if err != nil {
			log.Fatal(err)
		}

		err = generatePluginData(goPackage, s)
		if err != nil {
			log.Fatal(err)
		}
	}
}
