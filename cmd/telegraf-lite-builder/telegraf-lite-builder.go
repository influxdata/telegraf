package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"archive/zip"
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"go/build"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"text/template"
)

//go:generate go run ./scripts/templatebuilder/buildtemplate.go

const binaryName = "telegraf-lite"

type listFlags []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *listFlags) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *listFlags) Set(value string) error {
	*i = append(*i, strings.Split(value, ",")...)
	return nil
}

func main() {
	var cmdLnBuildOpt buildOptions
	var server serverCfg
	flag.Bool("h", false, "displays the help page")
	cmdLine := flag.NewFlagSet("command line flags", flag.ExitOnError)
	cmdLine.Var((*listFlags)(&cmdLnBuildOpt.Aggregators), "aggregators", "comma-separated list of aggregator plugins, defaults to `all`")
	cmdLine.Var((*listFlags)(&cmdLnBuildOpt.Inputs), "inputs", "comma-separated list of input plugins, defaults to `all`")
	cmdLine.Var((*listFlags)(&cmdLnBuildOpt.Outputs), "outputs", "comma-separated list of output plugins, defaults to `all`")
	cmdLine.Var((*listFlags)(&cmdLnBuildOpt.Processors), "processors", "comma-separated list of processor plugins, defaults to `all`")
	cmdLine.StringVar(&cmdLnBuildOpt.Compression, "compression", "", "which compression scheme to use, uncompressed is default")
	cmdLine.BoolVar(&cmdLnBuildOpt.Upx, "upx", true, "strip out debugging statements")

	serverFlgs := flag.NewFlagSet("server flags", flag.ExitOnError)
	serverFlgs.StringVar(&server.bind, "bind", "localhost:8080", "the address to bind the server to")
	flag.Parse()
	// TODO(docmerlin): add the command line args to the help.
	if len(os.Args) == 0 {
	}

	switch flag.Arg(0) {
	case "serve":
		serverFlgs.Parse(os.Args[2:])
		log.Fatal(server.Serve())
	case "":
		cmdLine.Parse(os.Args[1:])
		name := "telegraf-lite"
		switch cmdLnBuildOpt.Compression {
		case "gzip":
			name += ".gz"
		case "zip":
			name += ".zip"
		}
		f, err := os.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		b, err := newBuilder(mainTemplateBase64)
		if err != nil {
			log.Fatal(err)
		}
		b.build(context.Background(), cmdLnBuildOpt, f)
	default:
		//todo(docmerlin): display help page.
	}
}

type serverCfg struct {
	bind    string
	builder builder
	mux     http.ServeMux
}

func (s *serverCfg) Serve() error {
	//TODO(docmerlin): include build info, version etc on the download page
	b, err := newBuilder(mainTemplateBase64)
	if err != nil {
		return err
	}
	s.builder = *b
	tmpl := &template.Template{}
	tmpl, err = tmpl.Parse(page)
	if err != nil {
		return err
	}
	pageBuf := bytes.Buffer{}
	if err = tmpl.Execute(&pageBuf, s.builder); err != nil {
		return err
	}
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pageBuf.WriteTo(w)
	})
	s.mux.HandleFunc("/v1/download", func(w http.ResponseWriter, r *http.Request) {
		s.build(w, r)
	})
	return http.ListenAndServe(s.bind, &s.mux)
}

