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

// GetStackTracer will return the first StackTracer in the causer chain.
// This function is used by AddStack to avoid creating redundant stack traces.
//
// You can also use the StackTracer interface on the returned error to get the stack trace.
func GetStackTracer(origErr error) StackTracer {
	var stacked StackTracer
	WalkDeep(origErr, func(err error) bool {
		if stackTracer, ok := err.(StackTracer); ok {
			stacked = stackTracer
			return true
		}
		return false
	})
	return stacked
}
