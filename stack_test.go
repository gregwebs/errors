package errors

import (
	"testing"
)

func TestStackTrace(t *testing.T) {
	tests := []struct {
		err  error
		want []string
	}{{
		New("ooh"), []string{
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:12",
		},
	}, {
		Wrap(New("ooh"), "ahh"), []string{
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:17", // this is the stack of Wrap, not New
		},
	}, {
		Cause(Wrap(New("ooh"), "ahh")), []string{
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:22", // this is the stack of New
		},
	}, {
		func() error { return New("ooh") }(), []string{
			`github.com/gregwebs/errors.(func·009|TestStackTrace.func1)` +
				"\n\tgithub.com/gregwebs/errors/stack_test.go:27", // this is the stack of New
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:27", // this is the stack of New's caller
		},
	}, {
		func() error {
			return func() error {
				return Errorf("hello %s", "world")
			}()
		}(), []string{
			`github.com/gregwebs/errors.(func·010|TestStackTrace.func2.1)` +
				"\n\tgithub.com/gregwebs/errors/stack_test.go:36", // this is the stack of Errorf
			`github.com/gregwebs/errors.(func·011|TestStackTrace.func2)` +
				"\n\tgithub.com/gregwebs/errors/stack_test.go:37", // this is the stack of Errorf's caller
			"github.com/gregwebs/errors.TestStackTrace\n" +
				"\tgithub.com/gregwebs/errors/stack_test.go:38", // this is the stack of Errorf's caller's caller
		},
	}}
	for i, tt := range tests {
		ste := GetStackTracer(tt.err)
		if ste == nil {
			t.Fatalf("expected a stack trace from test %d error: %v", i+1, tt.err)
		}
		st := ste.StackTrace()
		for j, want := range tt.want {
			testFormatRegexp(t, i, st[j], "%+v", want)
		}
	}
}
