package gnmi

import (
	"sort"
)

// Try to find the alias for the given path
type aliasCandidate struct {
	path, alias string
}

func (h *Handler) AddAliasFromSubscription(s Subscription) error {
	// Build the subscription path without keys
	path, err := ParsePath(s.Origin, s.Path, "")
	if err != nil {
		return err
	}
	info := newInfoFromPathWithoutKeys(path)
	if h.enforceFirstNamespaceAsOrigin {
		info.enforceFirstNamespaceAsOrigin()
	}

	// If the user didn't provide a measurement name, use last path element
	name := s.Name
	if name == "" && len(info.segments) > 0 {
		name = info.segments[len(info.segments)-1].id
	}
	if name != "" && info != nil {
		h.aliases[info] = name
	}
	return nil
}

func (h *Handler) AddAlias(name, path string) {
	info := newInfoFromString(path)
	if h.enforceFirstNamespaceAsOrigin {
		info.enforceFirstNamespaceAsOrigin()
	}

	h.aliases[info] = name
}

func (h *Handler) lookupAlias(info *pathInfo) (aliasPath, alias string) {
	candidates := make([]aliasCandidate, 0, len(h.aliases))
	for i, a := range h.aliases {
		if !i.isSubPathOf(info) {
			continue
		}
		candidates = append(candidates, aliasCandidate{i.String(), a})
	}
	if len(candidates) == 0 {
		return "", ""
	}

	// Reverse sort the candidates by path length so we can use the longest match
	sort.SliceStable(candidates, func(i, j int) bool {
		return len(candidates[i].path) > len(candidates[j].path)
	})

	return candidates[0].path, candidates[0].alias
}
