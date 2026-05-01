package opcua

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

const recursiveSegment = "**"

// PathPattern is a compiled glob pattern over OPC UA browse-path segments.
//
// Patterns are split on "/". Within a segment, path.Match semantics apply:
// "*" matches any sequence of non-separator characters, "?" matches a single
// character, "[abc]" matches a character class, and "\" escapes the next
// character. A standalone "**" segment matches zero or more segments.
//
// Browse names containing a literal "/" cannot be matched: the separator is
// reserved.
type PathPattern struct {
	segments []string
	raw      string
}

// CompilePathPattern parses a pattern into a PathPattern. Empty patterns,
// empty segments, and segments with malformed glob syntax are rejected.
func CompilePathPattern(pattern string) (*PathPattern, error) {
	if pattern == "" {
		return nil, errors.New("empty pattern")
	}

	raw := strings.Split(pattern, "/")
	segments := make([]string, 0, len(raw))
	for _, s := range raw {
		if s == "" {
			return nil, fmt.Errorf("empty segment in pattern %q", pattern)
		}
		if s != recursiveSegment {
			if _, err := path.Match(s, ""); err != nil {
				return nil, fmt.Errorf("invalid segment %q in pattern %q: %w", s, pattern, err)
			}
		}
		// Adjacent ** segments collapse into one to keep matching linear.
		if s == recursiveSegment && len(segments) > 0 && segments[len(segments)-1] == recursiveSegment {
			continue
		}
		segments = append(segments, s)
	}

	return &PathPattern{segments: segments, raw: pattern}, nil
}

// Match reports whether the browse-path segments match the compiled pattern.
func (p *PathPattern) Match(segments []string) bool {
	return matchSegments(p.segments, segments)
}

// String returns the original pattern text.
func (p *PathPattern) String() string {
	return p.raw
}

func matchSegments(pattern, segments []string) bool {
	for len(pattern) > 0 {
		if pattern[0] == recursiveSegment {
			pattern = pattern[1:]
			if len(pattern) == 0 {
				return true
			}
			for i := 0; i <= len(segments); i++ {
				if matchSegments(pattern, segments[i:]) {
					return true
				}
			}
			return false
		}

		if len(segments) == 0 {
			return false
		}

		// path.Match was validated at compile time, but it can still surface
		// ErrBadPattern for inputs that exercise lazy validation paths.
		// Treat any error as no-match so the runtime never panics.
		ok, err := path.Match(pattern[0], segments[0])
		if err != nil || !ok {
			return false
		}
		pattern = pattern[1:]
		segments = segments[1:]
	}

	return len(segments) == 0
}
