package errwrap

import (
	stderrors "errors"
	"fmt"
	"io"
	"log"
)

// Cause returns the underlying cause of the error, if possible.
// Unwrap goes just one level deep, but Cause goes all the way to the bottom
// If nil is given, it will return nil
func Cause(err error) error {
	cause := Unwrap(err)
	if cause == nil {
		return err
	}
	return Cause(cause)
}

// Unwrap uses the Unwrap method to return the next error in the chain or nil.
// This is the same as the standard errors.Unwrap
func Unwrap(err error) error {
	u, ok := err.(unwrapper)
	if !ok {
		return nil
	}
	return u.Unwrap()
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
	var target Err
	return target, stderrors.As(err, &target)
}

// The same as the standard errors.Join
func Join(errs ...error) error {
	return stderrors.Join(errs...)
}

// A generic form of Joins
func JoinsG[T error](errs ...T) []T {
	n := 0
	for _, err := range errs {
		if !isNil(err) {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	newErrs := make([]T, 0, n)
	for _, err := range errs {
		if !isNil(err) {
			newErrs = append(newErrs, err)
		}
	}
	return newErrs
}

// The same as errors.Join but returns the array rather than wrapping it.
// Also uses isNil for a better nil check.
func Joins(errs ...error) []error {
	return JoinsG(errs...)
}

// errorGroup is an interface for multiple errors that are not a chain.
// This happens for example when executing multiple operations in parallel.
// The standard Go API is now Unwrap() []error
type errorGroup interface {
	Errors() []error
}

// If the error is not nil and an errorGroup or satisfies Unwraps() []error, return its list of errors
// otherwise return nil
func Unwraps(err error) []error {
	if group, ok := err.(unwraps); ok {
		return group.Unwrap()
	}
	return nil
}

// Deprecated: use Unwraps
func Errors(err error) []error {
	if err == nil {
		return nil
	}
	if group, ok := err.(errorGroup); ok {
		return group.Errors()
	} else if group, ok := err.(unwraps); ok {
		return group.Unwrap()
	}
	return nil
}

// ErrorUnwrap allows wrapped errors to give just the message of the individual error without any unwrapping.
//
// The existing Error() convention extends that output to all errors that are wrapped.
// ErrorNoUnwrap() has just the wrapping message without additional unwrapped messages.
//
// Existing Error() definitions look like this:
//
//	func (hasWrapped) Error() string { return hasWrapped.message + ": " + hasWrapped.Unwrap().Error() }
//
// An ErrorNoUnwrap() definitions look like this:
//
//	func (hasWrapped) ErrorNoUnwrap() string { return hasWrapped.message }
type ErrorUnwrap interface {
	error
	Unwrap() error
	// ErrorNoUnwrap is the error message component of the wrapping
	// It will be a prefix of Error()
	// If there is no message in the wrapping then this can return an empty string
	ErrorNoUnwrap() string
}

// The ErrorWrapper method allows for modifying the inner error while maintaining the same outer type.
// This is useful for wrapping types that implement an interface that extend errors.
type ErrorWrapper interface {
	error
	WrapError(func(error) error)
}

// Uses ErrorWrapper to wrap in place, if ErrorWrapper is available.
// Returns true if wrapped in place.
// Returns false if not wrapped in place, including if the given error is nil.
func WrapInPlace(err error, wrap func(error) error) bool {
	if isNil(err) {
		return false
	}
	if inPlace, ok := AsType[ErrorWrapper](err); ok {
		inPlace.WrapError(wrap)
		return true
	}
	return false
}

// ErrorWrap should be included as a pointer.
// If fulfills the ErrorWrapper interface.
// This allows for wrapping an inner error without changing the outer type.
type ErrorWrap struct {
	error
}

// NewErrorWrap returns a pointer because ErrorWrap should be used as a pointer.
func NewErrorWrap(err error) *ErrorWrap {
	return &ErrorWrap{err}
}

// This struct is designed to be used as an embeded error.
func (ew *ErrorWrap) Unwrap() error {
	return ew.error
}

func (ew *ErrorWrap) WrapError(wrap func(error) error) {
	ew.error = wrap(ew.error)
}

func (ew *ErrorWrap) Format(s fmt.State, verb rune) {
	forwardFormatting(ew.error, s, verb)
}

func formatErrorUnwrapStack(err ErrorUnwrap, s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			formatterPlusV(s, verb, err.Unwrap())
			if msg := err.ErrorNoUnwrap(); msg != "" {
				writeString(s, "\n"+msg)
			}

			if stackTracer, ok := err.(stackTraceFormatter); ok {
				stackTracer.FormatStackTrace(s, verb)
			}
			return
		}
		fallthrough
	case 's':
		writeString(s, err.Error())
	case 'q':
		fmt.Fprintf(s, "%q", err.Error())
	}
}

// Forward to a Formatter if it exists
func forwardFormatting(err error, s fmt.State, verb rune) {
	if formatter, ok := err.(fmt.Formatter); ok {
		formatter.Format(s, verb)
	} else if errUnwrap, ok := err.(ErrorUnwrap); ok {
		formatErrorUnwrapStack(errUnwrap, s, verb)
	} else {
		fmtString := fmt.FormatString(s, verb)
		// unwrap before calling forwamrdFormatting to avoid infinite recursion
		fmt.Fprintf(s, fmtString, err)
	}
}

var _ ErrorWrapper = (*ErrorWrap)(nil) // assert implements interface

type stackTraceFormatter interface {
	FormatStackTrace(s fmt.State, verb rune)
}

// HandleFmtWriteError handles (rare) errors when writing to fmt.State.
// It defaults to printing the errors.
func HandleFmtWriteError(handler func(err error)) {
	handleFmtWriteError = handler
}

var handleFmtWriteError = func(err error) {
	log.Println(err)
}

func writeString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		handleFmtWriteError(err)
	}
}

// Deprecated: WalkDeep was created before iterators.
// UnwrapGroups is now preferred for those using Go version >= 1.23.
// Note that WalkDeep uses the opposite convention for boolean return values compared to golang iterators.
// WalkDeep does a depth-first traversal of all errors.
// An error that defines Unwrap() []error is expanded and traversed
// The visitor function can return true to end the traversal early
// If iteration is ended early, WalkDeep will return true, otherwise false.
func WalkDeep(err error, visitor func(err error) bool) bool {
	if err == nil {
		return false
	}
	if done := visitor(err); done {
		return true
	}
	if done := WalkDeep(Unwrap(err), visitor); done {
		return true
	}

	// Go wide
	if errors := Errors(err); len(errors) > 0 {
		for _, err := range errors {
			if early := WalkDeep(err, visitor); early {
				return true
			}
		}
	}

	return false
}

// Deprecated: WalkDeepLevel was created before iterators.
// UnwrapGroupsLevel is now preferred for those using Go version >= 1.23.
// This operates the same as [WalkDeep] but adds a second parameter to the visitor: the level of depth in the error tree.
func WalkDeepLevel(err error, visitor func(error, int) bool) bool {
	return walkDeepLevel(err, visitor, 0)
}

func walkDeepLevel(err error, visitor func(error, int) bool, stack int) bool {
	if err == nil {
		return false
	}
	if done := visitor(err, stack); done {
		return true
	}
	if done := walkDeepLevel(Unwrap(err), visitor, stack+1); done {
		return true
	}

	// Go wide
	if errors := Errors(err); len(errors) > 0 {
		for _, err := range errors {
			if early := walkDeepLevel(err, visitor, stack+1); early {
				return true
			}
		}
	}

	return false
}
