package errors

import (
	"bytes"
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// HasSlogRecord is used to retrieve an slog.Record.
// A StructuredError defines it.
// Alternatively SlogMessager and LogValuer can be defined.
type HasSlogRecord interface {
	GetSlogRecord() slog.Record
}

// StructuredError is returned by Wraps and Slog
type StructuredError interface {
	error
	HasSlogRecord
}

// SlogMessager provides a string that can be used as the slog message.
// Normally LogValuer is defined as well.
// When defining error types this may be simpler to use than defining HasSlogRecord.
type SlogMessager interface {
	SlogMsg() string
}

type structuredErr struct {
	Record slog.Record
	err    error
	msg    string
	stack  Stack
}

func (se structuredErr) GetSlogRecord() slog.Record {
	return se.Record
}

func (se structuredErr) Error() string {
	return joinZero(": ", se.ErrorNoUnwrap(), se.err.Error())
}

func (se structuredErr) Unwrap() error   { return se.err }
func (se structuredErr) SlogMsg() string { return se.msg }
func (se structuredErr) HasStack() bool  { return true }
func (se structuredErr) ErrorNoUnwrap() string {
	msg := se.msg
	stext := structureAsText(se.Record)
	if stext != "" {
		msg = joinZero(" ", msg, stext)
	}
	return msg
}

func (se structuredErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			formatterPlusV(s, verb, se.err)
			if !hasStack(se.err) {
				var st StackTracer
				if stderrors.As(se.err, &st) {
					st.StackTrace().Format(s, verb)
				} else {
					se.stack.Format(s, verb)
				}
			}
			writeStringSlogerr(s, "\n"+joinZero(" ", se.msg, structureAsText(se.Record)))
			return
		}
		fallthrough
	case 's', 'q':
		writeStringSlogerr(s, se.Error())
	}
}

func (se structuredErr) FormatStackTrace(s fmt.State, verb rune) {
	se.stack.FormatStackTrace(s, verb)
}

// Slog creates an error that instead of generating a format string generates a structured slog Record.
// Accepts as args any valid slog args.
// Also accepts `[]slog.Attr` as a single argument to avoid having to cast that argument.
// The slog Record can be retrieved with SlogRecord.
// Structured errors are more often created by wrapping existing errors with Wraps.
func Slog(msg string, args ...interface{}) StructuredError {
	return wrapsSkip(stderrors.New(""), msg, 1, args...)
}

func wrapsSkip(err error, msg string, skip int, args ...interface{}) StructuredError {
	if err == nil {
		return nil
	}
	var pc uintptr
	stack := NewStackSkip(skip + 3)
	if hr, ok := err.(HasSlogRecord); ok {
		record := hr.GetSlogRecord()
		pc = record.PC
	} else {
		pc = stack[0]
	}
	record := slog.NewRecord(time.Now(), slog.LevelError, msg, pc)
	// support passing an array of Attr
	// otherwise would require a cast to any
	for i := 0; i < len(args) && len(args) > 0; {
		switch attrs := any(args[0]).(type) {
		case []slog.Attr:
			record.AddAttrs(attrs...)
			args = args[1:]
		default:
			i += 2 // k/v pairs
		}
	}
	if len(args) > 0 {
		record.Add(args...)
	}

	return structuredErr{
		Record: record,
		err:    err,
		msg:    msg,
		stack:  stack,
	}
}

// Wraps ends with an "s" to indicate it is Structured.
// Accepts as args any valid slog args. These will generate an slog Record
// Also accepts []slog.Attr as a single argument to avoid having to cast that argument.
func Wraps(err error, msg string, args ...interface{}) StructuredError {
	return wrapsSkip(err, msg, 1, args...)
}

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
	msgs := []string{}
	var record *slog.Record
	msgDone := false
	var msgUnrecognized string
	walkUnwrapLevel(inputErr, func(err error, stack int) bool {
		if hr, ok := err.(HasSlogRecord); ok {
			newRecord := hr.GetSlogRecord()
			if record == nil {
				cloned := newRecord.Clone()
				record = &cloned
			} else {
				newRecord.Attrs(func(attr slog.Attr) bool {
					record.AddAttrs(attr)
					return true
				})
				if record.PC == 0 && newRecord.PC != 0 {
					record.PC = newRecord.PC
				}
			}
		} else if logValuer, ok := err.(valuerError); ok {
			if record == nil {
				record = toSlogRecord(logValuer, 5+stack)
			} else {
				record.AddAttrs(attrsFromValue(logValuer.LogValue())...)
			}
		}

		// Gather messages until we reach a message that does not understand ErrorNotUnwrapped, SlogMessager, or HasSlogRecord
		if !msgDone {
			newErrMsg := err.Error()

			// We are recursing down and the error is still the same
			// This was just a simple wrapper
			if msgUnrecognized != "" && strings.HasPrefix(newErrMsg, msgUnrecognized) {
				msgUnrecognized = ""
			}

			if slogMsg := slogMsg(err); slogMsg != nil {
				if *slogMsg != "" {
					if msgUnrecognized == "" || strings.HasPrefix(*slogMsg, msgUnrecognized) {
						msgs = append(msgs, *slogMsg)
						msgUnrecognized = ""
					}
				}
			} else if noUnwrap, ok := err.(errorUnwrap); ok {
				if msg := noUnwrap.ErrorNoUnwrap(); msg != "" {
					if msgUnrecognized == "" || strings.HasPrefix(msgUnrecognized, msg) {
						msgs = append(msgs, msg)
						msgUnrecognized = ""
					}
				}
			} else if msgUnrecognized == "" {
				if len(msgs) == 0 || newErrMsg != msgs[len(msgs)-1] {
					msgUnrecognized = newErrMsg
				}
			} else {
				msgs = append(msgs, msgUnrecognized)
				msgDone = true
			}
		}

		return false // keep going
	}, 0)

	if record != nil {
		if !msgDone && msgUnrecognized != "" {
			msgs = append(msgs, msgUnrecognized)
		}
		record.Message = strings.Join(msgs, ": ")
	}
	return record
}

