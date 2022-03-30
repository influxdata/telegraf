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

	fails int
}

// func (t *T) assert(b bool) {
// 	if b {
// 		return
// 	}
// 	t.fails += 1

// 	t.printRule()
// 	t.printFile()
// }

// called by all assert functions that involve a node
func (t *T) printFailedAssert(n ast.Node, format string, args ...interface{}) {
	t.printFile(n)
	fmt.Printf(format+"\n", args...)
	//t.printRule(3)
	t.fails += 1
}

// Assert function that doesnt involve a node, for example if something is missing
func (t *T) assertf(format string, args ...interface{}) {
	fmt.Printf("%s:%d: ", t.filename, 1) //similar to printFile
	fmt.Printf(format+"\n", args...)
	//t.printRule(2)
	t.fails += 1
}

func (t *T) assertNodef(n ast.Node, format string, args ...interface{}) {
	t.printFailedAssert(n, format, args...)
}

func (t *T) printRule(callers int) {
	pc, codeFilename, codeLine, ok := runtime.Caller(callers)
	if !ok {
		panic("can't get caller")
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
	return sort.SearchInts(t.newlineOffsets, offset) + 1
}

func (t *T) printFile(n ast.Node) {
	lines := n.Lines()
	if lines.Len() == 0 {
		panic("can't get offset of node")
	}
	offset := lines.At(0).Start
	line := t.line(offset)

	fmt.Printf("%s:%d: ", t.filename, line)
	//fmt.Printf("offset: %d\n", offset)
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

	t.printFailedAssert(n, "expected %s, have %s", expected.String(), n.Kind().String())

	//n.Dump(t.markdown, 0)
}

func (t *T) assertFirstChildRegexp(expectedPattern string, n ast.Node) {
	var validRegexp = regexp.MustCompile(expectedPattern)

	if !n.HasChildren() {
		t.printFailedAssert(n, "expected children")
		return
	}
	c := n.FirstChild()

	actual := string(c.Text(t.markdown))

	if !validRegexp.MatchString(actual) {
		t.printFailedAssert(n, "'%s' doesn't match regexp '%s'", actual, expectedPattern)
		return
	}
}

func (t *T) assertHeadingLevel(expected int, n ast.Node) {

	h, ok := n.(*ast.Heading)
	if !ok {
		fmt.Printf("failed Heading type assertion\n")
		t.fails += 1
		return
	}

	if h.Level == expected {
		return
	}

	t.printFailedAssert(n, "expected header level %d, have %d", expected, h.Level)

	//n.Dump(t.markdown, 0)
}
