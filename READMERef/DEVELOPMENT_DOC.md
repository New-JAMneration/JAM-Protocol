# Development Guide

## 1. Environment

go 1.23.3

## 2. Available Third-Party Libraries

### 2.1 Cryptographic Primitives

- Hash:
  - [Blake2b256 and keccak256 from gossamer](https://github.com/ChainSafe/gossamer/blob/development/lib/common/hasher.go)
  - [golang crypto blake2b 256-bit](https://pkg.go.dev/golang.org/x/crypto/blake2b#New256)
  - [golang crypto sha3 keccak 256-bit](https://pkg.go.dev/golang.org/x/crypto/sha3#NewLegacyKeccak256) (use `NewLegacyKeccak256` to get a LegacyKeccak256 hash object)
  - You can verify Blake2b hash output at [Polkadot.js](https://polkadot.js.org/) → Develop → Utility → hash

- Erasure Coding:
  - [klauspost/reedsolomon](https://github.com/klauspost/reedsolomon) — efficient Reed-Solomon encoding for data recovery and error correction.

- Bandersnatch:
  - No mainstream Go implementation exists yet. (ref: https://github.com/Consensys/gnark-crypto/tree/master/ecc/bls12-381)

- Ed25519:
  - The Go standard library `crypto/ed25519` package provides key generation, signing, and verification.

- RingVRF:
  - [VRF and RingVRF Spec](https://github.com/davxy/bandersnatch-vrfs-spec)
  - [Davxy reference implementation (for testing)](https://github.com/davxy/ark-ec-vrfs)
  - [RingVRF proof by w3f](https://github.com/w3f/ring-proof)

### 2.2 Codec

- **SCALE (Simple Concatenated Aggregate Little-Endian):**
  - A binary encoding format designed by Parity Technologies for the Substrate framework. The [ChainSafe/gossamer](https://github.com/ChainSafe/gossamer) project provides a Go SCALE codec implementation.

### 2.3 Network Protocol

- **QUIC (Quick UDP Internet Connections):**
  - [quic-go](https://github.com/quic-go/quic-go) — a pure-Go QUIC implementation supporting HTTP/3 and several related RFCs.

#### JAMNP Stream Implementation Conventions

In the JAM Network Protocol (JAMNP), `FIN` means **clean termination of the QUIC stream's send half** (the QUIC FIN bit). It is **not** a message byte.

| Operation | Correct approach |
|---|---|
| Send FIN | `stream.Close()` |
| Detect remote FIN | `expectRemoteFIN(stream)` — reads 1 byte; `io.EOF` = OK, any data = error |
| Send a message | `stream.WriteMessage(payload)` — prepends a 4-byte LE size prefix per JAMNP spec |
| Read a message | `stream.ReadMessage()` — parses the 4-byte LE size prefix and returns the payload |

## 3. Milestones

### M0

- [Reference: davxy M1 conference](https://github.com/w3f/jamtestvectors/issues/21)

### M1 — Prove that the block stream is correctly converted to state

- [JAM Network Protocol](https://github.com/zdave-parity/jam-np/blob/main/simple.md)

### M2 — Network test environment (test sandbox); must demonstrate both correct state and correct outgoing messages
