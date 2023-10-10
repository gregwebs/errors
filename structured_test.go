package errors

import (
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
	if numAttrs := SlogRecord(err).NumAttrs(); numAttrs != 4 {
		t.Errorf("expected 4 attributes, got %d for %s", numAttrs, err.Error())
	}
}
