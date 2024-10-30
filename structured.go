package errors

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// HasSlogRecord is used to retrieve an slog.Record.
// A StructuredError defines it.
// Alternatively SlogMessager and SlogAttributer can be defined.
type HasSlogRecord interface {
	GetSlogRecord() slog.Record
}

// StructuredError is returned by Wraps and Slog
type StructuredError interface {
	error
	HasSlogRecord
}

// SlogMessager provides a string that can be used as the slog message.
// Normally SlogAttributer is defined as well.
// When defining error types this may be simpler to use than defining HasSlogRecord.
type SlogMessager interface {
	SlogMsg() string
}

// SlogAttributer provides a string that can be used as the slog message.
// Normally SlogMessager is defined as well.
// When defining error types this may be simpler to use than defining HasSlogRecord.
type SlogAttributer interface {
	SlogAttrs() []slog.Attr
}

type structuredErr struct {
	Record slog.Record
	err    error
	msg    string
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
			formatterPlusV(s, verb, se.Unwrap())
			writeString(s, "\n"+joinZero(" ", se.msg, structureAsText(se.Record)))
			return
		}
		fallthrough
	case 's', 'q':
		writeString(s, se.Error())
	}
}

// Slog creates an error that instead of generating a format string generates a structured slog Record.
// Accepts as args any valid slog args.
// Also accepts []slog.Attr as a single argument to avoid having to cast that argument.
// The slog Record can be retrieved with SlogRecord.
// Structured errors are more often created by wrapping existing errors with Wraps.
func Slog(msg string, args ...interface{}) StructuredError {
	return Wraps(New(""), msg, args...)
}

// Wraps ends with an "s" to indicate it is Structured.
// Accepts as args any valid slog args. These will generate an slog Record
// Also accepts []slog.Attr as a single argument to avoid having to cast that argument.
func Wraps(err error, msg string, args ...interface{}) StructuredError {
	if err == nil {
		return nil
	}
	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	pc = pcs[0]

	record := slog.NewRecord(time.Now(), slog.LevelError, msg, pc)
	// support passing an array of Attr: otherwise would require a cast
	for loop := true; loop && len(args) > 0; {
		switch attrs := any(args[0]).(type) {
		case []slog.Attr:
			record.AddAttrs(attrs...)
			args = args[1:]
		default:
			loop = false
		}
	}
	if len(args) > 0 {
		record.Add(args...)
	}

	// TODO: use the exact same stack for the error and the record
	return structuredErr{
		Record: record,
		err:    AddStackSkip(err, 1),
		msg:    msg,
	}
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
	walkDeepStack(inputErr, func(err error, stack int) bool {
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
		} else if sloga, ok := err.(attributeError); ok {
			if record == nil {
				record = toSlogRecord(sloga, 5+stack)
			} else {
				record.AddAttrs(sloga.SlogAttrs()...)
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
			} else if noUnwrap, ok := err.(ErrorNotUnwrapped); ok {
				if msg := noUnwrap.ErrorNoUnwrap(); msg != "" {
					if msgUnrecognized == "" || strings.HasPrefix(msgUnrecognized, msg) {
						msgs = append(msgs, msg)
						msgUnrecognized = ""
					}
				}
			} else if msgUnrecognized == "" {
				if len(msgs) > 0 && newErrMsg != msgs[len(msgs)-1] {
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
//			zap.S().Errorf("%+v", err)
//		} else {
//			zap.S().Error(output())
//		}
//	}
func SlogTextBuffer(opts *slog.HandlerOptions) (slog.Handler, func() string) {
	buf := bytes.NewBuffer([]byte{})
	h := slog.NewTextHandler(buf, opts)
	return h, func() string { return buf.String() }
}

// slogMsg returns a message from a SlogMessager or a HasSlogRecord.
// Otherwise it returns an empty string
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
// slogAttributes returns a message from a SlogAttributer or a HasSlogRecord.
// Otherwise it returns a nil array
func slogAttributes(err any) []slog.Attr {
	if sloga, ok := err.(SlogAttributer); ok {
		return sloga.SlogAttrs()
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

type attributeError interface {
	SlogAttributer
	error
}

func toSlogRecord(err attributeError, skip int) *slog.Record {
	if skip <= 0 {
		skip = 2
	}
	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])
	pc = pcs[0]
	var msg string
	if sm := slogMsg(err); sm != nil {
		msg = *sm
	} else {
		msg = err.Error()
	}
	record := slog.NewRecord(time.Now(), slog.LevelError, msg, pc)
	record.AddAttrs(err.SlogAttrs()...)
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

func textFromRecord(err attributeError) string {
	record := toSlogRecord(err, 0)
	if record == nil {
		return ""
	}
	return joinZero(" ", record.Message, structureAsText(*record))
}
