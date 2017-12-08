// +build mktemplates
// Build tag so that users who run `go get zombiezen.com/go/capnproto2/...` don't install this command.
// cd internal/cmd/mktemplates && go build -tags=mktemplates

// mktemplates is a command to regenerate capnpc-go/templates.go.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template/parse"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: mktemplates OUT DIR")
		os.Exit(64)
	}
	dir := os.Args[2]
	names, err := listdir(dir)
	if err != nil {
		fatalln(err)
	}
	ts := make([]template, len(names))
	for i, name := range names {
		src, err := ioutil.ReadFile(filepath.Join(dir, name))
		if err != nil {
			fatalf("reading template %s: %v", name, err)
		}
		compiled, err := compileTemplate(name, string(src))
		if err != nil {
			fatalf("compiling template %s: %v", name, err)
		}
		ts[i] = template{
			name:    name,
			content: compiled,
		}
	}
	genbuf := new(bytes.Buffer)
	err = generateGo(genbuf, os.Args, ts)
	if err != nil {
		fatalln("generating code:", err)
	}
	code, err := format.Source(genbuf.Bytes())
	if err != nil {
		fatalln("formatting code:", err)
	}
	outname := os.Args[1]
	out, err := os.Create(outname)
	if err != nil {
		fatalf("opening destination %s: %v", outname, err)
	}
	_, err = out.Write(code)
	cerr := out.Close()
	if err != nil {
		fatalf("write to %s: %v", outname, err)
	}
	if cerr != nil {
		fatalln(err)
	}
}

func compileTemplate(name, src string) (string, error) {
	tset, err := parse.Parse(name, src, "{{", "}}", funcStubs)
	if err != nil {
		return "", err
	}
	return tset[name].Root.String(), nil
}

func generateGo(w io.Writer, args []string, ts []template) error {
	src := new(bytes.Buffer)
	for _, t := range ts {
		fmt.Fprintf(src, "{{define %q}}", t.name)
		src.WriteString(t.content)
		src.WriteString("{{end}}")
	}

	// TODO(light): collect errors
	fmt.Fprintln(w, "//go:generate", strings.Join(args, " "))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "package main")
	fmt.Fprintln(w, "import (")
	fmt.Fprintln(w, "\t\"strings\"")
	fmt.Fprintln(w, "\t\"text/template\"")
	fmt.Fprintln(w, ")")
	fmt.Fprintln(w, "var templates = template.Must(template.New(\"\").Funcs(template.FuncMap{")
	fmt.Fprintln(w, "\t\"title\": strings.Title,")
	fmt.Fprintf(w, "}).Parse(\n\t%q))\n", src.Bytes())
	for _, t := range ts {
		if strings.HasPrefix(t.name, "_") {
			continue
		}
		fmt.Fprintf(w, "func render%s(r renderer, p %sParams) error {\n\treturn r.Render(%[2]q, p)\n}\n", strings.Title(t.name), t.name)
	}
	return nil
}

type template struct {
	name    string
	content string
}

func listdir(name string) ([]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	n := 0
	for _, name := range names {
		if !strings.HasPrefix(name, ".") {
			names[n] = name
			n++
		}
	}
	names = names[:n]
	sort.Strings(names)
	return names, nil
}

func fatalln(args ...interface{}) {
	var buf bytes.Buffer
	buf.WriteString("mktemplates: ")
	fmt.Fprintln(&buf, args...)
	os.Stderr.Write(buf.Bytes())
	os.Exit(1)
}

func fatalf(format string, args ...interface{}) {
	var buf bytes.Buffer
	buf.WriteString("mktemplates: ")
	fmt.Fprintf(&buf, format, args...)
	if !bytes.HasSuffix(buf.Bytes(), []byte{'\n'}) {
		buf.Write([]byte{'\n'})
	}
	os.Stderr.Write(buf.Bytes())
	os.Exit(1)
}

var funcStubs = map[string]interface{}{
	// Built-ins
	"and":      variadicBoolStub,
	"call":     func(interface{}, ...interface{}) (interface{}, error) { return nil, nil },
	"eq":       func(arg0 interface{}, args ...interface{}) (bool, error) { return false, nil },
	"ge":       cmpStub,
	"gt":       cmpStub,
	"html":     escaperStub,
	"index":    func(interface{}, ...interface{}) (interface{}, error) { return nil, nil },
	"js":       escaperStub,
	"le":       cmpStub,
	"len":      func(interface{}) (int, error) { return 0, nil },
	"lt":       cmpStub,
	"ne":       cmpStub,
	"not":      func(interface{}) bool { return false },
	"or":       variadicBoolStub,
	"print":    fmt.Sprint,
	"printf":   fmt.Sprintf,
	"println":  fmt.Sprintln,
	"urlquery": escaperStub,

	// App-specific
	"title": strings.Title,
}

func variadicBoolStub(arg0 interface{}, args ...interface{}) interface{} {
	return arg0
}

func cmpStub(interface{}, interface{}) (bool, error) {
	return false, nil
}

func escaperStub(...interface{}) string {
	return ""
}

func importStub() string {
	return ""
}
