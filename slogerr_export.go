package errors

import (
	stderrors "errors"
	"log/slog"

	"github.com/gregwebs/errors/slogerr"
)

// SlogRecord traverses the error chain, calling Unwrap(), to look for slog Records
// All records will be merged and mesage strings are joined together
// The message string may contain some of the structured information
// This depends on defining ErrorNoUnwrap from the interface ErrorNotUnwrapped
//
//	if record := errors.SlogRecord(errIn); record != nil {
//		record.Add(logArgs...)
//		if err := slog.Default().Handler().Handle(ctx, *record); err != nil {
//			slog.ErrorContext(ctx, fmt.Sprintf("%+v", err))
//		}
//	}
func SlogRecord(inputErr error) *slog.Record {
	return slogerr.SlogRecord(inputErr)
}

// Slog creates an error that instead of generating a format string generates a structured slog Record.
// Accepts as args any valid slog args.
// Also accepts `[]slog.Attr` as a single argument to avoid having to cast that argument.
// The slog Record can be retrieved with SlogRecord.
// Structured errors are more often created by wrapping existing errors with Wraps.
func Slog(msg string, args ...interface{}) slogerr.StructuredError {
	return slogerr.WrapsSkip(stderrors.New(""), msg, 1, args...)
}

// Wraps ends with an "s" to indicate it is Structured.
// Accepts as args any valid slog args. These will generate an slog Record
// Also accepts []slog.Attr as a single argument to avoid having to cast that argument.
func Wraps(err error, msg string, args ...interface{}) slogerr.StructuredError {
	return slogerr.WrapsSkip(err, msg, 1, args...)
}
