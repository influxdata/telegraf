package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	var err error
	filename := "/home/reim/go/src/github.com/influxdata/telegraf/plugins/inputs/modbus/README.md"
	//filename := "test.txt"
	err = checkFile(filename)
	if err != nil {
		panic(err)
	}
}

type ruleFunc func(*T, ast.Node) error

func checkFile(filename string) error {
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
	last := 0
	threshold := 80
	for i, cur := range newlineOffsets {
		len := cur - last - 1 // -1 to exclude the newline
		if len > threshold {
			fmt.Printf("%s:%d: long line\n", filename, i+1) // +1 because line numbers start with 1
		}
		last = cur
	}

	p := goldmark.DefaultParser()
	r := text.NewReader(md)
	root := p.Parse(r)

	rules := []ruleFunc{
		mainHeading,
	}

	for _, rule := range rules {
		tester := T{
			filename: filename,
			markdown: md,
		}
		err = rule(&tester, root)
		if err != nil {
			return err
		}
		tester.printPass()
		fmt.Printf("\n")
	}

	return nil
}
