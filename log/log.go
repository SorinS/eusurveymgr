package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

type LogLevel int

// Logger interface
type Logger struct {
	level      LogLevel
	writer     io.Writer
	timeFunc   func() time.Time
	levelNames []string
}

const (
	Debug   LogLevel = 0
	Info    LogLevel = 1
	Warning LogLevel = 2
	Error   LogLevel = 3
)

// Default log level
var logLevel LogLevel = Info
var logLevelNames = []string{"DEBUG", "INFO", "WARNING", "ERROR"}
var logger Logger

func init() {
	logger = Logger{
		level:      Info,
		writer:     os.Stdout,
		timeFunc:   time.Now,
		levelNames: logLevelNames,
	}
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel, writer io.Writer) *Logger {
	return &Logger{
		level:      level,
		writer:     writer,
		timeFunc:   time.Now,
		levelNames: logLevelNames,
	}
}

// Log functions for different levels

func (l *Logger) Debugf(format string, args ...any) {
	l.log(Debug, format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(Info, format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(Warning, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.log(Error, format, args...)
}

func Debugf(format string, args ...any) {
	logger.log(Debug, format, args...)
}

func Infof(format string, args ...any) {
	logger.log(Info, format, args...)
}

func Warnf(format string, args ...any) {
	logger.log(Warning, format, args...)
}

func Errorf(format string, args ...any) {
	logger.log(Error, format, args...)
}

// log is the internal function to log messages
func (l *Logger) log(level LogLevel, format string, args ...any) {
	if level >= l.level {
		stamp := l.timeFunc().Format(time.RFC3339Nano)
		fmt.Fprintf(l.writer, "["+stamp+"] "+l.levelNames[level]+": "+format+"\n", args...)
	}
}

func (l *Logger) SetLogWriter(writer io.Writer) {
	l.writer = writer
}

func (l *Logger) SetLogLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) GetLogLevel() LogLevel {
	return l.level
}

func SetLogLevel(level LogLevel) {
	logger.level = level
}