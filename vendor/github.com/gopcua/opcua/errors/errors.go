package errors

import (
	pkg_errors "github.com/pkg/errors"
)

// Prefix is the default error string prefix
const Prefix = "opcua: "

// Errorf is a wrapper for `errors.Errorf`
func Errorf(format string, a ...interface{}) error {
	return pkg_errors.Errorf(Prefix+format, a...)
}

// New is a wrapper for `errors.New`
func New(text string) error {
	return pkg_errors.New(Prefix + text)
}

// Equal returns true if the two errors have the same error message.
//
// todo(fs): the reason we need this function and cannot just use
// todo(fs): reflect.DeepEqual(err1, err2) is that by using github.com/pkg/errors
// todo(fs): the underlying stack traces change and because of this the errors
// todo(fs): are no longer comparable. This is a downside of basing our errors
// todo(fs): errors implementation on github.com/pkg/errors and we may want to
// todo(fs): revisit this.
// todo(fs): See https://play.golang.org/p/1WqB7u4BUf7 (by @kung-foo)
func Equal(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 != nil && err2 != nil {
		return err1.Error() == err2.Error()
	}
	return false
}
