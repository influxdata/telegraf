// Package capnptool provides an API for calling the capnp tool in tests.
package capnptool

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// Tool is the path to the capnp command-line tool.
// It can be used from multiple goroutines.
type Tool string

var cache struct {
	init sync.Once
	tool Tool
	err  error
}

// Find searches PATH for the capnp tool.
func Find() (Tool, error) {
	cache.init.Do(func() {
		path, err := exec.LookPath("capnp")
		if err != nil {
			cache.err = err
			return
		}
		cache.tool = Tool(path)
	})
	return cache.tool, cache.err
}

// Run executes the tool with the given stdin and arguments returns the stdout.
func (tool Tool) Run(stdin io.Reader, args ...string) ([]byte, error) {
	c := exec.Command(string(tool), args...)
	c.Stdin = stdin
	stderr := new(bytes.Buffer)
	c.Stderr = stderr
	out, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("run `%s`: %v; stderr:\n%s", strings.Join(c.Args, " "), err, stderr)
	}
	return out, nil
}

// Encode encodes Cap'n Proto text into the binary representation.
func (tool Tool) Encode(typ Type, text string) ([]byte, error) {
	return tool.Run(strings.NewReader(text), "encode", typ.SchemaPath, typ.Name)
}

// Decode decodes a Cap'n Proto message into text.
func (tool Tool) Decode(typ Type, r io.Reader) (string, error) {
	out, err := tool.Run(r, "decode", "--short", typ.SchemaPath, typ.Name)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// DecodePacked decodes a packed Cap'n Proto message into text.
func (tool Tool) DecodePacked(typ Type, r io.Reader) (string, error) {
	out, err := tool.Run(r, "decode", "--short", "--packed", typ.SchemaPath, typ.Name)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Type is a reference to a Cap'n Proto type in a schema.
type Type struct {
	SchemaPath string
	Name       string
}
