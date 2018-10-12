package globpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

var sepStr = fmt.Sprintf("%v", string(os.PathSeparator))

type GlobPath struct {
	path         string
	hasMeta      bool
	hasSuperMeta bool
	rootGlob     string
	g            glob.Glob
}

func Compile(path string) (*GlobPath, error) {
	out := GlobPath{
		hasMeta:      hasMeta(path),
		hasSuperMeta: hasSuperMeta(path),
		path:         path,
	}

	// if there are no glob meta characters in the path, don't bother compiling
	// a glob object
	if !out.hasMeta || !out.hasSuperMeta {
		return &out, nil
	}

	// find the root elements of the object path, the entry point for recursion
	// when you have a super-meta in your path (which are :
	// glob(/your/expression/until/first/star/of/super-meta))
	out.rootGlob = path[:strings.Index(path, "**")+1]
	var err error
	if out.g, err = glob.Compile(path, os.PathSeparator); err != nil {
		return nil, err
	}
	return &out, nil
}

func (g *GlobPath) Match() map[string]os.FileInfo {
	out := make(map[string]os.FileInfo)
	if !g.hasMeta {
		info, err := os.Stat(g.path)
		if err == nil {
			out[g.path] = info
		}
		return out
	}
	if !g.hasSuperMeta {
		files, _ := filepath.Glob(g.path)
		for _, file := range files {
			info, err := os.Stat(file)
			if err == nil {
				out[file] = info
			}
		}
		return out
	}
	roots, err := filepath.Glob(g.rootGlob)
	if err != nil {
		return out
	}
	walkfn := func(path string, info os.FileInfo, _ error) error {
		if g.g.Match(path) {
			out[path] = info
		}
		return nil

	}
	for _, root := range roots {
		filepath.Walk(root, walkfn)
	}
	return out
}

// hasMeta reports whether path contains any magic glob characters.
func hasMeta(path string) bool {
	return strings.IndexAny(path, "*?[") >= 0
}

// hasSuperMeta reports whether path contains any super magic glob characters (**).
func hasSuperMeta(path string) bool {
	return strings.Index(path, "**") >= 0
}
