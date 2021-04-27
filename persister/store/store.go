package store

type Store interface {
	Init() error
	SetState(id string, state interface{}) error
	GetState(id string) (interface{}, bool)
	GetStates() map[string]interface{}

	Read() error
	Write() error
}
