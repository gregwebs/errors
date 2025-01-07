package errors

import (
	"fmt"
	"io"
	"log"
	"path"
	"runtime"
	"strings"
)

// StackTracer retrieves the StackTrace
type StackTracer interface {
	StackTrace() StackTrace
}

// StackTraceFormatter is an alternative to StackTracer that only uses standard library types
// In practice the stack trace is usually only used for printing.
// With this definition a package can define a printing of a stack trace without importing this package.
type StackTraceFormatter interface {
	FormatStackTrace(s fmt.State, verb rune)
}

// Frame represents a program counter inside a stack frame.
type Frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// file returns the full path to the file that contains the
// function for this Frame's pc.
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line returns the line number of source code of the
// function for this Frame's pc.
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   function name and path of source file relative to the compile time
//	      GOPATH separated by \n\t (<funcname>\n\t<path>)
//	%+v   equivalent to %+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			pc := f.pc()
			fn := runtime.FuncForPC(pc)
			if fn == nil {
				writeStringStackfmt(s, "unknown")
			} else {
				file, _ := fn.FileLine(pc)
				fmt.Fprintf(s, "%s\n\t%s", fn.Name(), file)
			}
		default:
			writeStringStackfmt(s, path.Base(f.file()))
		}
	case 'd':
		fmt.Fprintf(s, "%d", f.line())
	case 'n':
		name := runtime.FuncForPC(f.pc()).Name()
		writeStringStackfmt(s, funcname(name))
	case 'v':
		f.Format(s, 's')
		writeStringStackfmt(s, ":")
		f.Format(s, 'd')
	}
}

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

type Stack []uintptr

func (st Stack) StackTrace() StackTrace {
	return st.Frames()
}

// Format formats the stack of Frames according to the fmt.Formatter interface.
//
//	%s	lists source files for each Frame in the stack
//	%v	lists the source file and line number for each Frame in the stack
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+v   Prints filename, function, and line number for each Frame in the stack.
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				fmt.Fprintf(s, "\n%+v", f)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []Frame(st))
		default:
			fmt.Fprintf(s, "%v", []Frame(st))
		}
	case 's':
		fmt.Fprintf(s, "%s", []Frame(st))
	}
}

func (st StackTrace) FormatStackTrace(s fmt.State, verb rune) {
	st.Format(s, verb)
}

func (st Stack) Format(s fmt.State, verb rune) {
	StackTrace(st.Frames()).Format(s, verb)
}

func (st Stack) FormatStackTrace(s fmt.State, verb rune) {
	StackTrace(st.Frames()).FormatStackTrace(s, verb)
}

func (s Stack) Frames() []Frame {
	f := make([]Frame, len(s))
	for i := 0; i < len(f); i++ {
		f[i] = Frame(s[i])
	}
	return f
}

func NewStack() Stack {
	return NewStackSkip(2)
}

func NewStackSkip(skip int) Stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2+skip, pcs[:])
	var st Stack = pcs[0:n]
	return st
}

// funcname removes the path prefix component of a function's name reported by func.Name().
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// HandleWriteErrorStackfmt handles (rare) errors when writing to fmt.State.
// It defaults to printing the errors.
func HandleWriteErrorStackfmt(handler func(err error)) {
	handleWriteErrorStackfmt = handler
}

var handleWriteErrorStackfmt = func(err error) {
	log.Println(err)
}

func writeStringStackfmt(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		handleWriteErrorStackfmt(err)
	}
}
