package globpath

import (
	"fmt"
	"io/ioutil"
	"log"
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

func (g *GlobPath) Match() (map[string]os.FileInfo, error) {
	out := make(map[string]os.FileInfo)
	if !g.hasMeta {
		info, err := os.Stat(g.path)
		if err == nil {
			out[g.path] = info
		} else if os.IsPermission(err) {
			return out, err
		}
		return out, nil
	}

	if !g.hasSuperMeta {
		var returnErr error
		files, _ := filepath.Glob(g.path)
		for _, file := range files {
			info, err := os.Stat(file)
			if err == nil {
				out[file] = info
			} else if os.IsPermission(err) {
				if returnErr.Error() != "" {
					returnErr = fmt.Errorf("%s; %s", returnErr.Error(), err.Error())
				} else {
					returnErr = err
				}
			}
		}
		return out, returnErr
	}

	roots, err := filepath.Glob(g.rootGlob)
	if err != nil {
		return out, err
	}

	walkfn := func(path string, info os.FileInfo, _ error) error {
		if g.g.Match(path) {
			out[path] = info
		} else {
			fi, err := os.Stat(path)
			if err != nil && os.IsPermission(err) {
				return err
			}
			if fi != nil && fi.IsDir() {
				_, err := ioutil.ReadDir(path)
				if err != nil && os.IsPermission(err) {
					return err
				}
			}
		}
		return nil
	}

	for _, root := range roots {
		err := filepath.Walk(root, walkfn)
		if err != nil {
			log.Printf("D! Failed to walk '%s' - %s", root, err.Error())
			continue
		}
	}

	return out, nil
}

// hasMeta reports whether path contains any magic glob characters.
func hasMeta(path string) bool {
	return strings.IndexAny(path, "*?[") >= 0
}

// hasSuperMeta reports whether path contains any super magic glob characters (**).
func hasSuperMeta(path string) bool {
	return strings.Index(path, "**") >= 0
}
