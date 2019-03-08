package globpath

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/karrick/godirwalk"
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
// If it's a static path, returns path
func (g *GlobPath) Match() []string {
	if !g.hasMeta {
		return []string{g.path}
	}
	if !g.HasSuperMeta {
		files, _ := filepath.Glob(g.path)
		return files
	}
	roots, err := filepath.Glob(g.rootGlob)
	if err != nil {
		return []string{}
	}
	out := []string{}
	walkfn := func(path string, _ *godirwalk.Dirent) error {
		if g.g.Match(path) {
			out = append(out, path)
		}
		return nil

	}
	for _, root := range roots {
		fileinfo, err := os.Stat(root)
		if err != nil {
			continue
		}
		if !fileinfo.IsDir() {
			if g.MatchString(root) {
				out = append(out, root)
			}
			continue
		}
		godirwalk.Walk(root, &godirwalk.Options{
			Callback: walkfn,
			Unsorted: true,
		})
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
