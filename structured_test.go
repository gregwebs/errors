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

func TestStructuredWrap(t *testing.T) {
	errInner := Wraps(
		New("cause1"),
		"structured1",
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
