package PVM

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type LogLevel int

const (
	FATAL LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
)

var (
	PVMLogger *Logger
	once      sync.Once
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
		PVMLogger, err = newLogger(level)
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
	if PVMLogger == nil {
		_ = Init(DEBUG)
	}
	return PVMLogger
}

func (l *Logger) log(level LogLevel, coreID types.CoreIndex, serviceID types.ServiceId, format string, args ...interface{}) {
	if level > l.level {
		fmt.Println(l.level)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	levelStr := []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG"}
	timestamp := time.Now().Format(l.timeFormat)

	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	fmt.Printf("%s [%s][core:%d][service:%d] %s\n", timestamp, levelStr[level], coreID, serviceID, message)

	if level == FATAL {
		os.Exit(1)
	}
}

func Debug(coreID types.CoreIndex, serviceID types.ServiceId, args ...interface{}) {
	getLogger().log(DEBUG, coreID, serviceID, strings.TrimSpace(fmt.Sprint(args...)))
}

func Debugf(coreID types.CoreIndex, serviceID types.ServiceId, format string, args ...interface{}) {
	getLogger().log(DEBUG, coreID, serviceID, format, args...)
}

func Info(coreID types.CoreIndex, serviceID types.ServiceId, args ...interface{}) {
	getLogger().log(INFO, coreID, serviceID, strings.TrimSpace(fmt.Sprint(args...)))
}

func Infof(coreID types.CoreIndex, serviceID types.ServiceId, format string, args ...interface{}) {
	getLogger().log(INFO, coreID, serviceID, format, args...)
}

func Warn(coreID types.CoreIndex, serviceID types.ServiceId, args ...interface{}) {
	getLogger().log(WARN, coreID, serviceID, strings.TrimSpace(fmt.Sprint(args...)))
}

func Warnf(coreID types.CoreIndex, serviceID types.ServiceId, format string, args ...interface{}) {
	getLogger().log(WARN, coreID, serviceID, format, args...)
}

func Error(coreID types.CoreIndex, serviceID types.ServiceId, args ...interface{}) {
	getLogger().log(ERROR, coreID, serviceID, strings.TrimSpace(fmt.Sprint(args...)))
}

func Errorf(coreID types.CoreIndex, serviceID types.ServiceId, format string, args ...interface{}) {
	getLogger().log(ERROR, coreID, serviceID, format, args...)
}

func Fatal(coreID types.CoreIndex, serviceID types.ServiceId, args ...interface{}) {
	getLogger().log(FATAL, coreID, serviceID, strings.TrimSpace(fmt.Sprint(args...)))
}

func Fatalf(coreID types.CoreIndex, serviceID types.ServiceId, format string, args ...interface{}) {
	getLogger().log(FATAL, coreID, serviceID, format, args...)
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
