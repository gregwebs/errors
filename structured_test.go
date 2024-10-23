package errors

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestStructuredBad(t *testing.T) {
	errBad := Slog(
		"cause1",
		"structured1",
		"key", "value",
		"int", 1,
	)
	record := SlogRecord(errBad)
	if numAttrs := record.NumAttrs(); numAttrs != 3 {
		t.Errorf("expected 3 attributes, got %d for %s", numAttrs, errBad.Error())
	}
	handler, getBuf := SlogTextBuffer(nil)
	if err := handler.Handle(context.Background(), *record); err != nil {
		t.Fatalf("error writing out record %+v", err)
	}
	if !strings.Contains(getBuf(), "!BADKEY") {
		t.Errorf("expected BADKEY from bad wrapping")
	}
}

func TestStructured(t *testing.T) {
	errInner := Slog(
		"cause1",
		"key", "value",
		"int", 1,
	)
	err := Wraps(
		errInner,
		"structured2",
		"key", "value",
		"int", 3,
	)

	if numAttrs := err.GetSlogRecord().NumAttrs(); numAttrs != 2 {
		t.Errorf("expected 2 attributes, got %d for %s", numAttrs, err.Error())
	}
	record := SlogRecord(err)
	if numAttrs := record.NumAttrs(); numAttrs != 4 {
		t.Errorf("expected 4 attributes, got %d for %s", numAttrs, err.Error())
	}
	if record.Message != "structured2: cause1" {
		t.Errorf("unexpected record message: %s", record.Message)
	}
	if err.Error() != "structured2 key=value int=3: cause1 key=value int=1" {
		t.Errorf("unexpected Error: %s", err.Error())
	}

	// Test stack trace
	hOpts := slog.HandlerOptions{
		AddSource: true,
	}
	handler, getBuf := SlogTextBuffer(&hOpts)
	if err := handler.Handle(context.Background(), *record); err != nil {
		t.Fatalf("error writing out record %+v", err)
	}
	if !strings.Contains(getBuf(), "structured_test.go") {
		t.Errorf("expected stack trace with file")
	}
}

type container struct {
	error
}

func (c container) Unwrap() error {
	return c.error
}

func TestStructuredWrap(t *testing.T) {
	errInner := Wraps(
		New("cause1"),
		"structured1",
		"key", "value",
		"int", 1,
	)
	err := container{Wraps(
		errInner,
		"structured2",
		"key", "value",
		"int", 3,
	)}

	if numAttrs := errInner.GetSlogRecord().NumAttrs(); numAttrs != 2 {
		t.Errorf("expected 2 attributes, got %d for %s", numAttrs, err.Error())
	}
	record := SlogRecord(err)
	if numAttrs := record.NumAttrs(); numAttrs != 4 {
		t.Errorf("expected 4 attributes, got %d for %s", numAttrs, err.Error())
	}
	if record.Message != "structured2: structured1: cause1" {
		t.Errorf("unexpected record message: %s", record.Message)
	}
	if err.Error() != "structured2 key=value int=3: structured1 key=value int=1: cause1" {
		t.Errorf("unexpected Error: %s", err.Error())
	}

	// Test stack trace
	hOpts := slog.HandlerOptions{
		AddSource: true,
	}
	handler, getBuf := SlogTextBuffer(&hOpts)
	if err := handler.Handle(context.Background(), *record); err != nil {
		t.Fatalf("error writing out record %+v", err)
	}
	if !strings.Contains(getBuf(), "structured_test.go") {
		t.Errorf("expected stack trace with file")
	}
}

func TestStructuredNil(t *testing.T) {
	if err := Wraps(nil, "testing nil error", "test", 1); err != nil {
		t.Errorf("expected nil")
	}
}

