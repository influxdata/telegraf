// generate_plugindata is a tool used to inject the sample configuration into all the plugins
// It extracts the sample configuration from the plugins README.md
// Then the plugin's main source file is used as a template, and {{ .SampleConfig }} is replaced with the configuration
// This tool is then also used to revert these changes with the `--clean` flag
package main

import (
	"flag"
	"fmt"
	"log" //nolint:revive
	"os"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

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

	return "", fmt.Errorf("No configuration found for plugin: %s", os.Getenv("GOPACKAGE"))
}

// generatePluginData parses the main source file of the plugin as a template and updates it with the sample configuration
// The original source file is saved so that these changes can be reverted
func generatePluginData(goPackage string, sampleConfig string) error {
	sourceName := fmt.Sprintf("%s.go", goPackage)

	plugin, err := os.ReadFile(sourceName)
	if err != nil {
		return err
	}

	err = os.Rename(sourceName, fmt.Sprintf("%s.tmp", sourceName))
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

// cleanGeneratedFiles will revert the changes made by generatePluginData
func cleanGeneratedFiles(goPackage string) error {
	sourceName := fmt.Sprintf("%s.go", goPackage)
	err := os.Remove(sourceName)
	if err != nil {
		return err
	}
	err = os.Rename(fmt.Sprintf("%s.tmp", sourceName), sourceName)
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
