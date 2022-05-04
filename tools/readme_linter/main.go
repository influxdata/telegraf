package main

import (
	"bufio"
	"bytes"
	"flag"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func main() {
	sourceFlag := flag.Bool("source", false, "include location of linter code that failed assertion")

	flag.Parse()

	var err error
	for _, filename := range flag.Args() {
		err = checkFile(filename, guessPluginType(filename), *sourceFlag)
		if err != nil {
			panic(err)
		}
	}
}

type ruleFunc func(*T, ast.Node) error

type rulesMap map[plugin][]ruleFunc

var rules rulesMap

func init() {
	rules = make(rulesMap)

	//rules for all plugin types
	all := []ruleFunc{
		firstSection,
		noLongLinesInParagraphs(80),
		configSection,
		relativeTelegrafLinks,
	}
	for i := pluginInput; i <= pluginParser; i++ {
		rules[i] = all
	}

	inputRules := []ruleFunc{
		requiredSectionsClose([]string{
			"Example Output",
			"Metrics",
		}),
	}
	rules[pluginInput] = append(rules[pluginInput], inputRules...)
}

func checkFile(filename string, pluginType plugin, sourceFlag bool) error {
	md, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Goldmark returns locations as offsets. We want line
	// numbers. Find the newlines in the file so we can translate
	// later.
	scanner := bufio.NewScanner(bytes.NewReader(md))
	scanner.Split(bufio.ScanRunes)
	offset := 0
	newlineOffsets := []int{}
	for scanner.Scan() {
		if scanner.Text() == "\n" {
			newlineOffsets = append(newlineOffsets, offset)
		}

		offset++
	}

	p := goldmark.DefaultParser()

	// We need goldmark to parse tables, otherwise they show up as
	// paragraphs. Since tables often have long lines and we check for long
	// lines in paragraphs, without table parsing there are false positive long
	// lines in tables.
	//
	// The tableParagraphTransformer is an extension and not part of the default
	// parser so we add it. There may be an easier way to do it, but this works:
	p.AddOptions(
		parser.WithParagraphTransformers(
			util.Prioritized(extension.NewTableParagraphTransformer(), 99),
		),
	)

	r := text.NewReader(md)
	root := p.Parse(r)

	rules := rules[pluginType]

	tester := T{
		filename:       filename,
		markdown:       md,
		newlineOffsets: newlineOffsets,
		sourceFlag:     sourceFlag,
	}
	for _, rule := range rules {
		err = rule(&tester, root)
		if err != nil {
			return err
		}
	}
	tester.printPassFail()

	return nil
}
