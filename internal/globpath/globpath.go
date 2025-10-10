package globpath

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type GlobPath struct {
	path         string
	hasMeta      bool
	HasSuperMeta bool
	rootGlob     string
}

func Compile(path string) (*GlobPath, error) {
	out := GlobPath{
		hasMeta:      hasMeta(path),
		HasSuperMeta: hasSuperMeta(path),
		path:         filepath.FromSlash(path),
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
	return &out, nil
}

// Match returns all files matching the expression.
// If it's a static path, returns path.
// All returned path will have the host platform separator.
func (g *GlobPath) Match() []string {
	pattern := g.path

	// Convert old-style ** patterns to doublestar v4 format
	// Old: **.txt -> New: **/*.txt
	pattern = strings.ReplaceAll(pattern, "**.", "**/*.")

	files, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		// Return nil on error to maintain backward compatibility
		return nil
	}

	// Special handling for "**" pattern to exclude the base directory itself
	// This maintains backward compatibility with existing tests
	if strings.HasSuffix(g.path, "**") && len(files) > 0 {
		baseDir := filepath.Dir(g.path)
		result := make([]string, 0, len(files))
		for _, f := range files {
			// Exclude the base directory itself from results
			if f != baseDir {
				result = append(result, f)
			}
		}
		return result
	}

	return files
}

// MatchString tests the path string against the glob.  The path should contain
// the host platform separator.
func (g *GlobPath) MatchString(path string) bool {
	if !g.HasSuperMeta {
		// Use standard library for simple patterns without **
		res, err := filepath.Match(g.path, path)
		if err != nil {
			// Invalid pattern, return false to maintain backward compatibility
			return false
		}
		return res
	}

	// Use doublestar for patterns with ** support
	matched, err := doublestar.PathMatch(g.path, path)
	if err != nil {
		// Invalid pattern, return false to maintain backward compatibility
		return false
	}
	return matched
}

// GetRoots returns a list of files and directories which should be optimal
// prefixes of matching files when you have a super-meta in your expression :
// - any directory under these roots may contain a matching file
// - no file outside of these roots can match the pattern
// Note that it returns both files and directories.
// All returned path will have the host platform separator.
func (g *GlobPath) GetRoots() []string {
	if !g.hasMeta {
		return []string{g.path}
	}

	if !g.HasSuperMeta {
		matches, err := filepath.Glob(g.path)
		if err != nil {
			// Invalid pattern, return nil
			return nil
		}
		return matches
	}

	roots, err := filepath.Glob(g.rootGlob)
	if err != nil {
		// Invalid pattern, return nil
		return nil
	}
	return roots
}

// hasMeta reports whether path contains any magic glob characters.
func hasMeta(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

// hasSuperMeta reports whether path contains any super magic glob characters (**).
func hasSuperMeta(path string) bool {
	return strings.Contains(path, "**")
}
