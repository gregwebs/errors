package stackfmt

import (
	"fmt"
	"testing"
)

// This comment helps to maintain original line numbers
// Perhaps this test is too fragile :)
func stackTrace() StackTrace {
	return NewStackSkip(0).StackTrace()
	// This comment helps to maintain original line numbers
	// Perhaps this test is too fragile :)
}

func TestStackTraceFormat(t *testing.T) {
	tests := []struct {
		StackTrace
		format string
		want   string
	}{{
		nil,
		"%s",
		`[]`,
	}, {
		nil,
		"%v",
		`[]`,
	}, {
		nil,
		"%+v",
		"",
	}, {
		nil,
		"%#v",
		`[]stackfmt.Frame(nil)`,
	}, {
		make(StackTrace, 0),
		"%s",
		`[]`,
	}, {
		make(StackTrace, 0),
		"%v",
		`[]`,
	}, {
		make(StackTrace, 0),
		"%+v",
		"",
	}, {
		make(StackTrace, 0),
		"%#v",
		`[]stackfmt.Frame{}`,
	}, {
		stackTrace()[:2],
		"%s",
		`[stackfmt_test.go stackfmt_test.go]`,
	}, {
		stackTrace()[:2],
		"%v",
		`[stackfmt_test.go:11 stackfmt_test.go:58]`,
	}, {
		stackTrace()[:2],
		"%+v",
		"\n" +
			"github.com/gregwebs/errors/stackfmt.stackTrace\n" +
			"\tgithub.com/gregwebs/errors/stackfmt/stackfmt_test.go:11\n" +
			"github.com/gregwebs/errors/stackfmt.TestStackTraceFormat\n" +
			"\tgithub.com/gregwebs/errors/stackfmt/stackfmt_test.go:62",
	}, {
		stackTrace()[:2],
		"%#v",
		`[]stackfmt.Frame{stackfmt_test.go:11, stackfmt_test.go:70}`,
	}}

	for i, tt := range tests {
		testFormatString(t, i, tt.StackTrace, tt.format, tt.want)
	}
}

func TestNewStack(t *testing.T) {
	got := NewStackSkip(1).StackTrace()
	want := NewStackSkip(1).StackTrace()
	if got[0] != want[0] {
		t.Errorf("NewStack(remove NewStack): want: %v, got: %v", want, got)
	}
	gotFirst := fmt.Sprintf("%+v", got[0])[0:15]
	if gotFirst != "testing.tRunner" {
		t.Errorf("NewStack(): want: %v, got: %+v", "testing.tRunner", gotFirst)
	}
}

func TestFuncname(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"", ""},
		{"runtime.main", "main"},
		{"github.com/gregwebs/errors.funcname", "funcname"},
		{"funcname", "funcname"},
		{"io.copyBuffer", "copyBuffer"},
		{"main.(*R).Write", "(*R).Write"},
	}

	for _, tt := range tests {
		got := funcname(tt.name)
		want := tt.want
		if got != want {
			t.Errorf("funcname(%q): want: %q, got %q", tt.name, want, got)
		}
	}
}
