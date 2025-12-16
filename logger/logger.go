package logger

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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

type Logger struct {
	level       LogLevel
	mu          sync.Mutex
	showLine    bool
	useColor    bool
	timeFormat  string
	pvmShowLine bool // Separate control for PVM logs
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Context for step tracking (headerHash, [slot, epoch])
var (
	HeaderHash  string
	Slot        types.TimeSlot
	Epoch       types.TimeSlot
	hasContext  bool
	currentStep string
)

// =============================================================================
// Initialization
// =============================================================================

func Init(level LogLevel) error {
	var err error
	once.Do(func() {
		defaultLogger, err = newLogger(level)
	})
	return err
}

func newLogger(level LogLevel) (*Logger, error) {
	return &Logger{
		level:       level,
		showLine:    true,           // Enable logging by default
		useColor:    true,           // Enable colors by default
		timeFormat:  "15:04:05.000", // Short format: HH:MM:SS.ms, optional: MM-DD HH:MM:SS
		pvmShowLine: false,          // PVM logging disabled by default
	}, nil
}

func getLogger() *Logger {
	if defaultLogger == nil {
		_ = Init(DEBUG)
	}
	return defaultLogger
}

// =============================================================================
// Context Management (headerHash, slot, step)
// =============================================================================

// SetContext sets the current block context for logging
// Epoch is auto-calculated: epoch = slot / EpochLength
// Format: [headerHash[:8]|slot:epoch] or [headerHash[:8]|slot:epoch|step]
func SetContext(headerHash types.HeaderHash, slot types.TimeSlot) {
	HeaderHash = hex.EncodeToString(headerHash[:])
	Slot = slot
	Epoch = slot / types.TimeSlot(types.EpochLength)
	hasContext = true
}

// ClearContext clears the current block context
func ClearContext() {
	HeaderHash = ""
	Slot = 0
	Epoch = 0
	hasContext = false
	currentStep = ""
}

// SetStep sets the current processing step (e.g., "safrole", "disputes", "accumulate")
func SetStep(step string) {
	currentStep = step
}

// =============================================================================
// Core Logging Function
// =============================================================================

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	// FATAL always logs regardless of level or showLine flag
	if level != FATAL && (level < l.level || !l.showLine) {
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

	// Build context prefix: [hash|slot:epoch|step] or [hash|slot:epoch]
	contextPrefix := ""
	if hasContext {
		hashStr := HeaderHash
		if len(HeaderHash) > 8 {
			hashStr = HeaderHash[:8]
		}
		if currentStep != "" {
			contextPrefix = fmt.Sprintf("[%s|(%d:%d)|%s] ", hashStr, Slot, Epoch, currentStep)
		} else {
			contextPrefix = fmt.Sprintf("[%s|(%d:%d)] ", hashStr, Slot, Epoch)
		}
	} else if currentStep != "" {
		contextPrefix = fmt.Sprintf("[%s] ", currentStep)
	}

	// Format output with or without color
	levelStr := levelStrings[level]
	if l.useColor {
		color := levelColors[level]
		// Color the level tag
		coloredLevel := fmt.Sprintf("%s[%s]%s", color, levelStr, Reset)
		// Color context in cyan
		if contextPrefix != "" {
			contextPrefix = fmt.Sprintf("%s%s%s", Cyan, contextPrefix, Reset)
		}
		fmt.Printf("%s %s %s%s\n", timestamp, coloredLevel, contextPrefix, message)
	} else {
		fmt.Printf("%s [%s] %s%s\n", timestamp, levelStr, contextPrefix, message)
	}

	if level == FATAL {
		os.Exit(1)
	}
}

// =============================================================================
// Standard Log Functions (DEBUG, INFO, WARN)
// =============================================================================

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

// =============================================================================
// Error Functions
// =============================================================================

func Error(args ...interface{}) {
	getLogger().log(ERROR, strings.TrimSpace(fmt.Sprint(args...)))
}

func Errorf(format string, args ...interface{}) {
	getLogger().log(ERROR, format, args...)
}

// =============================================================================
// Protocol Error Functions
// Use for protocol-level errors (block invalid) - logs error but DOES NOT exit
// =============================================================================

func ProtocolError(args ...interface{}) {
	l := getLogger()
	msg := strings.TrimSpace(fmt.Sprint(args...))
	if l.useColor {
		l.log(ERROR, "%s[PROTOCOL]%s %s", Magenta, Reset, msg)
	} else {
		l.log(ERROR, "[PROTOCOL] "+msg)
	}
}

func ProtocolErrorf(format string, args ...interface{}) {
	l := getLogger()
	msg := fmt.Sprintf(format, args...)
	if l.useColor {
		l.log(ERROR, "%s[PROTOCOL]%s %s", Magenta, Reset, msg)
	} else {
		l.log(ERROR, "[PROTOCOL] "+msg)
	}
}

// ProtocolErrorWithCode logs a protocol error with its error code
func ProtocolErrorWithCode(errCode interface{}, message string) {
	l := getLogger()
	if l.useColor {
		l.log(ERROR, "%s[PROTOCOL]%s %v: %s", Magenta, Reset, errCode, message)
	} else {
		l.log(ERROR, fmt.Sprintf("[PROTOCOL] %v: %s", errCode, message))
	}
}

// =============================================================================
// Fatal Functions
// Use for unexpected runtime errors - logs error and EXITS the program
// =============================================================================

func Fatal(args ...interface{}) {
	getLogger().log(FATAL, strings.TrimSpace(fmt.Sprint(args...)))
}

func Fatalf(format string, args ...interface{}) {
	getLogger().log(FATAL, format, args...)
}

// =============================================================================
// Step Functions
// Use for tracking processing steps with (headerHash, slot) context
// =============================================================================

func StepStart(step string) {
	SetStep(step)
	getLogger().log(INFO, fmt.Sprintf("Starting %s", step))
}

func StepEnd(step string) {
	getLogger().log(INFO, fmt.Sprintf("Completed %s", step))
	SetStep("")
}

func StepProtocolError(step string, err error) {
	SetStep(step)
	ProtocolErrorf("%s failed: %v", step, err)
}

func StepFatal(step string, err error) {
	SetStep(step)
	getLogger().log(FATAL, fmt.Sprintf("%s fatal error: %v", step, err))
}

// =============================================================================
// PVM Logger Functions (Separate Control)
// =============================================================================

func PVMDebug(args ...interface{}) {
	l := getLogger()
	if !l.pvmShowLine {
		return
	}
	msg := strings.TrimSpace(fmt.Sprint(args...))
	if l.useColor {
		l.log(DEBUG, "%s[PVM]%s %s", Gray, Reset, msg)
	} else {
		l.log(DEBUG, "[PVM] "+msg)
	}
}

func PVMDebugf(format string, args ...interface{}) {
	l := getLogger()
	if !l.pvmShowLine {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if l.useColor {
		l.log(DEBUG, "%s[PVM]%s %s", Gray, Reset, msg)
	} else {
		l.log(DEBUG, "[PVM] "+msg)
	}
}

func PVMInfo(args ...interface{}) {
	l := getLogger()
	if !l.pvmShowLine {
		return
	}
	msg := strings.TrimSpace(fmt.Sprint(args...))
	if l.useColor {
		l.log(INFO, "%s[PVM]%s %s", Gray, Reset, msg)
	} else {
		l.log(INFO, "[PVM] "+msg)
	}
}

func PVMInfof(format string, args ...interface{}) {
	l := getLogger()
	if !l.pvmShowLine {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if l.useColor {
		l.log(INFO, "%s[PVM]%s %s", Gray, Reset, msg)
	} else {
		l.log(INFO, "[PVM] "+msg)
	}
}

func PVMError(args ...interface{}) {
	l := getLogger()
	if !l.pvmShowLine {
		return
	}
	msg := strings.TrimSpace(fmt.Sprint(args...))
	if l.useColor {
		l.log(ERROR, "%s[PVM]%s %s", Gray, Reset, msg)
	} else {
		l.log(ERROR, "[PVM] "+msg)
	}
}

func PVMErrorf(format string, args ...interface{}) {
	l := getLogger()
	if !l.pvmShowLine {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if l.useColor {
		l.log(ERROR, "%s[PVM]%s %s", Gray, Reset, msg)
	} else {
		l.log(ERROR, "[PVM] "+msg)
	}
}

// =============================================================================
// Configuration Functions
// =============================================================================

func SetLevel(levelStr string) {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()

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
	l.level = level
}

// Enable enables logging output
func Enable() {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showLine = true
}

// Disable disables logging output (except FATAL which always logs)
func Disable() {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showLine = false
}

// SetShowLine sets whether logging is enabled
func SetShowLine(show bool) {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showLine = show
}

// IsShowLine returns whether logging is currently enabled
func IsShowLine() bool {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.showLine
}

// EnableColor enables colored output
func EnableColor() {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.useColor = true
}

// DisableColor disables colored output
func DisableColor() {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.useColor = false
}

// SetColor sets whether to use colored output
func SetColor(enabled bool) {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.useColor = enabled
}

// SetTimeFormat sets the time format for log timestamps
// Common formats:
//   - "15:04:05"       - HH:MM:SS
//   - "15:04:05.000"   - HH:MM:SS.ms (default)
//   - "01-02 15:04:05" - MM-DD HH:MM:SS
//   - time.RFC3339    - Full ISO format
func SetTimeFormat(format string) {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timeFormat = format
}

// =============================================================================
// PVM Logger Configuration
// =============================================================================

// EnablePVM enables PVM logging
func EnablePVM() {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pvmShowLine = true
}

// DisablePVM disables PVM logging
func DisablePVM() {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pvmShowLine = false
}

// SetPVMShowLine sets whether PVM logging is enabled
func SetPVMShowLine(show bool) {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pvmShowLine = show
}

// IsPVMEnabled returns whether PVM logging is enabled
func IsPVMEnabled() bool {
	l := getLogger()
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pvmShowLine
}
