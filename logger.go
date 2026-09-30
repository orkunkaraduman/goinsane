package goinsane

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
)

var _ Logger = (*log.Logger)(nil)

type Logger interface {
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
}

var _ Logger = (*SimpleLogger)(nil)

type SimpleLogger struct {
	Logger *log.Logger
}

func (l *SimpleLogger) Print(v ...any) {
	if l == nil || l.Logger == nil {
		return
	}
	buf := make([]byte, 0, 4096)
	l.output(fmt.Append(buf, v...))
}

func (l *SimpleLogger) Printf(format string, v ...any) {
	if l == nil || l.Logger == nil {
		return
	}
	buf := make([]byte, 0, 4096)
	l.output(fmt.Appendf(buf, format, v...))
}

func (l *SimpleLogger) Println(v ...any) {
	if l == nil || l.Logger == nil {
		return
	}
	buf := make([]byte, 0, 4096)
	l.output(fmt.Appendln(buf, v...))
}

func (l *SimpleLogger) Stack(msg string, all bool) {
	if l == nil || l.Logger == nil {
		return
	}
	buf := make([]byte, 1<<16)
	buf = buf[:runtime.Stack(buf[:len(buf)-1], all)]
	builder := new(strings.Builder)
	msg = strconv.Quote(msg)
	builder.WriteString(msg[1 : len(msg)-1])
	builder.WriteByte('\n')
	for b := range bytes.SplitSeq(buf, []byte{'\n'}) {
		builder.WriteByte('\t')
		builder.Write(b)
		builder.WriteByte('\n')
	}
	l.Logger.Output(2, builder.String())
}

func (l *SimpleLogger) output(buf []byte) {
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
