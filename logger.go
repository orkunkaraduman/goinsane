package goinsane

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
)

var _ Logger = (*log.Logger)(nil)

type Logger interface {
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
}

type StackLogger interface {
	Stack(msg string, all bool)
}

var _ Logger = SimpleLogger{}
var _ StackLogger = SimpleLogger{}

type SimpleLogger struct {
	Logger *log.Logger
	Prefix string
}

func (l SimpleLogger) Print(v ...any) {
	if l.Logger == nil {
		return
	}
	buf := make([]byte, 0, 4096)
	buf = append(buf, l.Prefix...)
	l.output(fmt.Append(buf, v...))
}

func (l SimpleLogger) Printf(format string, v ...any) {
	if l.Logger == nil {
		return
	}
	buf := make([]byte, 0, 4096)
	buf = append(buf, l.Prefix...)
	l.output(fmt.Appendf(buf, format, v...))
}

func (l SimpleLogger) Println(v ...any) {
	if l.Logger == nil {
		return
	}
	buf := make([]byte, 0, 4096)
	buf = append(buf, l.Prefix...)
	l.output(fmt.Appendln(buf, v...))
}

func (l SimpleLogger) Stack(msg string, all bool) {
	if l.Logger == nil {
		return
	}
	buf := make([]byte, 1<<16)
	buf = buf[:runtime.Stack(buf[:len(buf)-1], all)]
	s := strconv.Quote(l.Prefix + msg)
	builder := new(strings.Builder)
	builder.WriteString(s[1 : len(s)-1])
	builder.WriteByte('\n')
	for b := range bytes.SplitSeq(buf, []byte{'\n'}) {
		builder.WriteByte('\t')
		builder.Write(b)
		builder.WriteByte('\n')
	}
	l.Logger.Output(2, builder.String())
}

func (l SimpleLogger) output(buf []byte) {
	builder := new(strings.Builder)
	if n := len(buf); n > 0 && buf[n-1] == '\n' {
		buf = buf[:n-1]
	}
	for i, b := range bytes.Split(buf, []byte{'\n'}) {
		if i > 0 {
			builder.WriteByte('\t')
		}
		builder.Write(b)
		builder.WriteByte('\n')
	}
	l.Logger.Output(3, builder.String())
}

var defaultLoggerPointer atomic.Pointer[Logger]

func init() {
	SetLogger(nil)
}

func SetLogger(logger Logger) {
	if logger == nil {
		logger = SimpleLogger{Logger: log.Default()}
	}
	defaultLoggerPointer.Store(&logger)
}

func Log() Logger {
	return *defaultLoggerPointer.Load()
}

func LogStack(msg string, all bool) {
	if l, ok := (*defaultLoggerPointer.Load()).(StackLogger); ok {
		l.Stack(msg, all)
	}
}
