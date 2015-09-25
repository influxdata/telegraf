// Copyright (c) 2012 VMware, Inc.

// +build darwin freebsd netbsd openbsd

// Go interface to BSD kqueue process events.
package psnotify

import (
	"syscall"
)

const (
	// Flags (from <sys/event.h>)
	PROC_EVENT_FORK = syscall.NOTE_FORK // fork() events
	PROC_EVENT_EXEC = syscall.NOTE_EXEC // exec() events
	PROC_EVENT_EXIT = syscall.NOTE_EXIT // exit() events

	// Watch for all process events
	PROC_EVENT_ALL = PROC_EVENT_FORK | PROC_EVENT_EXEC | PROC_EVENT_EXIT
)

type kqueueListener struct {
	kq  int                 // The syscall.Kqueue() file descriptor
	buf [1]syscall.Kevent_t // An event buffer for Add/Remove watch
}

// Initialize bsd implementation of the eventListener interface
func createListener() (eventListener, error) {
	listener := &kqueueListener{}
	kq, err := syscall.Kqueue()
	listener.kq = kq
	return listener, err
}

// Initialize Kevent_t fields and propagate changelist for the given pid
func (w *Watcher) kevent(pid int, fflags uint32, flags int) error {
	listener, _ := w.listener.(*kqueueListener)
	event := &listener.buf[0]

	syscall.SetKevent(event, pid, syscall.EVFILT_PROC, flags)
	event.Fflags = fflags

	_, err := syscall.Kevent(listener.kq, listener.buf[:], nil, nil)

	return err
}

// Delete filter for given pid from the queue
func (w *Watcher) unregister(pid int) error {
	return w.kevent(pid, 0, syscall.EV_DELETE)
}

// Add and enable filter for given pid in the queue
func (w *Watcher) register(pid int, flags uint32) error {
	return w.kevent(pid, flags, syscall.EV_ADD|syscall.EV_ENABLE)
}

// Poll the kqueue file descriptor and dispatch to the Event channels
func (w *Watcher) readEvents() {
	listener, _ := w.listener.(*kqueueListener)
	events := make([]syscall.Kevent_t, 10)

	for {
		if w.isDone() {
			return
		}

		n, err := syscall.Kevent(listener.kq, nil, events, nil)
		if err != nil {
			w.Error <- err
			continue
		}

		for _, ev := range events[:n] {
			pid := int(ev.Ident)

			switch ev.Fflags {
			case syscall.NOTE_FORK:
				w.Fork <- &ProcEventFork{ParentPid: pid}
			case syscall.NOTE_EXEC:
				w.Exec <- &ProcEventExec{Pid: pid}
			case syscall.NOTE_EXIT:
				w.RemoveWatch(pid)
				w.Exit <- &ProcEventExit{Pid: pid}
			}
		}
	}
}

// Close our kqueue file descriptor; deletes any remaining filters
func (listener *kqueueListener) close() error {
	return syscall.Close(listener.kq)
}
