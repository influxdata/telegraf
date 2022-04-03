package main

import (
	"bufio"
	"bytes"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	var err error
	for _, filename := range os.Args[1:] {
		err = checkFile(filename, guessPluginType(filename))
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

func checkFile(filename string, pluginType plugin) error {
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

		offset += 1
	}

	p := goldmark.DefaultParser()
	r := text.NewReader(md)
	root := p.Parse(r)

	rules := rules[pluginType]

	tester := T{
		filename:       filename,
		markdown:       md,
		newlineOffsets: newlineOffsets,
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
