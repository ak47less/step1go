package step1go

import (
	"fmt"
	"os"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

type LogLevel byte

const (
	LogLevel0 LogLevel = iota
	LogLevelMin

	LogLevelTrace
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal

	LogLevelMax
)

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelInfo:
		return "INFO_"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelTrace:
		return "TRACE"
	case LogLevelWarn:
		return "WARN_"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	}
	return fmt.Sprintf("LogLevel(%d)", ll)
}

////////////////////////////////////////////////////////////////////////////////

type LogContext struct {
	tag string

	logger *Logger

	handler LogHandler

	levelGate LogLevel
}

////////////////////////////////////////////////////////////////////////////////

type Logger struct {
	ctx *LogContext
}

func (inst *Logger) Fatal(fmt string, args ...any) *Logger {
	const lv = LogLevelFatal
	return inst.innerWriteFmt(lv, fmt, args...)
}

func (inst *Logger) Error(fmt string, args ...any) *Logger {
	const lv = LogLevelError
	return inst.innerWriteFmt(lv, fmt, args...)
}

func (inst *Logger) Warn(fmt string, args ...any) *Logger {
	const lv = LogLevelWarn
	return inst.innerWriteFmt(lv, fmt, args...)
}

func (inst *Logger) Info(fmt string, args ...any) *Logger {
	const lv = LogLevelInfo
	return inst.innerWriteFmt(lv, fmt, args...)
}

func (inst *Logger) Debug(fmt string, args ...any) *Logger {
	const lv = LogLevelDebug
	return inst.innerWriteFmt(lv, fmt, args...)
}

func (inst *Logger) Trace(fmt string, args ...any) *Logger {
	const lv = LogLevelTrace
	return inst.innerWriteFmt(lv, fmt, args...)
}

func (inst *Logger) SetGate(l LogLevel) *Logger {
	inst.ctx.levelGate = l
	return inst
}

func (inst *Logger) innerWriteFmt(lv LogLevel, f string, args ...any) *Logger {

	rec := new(LogRecord)
	msg := fmt.Sprintf(f, args...)

	rec.Level = lv
	rec.Tag = inst.ctx.tag
	rec.Message = msg
	rec.Time = time.Now()

	return inst.innerWriteRec(rec)
}

func (inst *Logger) innerWriteRec(rec *LogRecord) *Logger {
	inst.ctx.handler.Handle(rec)
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// LogRecord

type LogRecord struct {
	Level LogLevel

	Time time.Time

	Tag string

	Message string
}

////////////////////////////////////////////////////////////////////////////////

type LogHandler interface {
	Handle(rec *LogRecord)
}

////////////////////////////////////////////////////////////////////////////////

type innerLogHandler struct {
	ctx *LogContext
}

// Handle implements [LogHandler].
func (inst *innerLogHandler) Handle(rec *LogRecord) {

	dst := os.Stdout
	ts := rec.Time
	msg := rec.Message
	tag := rec.Tag
	level := rec.Level
	gate := inst.ctx.levelGate

	if level < gate {
		return
	}

	if level >= LogLevelError {
		dst = os.Stderr
	}

	const nl = "\n"
	tstr := inst.formatTime(ts)

	fmt.Fprint(dst, "[", level, "] ", tstr, " (", tag, ") ", msg, nl)
}

func (inst *innerLogHandler) formatTime(t time.Time) string {
	return t.Format(time.DateTime)
}

////////////////////////////////////////////////////////////////////////////////

type innerLoggerHolder struct {
	ctx *LogContext
}

var theLoggerHolder innerLoggerHolder

func (inst *innerLoggerHolder) GetLogger() *Logger {
	return inst.GetContext().logger
}

func (inst *innerLoggerHolder) GetContext() *LogContext {
	ctx := inst.ctx
	if ctx == nil {
		ctx = inst.createContext()
		inst.ctx = ctx
	}
	return ctx
}

func (inst *innerLoggerHolder) createContext() *LogContext {

	ctx := new(LogContext)
	logger := new(Logger)
	handler := new(innerLogHandler)

	logger.ctx = ctx
	handler.ctx = ctx

	ctx.handler = handler
	ctx.levelGate = LogLevelInfo
	ctx.tag = "step1go"
	ctx.logger = logger

	return ctx
}

////////////////////////////////////////////////////////////////////////////////

func Log() *Logger {
	return theLoggerHolder.GetLogger()
}

////////////////////////////////////////////////////////////////////////////////
// EOF