func TestStructuredAttr(t *testing.T) {
	attrs := []slog.Attr{}
	attrs = append(attrs, slog.String("string", "test"))
	attrs = append(attrs, slog.Int("int", 1))
	err := Wraps(errors.New("error"), "testing attrs", attrs)
	got := err.Error()
	expected := "testing attrs string=test int=1: error"
	if got != expected {
		t.Errorf("\nexpected: '%s'\n but got: '%s'", expected, got)
	}
}

type slogAttrs struct {
	inner error
}

func (sa slogAttrs) Unwrap() error {
	return sa.inner
}

func (sa slogAttrs) Error() string {
	var inner string
	if sa.inner != nil {
		inner = sa.inner.Error()
	}
	return joinZero(": ", textFromRecord(sa), inner)
}

func (sa slogAttrs) SlogMsg() string {
	return "cause1"
}

func (sa slogAttrs) SlogAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("key", "value"),
		slog.Int("int", 1),
	}
}

func TestStructuredAttrsInner(t *testing.T) {
	err := Wraps(
		slogAttrs{},
		"structured2",
		"key", "value",
		"int", 3,
	)

	if numAttrs := err.GetSlogRecord().NumAttrs(); numAttrs != 2 {
		t.Errorf("expected 2 attributes, got %d for %s", numAttrs, err.Error())
	}
	record := SlogRecord(err)
	if numAttrs := record.NumAttrs(); numAttrs != 4 {
		t.Errorf("expected 4 attributes, got %d for %s", numAttrs, err.Error())
	}
	if record.Message != "structured2: cause1" {
		t.Errorf("unexpected record message: %s", record.Message)
	}
	if err.Error() != "structured2 key=value int=3: cause1 key=value int=1" {
		t.Errorf("unexpected Error: %s", err.Error())
	}

	// Test stack trace
	hOpts := slog.HandlerOptions{
		AddSource: true,
	}
	handler, getBuf := SlogTextBuffer(&hOpts)
	if err := handler.Handle(context.Background(), *record); err != nil {
		t.Fatalf("error writing out record %+v", err)
	}
	if !strings.Contains(getBuf(), "structured_test.go") {
		t.Errorf("expected stack trace with file")
	}
}

func TestStructuredAttrsOuter(t *testing.T) {
	errInner := Slog(
		"structured2",
		"key", "value",
		"int", 3,
	)
	err := container{slogAttrs{inner: errInner}}
	record := SlogRecord(err)
	if numAttrs := record.NumAttrs(); numAttrs != 4 {
		t.Errorf("expected 4 attributes, got %d for %s", numAttrs, err.Error())
	}
	if record.Message != "cause1: structured2" {
		t.Errorf("unexpected record message: %s", record.Message)
	}
	if err.Error() != "cause1 key=value int=1: structured2 key=value int=3" {
		t.Errorf("unexpected Error: %s", err.Error())
	}

	// Test stack trace
	hOpts := slog.HandlerOptions{
		AddSource: true,
	}
	handler, getBuf := SlogTextBuffer(&hOpts)
	if err := handler.Handle(context.Background(), *record); err != nil {
		t.Fatalf("error writing out record %+v", err)
	}
	bufOut := getBuf()
	if !strings.Contains(bufOut, "structured_test.go") {
		t.Errorf("expected stack trace with file, got %s", bufOut)
	}
}

func TestStructuredWrapping(t *testing.T) {
	errInner := Slog("structured2", "k", "v")
	wrapped := Wraps(errInner, "", "outer", 1)
	expectedErrorMsg := "outer=1: structured2 k=v"
	if wrapped.Error() != expectedErrorMsg {
		t.Errorf("Unexpected Error(): %s", wrapped.Error())
	}

	record := SlogRecord(wrapped)
	if record == nil {
		t.Errorf("no SlogRecord")
	} else {
		text := joinZero(" ", record.Message, structureAsText(*record))
		if text != "structured2 outer=1 k=v" {
			t.Errorf("Unexpected text: %s", text)
		}
	}
}
