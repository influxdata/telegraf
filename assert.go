package main

import (
	"fmt"
	"runtime"

	"github.com/yuin/goldmark/ast"
)

//type for all linter assert methods
type T struct {
	filename string
	markdown []byte

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

func (t *T) assertKind(expected ast.NodeKind, n ast.Node) {
	if n.Kind() == expected {
		return
	}

	t.printRule()
	t.printFile(n)
}

func (t *T) printRule() {
	pc, codeFilename, codeLine, ok := runtime.Caller(2)
	if !ok {
		panic("can't get caller")
	}

	f := runtime.FuncForPC(pc)
	var funcName string
	if f != nil {
		funcName = f.Name()
	}

	if len(funcName) == 0 {
		fmt.Printf("failed assert\n")
	} else {
		fmt.Printf("failed assert in function %s\n", funcName)
	}
	fmt.Printf("%s:%d:\n", codeFilename, codeLine)
}

func (t *T) printFile(n ast.Node) {
	fmt.Printf("in markdown file\n")
	fmt.Printf("%s\n", t.filename)
}

func (t *T) printPass() {
	if t.fails == 0 {
		fmt.Printf("Pass %s\n", t.filename)
	}
}
