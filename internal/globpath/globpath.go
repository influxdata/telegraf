package globpath

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

type GlobPath struct {
	path         string
	hasMeta      bool
	HasSuperMeta bool
	rootGlob     string
	g            glob.Glob
}

func Compile(path string) (*GlobPath, error) {
	out := GlobPath{
		hasMeta:      hasMeta(path),
		HasSuperMeta: hasSuperMeta(path),
		path:         path,
	}

	// if there are no glob meta characters in the path, don't bother compiling
	// a glob object
	if !out.hasMeta || !out.HasSuperMeta {
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

// Match returns all files matching the expression
func (g *GlobPath) Match() map[string]os.FileInfo {
	out := make(map[string]os.FileInfo)
	if !g.hasMeta {
		info, err := os.Stat(g.path)
		if err == nil {
			out[g.path] = info
		}
		return out
	}
	if !g.HasSuperMeta {
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

// MatchString test a string against the glob
func (g *GlobPath) MatchString(path string) bool {
	if !g.HasSuperMeta {
		res, _ := filepath.Match(g.path, path)
		return res
	}
	return g.g.Match(path)
}

// GetRoots returns a list of files and directories which should be optimal
// prefixes of matching files when you have a super-meta in your expression :
// - any directory under these roots may contain a matching file
// - no file outside of these roots can match the pattern
// Note that it returns both files and directories.
func (g *GlobPath) GetRoots() []string {
	if !g.hasMeta {
		return []string{g.path}
	}
	if !g.HasSuperMeta {
		matches, _ := filepath.Glob(g.path)
		return matches
	}
	roots, _ := filepath.Glob(g.rootGlob)
	return roots
}

// hasMeta reports whether path contains any magic glob characters.
func hasMeta(path string) bool {
	return strings.IndexAny(path, "*?[") >= 0
}

// hasSuperMeta reports whether path contains any super magic glob characters (**).
func hasSuperMeta(path string) bool {
	return strings.Index(path, "**") >= 0
}
