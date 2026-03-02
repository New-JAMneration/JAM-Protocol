# Folder Structure

> **Note:** This layout can change as we go; we try to keep the doc in sync when we touch the code.

This document describes the project folder structure.

---

## Project Layout

```
JAM-Protocol/
├── README.md              # Project entry point
├── Makefile               # Build and test commands
├── go.mod                 # Go module definition
├── VERSION_GP             # Gray Paper version
├── VERSION_TARGET         # Target version
│
├── cmd/                   # Application entry points
│   ├── fuzz/              # Fuzzing tool
│   └── node/              # Node application
│
├── config/                # Configuration handling
│
├── docker/                # Docker configuration
│   └── Dockerfile
│
├── internal/              # Internal packages (not for external use)
│   ├── blockchain/        # Blockchain core logic
│   ├── service_account/   # Service account handling
│   ├── types/             # Type definitions
│   └── ...                # Other internal modules
│
├── jamtests/              # JAM test implementations
│   ├── accumulate/
│   ├── assurances/
│   ├── authorizations/
│   ├── disputes/
│   ├── history/
│   ├── preimages/
│   ├── reports/
│   ├── safrole/
│   ├── statistics/
│   └── trace/
│
├── logger/                # Global logging utilities
│
├── pkg/                   # Reusable packages (for external use)
│   ├── codecs/            # Encoding/decoding (SCALE codec)
│   │   └── scale/
│   ├── erasure_coding/    # Erasure coding implementation
│   │   └── reed-solomon-ffi/
│   └── test_data/
│
├── PVM/                   # Polkavm implementation
│   ├── host_call_*.go     # Host call implementations
│   ├── instructions.go    # PVM instructions
│   ├── memory.go          # Memory management
│   └── ...
│
├── READMERef/             # Project documentation
│   ├── INDEX.md           # Documentation index
│   ├── DOCUMENTATION_POLICY.md
│   └── ...                # Other documentation files
│
├── scripts/               # Build and release scripts
│
└── testdata/              # Test data and runners
    ├── jam_test_vector/
    ├── jam_testnet/
    └── traces/
```

---

## Directory Descriptions

| Directory | Description |
|-----------|-------------|
| `cmd/` | Application entry points. Each subdirectory is a separate executable. |
| `config/` | Configuration handling and loading. |
| `docker/` | Docker-related files for containerization. |
| `internal/` | Internal packages. These are not exposed to external projects. |
| `jamtests/` | Test implementations for JAM protocol components. |
| `logger/` | Global logging utilities. See [LOGGER_USAGE.md](./LOGGER_USAGE.md) for usage. |
| `pkg/` | Reusable packages that can be imported by external projects. |
| `PVM/` | Polkavm (PVM) implementation including host calls and instructions. |
| `READMERef/` | Project documentation. See [INDEX.md](./INDEX.md) for full list. |
| `scripts/` | Build, release, and utility scripts. |
| `testdata/` | Test data, vectors, and test runners. |

---

## Module READMEs

For module-specific documentation, see:

- [cmd/fuzz/README.md](../cmd/fuzz/README.md) - Fuzzing tool usage
- [pkg/erasure_coding/reed-solomon-ffi/README.md](../pkg/erasure_coding/reed-solomon-ffi/README.md) - Reed-Solomon FFI
- [pkg/test_data/README.md](../pkg/test_data/README.md) - Test data information
