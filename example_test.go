package errors_test

import (
	"fmt"

	"github.com/gregwebs/errors"
)

func ExampleNew() {
	err := errors.New("whoops")
	fmt.Println(err)

	// Output: whoops
}

func ExampleNew_printf() {
	err := errors.New("whoops")
	fmt.Printf("%+v", err)

	// Example output:
	// whoops
	// github.com/gregwebs/errors_test.ExampleNew_printf
	//         /home/dfc/src/github.com/gregwebs/errors/example_test.go:17
	// testing.runExample
	//         /home/dfc/go/src/testing/example.go:114
	// testing.RunExamples
	//         /home/dfc/go/src/testing/example.go:38
	// testing.(*M).Run
	//         /home/dfc/go/src/testing/testing.go:744
	// main.main
	//         /github.com/gregwebs/errors/_test/_testmain.go:106
	// runtime.main
	//         /home/dfc/go/src/runtime/proc.go:183
	// runtime.goexit
	//         /home/dfc/go/src/runtime/asm_amd64.s:2059
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
	fmt.Printf("%+v", err)

	// Example Output:
	// whoops
	// github.com/gregwebs/errors_test.ExampleAddStack_printf
	//         /home/fabstu/go/src/github.com/gregwebs/errors/example_test.go:55
	// testing.runExample
	//         /usr/lib/go/src/testing/example.go:114
	// testing.RunExamples
	//         /usr/lib/go/src/testing/example.go:38
	// testing.(*M).Run
	//         /usr/lib/go/src/testing/testing.go:744
	// main.main
	//         github.com/gregwebs/errors/_test/_testmain.go:106
	// runtime.main
	//         /usr/lib/go/src/runtime/proc.go:183
	// runtime.goexit
	//         /usr/lib/go/src/runtime/asm_amd64.s:2086
	// github.com/gregwebs/errors_test.ExampleAddStack_printf
	//         /home/fabstu/go/src/github.com/gregwebs/errors/example_test.go:56
	// testing.runExample
	//         /usr/lib/go/src/testing/example.go:114
	// testing.RunExamples
	//         /usr/lib/go/src/testing/example.go:38
	// testing.(*M).Run
	//         /usr/lib/go/src/testing/testing.go:744
	// main.main
	//         github.com/gregwebs/errors/_test/_testmain.go:106
	// runtime.main
	//         /usr/lib/go/src/runtime/proc.go:183
	// runtime.goexit
	//         /usr/lib/go/src/runtime/asm_amd64.s:2086
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
	fmt.Printf("%+v\n", err)

	// Example output:
	// error
	// github.com/gregwebs/errors_test.fn
	//         /home/dfc/src/github.com/gregwebs/errors/example_test.go:47
	// github.com/gregwebs/errors_test.ExampleCause_printf
	//         /home/dfc/src/github.com/gregwebs/errors/example_test.go:63
	// testing.runExample
	//         /home/dfc/go/src/testing/example.go:114
	// testing.RunExamples
	//         /home/dfc/go/src/testing/example.go:38
	// testing.(*M).Run
	//         /home/dfc/go/src/testing/testing.go:744
	// main.main
	//         /github.com/gregwebs/errors/_test/_testmain.go:104
	// runtime.main
	//         /home/dfc/go/src/runtime/proc.go:183
	// runtime.goexit
	//         /home/dfc/go/src/runtime/asm_amd64.s:2059
	// github.com/gregwebs/errors_test.fn
	// 	  /home/dfc/src/github.com/gregwebs/errors/example_test.go:48: inner
	// github.com/gregwebs/errors_test.fn
	//        /home/dfc/src/github.com/gregwebs/errors/example_test.go:49: middle
	// github.com/gregwebs/errors_test.fn
	//      /home/dfc/src/github.com/gregwebs/errors/example_test.go:50: outer
}

func ExampleWrapf() {
	cause := errors.New("whoops")
	err := errors.Wrapf(cause, "oh noes #%d", 2)
	fmt.Println(err)

	// Output: oh noes #2: whoops
}

func ExampleErrorf_extended() {
	err := errors.Errorf("whoops: %s", "foo")
	fmt.Printf("%+v", err)

	// Example output:
	// whoops: foo
	// github.com/gregwebs/errors_test.ExampleErrorf
	//         /home/dfc/src/github.com/gregwebs/errors/example_test.go:101
	// testing.runExample
	//         /home/dfc/go/src/testing/example.go:114
	// testing.RunExamples
	//         /home/dfc/go/src/testing/example.go:38
	// testing.(*M).Run
	//         /home/dfc/go/src/testing/testing.go:744
	// main.main
	//         /github.com/gregwebs/errors/_test/_testmain.go:102
	// runtime.main
	//         /home/dfc/go/src/runtime/proc.go:183
	// runtime.goexit
	//         /home/dfc/go/src/runtime/asm_amd64.s:2059
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
	fmt.Printf("%+v", st[0:2]) // top two frames

	// Example output:
	// github.com/gregwebs/errors_test.fn
	//	/home/dfc/src/github.com/gregwebs/errors/example_test.go:47
	// github.com/gregwebs/errors_test.Example_stackTrace
	//	/home/dfc/src/github.com/gregwebs/errors/example_test.go:127
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
