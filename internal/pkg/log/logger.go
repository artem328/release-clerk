package log

import (
	"fmt"
	"io"
)

type Logger interface {
	Log(args ...any)
	Logf(format string, args ...any)
	Debug(args ...any)
	Debugf(format string, args ...any)
	Section(name string) Logger
}

type GenericLogger struct {
	out     io.Writer
	section string
	debug   bool
}

func NewGenericLogger(out io.Writer, section string, debug bool) GenericLogger {
	return GenericLogger{out: out, section: section, debug: debug}
}

func (l GenericLogger) Log(args ...any) {
	l.log(args, false)
}

func (l GenericLogger) Logf(format string, args ...any) {
	l.logf(format, args, false)
}

func (l GenericLogger) Debug(args ...any) {
	l.log(args, true)
}

func (l GenericLogger) Debugf(format string, args ...any) {
	l.logf(format, args, true)
}

func (l GenericLogger) Section(name string) Logger {
	if l.section != "" {
		name = l.section + "." + name
	}

	return NewGenericLogger(l.out, name, l.debug)
}

func (l GenericLogger) logPrefix(debug bool) {
	var needSpace bool

	if l.section != "" {
		_, _ = fmt.Fprint(l.out, "[")
		_, _ = fmt.Fprint(l.out, l.section)
		_, _ = fmt.Fprint(l.out, "]")
		needSpace = true
	}

	if debug && l.debug {
		_, _ = fmt.Fprint(l.out, "[debug]")
		needSpace = true
	}

	if needSpace {
		_, _ = fmt.Fprint(l.out, " ")
	}
}

func (l GenericLogger) log(args []any, debug bool) {
	l.logPrefix(debug)

	_, _ = fmt.Fprint(l.out, args...)
	_, _ = fmt.Fprintln(l.out)
}

func (l GenericLogger) logf(format string, args []any, debug bool) {
	l.logPrefix(debug)

	_, _ = fmt.Fprintf(l.out, format, args...)
	_, _ = fmt.Fprintln(l.out)
}
