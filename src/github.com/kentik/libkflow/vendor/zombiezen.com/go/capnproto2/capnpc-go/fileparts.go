package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"zombiezen.com/go/capnproto2"
)

type staticData struct {
	name string
	buf  []byte
}

func (sd *staticData) init(fileID uint64) {
	sd.name = fmt.Sprintf("x_%x", fileID)
	sd.buf = make([]byte, 0, 4096)
}

func (sd *staticData) copyData(obj capnp.Ptr) (staticDataRef, error) {
	m, _, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return staticDataRef{}, err
	}
	err = m.SetRootPtr(obj)
	if err != nil {
		return staticDataRef{}, err
	}
	data, err := m.Marshal()
	if err != nil {
		return staticDataRef{}, err
	}
	ref := staticDataRef{data: sd}
	ref.Start = len(sd.buf)
	sd.buf = append(sd.buf, data...)
	ref.End = len(sd.buf)
	return ref, nil
}

type staticDataRef struct {
	data       *staticData
	Start, End int
}

func (ref staticDataRef) IsValid() bool {
	return ref.Start < ref.End
}

func (ref staticDataRef) String() string {
	return fmt.Sprintf("%s[%d:%d]", ref.data.name, ref.Start, ref.End)
}

type imports struct {
	specs []importSpec
	used  map[string]bool // keyed on import path
}

var capnpImportSpec = importSpec{path: capnpImport, name: "capnp"}

func (i *imports) init() {
	i.specs = nil
	i.used = make(map[string]bool)

	i.reserve(capnpImportSpec)
	i.reserve(importSpec{path: schemasImport, name: "schemas"})
	i.reserve(importSpec{path: serverImport, name: "server"})
	i.reserve(importSpec{path: textImport, name: "text"})
	i.reserve(importSpec{path: contextImport, name: "context"})

	i.reserve(importSpec{path: "math", name: "math"})
	i.reserve(importSpec{path: "strconv", name: "strconv"})
}

func (i *imports) Capnp() string {
	return i.add(importSpec{path: capnpImport, name: "capnp"})
}

func (i *imports) Schemas() string {
	return i.add(importSpec{path: schemasImport, name: "schemas"})
}

func (i *imports) Server() string {
	return i.add(importSpec{path: serverImport, name: "server"})
}

func (i *imports) Text() string {
	return i.add(importSpec{path: textImport, name: "text"})
}

func (i *imports) Context() string {
	return i.add(importSpec{path: contextImport, name: "context"})
}

func (i *imports) Math() string {
	return i.add(importSpec{path: "math", name: "math"})
}

func (i *imports) Strconv() string {
	return i.add(importSpec{path: "strconv", name: "strconv"})
}

func (i *imports) usedImports() []importSpec {
	specs := make([]importSpec, 0, len(i.specs))
	for _, s := range i.specs {
		if i.used[s.path] {
			specs = append(specs, s)
		}
	}
	return specs
}

func (i *imports) byPath(path string) (spec importSpec, ok bool) {
	for _, spec = range i.specs {
		if spec.path == path {
			return spec, true
		}
	}
	return importSpec{}, false
}

func (i *imports) byName(name string) (spec importSpec, ok bool) {
	for _, spec = range i.specs {
		if spec.name == name {
			return spec, true
		}
	}
	return importSpec{}, false
}

func (i *imports) add(spec importSpec) (name string) {
	name = i.reserve(spec)
	i.used[spec.path] = true
	return name
}

// reserve adds an import spec without marking it as used.
func (i *imports) reserve(spec importSpec) (name string) {
	if ispec, ok := i.byPath(spec.path); ok {
		return ispec.name
	}
	if spec.name == "" {
		spec.name = pkgFromImport(spec.path)
	}
	if _, found := i.byName(spec.name); found {
		for base, n := spec.name, uint64(2); ; n++ {
			spec.name = base + strconv.FormatUint(n, 10)
			if _, found = i.byName(spec.name); !found {
				break
			}
		}
	}
	i.specs = append(i.specs, spec)
	return spec.name
}

func pkgFromImport(path string) string {
	if i := strings.LastIndex(path, "/"); i != -1 {
		path = path[i+1:]
	}
	p := []rune(path)
	n := 0
	for _, r := range p {
		if isIdent(r) {
			p[n] = r
			n++
		}
	}
	if n == 0 || !isLower(p[0]) {
		return "pkg" + string(p[:n])
	}
	return string(p[:n])
}

func isLower(r rune) bool {
	return 'a' <= r && r <= 'z' || r == '_'
}

func isIdent(r rune) bool {
	return isLower(r) || 'A' <= r && r <= 'Z' || r >= 0x80 && unicode.IsLetter(r)
}

type importSpec struct {
	path string
	name string
}

func (spec importSpec) String() string {
	if spec.name == "" {
		return strconv.Quote(spec.path)
	}
	return spec.name + " " + strconv.Quote(spec.path)
}
