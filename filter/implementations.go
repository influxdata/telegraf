package filter

import "github.com/gobwas/glob"

type filterSingle struct {
	s string
}

func (f *filterSingle) Match(s string) bool {
	return f.s == s
}

type filterNoGlob struct {
	m map[string]struct{}
}

func newFilterNoGlob(filters []string) Filter {
	out := filterNoGlob{m: make(map[string]struct{})}
	for _, filter := range filters {
		out.m[filter] = struct{}{}
	}
	return &out
}

func (f *filterNoGlob) Match(s string) bool {
	_, ok := f.m[s]
	return ok
}

type filterGlobMultiple struct {
	set []glob.Glob
}

func newFilterGlobMultiple(filters []string, separators ...rune) (Filter, error) {
	f := &filterGlobMultiple{set: make([]glob.Glob, 0, len(filters))}

	for _, pattern := range filters {
		g, err := glob.Compile(pattern, separators...)
		if err != nil {
			return nil, err
		}
		f.set = append(f.set, g)
	}

	return f, nil
}

func (f *filterGlobMultiple) Match(s string) bool {
	for _, g := range f.set {
		if g.Match(s) {
			return true
		}
	}

	return false
}
