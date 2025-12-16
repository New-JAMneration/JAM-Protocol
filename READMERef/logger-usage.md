# Logger Usage Guide

This document describes how to use the logger package in JAM-Protocol.

## Table of Contents

- [Log Levels](#log-levels)
- [Allowed Usage](#allowed-usage)
- [Forbidden Usage](#forbidden-usage)
- [Environment Variables](#environment-variables)
- [Config File](#config-file)
- [PVM Logger](#pvm-logger)
- [Context Logging](#context-logging)
- [Error Handling Strategy](#error-handling-strategy)
- [Quick Reference](#quick-reference)

---

## Log Levels

| Level | Function | Description |
|-------|----------|-------------|
| `DEBUG` | `logger.Debug()`, `logger.Debugf()` | Development debug messages, detailed execution info |
| `INFO` | `logger.Info()`, `logger.Infof()` | General execution info, important state changes |
| `WARN` | `logger.Warn()`, `logger.Warnf()` | Warning messages, expected anomalies |
| `ERROR` | `logger.Error()`, `logger.Errorf()` | Error messages, issues that need investigation |
| `FATAL` | `logger.Fatal()`, `logger.Fatalf()` | Critical errors, program terminates (`os.Exit(1)`) |

---

## Allowed Usage

### ✅ Basic Logging

// Debug: detailed execution info
logger.Debug("🚀 Store initialized")
logger.Debugf("Processing block %d", blockNum)

// Info: general execution info
logger.Info("Test passed")
logger.Infof("Total: %d, Passed: %d", total, passed)

// Warn: expected anomalies
logger.Warn("Config file not found, using default configuration")
logger.Warnf("Service %d not in state", serviceID)

// Error: issues that need investigation (program continues)
logger.Errorf("Failed to index block: %v", err)

// Fatal: critical errors, program terminates
logger.Fatal(err)  // No need to call os.Exit(1) after this
logger.Fatalf("Cannot start: %v", err)
```

### ✅ Creating Errors for Return

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

## Forbidden Usage

### ❌ Do NOT use standard `log` package

```go
// ❌ WRONG
log.Println("message")
log.Printf("msg: %v", x)
log.Fatal(err)

// ✅ CORRECT
logger.Info("message")
logger.Infof("msg: %v", x)
logger.Fatal(err)
```

### ❌ Do NOT use `fmt.Print` for logging

```go
// ❌ WRONG
fmt.Println("debug info")
fmt.Printf("error: %v\n", err)

// ✅ CORRECT
logger.Debug("debug info")
logger.Errorf("error: %v", err)
```

### ❌ Do NOT call `os.Exit` after `logger.Fatal`

```go
// ❌ WRONG - redundant os.Exit
logger.Fatal(err)
os.Exit(1)  // This line will never execute

// ✅ CORRECT
logger.Fatal(err)  // Already calls os.Exit(1)
```

### ❌ Do NOT use `log` for PVM code

```go
// ❌ WRONG in PVM code
log.Printf("instruction: %s", opcode)

// ✅ CORRECT in PVM code
logger.PVMDebugf("instruction: %s", opcode)
```

---

## Environment Variables

Control logging behavior via environment variables (higher priority than config file):

### PVM_LOG

Enable/disable PVM logging:

```bash
# Enable PVM logging
PVM_LOG=true make test-jam-test-vectors mode=accumulate
PVM_LOG=1 make run-target

# Disable PVM logging (default)
PVM_LOG=false make test-jam-test-vectors mode=accumulate
```

### LOG_LEVEL

Set log level (case-insensitive):

```bash
# All of these work
LOG_LEVEL=DEBUG make test-jam-test-vectors mode=safrole
LOG_LEVEL=debug make test-jam-test-vectors mode=safrole
LOG_LEVEL=Debug make test-jam-test-vectors mode=safrole

# Available levels: DEBUG, INFO, WARN, ERROR
LOG_LEVEL=INFO make test-jam-test-vectors mode=safrole
```

### LOG_COLOR

Enable/disable colored output:

```bash
# Disable colors (useful for log files)
LOG_COLOR=false make test-jam-test-vectors mode=safrole

# Enable colors (default)
LOG_COLOR=true make test-jam-test-vectors mode=safrole
```

### Combined Usage

```bash
PVM_LOG=true LOG_LEVEL=DEBUG make test-jam-test-vectors mode=accumulate
```

---

## Config File

Configure logging in `config.json`:

```json
{
  "log": {
    "level": "DEBUG",
    "color": true,
    "show_line": true,
    "pvm": false
  }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `level` | `"DEBUG"` | Log level: DEBUG, INFO, WARN, ERROR |
| `color` | `true` | Enable colored output |
| `show_line` | `true` | Enable logging |
| `pvm` | `false` | Enable PVM logging |

### Priority

```
Environment Variable > config.json > Default Value
```

---

## PVM Logger

PVM-related logs use separate `logger.PVM*` functions, controlled by `PVM_LOG` or config.

### Usage

```go
// PVM-specific logger (controlled by PVM_LOG)
logger.PVMDebug("Memory initialized")
logger.PVMDebugf("[%d]: pc: %d, %s", instrCount, pc, opcode)
logger.PVMErrorf("Decode error: %v", err)

// Fatal is NOT controlled by PVM_LOG (always prints)
logger.Fatalf("Critical PVM error: %v", err)
```

### When to Use

| Type | Function |
|------|----------|
| Instruction execution trace | `logger.PVMDebugf()` |
| Host-call execution | `logger.PVMDebugf()` |
| Memory map info | `logger.PVMDebugf()` |
| Decode errors | `logger.PVMErrorf()` |
| Critical errors (terminate) | `logger.Fatalf()` |

### Example Output

When `PVM_LOG=true`:

```
15:07:32.373 [PVM] [DEBUG] Memory Map RO data: 0x00001000 0x00002000
15:07:32.374 [PVM] [DEBUG] [123]: pc: 45, load_u32, r1 = 0x1234
15:07:32.375 [PVM] [DEBUG] [124]: pc: 50, add32, r2 = r1 + r3
```

---

## Context Logging

Add block context (headerHash, slot, epoch) to log messages.

### Usage

```go
// Set context (epoch is auto-calculated: epoch = slot / EpochLength)
logger.SetContext(headerHash, slot)
defer logger.ClearContext()

// Optional: set processing step
logger.SetStep("safrole")

// All subsequent logs will include context
logger.Info("Processing block")
```

### Output Format

```
[headerHash[:8]|(slot:epoch)]        - without step
[headerHash[:8]|(slot:epoch)|step]   - with step
```

### Example

```go
func (s *Service) ImportBlock(block *types.Block) error {
    headerHash, _ := hash.ComputeBlockHeaderHash(block.Header)
    
    logger.SetContext(headerHash, block.Header.Slot)
    defer logger.ClearContext()
    
    logger.SetStep("initialize")
    logger.Debug("Starting block import")
    // Output: 15:07:32.373 [DEBUG] [a1b2c3d4|(123:10)|initialize] Starting block import
    
    logger.SetStep("safrole")
    logger.Debug("Processing safrole")
    // Output: 15:07:32.380 [DEBUG] [a1b2c3d4|(123:10)|safrole] Processing safrole
    
    return nil
}
```

---

## Error Handling Strategy

### Error Categories

```
┌────────────────────────────────────────────────────────────────┐
│  Protocol Error (types.ErrorCode)                              │
│  ─────────────────────────────────────────────────────────────  │
│  • Block validation failure (expected error)                   │
│  • Program continues, returns error code                       │
│  • Use: return &types.ErrorCode{}                              │
├────────────────────────────────────────────────────────────────┤
│  Runtime Error                                                 │
│  ─────────────────────────────────────────────────────────────  │
│  • Program bug, unexpected error                               │
│  • Needs investigation and fix                                 │
│  • Use: logger.Errorf() + return err                           │
├────────────────────────────────────────────────────────────────┤
│  Fatal Error                                                   │
│  ─────────────────────────────────────────────────────────────  │
│  • Unrecoverable critical error                                │
│  • Program terminates                                          │
│  • Use: logger.Fatalf()                                        │
└────────────────────────────────────────────────────────────────┘
```

### Patterns

#### Pattern 1: Log + Return (Recommended)

```go
func ProcessData(data []byte) error {
    result, err := decode(data)
    if err != nil {
        logger.Errorf("Decode failed: %v", err)  // Log for debugging
        return fmt.Errorf("decode error: %w", err)  // Return for handling
    }
    return nil
}
```

#### Pattern 2: Protocol Error (Block Validation)

```go
func ValidateBlock(block Block) *types.ErrorCode {
    if block.Slot < 0 {
        errCode := ErrorCodes.BadSlot
        return &errCode  // Expected error, no logging needed
    }
    return nil
}
```

#### Pattern 3: Fatal (Unrecoverable)

```go
func Initialize() {
    config, err := loadConfig()
    if err != nil {
        logger.Fatalf("Cannot load config: %v", err)  // Program terminates
    }
}
```

### STF Error Handling

```go
// stf.RunSTF() returns (isProtocol bool, err error)
//
// isProtocol=true,  err!=nil  → Protocol error (block invalid, continue)
// isProtocol=false, err!=nil  → Runtime error (bug, investigate)
// isProtocol=false, err==nil  → Success

isProtocol, err := stf.RunSTF()
if err != nil {
    if isProtocol {
        logger.ProtocolErrorf("Block invalid: %v", err)
        // Continue processing next block
    } else {
        logger.Errorf("Unexpected error: %v", err)
        // Runtime error, needs investigation
    }
}
```

---

## Quick Reference

| Scenario | Use |
|----------|-----|
| General message | `logger.Info()` |
| Detailed debug | `logger.Debug()` |
| Expected warning | `logger.Warn()` |
| Error (investigate) | `logger.Errorf()` + `return err` |
| Critical error (terminate) | `logger.Fatalf()` |
| PVM trace | `logger.PVMDebugf()` |
| Create error | `fmt.Errorf()` |
| Protocol error | `return &types.ErrorCode{}` |

### Programmatic Configuration

```go
// Set log level
logger.SetLevel("DEBUG")  // DEBUG, INFO, WARN, ERROR

// Set color
logger.SetColor(true)
logger.SetColor(false)

// Set PVM logging
logger.SetPVMShowLine(true)
logger.SetPVMShowLine(false)

// Set context
logger.SetContext(headerHash, slot)
logger.ClearContext()
logger.SetStep("safrole")
```
