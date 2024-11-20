// Package errors provides simple error handling primitives.
//
// The traditional error handling idiom in Go is roughly akin to
//
//	if err != nil {
//	        return err
//	}
//
// which applied recursively up the call stack results in error reports
// without context or debugging information. The errors package allows
// programmers to add context to the failure path in their code in a way
// that does not destroy the original value of the error.
//
// # Adding context to an error
//
// The errors.Wrap function returns a new error that adds context to the
// original error by recording a stack trace at the point Wrap is called,
// and the supplied message. For example
//
//	_, err := ioutil.ReadAll(r)
//	if err != nil {
//	        return errors.Wrap(err, "read failed")
//	}
//
// If additional control is required the errors.AddStack and errors.WithMessage
// functions destructure errors.Wrap into its component operations of annotating
// an error with a stack trace and an a message, respectively.
//
// # Retrieving the cause of an error
//
// Using errors.Wrap constructs a stack of errors, adding context to the
// preceding error. Depending on the nature of the error it may be necessary
// to reverse the operation of errors.Wrap to retrieve the original error
// for inspection. Any error value which implements the Unwrap interface
// can be inspected one level deeper by the errors.Unwrap function. errors.Cause will recursively unwrap
// the error. For example:
//
//	switch err := errors.Cause(err).(type) {
//	case *MyError:
//	        // handle specifically
//	default:
//	        // unknown error
//	}
//
// # Formatted printing of errors
//
// All error values returned from this package implement fmt.Formatter and can
// be formatted by the fmt package. The following verbs are supported
//
//	%s    print the error. If the error has a Cause it will be
//	      printed recursively with colon separations
//	%v    see %s
//	%+v   extended format. Each Frame of the error's StackTrace will
//	      be printed in detail.
//	%-v   similar to %s but newline separated. No stack traces included.
//
// # Retrieving the stack trace of an error or wrapper
//
// New, Errorf, Wrap, and Wrapf record a stack trace at the point they are invoked.
// This information can be retrieved with the StackTracer interface that returns
// a StackTrace. Where errors.StackTrace is defined as
//
//	type StackTrace []Frame
//
// The Frame type represents a call site in the stack trace. Frame supports
// the fmt.Formatter interface that can be used for printing information about
// the stack trace of this error. For example:
//
//	if stacked := errors.GetStackTracer(err); stacked != nil {
//	        for _, f := range stacked.StackTrace() {
//	                fmt.Printf("%+s:%d", f)
//	        }
//	}
//
// See the documentation for Frame.Format for more details.
//
// errors.Find can be used to search for an error in the error chain.
package errors

import (
	stderrors "errors"
	"fmt"
	"io"
	"log"
)

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return &fundamental{stderrors.New(message), callers()}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	if _, ok := err.(unwrapper); ok {
		return &addStack{withStack{err, callers()}}
	} else if _, ok := err.(unwraps); ok {
		return &addStack{withStack{err, callers()}}
	}
	return &fundamental{err, callers()}
}

// fundamental is a base error that doesn't wrap other errors
// It stores an error rather than just a string. This allows for:
// * reuse of existing patterns
// * usage of Errorf to support any formatting
// The latter is done in part to support %w, but note that if %w is used we don't use fundamental
type fundamental struct {
	error
	*stack
}

func (f *fundamental) StackTrace() StackTrace { return f.stack.StackTrace() }
func (f *fundamental) HasStack() bool         { return true }
func (f *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			writeString(s, f.Error())
			f.StackTrace().Format(s, verb)
			return
		}
		fallthrough
	case 's':
		writeString(s, f.Error())
	case 'q':
		fmt.Fprintf(s, "%q", f.Error())
	}
}

// StackTraceAware is an optimization to avoid repetitive traversals of an error chain.
// HasStack checks for this marker first.
// Wrap and Wrapf will produce this marker.
type StackTraceAware interface {
	HasStack() bool
}

// HasStack tells whether a StackTracer exists in the error chain
func HasStack(err error) bool {
	if errWithStack, ok := err.(StackTraceAware); ok {
		return errWithStack.HasStack()
	}
	return GetStackTracer(err) != nil
}

// AddStack annotates err with a stack trace at the point WithStack was called.
// It will first check with HasStack to see if a stack trace already exists before creating another one.
func AddStack(err error) error {
	if err == nil {
		return nil
	}
	if HasStack(err) {
		return err
	}
	return &addStack{withStack{err, callers()}}
}

// Same as AddStack but specify an additional number of callers to skip
func AddStackSkip(err error, skip int) error {
	if err == nil {
		return nil
	}
	if HasStack(err) {
		return err
	}
	return &addStack{withStack{err, callersSkip(skip + 3)}}
}

