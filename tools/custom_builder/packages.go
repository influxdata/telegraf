package main

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/influxdata/telegraf/filter"
)

// Define the categories we can handle and package filters
var packageFilter = filter.MustCompile([]string{
	"*/all",
	"*/*_test",
	"inputs/example",
	"inputs/main",
})

type packageInfo struct {
	Category      string
	Plugin        string
	Path          string
	Tag           string
	DefaultParser string
}

type packageCollection struct {
	packages       map[string][]packageInfo
	defaultParsers []string
}

// Define the package exceptions
var exceptions = map[string][]packageInfo{
	"parsers": {
		{
			Category: "parsers",
			Plugin:   "influx_upstream",
			Path:     "plugins/parsers/influx/influx_upstream",
			Tag:      "parsers.influx",
		},
	},
	"processors": {
		{
			Category: "processors",
			Plugin:   "aws_ec2",
			Path:     "plugins/processors/aws/ec2",
			Tag:      "processors.aws_ec2",
		},
	},
}

func (p *packageCollection) collectPackagesForCategory(category string) error {
	var entries []packageInfo
	pluginDir := filepath.Join("plugins", category)

	// Add exceptional packages if any
	if pkgs, found := exceptions[category]; found {
		entries = append(entries, pkgs...)
	}

	// Walk the directory and get the packages
	elements, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	for _, element := range elements {
		path := filepath.Join(pluginDir, element.Name())
		if !element.IsDir() {
			continue
		}

		var fset token.FileSet
		pkgs, err := parser.ParseDir(&fset, path, sourceFileFilter, parser.ParseComments)
		if err != nil {
			log.Printf("parsing directory %q failed: %v", path, err)
			continue
		}

		for name, pkg := range pkgs {
			if packageFilter.Match(category + "/" + name) {
				continue
			}

			// Extract the names of the plugins registered by this package
			registeredNames := extractRegisteredNames(pkg, category)
			if len(registeredNames) == 0 {
				log.Printf("WARN: Could not extract information from package %q", name)
				continue
			}

			// Extract potential default parsers for input and processor packages
			var defaultParser string
			switch category {
			case "inputs", "processors":
				var err error
				defaultParser, err = extractDefaultParser(path)
				if err != nil {
					log.Printf("Getting default parser for %s.%s failed: %v", category, name, err)
				}
			}

			for _, plugin := range registeredNames {
				path := filepath.Join("plugins", category, element.Name())
				tag := category + "." + element.Name()
				entries = append(entries, packageInfo{
					Category:      category,
					Plugin:        plugin,
					Path:          filepath.ToSlash(path),
					Tag:           tag,
					DefaultParser: defaultParser,
				})
			}
		}
	}
	p.packages[category] = entries

	return nil
}

func (p *packageCollection) FillDefaultParsers() {
	// Make sure we ignore all empty-named parsers which indicate
	// that there is no parser used by the plugin.
	parsers := map[string]bool{"": true}

	// Iterate over all plugins that may have parsers and collect
	// the defaults
	p.defaultParsers = make([]string, 0)
	for _, category := range []string{"inputs", "processors"} {
		for _, pkg := range p.packages[category] {
			name := pkg.DefaultParser
			if seen := parsers[name]; seen {
				continue
			}
			p.defaultParsers = append(p.defaultParsers, name)
			parsers[name] = true
		}
	}
}

func (p *packageCollection) CollectAvailable() error {
	p.packages = make(map[string][]packageInfo)

	for _, category := range categories {
		if err := p.collectPackagesForCategory(category); err != nil {
			return err
		}
	}

	p.FillDefaultParsers()

	return nil
}

func (p *packageCollection) ExtractTags() []string {
	var tags []string
	for category, pkgs := range p.packages {
		_ = category
		for _, pkg := range pkgs {
			tags = append(tags, pkg.Tag)
		}
	}
	sort.Strings(tags)

	return tags
}

