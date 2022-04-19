package main

import (
	"fmt"
	"regexp"
	"runtime"
	"sort"

	"github.com/yuin/goldmark/ast"
)

//type for all linter assert methods
type T struct {
	filename       string
	markdown       []byte
	newlineOffsets []int
	sourceFlag     bool

	fails int
}

// called by all assert functions that involve a node
func (t *T) printFailedAssertf(n ast.Node, format string, args ...interface{}) {
	t.printFile(n)
	fmt.Printf(format+"\n", args...)
	t.printRule(3)
	t.fails++
}

// Assert function that doesnt involve a node, for example if something is missing
func (t *T) assertf(format string, args ...interface{}) {
	t.assertLine2f(0, format, args...) // There's no line number associated, so use the first
}

func (t *T) assertNodef(n ast.Node, format string, args ...interface{}) {
	t.printFailedAssertf(n, format, args...)
}

func (t *T) assertLinef(line int, format string, args ...interface{}) {
	//this func only exists to make the call stack to t.printRule the same depth
	//as when called through assertf

	t.assertLine2f(line, format, args...)
}

func (t *T) assertLine2f(line int, format string, args ...interface{}) {
	t.printFileLine(line)
	fmt.Printf(format+"\n", args...)
	t.printRule(3)
	t.fails++
}

func (t *T) printRule(callers int) {
	if !t.sourceFlag {
		return
	}

	pc, codeFilename, codeLine, ok := runtime.Caller(callers)
	if !ok {
		panic("can not get caller")
	}

	f := runtime.FuncForPC(pc)
	var funcName string
	if f != nil {
		funcName = f.Name()
	}

	fmt.Printf("%s:%d: ", codeFilename, codeLine)
	if len(funcName) == 0 {
		fmt.Printf("failed assert\n")
	} else {
		fmt.Printf("failed assert in function %s\n", funcName)
	}
}

func (t *T) line(offset int) int {
	return sort.SearchInts(t.newlineOffsets, offset)
}

func (t *T) printFile(n ast.Node) {
	lines := n.Lines()
	if lines == nil || lines.Len() == 0 {
		t.printFileLine(0)
		return
	}
	offset := lines.At(0).Start
	line := t.line(offset)
	t.printFileLine(line)
}

func (t *T) printFileLine(line int) {
	fmt.Printf("%s:%d: ", t.filename, line+1) // Lines start with 1
}

func (t *T) printPassFail() {
	if t.fails == 0 {
		fmt.Printf("Pass %s\n", t.filename)
	} else {
		fmt.Printf("Fail %s, %d failed assertions\n", t.filename, t.fails)
	}
}

func (t *T) assertKind(expected ast.NodeKind, n ast.Node) {
	if n.Kind() == expected {
		return
	}

	t.printFailedAssertf(n, "expected %s, have %s", expected.String(), n.Kind().String())
}

func (t *T) assertFirstChildRegexp(expectedPattern string, n ast.Node) {
	var validRegexp = regexp.MustCompile(expectedPattern)

	if !n.HasChildren() {
		t.printFailedAssertf(n, "expected children")
		return
	}
	c := n.FirstChild()

	actual := string(c.Text(t.markdown))

	if !validRegexp.MatchString(actual) {
		t.printFailedAssertf(n, "'%s' does not match regexp '%s'", actual, expectedPattern)
		return
	}
}

func (t *T) assertHeadingLevel(expected int, n ast.Node) {
	h, ok := n.(*ast.Heading)
	if !ok {
		fmt.Printf("failed Heading type assertion\n")
		t.fails++
		return
	}

	if h.Level == expected {
		return
	}

	t.printFailedAssertf(n, "expected header level %d, have %d", expected, h.Level)
}
