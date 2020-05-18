package telegraf

// StateStore ...
type State interface {
	Open() error
	Close()
	Store(string, interface{}) error
	Load(string) (interface{}, error)
	Flush()
}
