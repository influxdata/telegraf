// Package queue implements a generic queue using a ring buffer.
package queue

// A Queue wraps an Interface to provide queue operations.
type Queue struct {
	q     Interface
	start int
	n     int
	cap   int
}

// New creates a new queue that starts with n elements.  The interface's
// length must not change over the course of the queue's usage.
func New(q Interface, n int) *Queue {
	qq := new(Queue)
	qq.Init(q, n)
	return qq
}

// Init initializes a queue.  The old queue is untouched.
func (q *Queue) Init(r Interface, n int) {
	q.q = r
	q.start = 0
	q.n = n
	q.cap = r.Len()
}

// Len returns the length of the queue.  This is different from the
// underlying interface's length, which is the queue's capacity.
func (q *Queue) Len() int {
	return q.n
}

// Push reserves space for an element on the queue, returning its index.
// If the queue is full, Push returns -1.
func (q *Queue) Push() int {
	if q.n >= q.cap {
		return -1
	}
	i := (q.start + q.n) % q.cap
	q.n++
	return i
}

// Front returns the index of the front of the queue, or -1 if the queue is empty.
func (q *Queue) Front() int {
	if q.n == 0 {
		return -1
	}
	return q.start
}

// Pop pops an element from the queue, returning whether it succeeded.
func (q *Queue) Pop() bool {
	if q.n == 0 {
		return false
	}
	q.q.Clear(q.start)
	q.start = (q.start + 1) % q.cap
	q.n--
	return true
}

// A type implementing Interface can be used to store elements in a Queue.
type Interface interface {
	// Len returns the number of elements available.
	Len() int
	// Clear removes the element at i.
	Clear(i int)
}
