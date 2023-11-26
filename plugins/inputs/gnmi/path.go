package gnmi

import (
	"regexp"
	"strings"

	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
)

// Regular expression to see if a path element contains an origin
var originPattern = regexp.MustCompile(`^([\w-]+):`)

type keySegment struct {
	name string
	path string
	kv   map[string]string
}

type pathInfo struct {
	origin    string
	target    string
	segments  []string
	keyValues []keySegment
}

func newInfoFromString(path string) *pathInfo {
	if path == "" {
		return &pathInfo{}
	}

	info := &pathInfo{}
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			info.segments = append(info.segments, s)
		}
	}
	info.normalize()

	return info
}

func newInfoFromPathWithoutKeys(path *gnmiLib.Path) *pathInfo {
	info := &pathInfo{
		origin:   path.Origin,
		segments: make([]string, 0, len(path.Elem)),
	}
	for _, elem := range path.Elem {
		if elem.Name == "" {
			continue
		}
		info.segments = append(info.segments, elem.Name)
	}
	info.normalize()

	return info
}

func newInfoFromPath(paths ...*gnmiLib.Path) *pathInfo {
	if len(paths) == 0 {
		return nil
	}

	info := &pathInfo{}
	if paths[0] != nil {
		info.origin = paths[0].Origin
		info.target = paths[0].Target
	}

	for _, p := range paths {
		if p == nil {
			continue
		}
		for _, elem := range p.Elem {
			if elem.Name == "" {
				continue
			}
			info.segments = append(info.segments, elem.Name)

			if len(elem.Key) == 0 {
				continue
			}
			keyInfo := keySegment{
				name: elem.Name,
				path: info.String(),
				kv:   make(map[string]string, len(elem.Key)),
			}
			for k, v := range elem.Key {
				keyInfo.kv[k] = v
			}
			info.keyValues = append(info.keyValues, keyInfo)
		}
	}
	info.normalize()

	return info
}

func (pi *pathInfo) empty() bool {
	return len(pi.segments) == 0
}

func (pi *pathInfo) append(paths ...*gnmiLib.Path) *pathInfo {
	// Copy the existing info
	path := &pathInfo{
		origin:    pi.origin,
		target:    pi.target,
		segments:  append([]string{}, pi.segments...),
		keyValues: make([]keySegment, 0, len(pi.keyValues)),
	}
	for _, elem := range pi.keyValues {
		keyInfo := keySegment{
			name: elem.name,
			path: elem.path,
			kv:   make(map[string]string, len(elem.kv)),
		}
		for k, v := range elem.kv {
			keyInfo.kv[k] = v
		}
		path.keyValues = append(path.keyValues, keyInfo)
	}

	// Add the new segments
	for _, p := range paths {
		for _, elem := range p.Elem {
			if elem.Name == "" {
				continue
			}
			path.segments = append(path.segments, elem.Name)

			if len(elem.Key) == 0 {
				continue
			}
			keyInfo := keySegment{
				name: elem.Name,
				path: path.String(),
				kv:   make(map[string]string, len(elem.Key)),
			}
			for k, v := range elem.Key {
				keyInfo.kv[k] = v
			}
			path.keyValues = append(path.keyValues, keyInfo)
		}
	}

	return path
}

func (pi *pathInfo) appendSegments(segments ...string) *pathInfo {
	// Copy the existing info
	path := &pathInfo{
		origin:    pi.origin,
		target:    pi.target,
		segments:  append([]string{}, pi.segments...),
		keyValues: make([]keySegment, 0, len(pi.keyValues)),
	}
	for _, elem := range pi.keyValues {
		keyInfo := keySegment{
			name: elem.name,
			path: elem.path,
			kv:   make(map[string]string, len(elem.kv)),
		}
		for k, v := range elem.kv {
			keyInfo.kv[k] = v
		}
		path.keyValues = append(path.keyValues, keyInfo)
	}

	// Add the new segments
	for _, s := range segments {
		if s == "" {
			continue
		}
		path.segments = append(path.segments, s)
	}

	return path
}

func (pi *pathInfo) normalize() {
	if len(pi.segments) == 0 {
		return
	}

	// Some devices supply the origin as part of the first path element,
	// so try to find and extract it there.
	groups := originPattern.FindStringSubmatch(pi.segments[0])
	if len(groups) == 2 {
		pi.origin = groups[1]
		pi.segments[0] = pi.segments[0][len(groups[1])+1:]
	}
}

func (pi *pathInfo) equalsPathNoKeys(path *gnmiLib.Path) bool {
	if len(pi.segments) != len(path.Elem) {
		return false
	}
	for i, s := range pi.segments {
		if s != path.Elem[i].Name {
			return false
		}
	}
	return true
}

func (pi *pathInfo) isSubPathOf(path *pathInfo) bool {
	// If both set an origin it has to match. Otherwise we ignore the origin
	if pi.origin != "" && path.origin != "" && pi.origin != path.origin {
		return false
	}

	// The "parent" path should have the same length or be shorter than the
	// sub-path to have a chance to match
	if len(pi.segments) > len(path.segments) {
		return false
	}

	// Compare the elements and exit if we find a mismatch
	for i, p := range pi.segments {
		if p != path.segments[i] {
			return false
		}
	}

	return true
}

func (pi *pathInfo) keepCommonPart(path *pathInfo) {
	shortestLen := len(pi.segments)
	if len(path.segments) < shortestLen {
		shortestLen = len(path.segments)
	}

	// Compare the elements and stop as soon as they do mismatch
	var matchLen int
	for i, p := range pi.segments[:shortestLen] {
		if p != path.segments[i] {
			break
		}
		matchLen = i + 1
	}
	if matchLen < 1 {
		pi.segments = nil
		return
	}
	pi.segments = pi.segments[:matchLen]
}

func (pi *pathInfo) split() (dir, base string) {
	if len(pi.segments) == 0 {
		return "", ""
	}
	if len(pi.segments) == 1 {
		return "", pi.segments[0]
	}

	dir = "/" + strings.Join(pi.segments[:len(pi.segments)-1], "/")
	if pi.origin != "" {
		dir = pi.origin + ":" + dir
	}
	return dir, pi.segments[len(pi.segments)-1]
}

func (pi *pathInfo) String() string {
	if len(pi.segments) == 0 {
		return ""
	}

	out := "/" + strings.Join(pi.segments, "/")
	if pi.origin != "" {
		out = pi.origin + ":" + out
	}
	return out
}

func (pi *pathInfo) Tags() map[string]string {
	tags := make(map[string]string, len(pi.keyValues))
	for _, s := range pi.keyValues {
		for k, v := range s.kv {
			key := strings.ReplaceAll(k, "-", "_")

			// Use short-form of key if possible
			if _, exists := tags[key]; !exists {
				tags[key] = v
				continue
			}
			tags[s.path+"/"+key] = v
		}
	}

	return tags
}
