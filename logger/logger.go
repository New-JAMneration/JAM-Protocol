package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// ANSI Color Codes
// =============================================================================

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
	Bold    = "\033[1m"
)

// =============================================================================
// Log Levels
// =============================================================================

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Level colors mapping
var levelColors = map[LogLevel]string{
	DEBUG: Blue,
	INFO:  Green,
	WARN:  Yellow,
	ERROR: Red,
	FATAL: Red + Bold,
}

var levelStrings = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

// =============================================================================
// Logger Configuration
// =============================================================================

// Logger represents a named logger instance
type Logger struct {
	name       string
	level      LogLevel
	mu         sync.Mutex
	enabled    bool
	useColor   bool
	timeFormat string
}

// LoggerConfig holds configuration for a logger
type LoggerConfig struct {
	Level      string `json:"level"`
	Enabled    bool   `json:"enabled"`
	Color      bool   `json:"color"`
	TimeFormat string `json:"time_format"` // Optional: if empty, uses global default
}

// =============================================================================
// Logger Registry (manages named loggers)
// =============================================================================

var (
	loggers      = make(map[string]*Logger)
	loggersMutex sync.RWMutex
	globalConfig = struct {
		timeFormat string
	}{
		timeFormat: "15:04:05.000",
	}
)

// GetLogger returns a logger by name, creating it if it doesn't exist
// Usage:
//
//	logger := logger.GetLogger("pvm")
//	logger.Debug("message")
func GetLogger(name string) *Logger {
	loggersMutex.RLock()
	if l, exists := loggers[name]; exists {
		loggersMutex.RUnlock()
		return l
	}
	loggersMutex.RUnlock()

	// Create new logger with defaults
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	// Double-check after acquiring write lock
	if l, exists := loggers[name]; exists {
		return l
	}

	l := &Logger{
		name:       name,
		level:      DEBUG,
		enabled:    name == "main", // Only main logger enabled by default
		useColor:   true,
		timeFormat: globalConfig.timeFormat,
	}
	loggers[name] = l
	return l
}

// ConfigureLogger configures a named logger
func ConfigureLogger(name string, cfg LoggerConfig) {
	l := GetLogger(name)
	l.mu.Lock()
	defer l.mu.Unlock()

	l.enabled = cfg.Enabled
	l.useColor = cfg.Color
	l.level = parseLevel(cfg.Level)
}

// parseLevel converts string to LogLevel
func parseLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return DEBUG
	}
}

// =============================================================================
// Default Logger (backward compatibility)
// =============================================================================

func getDefaultLogger() *Logger {
	return GetLogger("main")
}

// Initialize default logger on package load
func init() {
	l := GetLogger("main")
	l.enabled = true
}

// =============================================================================
// Logger Methods
// =============================================================================

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	// FATAL always logs regardless of level or enabled flag
	if level != FATAL && (level < l.level || !l.enabled) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(l.timeFormat)

	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	// Build prefix with logger name (except for "main")
	prefix := ""
	if l.name != "main" {
		prefix = fmt.Sprintf("[%s] ", strings.ToUpper(l.name))
	}

	// Format output with or without color
	levelStr := levelStrings[level]
	if l.useColor {
		color := levelColors[level]
		coloredLevel := fmt.Sprintf("%s[%s]%s", color, levelStr, Reset)
		if prefix != "" {
			prefix = fmt.Sprintf("%s%s%s", Gray, prefix, Reset)
		}
		fmt.Printf("%s %s %s%s\n", timestamp, coloredLevel, prefix, message)
	} else {
		fmt.Printf("%s [%s] %s%s\n", timestamp, levelStr, prefix, message)
	}

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.log(DEBUG, strings.TrimSpace(fmt.Sprint(args...)))
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.log(INFO, strings.TrimSpace(fmt.Sprint(args...)))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.log(WARN, strings.TrimSpace(fmt.Sprint(args...)))
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.log(ERROR, strings.TrimSpace(fmt.Sprint(args...)))
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.log(FATAL, strings.TrimSpace(fmt.Sprint(args...)))
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// =============================================================================
// Logger Configuration Methods
// =============================================================================