func (p *packageCollection) Print() {
	fmt.Println("-------------------------------------------------------------------------------")
	fmt.Println("Enabled plugins:")
	fmt.Println("-------------------------------------------------------------------------------")
	for _, category := range categories {
		pkgs := p.packages[category]
		sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Plugin < pkgs[j].Plugin })

		fmt.Printf("%s (%d):\n", category, len(pkgs))
		for _, pkg := range pkgs {
			fmt.Printf("  %-30s  %s\n", pkg.Plugin, pkg.Path)
		}
		fmt.Println("-------------------------------------------------------------------------------")
	}
}

func sourceFileFilter(d fs.FileInfo) bool {
	return strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go")
}

func findFunctionDecl(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		d, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if d.Name.Name == name && d.Recv == nil {
			return d
		}
	}
	return nil
}

func findAddStatements(decl *ast.FuncDecl, pluginType string) []*ast.CallExpr {
	var statements []*ast.CallExpr
	for _, stmt := range decl.Body.List {
		s, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := s.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		fun, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		e, ok := fun.X.(*ast.Ident)
		if !ok {
			continue
		}
		if e.Name == pluginType && (fun.Sel.Name == "Add" || fun.Sel.Name == "AddStreaming") {
			statements = append(statements, call)
		}
	}

	return statements
}

func extractPluginInfo(file *ast.File, pluginType string, declarations map[string]string) ([]string, error) {
	var registeredNames []string

	decl := findFunctionDecl(file, "init")
	if decl == nil {
		return nil, nil
	}
	calls := findAddStatements(decl, pluginType)
	if len(calls) == 0 {
		return nil, nil
	}
	for _, call := range calls {
		switch arg := call.Args[0].(type) {
		case *ast.Ident:
			resval, found := declarations[arg.Name]
			if !found {
				return nil, fmt.Errorf("cannot resolve registered name variable %q", arg.Name)
			}
			registeredNames = append(registeredNames, strings.Trim(resval, "\""))
		case *ast.BasicLit:
			if arg.Kind != token.STRING {
				return nil, errors.New("registered name is not a string")
			}
			registeredNames = append(registeredNames, strings.Trim(arg.Value, "\""))
		default:
			return nil, fmt.Errorf("unhandled argument type: %v (%T)", arg, arg)
		}
	}
	return registeredNames, nil
}

func extractPackageDeclarations(pkg *ast.Package) map[string]string {
	declarations := make(map[string]string)

	for _, file := range pkg.Files {
		for _, d := range file.Decls {
			gendecl, ok := d.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gendecl.Specs {
				spec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, id := range spec.Names {
					valspec, ok := id.Obj.Decl.(*ast.ValueSpec)
					if !ok || len(valspec.Values) != 1 {
						continue
					}
					valdecl, ok := valspec.Values[0].(*ast.BasicLit)
					if !ok || valdecl.Kind != token.STRING {
						continue
					}
					declarations[id.Name] = strings.Trim(valdecl.Value, "\"")
				}
			}
		}
	}
	return declarations
}

func extractRegisteredNames(pkg *ast.Package, pluginType string) []string {
	var registeredNames []string

	// Extract all declared variables of all files. This might be necessary when
	// using references across multiple files
	declarations := extractPackageDeclarations(pkg)

	// Find the registry Add statement and extract all registered names
	for fn, file := range pkg.Files {
		names, err := extractPluginInfo(file, pluginType, declarations)
		if err != nil {
			log.Printf("%q error: %v", fn, err)
			continue
		}
		registeredNames = append(registeredNames, names...)
	}
	return registeredNames
}

func extractDefaultParser(pluginDir string) (string, error) {
	re := regexp.MustCompile(`^\s*#?\s*data_format\s*=\s*"(.*)"\s*$`)

	// Exception for exec which uses JSON by default
	if filepath.Base(pluginDir) == "exec" {
		return "json", nil
	}

	// Walk all config files in the package directory
	elements, err := os.ReadDir(pluginDir)
	if err != nil {
		return "", err
	}

	for _, element := range elements {
		path := filepath.Join(pluginDir, element.Name())
		if element.IsDir() || filepath.Ext(element.Name()) != ".conf" {
			continue
		}

		// Read the config and search for a "data_format" entry
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			match := re.FindStringSubmatch(scanner.Text())
			if len(match) == 2 {
				return match[1], nil
			}
		}
	}

	return "", nil
}