type withStack struct {
	error
	*stack
}

func (w *withStack) StackTrace() StackTrace { return w.stack.StackTrace() }
func (w *withStack) Unwrap() error          { return w.error }
func (w *withStack) ErrorNoUnwrap() string  { return "" }
func (w *withStack) HasStack() bool         { return true }
func (w *withStack) Format(s fmt.State, verb rune) {
	formatError(w, s, verb)
}

func formatError(err ErrorUnwrap, s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			formatterPlusV(s, verb, err.Unwrap())
			if msg := err.ErrorNoUnwrap(); msg != "" {
				writeString(s, "\n"+msg)
			}
			if stackTracer, ok := err.(StackTracer); ok {
				stackTracer.StackTrace().Format(s, verb)
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

// addStack is returned directly whereas withStack is always used composed
// they Unwrap differently
type addStack struct {
	withStack
}

func (a *addStack) Unwrap() error { return a.error }
func (a *addStack) Format(s fmt.State, verb rune) {
	formatError(a, s, verb)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		msg:       message,
		withStack: withStack{err, callers()},
	}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Annotatef is call, and the format specifier.
// If err is nil, Annotatef returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		msg:       fmt.Sprintf(format, args...),
		withStack: withStack{err, callers()},
	}
}

// WrapoNoStack annotates err with a new message.
// If err is nil, returns nil.
// WrapNoStack does not add a new stack trace.
// When used consecutively, it will append the message strings rather than creating a new error
func WrapNoStack(err error, message string) error {
	if err == nil {
		return nil
	}
	if ns, ok := err.(*withMessageNoStack); ok {
		ns.msg = message + ": " + ns.msg
		return ns
	}
	return &withMessageNoStack{
		msg:   message,
		error: err,
	}
}

type withMessage struct {
	msg string
	withStack
}

func (w *withMessage) Error() string         { return w.msg + ": " + w.error.Error() }
func (w *withMessage) ErrorNoUnwrap() string { return w.msg }
func (w *withMessage) Format(s fmt.State, verb rune) {
	formatError(w, s, verb)
}

type withMessageNoStack struct {
	msg string
	error
}

func (w *withMessageNoStack) Error() string         { return w.msg + ": " + w.error.Error() }
func (w *withMessageNoStack) Unwrap() error         { return w.error }
func (w *withMessageNoStack) ErrorNoUnwrap() string { return w.msg }
func (w *withMessageNoStack) Format(s fmt.State, verb rune) {
	formatError(w, s, verb)
}

func formatterPlusV(s fmt.State, verb rune, err error) {
	if f, ok := err.(fmt.Formatter); ok {
		f.Format(s, verb)
	} else {
		fmt.Fprintf(s, "%+v", err)
	}
}

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

type unwrapper interface {
	Unwrap() error
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

// HandleWriteError handles errors when writing to fmt.State
var HandleWriteError = func(err error) {
	log.Println(err)
}

func writeString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		HandleWriteError(err)
	}
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
	if err == nil {
		return false
	}
	if inPlace, ok := AsType[ErrorWrapper](err); ok {
		inPlace.WrapError(wrap)
		return true
	}
	return false
}

// ErrorWrap should be included as a pointer.
// If fulfills the WrapError interface.
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

func (ew *ErrorWrap) HasStack() bool {
	return HasStack(ew.error)
}

func (ew *ErrorWrap) Format(s fmt.State, verb rune) {
	forwardFormatting(ew.error, s, verb)
}

// Forward to a Formatter if it exists
func forwardFormatting(err error, s fmt.State, verb rune) {
	if formatter, ok := err.(fmt.Formatter); ok {
		formatter.Format(s, verb)
	} else if errUnwrap, ok := err.(ErrorUnwrap); ok {
		formatError(errUnwrap, s, verb)
	} else {
		fmtString := fmt.FormatString(s, verb)
		// unwrap before calling forwamrdFormatting to avoid infinite recursion
		fmt.Fprintf(s, fmtString, err)
	}
}

var _ ErrorWrapper = (*ErrorWrap)(nil) // assert implements interface

// WrapFn returns a wrapping function that calls Wrap
func WrapFn(msg string) func(error) error {
	return func(err error) error { return Wrap(err, msg) }
}

// WrapFn returns a wrapping function that calls Wrapf
func WrapfFn(msg string, args ...interface{}) func(error) error {
	return func(err error) error { return Wrapf(err, msg, args...) }
}

// WrapFn returns a wrapping function that calls Wraps
func WrapsFn(msg string, args ...interface{}) func(error) error {
	return func(err error) error { return Wraps(err, msg, args...) }
}
