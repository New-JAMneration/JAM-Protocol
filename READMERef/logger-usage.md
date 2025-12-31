# Logger Usage Guide

This document describes how to use the logger package in JAM-Protocol.

## Table of Contents

- [Log Levels](#log-levels)
- [Basic Usage](#basic-usage)
- [Named Loggers](#named-loggers)
- [PVM Logger](#pvm-logger)
- [Config File](#config-file)
- [Forbidden Usage](#forbidden-usage)
- [Error Handling Strategy](#error-handling-strategy)
- [Quick Reference](#quick-reference)

---

## Log Levels

| Level | Function | Description |
|-------|----------|-------------|
| `DEBUG` | `logger.Debug()`, `logger.Debugf()` | Development debug messages |
| `INFO` | `logger.Info()`, `logger.Infof()` | General execution info |
| `WARN` | `logger.Warn()`, `logger.Warnf()` | Warning messages |
| `ERROR` | `logger.Error()`, `logger.Errorf()` | Error messages |
| `FATAL` | `logger.Fatal()`, `logger.Fatalf()` | Critical errors (program exits) |

---

## Basic Usage

### Standard Logging

```go
import "github.com/New-JAMneration/JAM-Protocol/logger"

// Debug: detailed execution info
logger.Debug("🚀 Store initialized")
logger.Debugf("Processing block %d", blockNum)

// Info: general execution info
logger.Info("Test passed")
logger.Infof("Total: %d, Passed: %d", total, passed)

// Warn: expected anomalies
logger.Warn("Config file not found")
logger.Warnf("Service %d not in state", serviceID)

// Error: issues that need investigation
logger.Errorf("Failed to index block: %v", err)

// Fatal: critical errors, program terminates
logger.Fatal(err)  // Calls os.Exit(1)
logger.Fatalf("Cannot start: %v", err)
```

### Creating Errors for Return

```go
import "fmt"

// Use fmt.Errorf to create errors for return
func ProcessData(data []byte) error {
    if len(data) == 0 {
        return fmt.Errorf("empty data")
    }
    
    result, err := decode(data)
    if err != nil {
        logger.Errorf("Decode failed: %v", err)  // Log for debugging
        return fmt.Errorf("decode error: %w", err)  // Return for handling
    }
    return nil
}
```

---

## Named Loggers

You can create named logger instances with independent configuration.

### Creating a Named Logger

```go
// Get or create a named logger
myLogger := logger.GetLogger("mymodule")

// Use it like the default logger
myLogger.Debug("Starting module")
myLogger.Infof("Processing %d items", count)
myLogger.Errorf("Error: %v", err)
```

### Output Format

Named loggers (except "main") include their name in the output:

```
15:07:32.373 [DEBUG] [MYMODULE] Starting module
15:07:32.374 [INFO] [MYMODULE] Processing 100 items
```

The main logger has no prefix:

```
15:07:32.373 [DEBUG] Starting application
```

---

## PVM Logger

PVM (Polka Virtual Machine) has its own named logger that can be independently enabled/disabled.

### Usage in PVM Code

```go
// PVM package already has pvmLogger defined in const.go:
// var pvmLogger = logger.GetLogger("pvm")

// Use directly in PVM code:
pvmLogger.Debug("Memory initialized")
pvmLogger.Debugf("[%d]: pc: %d, %s", instrCount, pc, opcode)
pvmLogger.Errorf("Decode error: %v", err)
```

### Output Format

```
15:07:32.373 [DEBUG] [PVM] Memory initialized
15:07:32.374 [DEBUG] [PVM] [123]: pc: 45, load_u32, r1 = 0x1234
```

---

## Config File

Configure logging in `config.json`:

```json
{
  "log": {
    "level": "DEBUG",
    "color": true,
    "enabled": true,
    "pvm": false
  }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `level` | `"DEBUG"` | Log level: DEBUG, INFO, WARN, ERROR (case-insensitive) |
| `color` | `true` | Enable colored output |
| `enabled` | `true` | Enable main logger |
| `pvm` | `false` | Enable PVM logger |

### Programmatic Configuration

```go
// Configure a named logger
logger.ConfigureLogger("pvm", logger.LoggerConfig{
    Level:   "DEBUG",
    Enabled: true,
    Color:   true,
})

// Or use individual methods
logger.SetLevel("DEBUG")
logger.SetColor(true)
logger.SetEnabled(true)
```

---

## Forbidden Usage

### Do NOT use standard `log` package

```go
// WRONG
log.Println("message")
log.Printf("msg: %v", x)
log.Fatal(err)

// CORRECT
logger.Info("message")
logger.Infof("msg: %v", x)
logger.Fatal(err)
```

### Do NOT use `fmt.Print` for logging

```go
// WRONG
fmt.Println("debug info")
fmt.Printf("error: %v\n", err)

// CORRECT
logger.Debug("debug info")
logger.Errorf("error: %v", err)
```

### Do NOT call `os.Exit` after `logger.Fatal`

```go
// WRONG - redundant os.Exit
logger.Fatal(err)
os.Exit(1)  // Never executes

// CORRECT
logger.Fatal(err)  // Already calls os.Exit(1)
```

---

## Error Handling Strategy

### Error Categories

| Category | Description | Action |
|----------|-------------|--------|
| **Protocol Error** | Block validation failure (expected) | Log + return error code |
| **Runtime Error** | Unexpected bug | Log + return error |
| **Fatal Error** | Unrecoverable | Log + exit |

### Pattern: Log + Return

```go
func ProcessData(data []byte) error {
    result, err := decode(data)
    if err != nil {
        logger.Errorf("Decode failed: %v", err)  // Log
        return fmt.Errorf("decode error: %w", err)  // Return
    }
    return nil
}
```

### Pattern: Protocol Error

```go
func ValidateBlock(block Block) *types.ErrorCode {
    if block.Slot < 0 {
        logger.Errorf("[PROTOCOL] invalid slot: %d", block.Slot)
        errCode := ErrorCodes.BadSlot
        return &errCode
    }
    return nil
}
```

### Pattern: Fatal Error

```go
func Initialize() {
    config, err := loadConfig()
    if err != nil {
        logger.Fatalf("Cannot load config: %v", err)
    }
}
```

---

## Quick Reference

### When to Use Which

| Scenario | Use |
|----------|-----|
| General message | `logger.Info()` |
| Detailed debug | `logger.Debug()` |
| Expected warning | `logger.Warn()` |
| Error (investigate) | `logger.Errorf()` + `return err` |
| Critical (terminate) | `logger.Fatalf()` |
| PVM trace | `pvmLogger.Debugf()` (in PVM package) |
| Create error | `fmt.Errorf()` |
| Protocol error | `logger.Errorf("[PROTOCOL] ...")` + `return &ErrorCode{}` |

### Logger Instance Methods

```go
l := logger.GetLogger("mymodule")

l.Debug(args...)
l.Debugf(format, args...)
l.Info(args...)
l.Infof(format, args...)
l.Warn(args...)
l.Warnf(format, args...)
l.Error(args...)
l.Errorf(format, args...)
l.Fatal(args...)
l.Fatalf(format, args...)

l.SetLevel("DEBUG")
l.SetEnabled(true)
l.SetColor(true)
l.Enable()
l.Disable()
l.IsEnabled()
```

### Package-Level Functions (Default Logger)

```go
logger.Debug(args...)
logger.Debugf(format, args...)
logger.Info(args...)
logger.Infof(format, args...)
logger.Warn(args...)
logger.Warnf(format, args...)
logger.Error(args...)
logger.Errorf(format, args...)
logger.Fatal(args...)
logger.Fatalf(format, args...)

logger.SetLevel("DEBUG")
logger.SetEnabled(true)
logger.SetColor(true)
```
