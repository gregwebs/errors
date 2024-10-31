package errors_test

import (
	"fmt"
	"strings"

	"github.com/gregwebs/errors"
)

func ExampleNew() {
	err := errors.New("whoops")
	fmt.Println(err)

	// Output: whoops
}

func firstLines(s string) string {
	allLines := strings.SplitAfterN(s, "\n", 10)
	lines := allLines[0:1]
	for _, line := range allLines[1 : len(allLines)-2] {
		// function
		if strings.HasPrefix(line, "github.com") {
			lines = append(lines, line)
			// File (different on different machine)
		} else if strings.HasPrefix(line, "\t") {
			continue
		} else {
			break
		}
	}
	return strings.Join(lines, "")
}

func ExampleNew_printf() {
	err := errors.Errorf("whoops, %d", 2)
	fmt.Print(firstLines(fmt.Sprintf("%+v", err)))

	// Output:
	// whoops, 2
	// github.com/gregwebs/errors_test.ExampleNew_printf
}

func ExampleWithMessage() {
	cause := errors.New("whoops")
	err := errors.WithMessage(cause, "oh noes")
	fmt.Println(err)

	// Output: oh noes: whoops
}

func ExampleAddStack() {
	cause := errors.New("whoops")
	err := errors.AddStack(cause)
	fmt.Println(err)

	// Output: whoops
}

func ExampleAddStack_printf() {
	cause := errors.New("whoops")
	err := errors.AddStack(cause)
	fmt.Print(firstLines(fmt.Sprintf("%+v", err)))

	// Output:
	// whoops
	// github.com/gregwebs/errors_test.ExampleAddStack_printf
}

func ExampleWrap() {
	cause := errors.New("whoops")
	err := errors.Wrap(cause, "oh noes")
	fmt.Println(err)

	// Output: oh noes: whoops
}

func fn() error {
	e1 := errors.New("error")
	e2 := errors.Wrap(e1, "inner")
	e3 := errors.Wrap(e2, "middle")
	return errors.Wrap(e3, "outer")
}

func ExampleCause() {
	err := fn()
	fmt.Println(err)
	fmt.Println(errors.Cause(err))

	// Output: outer: middle: inner: error
	// error
}

func ExampleWrap_extended() {
	err := fn()
	fmt.Print(firstLines(fmt.Sprintf("%+v", err)))

	// output:
	// error
	// github.com/gregwebs/errors_test.fn
	// github.com/gregwebs/errors_test.ExampleWrap_extended
}

func ExampleWrapf() {
	cause := errors.New("whoops")
	err := errors.Wrapf(cause, "oh noes #%d", 2)
	fmt.Println(err)

	// Output: oh noes #2: whoops
}

func ExampleErrorf_extended() {
	err := errors.Errorf("whoops: %s", "foo")
	fmt.Print(firstLines(fmt.Sprintf("%+v", err)))

	// Output:
	// whoops: foo
	// github.com/gregwebs/errors_test.ExampleErrorf_extended
}

func Example_stackTrace() {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	err, ok := errors.Cause(fn()).(stackTracer)
	if !ok {
		panic("oops, err does not implement stackTracer")
	}

	st := err.StackTrace()
	fmt.Print(firstLines(fmt.Sprintf("%+v", st)))

	// Output:
	// github.com/gregwebs/errors_test.fn
	// github.com/gregwebs/errors_test.Example_stackTrace
}

func ExampleCause_printf() {
	err := errors.Wrap(func() error {
		return func() error {
			return errors.Errorf("hello %s", "world")
		}()
	}(), "failed")

	fmt.Printf("%v", err)

	// Output: failed: hello world
}

func ExampleStructuredError() {
	err := errors.Wraps(
		errors.New("cause"),
		"structured",
		"key", "value",
		"int", 1,
	)

	fmt.Println(err.Error())
	// Output: structured key=value int=1: cause
}

func ExampleSlog() {
	err := errors.Slog(
		"cause",
		"key", "value",
		"int", 1,
	)

	fmt.Println(err.Error())
	// Output: cause key=value int=1
}

func ExampleSlogRecord() {
	err := errors.Wraps(
		errors.New("cause"),
		"structured",
		"key", "value",
		"int", 1,
	)

	rec := errors.SlogRecord(err)
	fmt.Println(rec.Message)
	// Output:
	// structured: cause
}