func (s *serverCfg) build(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	q := r.URL.Query()
	bo := buildOptions{}
	bo.Aggregators = q["a"]
	bo.Inputs = q["i"]
	bo.Outputs = q["o"]
	bo.Processors = q["p"]
	bo.GOOS = q.Get("GOOS")
	bo.GOARCH = q.Get("GOARCH")
	bo.Compression = q.Get("c")
	bo.Upx = len(q["upx"]) > 0
	log.Println("here", bo)
	if err := s.builder.Validate(&bo); err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(bo)
	buf := &bytes.Buffer{}
	if err := s.builder.build(r.Context(), bo, buf); err != nil {
		log.Println("E! " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	switch bo.Compression {
	case "gzip":
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", `attachment; filename="telegraf-lite.gzip"`)
	case "zip":
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="telegraf-lite.zip"`)
	case "tgz":
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", `attachment; filename="telegraf-lite.tgz"`)
	}
	if _, err := buf.WriteTo(w); err != nil {
		log.Println(err)
	}
}

func newBuilder(encodedTemplate string) (*builder, error) {
	//TODO(docmerlin): parse the Readme for a description of each plugin
	mainTemplate := &template.Template{}
	tmpl, err := base64.StdEncoding.DecodeString(encodedTemplate)
	if err != nil {
		return nil, err
	}
	mainTemplate, err = mainTemplate.Parse(string(tmpl))
	if err != nil {
		return nil, err
	}
	decodedgopkglock, err := base64.StdEncoding.DecodeString(gopkgLock)
	if err != nil {
		return nil, err
	}
	decodedgopkgtoml, err := base64.StdEncoding.DecodeString(gopkgtoml)
	if err != nil {
		return nil, err
	}

	b := &builder{
		template:    mainTemplate,
		gopkglock:   decodedgopkglock,
		gopkgtoml:   decodedgopkgtoml,
		Aggregators: map[string]string{},
		Inputs:      map[string]string{},
		Outputs:     map[string]string{},
		Processors:  map[string]string{},
		Compression map[string]string{},
	}
	p, err := build.Default.Import("github.com/influxdata/telegraf/plugins/aggregators/all", "all", build.IgnoreVendor)
	if err != nil {
		return nil, err
	}
	for _, x := range p.Imports {
		y := strings.TrimPrefix(x, "github.com/influxdata/telegraf/plugins/aggregators/")
		b.Aggregators[y] = ""
	}
	p, err = build.Default.Import("github.com/influxdata/telegraf/plugins/inputs/all", "all", build.IgnoreVendor)
	if err != nil {
		return nil, err
	}
	for _, x := range p.Imports {
		y := strings.TrimPrefix(x, "github.com/influxdata/telegraf/plugins/inputs/")
		b.Inputs[y] = ""
	}
	p, err = build.Default.Import("github.com/influxdata/telegraf/plugins/outputs/all", "all", build.IgnoreVendor)
	if err != nil {
		return nil, err
	}
	for _, x := range p.Imports {
		y := strings.TrimPrefix(x, "github.com/influxdata/telegraf/plugins/outputs/")
		b.Outputs[y] = ""
	}
	p, err = build.Default.Import("github.com/influxdata/telegraf/plugins/processors/all", "all", build.IgnoreVendor)
	if err != nil {
		return nil, err
	}
	for _, x := range p.Imports {
		y := strings.TrimPrefix(x, "github.com/influxdata/telegraf/plugins/processors/")
		b.Processors[y] = ""
	}
	// TODO(docmerlin): get a description or exceptions for each one and add it here.
	b.GOOS = map[string]string{
		"darwin":    "darwin",
		"dragonfly": "dragonfly",
		"freebsd":   "freebsd",
		"linux":     "linux",
		"nacl":      "nacl",
		"netbsd":    "netbsd",
		"openbsd":   "openbsd",
		"plan9":     "plan9",
		"solaris":   "solaris",
		"windows":   "windows",
	}
	b.GOARCH = map[string]string{
		"amd64":    "64 bit Intel and AMD",
		"386":      "32 bit Intel",
		"amd64p32": "64 bit intel and AMD with 32 bit pointers",
		"arm":      "32 bit ARM",
		"arm64":    "64 bit ARM",
		"ppc64":    "64 bit PowerPC",
		"ppc64le":  "64 bit PowerPC in little endian mode",
		"mips":     "32 bit MIPS",
		"mipsle":   "32 bit MIPS in little endian mode",
		"mips64":   "64 bit MIPS",
		"mips64le": "64 bit MIPS in little endian mode",
		"s390x":    "IBM s390x mainframe architecture",
	}
	b.Packaging = map[string]string{
		"gzip": "gzip",
		"tgz":"tgz",
		"zip":"zip",
	}
	// TODO(upx): check for UPX
	return b, nil
}

type builder struct {
	template    *template.Template
	gopkgtoml   []byte
	gopkglock   []byte
	Aggregators map[string]string `json:"a"`
	Inputs      map[string]string `json:"i"`
	Outputs     map[string]string	`json:"o"`
	Processors  map[string]string	`json:"p"`
	GOOS        map[string]string	`json:"GOOS"`
	GOARCH      map[string]string	`json:"GOARCH"`
	Packaging map[string]string `json:"packaging"`
	Upx         bool	`json:"upx"`
}

type buildOptions struct {
	Aggregators []string `json:"aggregators"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	Processors  []string `json:"processors"`
	GOOS        string   `json:"goos"`
	GOARCH      string   `json:"goarch"`
	Packaging string   `json:"packaging"`
	Upx         bool     `json:"upx"`
}

func (bo *buildOptions) SetDefaults() {
	if len(bo.Aggregators) == 0 {
		bo.Aggregators = []string{"all"}
	}
	if len(bo.Inputs) == 0 {
		bo.Inputs = []string{"all"}
	}
	if len(bo.Outputs) == 0 {
		bo.Outputs = []string{"all"}
	}
	if len(bo.Processors) == 0 {
		bo.Processors = []string{"all"}
	}
}

func (b *builder) Validate(bo *buildOptions) error {
	// if any of the options are "all" then just use "all"
	for i := range bo.Aggregators {
		if bo.Aggregators[i] == "all" {
			bo.Aggregators = append(bo.Aggregators[:0], "all")
			break
		}
		if _, ok := b.Aggregators[bo.Aggregators[i]]; !ok {
			return optionError{fmt.Errorf("aggregator plugin %s does not exist", bo.Aggregators[i])}
		}
	}

	// if any of the options are "all" then just use "all"
	for i := range bo.Inputs {
		if bo.Inputs[i] == "all" {
			bo.Inputs = append(bo.Inputs[:0], "all")
			break
		}
		if _, ok := b.Inputs[bo.Inputs[i]]; !ok {
			return optionError{fmt.Errorf("input plugin %s does not exist", bo.Inputs[i])}
		}
	}
	// if any of the options are "all" then just use "all"
	for i := range bo.Outputs {
		if bo.Outputs[i] == "all" {
			bo.Outputs = append(bo.Outputs[:0], "all")
			break
		}
		if _, ok := b.Outputs[bo.Outputs[i]]; !ok {
			return optionError{fmt.Errorf("output plugin %s does not exist", bo.Outputs[i])}
		}
	}

	for i := range bo.Processors {
		if bo.Processors[i] == "all" {
			bo.Processors = append(bo.Processors[:0], "all")
			break
		}
		if _, ok := b.Processors[bo.Processors[i]]; !ok {
			return optionError{fmt.Errorf("processor plugin %s does not exist", bo.Processors[i])}
		}
	}
	if _, ok := b.GOOS[bo.GOOS]; !ok {
		return optionError{fmt.Errorf("%s is not a supported operating system", bo.GOOS)}
	}
	if _, ok := b.GOARCH[bo.GOARCH]; !ok {
		return optionError{fmt.Errorf("%s is not a supported architecture", bo.GOARCH)}
	}
	if _, ok:= b.Packaging[bo.Packaging];!ok{
		return optionError{errors.New("%s packaging not supported", bo.Packaging)}
	}
}

type optionError struct {
	error
}

func (o *optionError) Unwrap() error {
	return o.error
}

func (b *builder) build(ctx context.Context, bo buildOptions, w io.Writer) error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir(dir, "tg-lite-build-")
	if err != nil {
		return fmt.Errorf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	in := filepath.Join(tmpDir, "main.go")
	f, err := os.Create(in)
	if err != nil {
		return fmt.Errorf("error creating temp file %q: %v", in, err)
	}
	if err := b.template.Execute(f, bo); err != nil {
		f.Close()
		return fmt.Errorf("error templating file %q: %v", in, err)
	}
	f.Close()

	// create gopkg files
	f, err = os.Create(filepath.Join(tmpDir, "Gopkg.lock"))
	if err != nil {
		return fmt.Errorf("error creating temp Gopkg.lock: %v", err)
	}
	if _, err = f.Write(b.gopkglock); err != nil {
		f.Close()
		return fmt.Errorf("error writing temp file Gopkg.lock %v", err)

	}
	f.Close()

	// create gopkg files
	f, err = os.Create(filepath.Join(tmpDir, "Gopkg.toml"))
	if err != nil {
		return fmt.Errorf("error creating temp Gopkg.toml: %v", err)
	}
	if _, err = f.Write(b.gopkgtoml); err != nil {
		f.Close()
		return fmt.Errorf("error writing temp file Gopkg.toml %v", err)
	}
	f.Close()

	goCache := filepath.Join(tmpDir, "gocache")
	var cmd *exec.Cmd
	if bo.Upx {
		// here we cheat and shell out.
		command := []string{"-c", `cd ` + tmpDir + `&&go build -ldflags="-s -w" -o a.out && upx a.out -obuilt.out > /dev/null && cat built.out > /dev/stdout`, in}
		cmd = exec.CommandContext(ctx, "sh", command...)
	} else {
		cmd = exec.CommandContext(ctx, "go", "build", "-ldflags=-s -w", "-o", "/dev/stdout", in)
	}
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"), // we need this for clang.
		"GOOS=" + bo.GOOS,
		"GOARCH=" + bo.GOARCH,
		"GOPATH=" + os.Getenv("GOPATH"),
		"GOCACHE=" + goCache}

	if err != nil {
		return fmt.Errorf("error reading result of build %v", err)
	}

	switch bo.Compression {
	case "gzip":
		gw := gzip.NewWriter(w)
		defer gw.Close()
		cmd.Stdout = gw
	case "zip":
		z := zip.NewWriter(w)
		defer z.Close()
		zfile, err := z.Create(binaryName)
		if err != nil {
			return fmt.Errorf("error writing zip file %v", err)
		}
		cmd.Stdout = zfile
	case "tgz":
		z:= tar.NewWriter(w)
		defer z.Close()
		if err := tw.WriteHeader(&tar.Header{
			Name:
		}); err != nil {
			return err
		}

	}
	//TODO(docmerlin): make this provide more visibility into this sort of error
	// go func() {
	// 	r, err := cmd.StderrPipe()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	res, _ := ioutil.ReadAll(r)
	// 	fmt.Println(string(res))
	// }()
	return cmd.Run()
}

const page = `<!DOCTYPE html>
<html>
  <head>
    <title>Telegraph-lite Downloader</title>
  </head>
  <body>
    <h1>Welcome to Telegraf-lite Downloader</h1>
    <h2>pick your plugins!</h2>
    <form action="/v1/download">
	  <h3>Aggregators</h3>
	  <p>
		<input type="checkbox" name="a" value="all">all
	  </p>{{range $plugin, $desc := .Aggregators}}
	  <p>
        <input type="checkbox" name="a" value="{{- $plugin -}}">{{ $plugin }}
      </p>{{end}}
	  <h3>Inputs</h3>
	  <p>
		<input type="checkbox" name="i" value="all">all
	  </p>{{range $plugin, $desc := .Inputs}}
      <p>
        <input type="checkbox" name="i" value="{{- $plugin -}}">{{ $plugin }}
      </p>{{end}}
	  <h3>Outputs</h3>
	  <p>
		<input type="checkbox" name="o" value="all">all
	  </p>{{range $plugin, $desc := .Outputs}}
      <p>
        <input type="checkbox" name="o" value="{{- $plugin -}}">{{ $plugin }}
      </p>{{end}}
	  <h3>Processors</h3>
	  <p>
		<input type="checkbox" name="p" value="all">all
	  </p>{{range $plugin, $desc := .Processors}}
      <p>
        <input type="checkbox" name="p" value="{{- $plugin -}}">{{ $plugin }}
      </p>{{end}}
	  <h3>Packaging</h3>
	  {{}}
      <p>
        <input type="radio" name="c" value="gzip" checked>gzip
      </p>
      <h3>Architecture</h3>{{range $plugin, $desc := .GOARCH}}
      <p>
        <input type="radio" name="GOARCH" value="{{- $plugin -}}">{{ $desc }}
      </p>{{end}}
      <h3>Operating System</h3>{{range $plugin, $desc := .GOOS}}
      <p>
        <input type="radio" name="GOOS" value="{{- $plugin -}}">{{ $desc }}
	  </p>{{end}}
	  <h3>Packaging</h3>{{range $plugin, $desc := .Packaging}}
	  <p>
	    <input type="radio" name="packaging" value="{{- $plugin -}}">{{ $desc }}
	  </p>
	  {{ if .Upx }}
	  <h3>Compress with upx (slightly slower startup times, but much smaller binary).</h3> 
	  <p>
		<input type="checkbox" name="upx" checked>
	  </p>{{ end }}
      <p>
        <input type="submit" value="download">
      </p>
    </form>
  </body>
</html>`
