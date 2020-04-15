package consumer

// Counter interface is used for exposing basic metrics from the scanner
type Counter interface {
	Add(string, int64)
}

// noopCounter implements counter interface with discard
type noopCounter struct{}

func (n noopCounter) Add(string, int64) {}
