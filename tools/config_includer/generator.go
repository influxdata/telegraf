package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"
)

func extractIncludes(tmpl *template.Template) []string {
	var includes []string
	for _, node := range tmpl.Root.Nodes {
		if n, ok := node.(*parse.TemplateNode); ok {
			includes = append(includes, n.Name)
		}
	}
	return includes
}

func absolutePath(root, fn string) (string, error) {
	pwd, err := filepath.Abs(fn)
	if err != nil {
		return "", fmt.Errorf("cannot determine absolute location of %q: %w", fn, err)
	}
	pwd, err = filepath.Rel(root, filepath.Dir(pwd))
	if err != nil {
		return "", fmt.Errorf("Cannot determine location of %q relative to %q: %w", pwd, root, err)
	}
	return string(filepath.Separator) + pwd, nil
}

func main() {
	// Estimate Telegraf root to be able to handle absolute paths
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get working directory: %v", err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		log.Fatalf("Cannot resolve working directory: %v", err)
	}

	var root string
	idx := strings.LastIndex(cwd, filepath.FromSlash("/plugins/"))
	if idx <= 0 {
		log.Fatalln("Cannot determine include root!")
	}
	root = cwd[:idx]

	var parent, inputFilename, outputFilename string
	switch len(os.Args) {
	case 1:
		parent = strings.TrimPrefix(filepath.ToSlash(cwd[idx:]), "/plugins/")
		parent = strings.ReplaceAll(parent, "/", ".")
		inputFilename = "sample.conf.in"
		outputFilename = "sample.conf"
	case 2:
		parent = os.Args[1]
		inputFilename = "sample.conf.in"
		outputFilename = "sample.conf"
	case 3:
		parent = os.Args[1]
		inputFilename = os.Args[2]
		if !strings.HasSuffix(inputFilename, ".in") {
			log.Fatalf("Template filename %q does not have '.in' suffix!", inputFilename)
		}
		outputFilename = strings.TrimSuffix(inputFilename, ".in")
	case 4:
		parent = os.Args[1]
		inputFilename = os.Args[2]
		outputFilename = os.Args[3]
	default:
		log.Fatalln("Invalid number of arguments")
	}

	roottmpl := template.New(inputFilename)
	known := make(map[string]bool)
	inroot, err := absolutePath(root, inputFilename)
	if err != nil {
		log.Fatal(err)
	}
	unresolved := map[string]string{inputFilename: filepath.Join(inroot, inputFilename)}
	for {
		if len(unresolved) == 0 {
			break
		}

		newUnresolved := make(map[string]string)
		for name, fn := range unresolved {
			if strings.HasPrefix(filepath.ToSlash(fn), "/") {
				fn = filepath.Join(root, fn)
			}

			if known[name] {
				// Include already resolved, skipping
				continue
			}

			tmpl, err := template.ParseFiles(fn)
			if err != nil {
				log.Fatalf("Reading template %q failed: %v", fn, err)
			}
			known[name] = true
			if _, err := roottmpl.AddParseTree(name, tmpl.Tree); err != nil {
				log.Fatalf("Adding include %q failed: %v", fn, err)
			}

			// For relative paths we need to make it relative to the include
			pwd, err := filepath.Abs(fn)
			if err != nil {
				log.Fatalf("Cannot determine absolute location of %q: %v", fn, err)
			}
			pwd, err = filepath.Rel(root, filepath.Dir(pwd))
			if err != nil {
				log.Fatalf("Cannot determine location of %q relative to %q: %v", pwd, root, err)
			}
			pwd = string(filepath.Separator) + pwd
			for _, iname := range extractIncludes(tmpl) {
				ifn := iname
				if !strings.HasPrefix(ifn, "/") {
					ifn = filepath.Join(pwd, ifn)
				}
				newUnresolved[iname] = ifn
			}
		}
		unresolved = newUnresolved
	}

	defines := map[string]string{"parent": parent}
	var buf bytes.Buffer
	if err := roottmpl.Execute(&buf, defines); err != nil {
		log.Fatalf("Executing template failed: %v", err)
	}

	if err := os.WriteFile(outputFilename, buf.Bytes(), 0640); err != nil {
		log.Fatalf("Writing output %q failed: %v", outputFilename, err)
	}
}
