// This must be package main
package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type analyzerPlugin struct{}

// This must be implemented
func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		TelegrafAnalyzer,
	}
}

// This must be defined and named 'AnalyzerPlugin'
var AnalyzerPlugin analyzerPlugin

var TelegrafAnalyzer = &analysis.Analyzer{
	Name: "telegraflinter",
	Doc:  "Find Telegraf specific review criteria, more info: https://github.com/influxdata/telegraf/wiki/Review",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			checkLogImport(n, pass)
			return true
		})
	}
	return nil, nil
}

func checkLogImport(n ast.Node, pass *analysis.Pass) {
	if !strings.HasPrefix(pass.Pkg.Path(), "github.com/influxdata/telegraf/plugins/") {
		return
	}
	if importSpec, ok := n.(*ast.ImportSpec); ok {
		if importSpec.Path != nil && strings.HasPrefix(importSpec.Path.Value, "\"log\"") {
			pass.Report(analysis.Diagnostic{
				Pos:            importSpec.Pos(),
				End:            0,
				Category:       "log",
				Message:        "Don't use log package in plugin, use the Telegraf logger.",
				SuggestedFixes: nil,
			})
		}
	}
}