// SetLevel sets the log level
func (l *Logger) SetLevel(levelStr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = parseLevel(levelStr)
}

// Enable enables the logger
func (l *Logger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = true
}

// Disable disables the logger
func (l *Logger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = false
}

// SetEnabled sets whether the logger is enabled
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// IsEnabled returns whether the logger is enabled
func (l *Logger) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

// SetColor sets whether to use colored output
func (l *Logger) SetColor(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.useColor = enabled
}

// =============================================================================
// Package-Level Functions (backward compatibility with logger.Debug(), etc.)
// =============================================================================

func Debug(args ...interface{}) {
	getDefaultLogger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	getDefaultLogger().Debugf(format, args...)
}

func Info(args ...interface{}) {
	getDefaultLogger().Info(args...)
}

func Infof(format string, args ...interface{}) {
	getDefaultLogger().Infof(format, args...)
}

func Warn(args ...interface{}) {
	getDefaultLogger().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	getDefaultLogger().Warnf(format, args...)
}

func Error(args ...interface{}) {
	getDefaultLogger().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	getDefaultLogger().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	getDefaultLogger().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	getDefaultLogger().Fatalf(format, args...)
}

// =============================================================================
// Package-Level Configuration Functions (backward compatibility)
// =============================================================================

// SetLevel sets the default logger's level
func SetLevel(levelStr string) {
	getDefaultLogger().SetLevel(levelStr)
}

// Enable enables the default logger
func Enable() {
	getDefaultLogger().Enable()
}

// Disable disables the default logger
func Disable() {
	getDefaultLogger().Disable()
}

// SetEnabled sets whether the default logger is enabled
func SetEnabled(enabled bool) {
	getDefaultLogger().SetEnabled(enabled)
}

// IsEnabled returns whether the default logger is enabled
func IsEnabled() bool {
	return getDefaultLogger().IsEnabled()
}

// SetColor sets whether the default logger uses colored output
func SetColor(enabled bool) {
	getDefaultLogger().SetColor(enabled)
}

// EnableColor enables colored output for the default logger
func EnableColor() {
	getDefaultLogger().SetColor(true)
}

// DisableColor disables colored output for the default logger
func DisableColor() {
	getDefaultLogger().SetColor(false)
}

// =============================================================================
// Context Formatting Helpers
// =============================================================================

// BlockContext holds block context information for logging
type BlockContext struct {
	HeaderHash string // First 8 chars of hex-encoded header hash
	Slot       uint32
	Epoch      uint32
	Method     string // Optional: current method/step name
}

// String returns the formatted context string: [hash|slot:epoch] or [hash|slot:epoch|method]
func (c BlockContext) String() string {
	hash := c.HeaderHash
	if len(hash) > 8 {
		hash = hash[:8]
	}
	if c.Method != "" {
		return fmt.Sprintf("[%s|(%d:%d)|%s]", hash, c.Slot, c.Epoch, c.Method)
	}
	return fmt.Sprintf("[%s|(%d:%d)]", hash, c.Slot, c.Epoch)
}

// FormatContext creates a formatted context prefix string
// Usage: logger.Debugf("%s Processing block", logger.FormatContext(hash, slot, epoch, ""))
func FormatContext(headerHash string, slot, epoch uint32, method string) string {
	hash := headerHash
	if len(hash) > 8 {
		hash = hash[:8]
	}
	if method != "" {
		return fmt.Sprintf("[%s|(%d:%d)|%s]", hash, slot, epoch, method)
	}
	return fmt.Sprintf("[%s|(%d:%d)]", hash, slot, epoch)
}

// FormatContextWithMethod creates a formatted context prefix with method name
// Usage: logger.Debugf("%s Processing", logger.FormatContextWithMethod(hash, slot, epoch, "ImportBlock"))
func FormatContextWithMethod(headerHash string, slot, epoch uint32, method string) string {
	return FormatContext(headerHash, slot, epoch, method)
}
