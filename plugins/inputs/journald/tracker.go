//go:build linux

package journald

import (
	"sync"

	"github.com/influxdata/telegraf"
)

type pendingEntry struct {
	cursor    string
	delivered bool
}

// cursorTracker keeps track of the entries that are not yet delivered by the
// outputs. The committed cursor only advances to an entry once this entry and
// all entries read before it are delivered, so resuming at the committed
// cursor never skips an entry. Deliveries may happen in any order.
type cursorTracker struct {
	pending []*pendingEntry
	index   map[telegraf.TrackingID]*pendingEntry
	early   map[telegraf.TrackingID]bool
	cursor  string

	sync.Mutex
}

func newCursorTracker() *cursorTracker {
	return &cursorTracker{
		index: make(map[telegraf.TrackingID]*pendingEntry),
		early: make(map[telegraf.TrackingID]bool),
	}
}

// add registers a new entry in reading order
func (t *cursorTracker) add(id telegraf.TrackingID, cursor string) {
	t.Lock()
	defer t.Unlock()

	e := &pendingEntry{cursor: cursor}
	// The delivery might be reported before the entry is registered
	if t.early[id] {
		delete(t.early, id)
		e.delivered = true
	} else {
		t.index[id] = e
	}
	t.pending = append(t.pending, e)
	t.advance()
}

// done marks the entry with the given ID as delivered
func (t *cursorTracker) done(id telegraf.TrackingID) {
	t.Lock()
	defer t.Unlock()

	e, found := t.index[id]
	if !found {
		t.early[id] = true
		return
	}
	delete(t.index, id)
	e.delivered = true
	t.advance()
}

func (t *cursorTracker) advance() {
	var n int
	for n < len(t.pending) && t.pending[n].delivered {
		if c := t.pending[n].cursor; c != "" {
			t.cursor = c
		}
		t.pending[n] = nil
		n++
	}
	t.pending = t.pending[n:]
}

func (t *cursorTracker) committed() string {
	t.Lock()
	defer t.Unlock()
	return t.cursor
}

func (t *cursorTracker) setCommitted(cursor string) {
	t.Lock()
	defer t.Unlock()
	t.cursor = cursor
}
