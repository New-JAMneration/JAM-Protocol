package logger

import (
	"io"
	"log"
	"os"
)

// Note: Color constants are now defined in logger.go
// This file provides backward compatibility for ColorLogger usage

// ColorLogger wraps the standard log.Logger and provides colored output functionality
type ColorLogger struct {
	*log.Logger
}

// NewColorLogger creates a new ColorLogger instance
func NewColorLogger(prefix string, flag int) *ColorLogger {
	return &ColorLogger{
		Logger: log.New(os.Stdout, prefix, flag),
	}
}

// NewColorLoggerToWriter creates a ColorLogger that outputs to the specified Writer
func NewColorLoggerToWriter(w io.Writer, prefix string, flag int) *ColorLogger {
	return &ColorLogger{
		Logger: log.New(w, prefix, flag),
	}
}

// colorize adds color codes to text
func (cl *ColorLogger) colorize(color, text string) string {
	return color + text + Reset
}

// Red outputs text in red color
func (cl *ColorLogger) Red(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Red, format)
	cl.Printf(coloredFormat, v...)
}

// Green outputs text in green color
func (cl *ColorLogger) Green(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Green, format)
	cl.Printf(coloredFormat, v...)
}

// Yellow outputs text in yellow color
func (cl *ColorLogger) Yellow(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Yellow, format)
	cl.Printf(coloredFormat, v...)
}

// Blue outputs text in blue color
func (cl *ColorLogger) Blue(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Blue, format)
	cl.Printf(coloredFormat, v...)
}

// Magenta outputs text in magenta color
func (cl *ColorLogger) Magenta(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Magenta, format)
	cl.Printf(coloredFormat, v...)
}

// Cyan outputs text in cyan color
func (cl *ColorLogger) Cyan(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Cyan, format)
	cl.Printf(coloredFormat, v...)
}

// Gray outputs text in gray color
func (cl *ColorLogger) Gray(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Gray, format)
	cl.Printf(coloredFormat, v...)
}

// White outputs text in white color
func (cl *ColorLogger) White(format string, v ...interface{}) {
	coloredFormat := cl.colorize(White, format)
	cl.Printf(coloredFormat, v...)
}

// Bold outputs text in bold format
func (cl *ColorLogger) Bold(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Bold, format)
	cl.Printf(coloredFormat, v...)
}

// Error outputs error level message in red color
func (cl *ColorLogger) Error(format string, v ...interface{}) {
	cl.Red("[ERROR] "+format, v...)
}

// Warning outputs warning level message in yellow color
func (cl *ColorLogger) Warning(format string, v ...interface{}) {
	cl.Yellow("[WARN] "+format, v...)
}

// Info outputs info level message in green color
func (cl *ColorLogger) Info(format string, v ...interface{}) {
	cl.Green("[INFO] "+format, v...)
}

// Debug outputs debug level message in blue color
func (cl *ColorLogger) Debug(format string, v ...interface{}) {
	cl.Blue("[DEBUG] "+format, v...)
}

// Success outputs success message in green bold color
func (cl *ColorLogger) Success(format string, v ...interface{}) {
	coloredFormat := cl.colorize(Green+Bold, "[SUCCESS] "+format)
	cl.Printf(coloredFormat, v...)
}

// Global ColorLogger instance
var (
	DefaultColorLogger = NewColorLogger("", log.LstdFlags)
)

// Global functions that use the default ColorLogger
func ColorRed(format string, v ...interface{}) {
	DefaultColorLogger.Red(format, v...)
}

func ColorGreen(format string, v ...interface{}) {
	DefaultColorLogger.Green(format, v...)
}

func ColorYellow(format string, v ...interface{}) {
	DefaultColorLogger.Yellow(format, v...)
}

func ColorBlue(format string, v ...interface{}) {
	DefaultColorLogger.Blue(format, v...)
}

func ColorError(format string, v ...interface{}) {
	DefaultColorLogger.Error(format, v...)
}

func ColorWarning(format string, v ...interface{}) {
	DefaultColorLogger.Warning(format, v...)
}

func ColorInfo(format string, v ...interface{}) {
	DefaultColorLogger.Info(format, v...)
}

func ColorDebug(format string, v ...interface{}) {
	DefaultColorLogger.Debug(format, v...)
}

func ColorSuccess(format string, v ...interface{}) {
	DefaultColorLogger.Success(format, v...)
}