// SlogTextBuffer produces a Handler that writes to a buffer
// The buffer output can be retrieved with the returned function
//
//	if record := errors.SlogRecord(err); record != nil {
//		handler, output := errors.SlogTextBuffer(slog.HandlerOptions{AddSource: false})
//		if err := handler.Handle(ctx, *record); err != nil {
//			fmt.Println(fmt.Sprintf("%+v", err))
//		} else {
//			fmt.Println(output())
//		}
//	}
func SlogTextBuffer(opts *slog.HandlerOptions) (slog.Handler, func() string) {
	buf := bytes.NewBuffer([]byte{})
	h := slog.NewTextHandler(buf, opts)
	return h, func() string { return buf.String() }
}

func slogMsgOrError(err error) string {
	msg := slogMsg(err)
	if msg != nil {
		return *msg
	} else {
		return err.Error()
	}
}

// slogMsg returns a message from a SlogMessager or a HasSlogRecord.
// Otherwise it returns nil
func slogMsg(err any) *string {
	if slogm, ok := err.(SlogMessager); ok {
		s := slogm.SlogMsg()
		return &s
	} else if hr, ok := err.(HasSlogRecord); ok {
		s := hr.GetSlogRecord().Message
		return &s
	} else {
		return nil
	}
}

/*
// slogAttributes returns a message from a LogValuer or a HasSlogRecord.
// Otherwise it returns a nil array
func slogAttributes(err any) []slog.Attr {
	if slogv, ok := err.(LogValuer); ok {
		return attrsFromValue(slogv.Slogv.Value())
	} else if hr, ok := err.(HasSlogRecord); ok {
		record := hr.GetSlogRecord()
		attrs := make([]slog.Attr, 0, record.NumAttrs())
		record.Attrs(func(attr slog.Attr) bool {
			attrs = append(attrs, attr)
			return true
		})
		return attrs
	} else {
		return nil
	}
}
*/

type valuerError interface {
	slog.LogValuer
	error
}

func attrsFromValue(value slog.Value) []slog.Attr {
	if value.Kind() == slog.KindGroup {
		return value.Group()
	} else {
		return []slog.Attr{slog.Any("error", value.Any())}
	}
}

func toSlogRecord(err valuerError, skip int) *slog.Record {
	if skip <= 0 {
		skip = 2
	}
	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])
	pc = pcs[0]
	msg := slogMsgOrError(err)
	record := slog.NewRecord(time.Now(), slog.LevelError, msg, pc)
	record.AddAttrs(attrsFromValue(err.LogValue())...)
	return &record
}

// Checks to see if an argument is the empty string.
// In that case just return the non-zero argument.
// Otherwise join the strings with the deliminator
func joinZero(delim string, str1 string, str2 string) string {
	if str1 == "" {
		return str2
	}
	if str2 == "" {
		return str1
	}
	return str1 + delim + str2
}

// TODO: should make own slog handler instead of re-using text
// This would avoid using ReplaceAttr and removing a newline
func structureAsText(record slog.Record) string {
	buf := new(bytes.Buffer)
	hOpts := slog.HandlerOptions{
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey || a.Key == slog.TimeKey || a.Key == slog.LevelKey {
				return slog.Attr{}
			}
			return a
		},
	}
	if err := slog.NewTextHandler(buf, &hOpts).Handle(context.Background(), record); err != nil {
		panic(err)
	}
	str := buf.String()
	return str[:len(str)-1]
}

func textFromRecord(err valuerError) string {
	record := toSlogRecord(err, 0)
	if record == nil {
		return ""
	}
	return joinZero(" ", record.Message, structureAsText(*record))
}

// WrapFn returns a wrapping function that calls Wraps
func WrapsFn(msg string, args ...interface{}) func(error) error {
	return func(err error) error { return Wraps(err, msg, args...) }
}

func walkUnwrapLevel(err error, visitor func(error, int) bool, stack int) bool {
	if err == nil {
		return false
	}
	if done := visitor(err, stack); done {
		return true
	}
	if unwrapped, ok := err.(unwrapper); ok {
		if done := walkUnwrapLevel(unwrapped.Unwrap(), visitor, stack+1); done {
			return true
		}
	}

	// Go wide
	if group, ok := err.(unwraps); ok {
		for _, err := range group.Unwrap() {
			if early := walkUnwrapLevel(err, visitor, stack+1); early {
				return true
			}
		}
	}

	return false
}

func hasStack(err error) bool {
	if errWithStack, ok := err.(stackTraceAware); ok {
		return errWithStack.HasStack()
	}
	if _, ok := err.(StackTracer); ok {
		return true
	}
	if _, ok := err.(StackTraceFormatter); ok {
		return true
	}
	return false
}

// HandleWriteErrorSlogerr handles (rare) errors when writing to fmt.State.
// It defaults to printing the errors.
func HandleWriteErrorSlogerr(handler func(err error)) {
	handleWriteErrorSlogerr = handler
}

var handleWriteErrorSlogerr = func(err error) {
	log.Println(err)
}

func writeStringSlogerr(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		handleWriteErrorSlogerr(err)
	}
}
