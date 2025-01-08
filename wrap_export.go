package errors

import (
	stderrors "errors"

	"github.com/gregwebs/errors/errwrap"
)

// Unwrap uses the Unwrap method to return the next error in the chain or nil.
// This is the same as the standard errors.Unwrap
func Unwrap(err error) error {
	return errwrap.Unwrap(err)
}

// Cause returns the underlying cause of the error, if possible.
// Unwrap goes just one level deep, but Cause goes all the way to the bottom
// If nil is given, it will return nil
func Cause(err error) error {
	return errwrap.Cause(err)
}

// A re-export of the standard library errors.Is
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// A re-export of the standard library errors.As
func As(err error, target any) bool {
	return stderrors.As(err, target)
}

// AsType is equivalient to As and returns the same boolean.
// Instead of instantiating a struct and passing it by pointer,
// the type of the error is given as the generic argument
// It is instantiated and returned.
func AsType[Err error](err error) (Err, bool) {
	return errwrap.AsType[Err](err)
}

func Unwraps(err error) []error {
	return errwrap.Unwraps(err)
}

// The same as the standard errors.Join
func Join(errs ...error) error {
	return stderrors.Join(errs...)
}

// The same as errors.Join but returns the array rather than wrapping it.
// Also uses isNil for a better nil check.
func Joins(errs ...error) []error {
	return errwrap.Joins(errs...)
}
