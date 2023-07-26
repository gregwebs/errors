package errors

import (
	"fmt"
	"runtime"
	"testing"
)

var initpc, _, _, _ = runtime.Caller(0)

func TestFrameLine(t *testing.T) {
	var tests = []struct {
		Frame
		want int
	}{{
		Frame(initpc),
		9,
	}, {
		func() Frame {
			var pc, _, _, _ = runtime.Caller(0)
			return Frame(pc)
		}(),
		20,
	}, {
		func() Frame {
			var pc, _, _, _ = runtime.Caller(1)
			return Frame(pc)
		}(),
		28,
	}, {
		Frame(0), // invalid PC
		0,
	}}

	for _, tt := range tests {
		got := tt.Frame.line()
		want := tt.want
		if want != got {
			t.Errorf("Frame(%v): want: %v, got: %v", uintptr(tt.Frame), want, got)
		}
	}
}

type X struct{}

func (x X) val() Frame {
	var pc, _, _, _ = runtime.Caller(0)
	return Frame(pc)
}

func (x *X) ptr() Frame {
	var pc, _, _, _ = runtime.Caller(0)
	return Frame(pc)
}

func TestFrameFormat(t *testing.T) {
	var tests = []struct {
		Frame
		format string
		want   string
	}{{
		Frame(initpc),
		"%s",
		"stack_test.go",
	}, {
		Frame(initpc),
		"%+s",
		"github.com/gregwebs/errors.init\n" +
			"\tgithub.com/gregwebs/errors/stack_test.go",
	}, {
		Frame(0),
		"%s",
		"unknown",
	}, {
		Frame(0),
		"%+s",
		"unknown",
	}, {
		Frame(initpc),
		"%d",
		"9",
	}, {
		Frame(0),
		"%d",
		"0",
	}, {
		Frame(initpc),
		"%n",
		"init",
	}, {
		func() Frame {
			var x X
			return x.ptr()
		}(),
		"%n",
		`\(\*X\).ptr`,
	}, {
		func() Frame {
			var x X
			return x.val()
		}(),
		"%n",
		"X.val",
	}, {
		Frame(0),
		"%n",
		"",
	}, {
		Frame(initpc),
		"%v",
		"stack_test.go:9",
	}, {
		Frame(initpc),
		"%+v",
		"github.com/gregwebs/errors.init\n" +
			"\tgithub.com/gregwebs/errors/stack_test.go:9",
	}, {
		Frame(0),
		"%v",
		"unknown:0",
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.Frame, tt.format, tt.want)
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

func TestStackTrace(t *testing.T) {
	tests := []struct {
		err  error
		want []string
	}{{
		New("ooh"), []string{
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:154",
		},
	}, {
		Wrap(New("ooh"), "ahh"), []string{
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:159", // this is the stack of Wrap, not New
		},
	}, {
		Cause(Wrap(New("ooh"), "ahh")), []string{
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:164", // this is the stack of New
		},
	}, {
		func() error { return New("ooh") }(), []string{
			`github.com/gregwebs/errors.(func·009|TestStackTrace.func1)` +
				"\n\tgithub.com/gregwebs/errors/stack_test.go:169", // this is the stack of New
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:169", // this is the stack of New's caller
		},
	}, {
		func() error {
			return func() error {
				return Errorf("hello %s", "world")
			}()
		}(), []string{
			`github.com/gregwebs/errors.(func·010|TestStackTrace.func2.1)` +
				"\n\tgithub.com/gregwebs/errors/stack_test.go:178", // this is the stack of Errorf
			`github.com/gregwebs/errors.(func·011|TestStackTrace.func2)` +
				"\n\tgithub.com/gregwebs/errors/stack_test.go:179", // this is the stack of Errorf's caller
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:180", // this is the stack of Errorf's caller's caller
		},
	}}
	for i, tt := range tests {
		ste := GetStackTracer(tt.err)
		st := ste.StackTrace()
		for j, want := range tt.want {
			testFormatRegexp(t, i, st[j], "%+v", want)
		}
	}
}

// This comment helps to maintain original line numbers
// Perhaps this test is too fragile :)
func stackTrace() StackTrace {
	return NewStack(0).StackTrace()
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
		`\[\]`,
	}, {
		nil,
		"%v",
		`\[\]`,
	}, {
		nil,
		"%+v",
		"",
	}, {
		nil,
		"%#v",
		`\[\]errors.Frame\(nil\)`,
	}, {
		make(StackTrace, 0),
		"%s",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%v",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%+v",
		"",
	}, {
		make(StackTrace, 0),
		"%#v",
		`\[\]errors.Frame{}`,
	}, {
		stackTrace()[:2],
		"%s",
		`\[stack_test.go stack_test.go\]`,
	}, {
		stackTrace()[:2],
		"%v",
		`[stack_test.go:207 stack_test.go:254]`,
	}, {
		stackTrace()[:2],
		"%+v",
		"\n" +
			"github.com/gregwebs/errors.stackTrace\n" +
			"\tgithub.com/gregwebs/errors/stack_test.go:201\n" +
			"github.com/gregwebs/errors.TestStackTraceFormat\n" +
			"\tgithub.com/gregwebs/errors/stack_test.go:252",
	}, {
		stackTrace()[:2],
		"%#v",
		`\[\]errors.Frame{stack_test.go:210, stack_test.go:260}`,
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.StackTrace, tt.format, tt.want)
	}
}

func TestNewStack(t *testing.T) {
	got := NewStack(1).StackTrace()
	want := NewStack(1).StackTrace()
	if got[0] != want[0] {
		t.Errorf("NewStack(remove NewStack): want: %v, got: %v", want, got)
	}
	gotFirst := fmt.Sprintf("%+v", got[0])[0:15]
	if gotFirst != "testing.tRunner" {
		t.Errorf("NewStack(): want: %v, got: %+v", "testing.tRunner", gotFirst)
	}
}
