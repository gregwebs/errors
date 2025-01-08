// Package errors provides error handling primitives that add stack traces and metadata to errors.
//
// Key Concepts:
//
//   - Adding Stack traces: All the error creation and wrapping functions ensure a stack trace is recorded for the error.
//   - Adding Context: The `errors.Wrap` and `errors.Wrapf` functions adds an additional string context to an error.
//   - Adding Structured data: The `errors.Wraps` and `errors.Slog` functions adds structured data to errors.
//   - Formatted Printing: Errors returned from this package implement the `fmt.Formatter` interface- verbose printing options will show the stack trace.
//   - Retrieving underlying errors: In addition to standard `errors.Unwrap`, `errors.Is`, and `errors.As`, there are `errors.AsType`, `errors.Cause`, and `errors.UnwrapGroups`.
//   - Retrieving the Stack Trace: `errors.GetStackTracer` retrieves the stack trace from the error.
//   - Retrieving the structured data: `errors.SlogRecord` retrieves structured data as an slog.Record.
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
package errors

import (
	"io"
	"log"

	"github.com/gregwebs/errors/errwrap"
	"github.com/gregwebs/errors/slogerr"
	"github.com/gregwebs/errors/stackfmt"
)

// GetStackTracer will return the first StackTracer in the causer chain.
// This function is used by AddStack to avoid creating redundant stack traces.
//
// You can also use the StackTracer interface on the returned error to get the stack trace.
func GetStackTracer(origErr error) stackfmt.StackTracer {
	var stacked stackfmt.StackTracer
	errwrap.WalkDeep(origErr, func(err error) bool {
		if stackTracer, ok := err.(stackfmt.StackTracer); ok {
			stacked = stackTracer
			return true
		}
		return false
	})
	return stacked
}

// IsNil performs additional checks besides == nil
// This helps deal with a design issue with Go interfaces: https://go.dev/doc/faq#nil_error
// It will return true if the error interface contains a nil pointer, interface, slice, array or map
// It will return true if the slice or array or map is empty
func IsNil(err error) bool {
	return isNil(err)
}

// HandleFmtWriteError handles (rare) errors when writing to fmt.State.
// It defaults to printing the errors.
func HandleFmtWriteError(handler func(err error)) {
	handleFmtWriteError = handler
	errwrap.HandleFmtWriteError(handler)
	stackfmt.HandleFmtWriteError(handler)
	slogerr.HandleFmtWriteError(handler)
}

var handleFmtWriteError = func(err error) {
	log.Println(err)
}

func writeString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		handleFmtWriteError(err)
	}
}
