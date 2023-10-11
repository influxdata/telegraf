package gnmi

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
)

// Regular expression to see if a path element contains an origin
var originPattern = regexp.MustCompile(`^([\w-_]+):`)

// Convert a string to a path
func jsonKeyToPath(p string) *gnmiLib.Path {
	elems := strings.Split(strings.TrimSpace(p), "/")

	path := &gnmiLib.Path{
		Elem: make([]*gnmiLib.PathElem, 0, len(elems)),
	}

	for _, e := range elems {
		path.Elem = append(path.Elem, &gnmiLib.PathElem{Name: e})
	}
	normalizePath(path)

	return path
}

// Normalize a path to remove special device oddities
func normalizePath(path *gnmiLib.Path) {
	if path.Origin == "" && len(path.Elem) > 0 {
		groups := originPattern.FindStringSubmatch(path.Elem[0].Name)
		if len(groups) == 2 {
			path.Origin = groups[1]
			path.Elem[0].Name = path.Elem[0].Name[len(groups[1])+1:]
		}
	}
}

// Parse path to path-buffer and tag-field
//
//nolint:revive //function-result-limit conditionally 4 return results allowed
func handlePath(gnmiPath *gnmiLib.Path, tags map[string]string, aliases map[string]string, prefix string) (origin, path, alias string, err error) {
	builder := bytes.NewBufferString(prefix)

	// Some devices do report the origin in the first path element
	// so try to find out if this is the case.
	normalizePath(gnmiPath)

	// Prefix with origin
	if len(gnmiPath.Origin) > 0 {
		origin = gnmiPath.Origin + ":"
	}

	// Parse generic keys from prefix
	for _, elem := range gnmiPath.Elem {
		if len(elem.Name) > 0 {
			if _, err := builder.WriteRune('/'); err != nil {
				return "", "", "", err
			}
			if _, err := builder.WriteString(elem.Name); err != nil {
				return "", "", "", err
			}
		}
		name := builder.String()

		if _, exists := aliases[origin+name]; exists {
			alias = origin + name
		} else if _, exists := aliases[name]; exists {
			alias = name
		}

		if tags != nil {
			for key, val := range elem.Key {
				key = strings.ReplaceAll(key, "-", "_")

				// Use short-form of key if possible
				if _, exists := tags[key]; exists {
					tags[name+"/"+key] = val
				} else {
					tags[key] = val
				}
			}
		}
	}

	return origin, builder.String(), alias, nil
}

// equalPathNoKeys checks if two gNMI paths are equal, without keys
func equalPathNoKeys(a *gnmiLib.Path, b *gnmiLib.Path) bool {
	if len(a.Elem) != len(b.Elem) {
		return false
	}
	for i := range a.Elem {
		if a.Elem[i].Name != b.Elem[i].Name {
			return false
		}
	}
	return true
}

func extractPathKeys(gpath *gnmiLib.Path) []*gnmiLib.PathElem {
	var newPath []*gnmiLib.PathElem
	for _, elem := range gpath.Elem {
		if elem.Key != nil {
			newPath = append(newPath, elem)
		}
	}
	return newPath
}

func extractTagsFromPath(gpath *gnmiLib.Path) map[string]string {
	fmt.Printf("# path: %v\n", pathToStringNoKeys(gpath))
	tags := make(map[string]string)
	var name string
	for _, elem := range gpath.Elem {
		if len(elem.Name) > 0 {
			name += "/" + elem.Name
		}

		for key, val := range elem.Key {
			key = strings.ReplaceAll(key, "-", "_")

			// Use short-form of key if possible
			if _, exists := tags[key]; exists {
				fmt.Printf("* key %q exists\n", key)
				tags[name+"/"+key] = val
			} else {
				fmt.Printf("* key %q does not exist\n", key)
				tags[key] = val
			}
		}
	}
	return tags
}

func pathToStringNoKeys(p *gnmiLib.Path) string {
	if p == nil {
		return ""
	}

	segments := make([]string, 0, len(p.Elem))
	for _, e := range p.Elem {
		if e.Name != "" {
			segments = append(segments, e.Name)
		}
	}
	if len(segments) == 0 {
		return ""
	}

	out := "/" + strings.Join(segments, "/")
	if p.Origin != "" {
		out = p.Origin + ":" + out
	}
	return out
}

func isSubPath(path, sub *gnmiLib.Path) bool {
	// If both set an origin it has to match. Otherwise we ignore the origin
	if path.Origin != "" && sub.Origin != "" && path.Origin != sub.Origin {
		return false
	}

	// The "parent" path should have the same length or be shorter than the
	// sub-path to have a chance to match
	if len(path.Elem) > len(sub.Elem) {
		return false
	}

	// Compare the elements and exit if we find a mismatch
	for i, p := range path.Elem {
		if p.Name != sub.Elem[i].Name {
			return false
		}
	}

	return true
}
