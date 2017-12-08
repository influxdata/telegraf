// Package logutil provides functions that can print to a logger.
// Any function in this package that takes in a *log.Logger can be
// passed nil to use the log package's default logger.
package logutil

import "log"

// Print calls Print on a logger or the default logger.
// Arguments are handled in the manner of fmt.Print.
func Print(l *log.Logger, v ...interface{}) {
	if l == nil {
		log.Print(v...)
	} else {
		l.Print(v...)
	}
}

// Printf calls Printf on a logger or the default logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(l *log.Logger, format string, v ...interface{}) {
	if l == nil {
		log.Printf(format, v...)
	} else {
		l.Printf(format, v...)
	}
}

// Println calls Println on a logger or the default logger.
// Arguments are handled in the manner of fmt.Println.
func Println(l *log.Logger, v ...interface{}) {
	if l == nil {
		log.Println(v...)
	} else {
		l.Println(v...)
	}
}
