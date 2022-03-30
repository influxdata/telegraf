package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type plugin int

const (
	none plugin = iota
	input
	output
	processor
	aggregator
	parser
)

func guessPluginType(filename string) plugin {
	switch {
	case strings.Contains(filename, "plugins/inputs/"):
		return input
	case strings.Contains(filename, "plugins/outputs/"):
		return output
	case strings.Contains(filename, "plugins/processors/"):
		return processor
	case strings.Contains(filename, "plugins/aggregators/"):
		return aggregator
	case strings.Contains(filename, "plugins/parsers/"):
		return parser
	default:
		return none
	}
}

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
		mainHeading,
		requiredHeadingsClose([]string{
			"Configuration",
		}),
	}
	for i := input; i <= parser; i++ {
		rules[i] = all
	}

	inputRules := []ruleFunc{
		requiredHeadingsClose([]string{
			"Example Output",
			"Metrics",
		}),
	}
	rules[input] = append(rules[input], inputRules...)

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

	// Find long lines
	// last := 0
	// threshold := 80
	// for i, cur := range newlineOffsets {
	// 	len := cur - last - 1 // -1 to exclude the newline
	// 	if len > threshold {
	// 		fmt.Printf("%s:%d: long line\n", filename, i+1) // +1 because line numbers start with 1
	// 	}
	// 	last = cur
	// }

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
