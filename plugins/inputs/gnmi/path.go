package gnmi

import (
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
)

type keySegment struct {
	name string
	path string
	kv   map[string]string
}

type segment struct {
	namespace string
	id        string
}

type pathInfo struct {
	origin    string
	target    string
	segments  []segment
	keyValues []keySegment
}

func newInfoFromString(path string) *pathInfo {
	if path == "" {
		return &pathInfo{}
	}

	parts := strings.Split(path, "/")

	var origin string
	if strings.HasSuffix(parts[0], ":") {
		origin = strings.TrimSuffix(parts[0], ":")
		parts = parts[1:]
	}

	info := &pathInfo{origin: origin}
	for _, part := range parts {
		if part == "" {
			continue
		}
		info.segments = append(info.segments, segment{id: part})
	}
	info.normalize()

	return info
}

func newInfoFromPathWithoutKeys(path *gnmi.Path) *pathInfo {
	info := &pathInfo{
		origin:   path.Origin,
		segments: make([]segment, 0, len(path.Elem)),
	}
	for _, elem := range path.Elem {
		if elem.Name == "" {
			continue
		}
		info.segments = append(info.segments, segment{id: elem.Name})
	}
	info.normalize()

	return info
}

func newInfoFromPath(paths ...*gnmi.Path) *pathInfo {
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
			if elem.Name != "" {
				info.segments = append(info.segments, segment{id: elem.Name})
			}

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

func (pi *pathInfo) append(paths ...*gnmi.Path) *pathInfo {
	// Copy the existing info
	segments := make([]segment, 0, len(pi.segments))
	path := &pathInfo{
		origin:    pi.origin,
		target:    pi.target,
		segments:  append(segments, pi.segments...),
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
			if elem.Name != "" {
				path.segments = append(path.segments, segment{id: elem.Name})
			}

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
	path.normalize()

	return path
}

func (pi *pathInfo) appendSegments(segments ...string) *pathInfo {
	// Copy the existing info
	seg := make([]segment, 0, len(segments))
	path := &pathInfo{
		origin:    pi.origin,
		target:    pi.target,
		segments:  append(seg, pi.segments...),
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
		path.segments = append(path.segments, segment{id: s})
	}
	path.normalize()

	return path
}

func (pi *pathInfo) normalize() {
	if len(pi.segments) == 0 {
		return
	}

	// Extract namespaces from segments
	for i, s := range pi.segments {
		if ns, id, found := strings.Cut(s.id, ":"); found && strings.Count(s.id, ":") == 1 {
			pi.segments[i].namespace = ns
			pi.segments[i].id = id
		}
	}

	// Remove empty segments
	segments := make([]segment, 0, len(pi.segments))
	for _, s := range pi.segments {
		if s.id != "" {
			segments = append(segments, s)
		}
	}
	pi.segments = segments
}

func (pi *pathInfo) enforceFirstNamespaceAsOrigin() {
	if len(pi.segments) == 0 {
		return
	}

	// Some devices supply the origin as part of the first path element,
	// so try to find and extract it there.
	if pi.segments[0].namespace != "" {
		pi.origin = pi.segments[0].namespace
		pi.segments[0].namespace = ""
	}
}

func (pi *pathInfo) equalsPathNoKeys(path *gnmi.Path) bool {
	if len(pi.segments) != len(path.Elem) {
		return false
	}
	for i, s := range pi.segments {
		if s.id != path.Elem[i].Name {
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
		ps := path.segments[i]
		if p.namespace != "" && ps.namespace != "" && p.namespace != ps.namespace {
			return false
		}
		if p.id != ps.id {
			return false
		}
	}

	return true
}

func (pi *pathInfo) relative(path *pathInfo, withNamespace bool) string {
	if !pi.isSubPathOf(path) || len(pi.segments) == len(path.segments) {
		return ""
	}

	segments := path.segments[len(pi.segments):len(path.segments)]
	var b strings.Builder
	if withNamespace && segments[0].namespace != "" {
		b.WriteString(segments[0].namespace)
		b.WriteString(":")
		b.WriteString(segments[0].id)
	} else {
		b.WriteString(segments[0].id)
	}
	for _, s := range segments[1:] {
		b.WriteString("/")
		if withNamespace && s.namespace != "" {
			b.WriteString(s.namespace)
			b.WriteString(":")
			b.WriteString(s.id)
		} else {
			b.WriteString(s.id)
		}
	}
	return b.String()
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

func (pi *pathInfo) dir() string {
	if len(pi.segments) <= 1 {
		return ""
	}

	var b strings.Builder
	if pi.origin != "" {
		b.WriteString(pi.origin)
		b.WriteString(":")
	}
	for _, s := range pi.segments[:len(pi.segments)-1] {
		b.WriteString("/")
		if s.namespace != "" {
			b.WriteString(s.namespace)
			b.WriteString(":")
			b.WriteString(s.id)
		} else {
			b.WriteString(s.id)
		}
	}
	return b.String()
}

func (pi *pathInfo) base() string {
	if len(pi.segments) == 0 {
		return ""
	}

	s := pi.segments[len(pi.segments)-1]
	if s.namespace != "" {
		return s.namespace + ":" + s.id
	}
	return s.id
}

func (pi *pathInfo) path() (origin, path string) {
	if len(pi.segments) == 0 {
		return pi.origin, "/"
	}

	var b strings.Builder
	for _, s := range pi.segments {
		b.WriteString("/")
		b.WriteString(s.id)
	}

	return pi.origin, b.String()
}

func (pi *pathInfo) fullPath() string {
	var b strings.Builder
	if pi.origin != "" {
		b.WriteString(pi.origin)
		b.WriteString(":")
	}
	if len(pi.segments) == 0 {
		return b.String()
	}

	for _, s := range pi.segments {
		b.WriteString("/")
		if s.namespace != "" {
			b.WriteString(s.namespace)
			b.WriteString(":")
			b.WriteString(s.id)
		} else {
			b.WriteString(s.id)
		}
	}

	return b.String()
}

func (pi *pathInfo) String() string {
	if len(pi.segments) == 0 {
		return ""
	}

	origin, path := pi.path()
	var b strings.Builder
	if origin != "" {
		b.WriteString(origin)
		b.WriteString(":")
	}
	b.WriteString(path)
	return b.String()
}

func (pi *pathInfo) tags(pathPrefix bool) map[string]string {
	tags := make(map[string]string, len(pi.keyValues))
	for _, s := range pi.keyValues {
		var prefix string
		if pathPrefix && s.name != "" {
			prefix = s.name + "_"
		}
		// precompute constant path prefix for this keySegment
		pathPrefixStr := s.path + "/"

		for k, v := range s.kv {
			// build the key (prefix + k) and sanitize in one builder
			var kb strings.Builder
			kb.WriteString(prefix)
			kb.WriteString(k)
			key := strings.ReplaceAll(kb.String(), "-", "_")

			// Use short-form of key if possible
			if _, exists := tags[key]; !exists {
				tags[key] = v
				continue
			}
			// build full path/key only when needed
			var fb strings.Builder
			fb.WriteString(pathPrefixStr)
			fb.WriteString(key)
			tags[fb.String()] = v
		}
	}

	return tags
}
