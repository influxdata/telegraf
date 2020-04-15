package senders

type SpanTag struct {
	Key   string
	Value string
}

type SpanLog struct {
	Timestamp int64
	Fields    map[string]string
}
