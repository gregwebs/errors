package errors

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

type HasSlogRecord interface {
	GetSlogRecord() slog.Record
}

type StructuredErr struct {
	Record  slog.Record
	err     error
	msg     string
	Context context.Context
}

func (se StructuredErr) GetSlogRecord() slog.Record {
	return se.Record
}

func (se StructuredErr) Error() string {
	msg := se.msg

	stext := structureAsText(se.Record)
	if stext != "" {
		msg += " " + stext
	}

	return msg + ": " + se.err.Error()
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

func (se StructuredErr) Unwrap() error         { return se.err }
func (se StructuredErr) ErrorNoUnwrap() string { return se.msg }
func (se StructuredErr) HasStack() bool        { return true }

func (se StructuredErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", se.Unwrap())
			writeString(s, se.msg+" "+structureAsText(se.Record))
			return
		}
		fallthrough
	case 's', 'q':
		writeString(s, se.Error())
	}
}

// S=structured
// Accepts as args any valid slog args.  These will generate an slog Record
func Wraps(err error, msg string, args ...interface{}) StructuredErr {
	return WrapsCtx(context.Background(), err, msg, args...)
}

// S=structured
// Ctx=Context
// Accepts as args any valid slog args.  These will generate an slog Record
func WrapsCtx(ctx context.Context, err error, msg string, args ...interface{}) StructuredErr {
	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(4, pcs[:])
	pc = pcs[0]

	record := slog.NewRecord(time.Now(), slog.LevelError, msg+": "+err.Error(), pc)
	record.Add(args...)

	// TODO: use the exact same stack for the error and the record
	return StructuredErr{
		Record:  record,
		err:     AddStackSkip(err, 1),
		msg:     msg,
		Context: ctx,
	}
}

// SlogRecord traverses the error chain, calling Unwrap(), to look for slog Records
// All records will be merged
// An error string is given as well
// This string may contain some of the structured information
// if the error does not defined ErrorNoUnwrap from the interface ErrorNotUnwrapped
func SlogRecord(inputErr error) *slog.Record {
	var msg string
	var record *slog.Record
	msgDone := false
	WalkDeep(inputErr, func(err error) bool {
		// Once we reach a message that does not understand ErrorNotUnwrapped
		// We must stop the traversal for the messages
		// Still we will collect the records
		if !msgDone {
			if nu, ok := err.(ErrorNotUnwrapped); ok {
				if msg == "" {
					msg = nu.ErrorNoUnwrap()
				} else {
					msg += ": " + nu.ErrorNoUnwrap()
				}
			} else {
				msgDone = true
				if msg == "" {
					msg = err.Error()
				} else {
					msg += ": " + err.Error()
				}
			}
		}

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
			}
		}

		return false // keep going
	})
	if record != nil {
		record.Message = msg
	}
	return record
}
