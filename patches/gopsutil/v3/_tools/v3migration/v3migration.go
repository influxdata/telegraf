package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/ast/astutil"
)

// https://github.com/shirou/gopsutil/issues/429
func issue429() error {
	f := func(filename string) error {
		fset := token.NewFileSet()
		expr, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		n := astutil.Apply(expr, func(cr *astutil.Cursor) bool {
			if cr.Name() == "Decls" {
				switch n := cr.Node().(type) {
				case *ast.FuncDecl:
					if n.Name.Name == "NetIOCounters" || n.Name.Name == ("NetIOCountersWithContext") {
						cr.Delete()
					}
				}
			}
			return true
		}, nil)
		return replace(filename, fset, n)
	}

	root := "process/"
	fnames := []string{"process.go", "process_darwin.go", "process_fallback.go", "process_freebsd.go", "process_linux.go", "process_openbsd.go", "process_bsd.go", "process_posix.go", "process_windows.go", "process_test.go"}
	for _, fname := range fnames {
		if err := f(root + fname); err != nil {
			log.Fatalln("run 429:", err)
		}
	}
	return nil
}

func issueRemoveUnusedValue() error {
	f := func(filename string) error {
		fset := token.NewFileSet()
		expr, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		n := astutil.Apply(expr, func(cr *astutil.Cursor) bool {
			if cr.Name() == "Decls" {
				switch n := cr.Node().(type) {
				case *ast.GenDecl:
					if n.Tok != token.TYPE {
						break
					}
					ts := n.Specs[0].(*ast.TypeSpec)
					if ts.Name.Name == "SystemProcessInformation" {
						cr.Delete()
					}
				}
			}
			return true
		}, nil)
		return replace(filename, fset, n)
	}

	if err := f("process/process_windows.go"); err != nil {
		log.Fatalln("run 429:", err)
	}
	return nil
}

func replace(filename string, fset *token.FileSet, n ast.Node) error {
	if err := os.Remove(filename); err != nil {
		return err
	}
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	if err := format.Node(fp, fset, n); err != nil {
		return err
	}
	fp.WriteString("\n")
	return nil
}

func main() {
	flag.Parse()
	for _, n := range flag.Args() {
		fmt.Println("issue:" + n)
		switch n {
		case "429":
			issue429()
		case "issueRemoveUnusedValue":
			issueRemoveUnusedValue()
		}
	}
}
