package globpath

import (
	"fmt"
	"io/ioutil"
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
	g            glob.Glob
	root         string
}

func Compile(path string) (*GlobPath, error) {
	out := GlobPath{
		hasMeta:      hasMeta(path),
		hasSuperMeta: hasSuperMeta(path),
		path:         path,
	}

	// if there are no glob meta characters in the path, don't bother compiling
	// a glob object or finding the root directory. (see short-circuit in Match)
	if !out.hasMeta || !out.hasSuperMeta {
		return &out, nil
	}

	var err error
	if out.g, err = glob.Compile(path, os.PathSeparator); err != nil {
		return nil, err
	}
	// Get the root directory for this filepath
	out.root = findRootDir(path)
	return &out, nil
}

func (g *GlobPath) Match() (map[string]os.FileInfo, error) {
	if !g.hasMeta {
		out := make(map[string]os.FileInfo)
		info, err := os.Stat(g.path)
		if err == nil {
			out[g.path] = info
		} else if os.IsPermission(err) {
			return out, err
		}
		return out, nil
	}
	if !g.hasSuperMeta {
		out := make(map[string]os.FileInfo)
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
	return walkFilePath(g.root, g.g)
}

// walk the filepath from the given root and return a list of files that match
// the given glob.
func walkFilePath(root string, g glob.Glob) (map[string]os.FileInfo, error) {
	matchedFiles := make(map[string]os.FileInfo)
	walkfn := func(path string, info os.FileInfo, _ error) error {
		if g.Match(path) {
			matchedFiles[path] = info
		} else {
			fi, err := os.Stat(path)
			if err != nil && os.IsPermission(err) {
				return err
			}
			if fi.IsDir() {
				_, err := ioutil.ReadDir(path)
				if err != nil && os.IsPermission(err) {
					return err
				}
			}
		}
		return nil
	}
	err := filepath.Walk(root, walkfn)
	return matchedFiles, err
}

// find the root dir of the given path (could include globs).
// ie:
//   /var/log/telegraf.conf -> /var/log
//   /home/** ->               /home
//   /home/*/** ->             /home
//   /lib/share/*/*/**.txt ->  /lib/share
func findRootDir(path string) string {
	pathItems := strings.Split(path, sepStr)
	out := sepStr
	for i, item := range pathItems {
		if i == len(pathItems)-1 {
			break
		}
		if item == "" {
			continue
		}
		if hasMeta(item) {
			break
		}
		out += item + sepStr
	}
	if out != "/" {
		out = strings.TrimSuffix(out, "/")
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
