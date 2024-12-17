package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	defaultLogger *Logger
	once          sync.Once
)

type Logger struct {
	level      LogLevel
	mu         sync.Mutex
	showLine   bool
	timeFormat string
}

func Init(level LogLevel) error {
	var err error
	once.Do(func() {
		defaultLogger, err = newLogger(level)
	})
	return err
}

func newLogger(level LogLevel) (*Logger, error) {
	return &Logger{
		level:      level,
		showLine:   true,
		timeFormat: time.RFC3339,
	}, nil
}

func getLogger() *Logger {
	if defaultLogger == nil {
		_ = Init(DEBUG)
	}
	return defaultLogger
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	timestamp := time.Now().Format(l.timeFormat)

	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	fmt.Printf("%s [%s] %s\n", timestamp, levelStr[level], message)

	if level == FATAL {
		os.Exit(1)
	}
}

func Debug(args ...interface{}) {
	getLogger().log(DEBUG, strings.TrimSpace(fmt.Sprint(args...)))
}

func Debugf(format string, args ...interface{}) {
	getLogger().log(DEBUG, format, args...)
}

func Info(args ...interface{}) {
	getLogger().log(INFO, strings.TrimSpace(fmt.Sprint(args...)))
}

func Infof(format string, args ...interface{}) {
	getLogger().log(INFO, format, args...)
}

func Warn(args ...interface{}) {
	getLogger().log(WARN, strings.TrimSpace(fmt.Sprint(args...)))
}

func Warnf(format string, args ...interface{}) {
	getLogger().log(WARN, format, args...)
}

func Error(args ...interface{}) {
	getLogger().log(ERROR, strings.TrimSpace(fmt.Sprint(args...)))
}

func Errorf(format string, args ...interface{}) {
	getLogger().log(ERROR, format, args...)
}

func Fatal(args ...interface{}) {
	getLogger().log(FATAL, strings.TrimSpace(fmt.Sprint(args...)))
}

func Fatalf(format string, args ...interface{}) {
	getLogger().log(FATAL, format, args...)
}

func SetLevel(levelStr string) {
	getLogger().mu.Lock()
	defer getLogger().mu.Unlock()

	var level LogLevel
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	case "FATAL":
		level = FATAL
	default:
		level = INFO
	}

	getLogger().level = level
}

func SetShowLine(show bool) {
	getLogger().mu.Lock()
	defer getLogger().mu.Unlock()
	getLogger().showLine = show
}

func SetTimeFormat(format string) {
	getLogger().mu.Lock()
	defer getLogger().mu.Unlock()
	getLogger().timeFormat = format
}
