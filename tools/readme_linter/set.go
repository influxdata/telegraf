package main

type set struct {
	m map[string]struct{}
}

func (s *set) add(key string) {
	s.m[key] = struct{}{}
}

func (s *set) has(key string) bool {
	var ok bool
	_, ok = s.m[key]
	return ok
}

func (s *set) forEach(f func(string)) {
	for key := range s.m {
		f(key)
	}
}

func newSet(elems []string) *set {
	s := &set{
		m: make(map[string]struct{}),
	}

	for _, elem := range elems {
		s.add(elem)
	}
	return s
}
