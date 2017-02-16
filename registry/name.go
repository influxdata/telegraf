package registry

import (
	"sync"
)

var nameOverride string
var mu sync.Mutex

func SetName(s string) {
	mu.Lock()
	nameOverride = s
	mu.Unlock()
}

func GetName() string {
	mu.Lock()
	defer mu.Unlock()
	return nameOverride
}
