package compare

type Expected struct {
	Fetch      bool
	Partial    bool
	Ordered    bool
	FetchCount int
	Val        interface{}
}

func (expected Expected) SetOrdered(ordered bool) Expected {
	expected.Ordered = ordered

	return expected
}

func (expected Expected) SetPartial(partial bool) Expected {
	expected.Partial = partial

	return expected
}

type Regex string

func IsUUID() Regex {
	return Regex("[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}")
}

func MatchesRegexp(expr string) Regex {
	return Regex(expr)
}

func UnorderedMatch(v interface{}) Expected {
	return Expected{
		Ordered: false,
		Partial: false,
		Val:     v,
	}
}

func PartialMatch(v interface{}) Expected {
	return Expected{
		Ordered: false,
		Partial: true,
		Val:     v,
	}
}
