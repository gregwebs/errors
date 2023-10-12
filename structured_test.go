package errors

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestStructured(t *testing.T) {
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

	if numAttrs := err.Record.NumAttrs(); numAttrs != 2 {
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
	buf := new(bytes.Buffer)
	if err := slog.NewTextHandler(buf, &hOpts).Handle(context.Background(), *record); err != nil {
		t.Fatalf("error writing out record %+v", err)
	}
	if !strings.Contains(buf.String(), "structured_test.go") {
		t.Errorf("expected stack trace with file")
	}
}
