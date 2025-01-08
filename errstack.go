package errors

import (
	stderrors "errors"
	"fmt"
	"io"
	"log"

	"github.com/gregwebs/errors/stackfmt"
)

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return &fundamental{stderrors.New(message), stackfmt.NewStack()}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	if _, ok := err.(unwrapper); ok {
		return &addStack{withStack{err, stackfmt.NewStack()}}
	} else if _, ok := err.(unwraps); ok {
		return &addStack{withStack{err, stackfmt.NewStack()}}
	}
	return &fundamental{err, stackfmt.NewStack()}
}

// fundamental is a base error that doesn't wrap other errors
// It stores an error rather than just a string. This allows for:
// * reuse of existing patterns
// * usage of Errorf to support any formatting
// The latter is done in part to support %w, but note that if %w is used we don't use fundamental
type fundamental struct {
	error
	stackfmt.Stack
}

func (f *fundamental) StackTrace() stackfmt.StackTrace { return f.Stack.StackTrace() }
func (f *fundamental) HasStack() bool                  { return true }
func (f *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			writeStringErrstack(s, f.Error())
			f.StackTrace().Format(s, verb)
			return
		}
		fallthrough
	case 's':
		writeStringErrstack(s, f.Error())
	case 'q':
		fmt.Fprintf(s, "%q", f.Error())
	}
}

// AddStack annotates err with a stack trace at the point WithStack was called.
// It will first check with HasStack to see if a stack trace already exists before creating another one.
func AddStack(err error) error {
	if isNil(err) {
		return nil
	}
	if HasStack(err) {
		return err
	}
	return &addStack{withStack{err, stackfmt.NewStack()}}
}

// Same as AddStack but specify an additional number of callers to skip
func AddStackSkip(err error, skip int) error {
	if isNil(err) {
		return nil
	}
	if HasStack(err) {
		return err
	}
	return &addStack{withStack{err, stackfmt.NewStackSkip(skip + 1)}}
}

type withStack struct {
	error
	stackfmt.Stack
}

func (w *withStack) StackTraceFormat(s fmt.State, v rune) { w.Stack.FormatStackTrace(s, v) }
func (w *withStack) StackTrace() stackfmt.StackTrace      { return w.Stack.StackTrace() }
func (w *withStack) Unwrap() error                        { return w.error }
func (w *withStack) ErrorNoUnwrap() string                { return "" }
func (w *withStack) HasStack() bool                       { return true }
func (w *withStack) Format(s fmt.State, verb rune) {
	formatErrorUnwrap(w, s, verb)
}

var _ stackfmt.StackTracer = &withStack{}
var _ stackfmt.StackTraceFormatter = &withStack{}

// addStack is returned directly whereas withStack is always used composed
// they Unwrap differently
type addStack struct {
	withStack
}

func (a *addStack) Unwrap() error { return a.error }
func (a *addStack) Format(s fmt.State, verb rune) {
	formatErrorUnwrap(a, s, verb)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if isNil(err) {
		return nil
	}
	if wm, ok := err.(*withMessage); ok {
		wm.msg = message + ": " + wm.msg
		return wm
	}
	return &withMessage{
		msg:       message,
		withStack: withStack{err, stackfmt.NewStack()},
	}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is call, and the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if isNil(err) {
		return nil
	}
	if wm, ok := err.(*withMessage); ok {
		wm.msg = fmt.Sprintf(format, args...) + ": " + wm.msg
		return wm
	}
	return &withMessage{
		msg:       fmt.Sprintf(format, args...),
		withStack: withStack{err, stackfmt.NewStack()},
	}
}

// WrapNoStack does not add a new stack trace.
// WrapNoStack annotates err with a new message.
// If err is nil, returns nil.
// When used consecutively, it will append the message strings rather than creating a new error
func WrapNoStack(err error, message string) error {
	if isNil(err) {
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
	formatErrorUnwrap(w, s, verb)
}

type withMessageNoStack struct {
	msg string
	error
}

func (w *withMessageNoStack) Error() string         { return w.msg + ": " + w.error.Error() }
func (w *withMessageNoStack) Unwrap() error         { return w.error }
func (w *withMessageNoStack) ErrorNoUnwrap() string { return w.msg }
func (w *withMessageNoStack) HasStack() bool        { return HasStack(w.error) }
func (w *withMessageNoStack) Format(s fmt.State, verb rune) {
	formatErrorUnwrap(w, s, verb)
}

// WrapFn returns a wrapping function that calls Wrap
func WrapFn(msg string) func(error) error {
	return func(err error) error { return Wrap(err, msg) }
}

// WrapfFn returns a wrapping function that calls Wrapf
func WrapfFn(msg string, args ...interface{}) func(error) error {
	return func(err error) error { return Wrapf(err, msg, args...) }
}

// StackTraceAware can be used to avoid repetitive traversals of an error chain.
// HasStack checks for this marker first.
type StackTraceAware interface {
	HasStack() bool
}

// HasStack returns true if the error will find a stack trace
// It does not unwrap errors
func HasStack(err error) bool {
	if errWithStack, ok := err.(StackTraceAware); ok {
		return errWithStack.HasStack()
	}
	if _, ok := err.(stackfmt.StackTracer); ok {
		return true
	}
	if _, ok := err.(stackfmt.StackTraceFormatter); ok {
		return true
	}
	return false
}

func formatErrorUnwrap(err error, s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if uwErr := err.(errorUnwrap); uwErr != nil {
				formatterPlusV(s, verb, uwErr.Unwrap())
				if msg := uwErr.ErrorNoUnwrap(); msg != "" {
					writeStringErrstack(s, "\n"+msg)
				}
			} else {
				writeStringErrstack(s, err.Error())
			}
			if stackTracer, ok := err.(stackfmt.StackTracer); ok {
				stackTracer.StackTrace().Format(s, verb)
			} else if stackTracer, ok := err.(stackfmt.StackTraceFormatter); ok {
				stackTracer.FormatStackTrace(s, verb)
			}
			return
		}
		fallthrough
	case 's':
		writeStringErrstack(s, err.Error())
	case 'q':
		fmt.Fprintf(s, "%q", err.Error())
	}
}

// HandleWriteErrorErrstack handles (rare) errors when writing to fmt.State.
// It defaults to printing the errors.
func HandleWriteErrorErrstack(handler func(err error)) {
	handleWriteErrorErrstack = handler
}

var handleWriteErrorErrstack = func(err error) {
	log.Println(err)
}

func writeStringErrstack(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		handleWriteErrorErrstack(err)
	}
}
