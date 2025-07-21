package core

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
)

// ConfigLogger is the key for the pipeline's logger
const ConfigLogger = "Core.Logger"

// Logger defines the output interface used by Hercules components.
type Logger interface {
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Critical(...interface{})
	Criticalf(string, ...interface{})
}

// NoOpLogger disables all logging.
type NoOpLogger struct{}

func (n *NoOpLogger) Info(...interface{})              {}
func (n *NoOpLogger) Infof(string, ...interface{})     {}
func (n *NoOpLogger) Warn(...interface{})              {}
func (n *NoOpLogger) Warnf(string, ...interface{})     {}
func (n *NoOpLogger) Error(...interface{})             {}
func (n *NoOpLogger) Errorf(string, ...interface{})    {}
func (n *NoOpLogger) Critical(...interface{})          {}
func (n *NoOpLogger) Criticalf(string, ...interface{}) {}

// FileLogger writes logs to a file.
type FileLogger struct {
	I *log.Logger
	W *log.Logger
	E *log.Logger
}

func NewFileLogger(w io.Writer) *FileLogger {
	return &FileLogger{
		I: log.New(w, "[INFO] ", log.LstdFlags),
		W: log.New(w, "[WARN] ", log.LstdFlags),
		E: log.New(w, "[ERROR] ", log.LstdFlags),
	}
}
func (f *FileLogger) Info(v ...interface{})                 { f.I.Println(v...) }
func (f *FileLogger) Infof(fm string, v ...interface{})     { f.I.Printf(fm, v...) }
func (f *FileLogger) Warn(v ...interface{})                 { f.W.Println(v...) }
func (f *FileLogger) Warnf(fm string, v ...interface{})     { f.W.Printf(fm, v...) }
func (f *FileLogger) Error(v ...interface{})                { f.E.Println(v...) }
func (f *FileLogger) Errorf(fm string, v ...interface{})    { f.E.Printf(fm, v...) }
func (f *FileLogger) Critical(v ...interface{})             { f.E.Println(v...) }
func (f *FileLogger) Criticalf(fm string, v ...interface{}) { f.E.Printf(fm, v...) }

// SlogLogger uses slog for JSON logging.
type SlogLogger struct {
	l *slog.Logger
}

func NewSlogLogger(w io.Writer) *SlogLogger {
	return &SlogLogger{l: slog.New(slog.NewJSONHandler(w, nil))}
}
func (s *SlogLogger) Info(v ...interface{}) { s.l.Info("info", "msg", v) }
func (s *SlogLogger) Infof(fm string, v ...interface{}) {
	s.l.Info("info", "msg", fmt.Sprintf(fm, v...))
}
func (s *SlogLogger) Warn(v ...interface{}) { s.l.Warn("warn", "msg", v) }
func (s *SlogLogger) Warnf(fm string, v ...interface{}) {
	s.l.Warn("warn", "msg", fmt.Sprintf(fm, v...))
}
func (s *SlogLogger) Error(v ...interface{}) { s.l.Error("error", "msg", v) }
func (s *SlogLogger) Errorf(fm string, v ...interface{}) {
	s.l.Error("error", "msg", fmt.Sprintf(fm, v...))
}
func (s *SlogLogger) Critical(v ...interface{}) { s.l.Error("critical", "msg", v) }
func (s *SlogLogger) Criticalf(fm string, v ...interface{}) {
	s.l.Error("critical", "msg", fmt.Sprintf(fm, v...))
}

// Global logger setter
var activeLogger Logger = &NoOpLogger{}

func SetLogger(l Logger) { activeLogger = l }
func GetLogger() Logger  { return activeLogger }

// DefaultLogger is the default logger used by a pipeline, and wraps the standard
// log library.
type DefaultLogger struct {
	I *log.Logger
	W *log.Logger
	E *log.Logger
}

// NewLogger returns a configured default logger.
func NewLogger() *DefaultLogger {
	return &DefaultLogger{
		I: log.New(os.Stderr, "[INFO] ", log.LstdFlags),
		W: log.New(os.Stderr, "[WARN] ", log.LstdFlags),
		E: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
}

// Info writes to "info" logger.
func (d *DefaultLogger) Info(v ...interface{}) { d.I.Println(v...) }

// Infof writes to "info" logger with printf-style formatting.
func (d *DefaultLogger) Infof(f string, v ...interface{}) { d.I.Printf(f, v...) }

// Warn writes to the "warning" logger.
func (d *DefaultLogger) Warn(v ...interface{}) { d.W.Println(v...) }

// Warnf writes to the "warning" logger with printf-style formatting.
func (d *DefaultLogger) Warnf(f string, v ...interface{}) { d.W.Printf(f, v...) }

// Error writes to the "error" logger.
func (d *DefaultLogger) Error(v ...interface{}) { d.E.Println(v...) }

// Errorf writes to the "error" logger with printf-style formatting.
func (d *DefaultLogger) Errorf(f string, v ...interface{}) { d.E.Printf(f, v...) }

// Critical writes to the "error" logger and logs the current stacktrace.
func (d *DefaultLogger) Critical(v ...interface{}) {
	d.E.Println(v...)
	d.logStacktraceToErr()
}

// Criticalf writes to the "error" logger with printf-style formatting and logs the
// current stacktrace.
func (d *DefaultLogger) Criticalf(f string, v ...interface{}) {
	d.E.Printf(f, v...)
	d.logStacktraceToErr()
}

// logStacktraceToErr prints a stacktrace to the logger's error output.
// It skips 4 levels that aren't meaningful to a logged stacktrace:
// * debug.Stack()
// * core.captureStacktrace()
// * DefaultLogger::logStacktraceToErr()
// * DefaultLogger::Error() or DefaultLogger::Errorf()
func (d *DefaultLogger) logStacktraceToErr() {
	d.E.Println("stacktrace:\n" + strings.Join(captureStacktrace(4), "\n"))
}

func captureStacktrace(skip int) []string {
	stack := string(debug.Stack())
	lines := strings.Split(stack, "\n")
	linesToSkip := 2*skip + 1
	if linesToSkip > len(lines) {
		return lines
	}
	return lines[linesToSkip:]
}
