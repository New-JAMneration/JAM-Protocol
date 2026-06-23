# Rust-VRF FFI Usage

This document describes the safe Go-side usage pattern for
`pkg/Rust-VRF/vrf-func-ffi`, which owns Rust heap objects through CGO pointers.

## Handler Lifecycle

`vrf.NewHandler`, `vrf.NewProver`, and `vrf.NewVerifier` allocate Rust heap
objects. Always release them when the owner is done:

```go
handler, err := vrf.NewHandler(ring, secretKey, ringSize, proverIndex)
if err != nil {
	return err
}
defer handler.Free()
```

`Free()` is idempotent on the Go wrapper, but callers should still treat the
object as closed after calling it. Do not call sign/verify methods after `Free()`.

## Ownership Rules

- Each `Handler`, `Prover`, or `Verifier` has one logical Go owner.
- Prefer `defer h.Free()` immediately after successful construction.
- Do not copy wrappers by value. Pass pointers.
- Do not call `Free()` while another goroutine may be using the same wrapper.
- Do not retain raw `unsafe.Pointer` values from the wrappers.

The wrapper has a finalizer as a last-resort safety net, but finalizers are not a
replacement for explicit `Free()`. Finalizers run at an unspecified time and may
not run before process exit.

## FFI Input Slices

The Go wrapper converts `[]byte` to a C-compatible `GoSlice` and calls into Rust.
The wrapper must keep all input and output backing arrays alive until the CGO call
has returned. When adding a new FFI call, follow this pattern:

```go
inputSlice := handleByteToGoSlice(input)
output := make([]byte, 32)
var outputLen C.size_t

result := C.some_vrf_function(
	inputSlice,
	(*C.uint8_t)(unsafe.Pointer(&output[0])),
	C.size_t(len(output)), // output_cap: buffer capacity
	&outputLen,
)
runtime.KeepAlive(input)
runtime.KeepAlive(output)
runtime.KeepAlive(wrapper)
```

Keep the receiver (`wrapper`) alive as well if the C call uses `wrapper.ptr`.
This prevents the finalizer from freeing the Rust object while C/Rust still uses
the pointer.

## Output Buffer Capacity

Every FFI function that writes into a Go-allocated buffer takes an explicit
`output_cap` (capacity in bytes) parameter immediately before the
`output_len` out-pointer. On the Rust side, `write_output`:

- returns `InvalidInput` if the output pointer or length pointer is null, and
- returns `SerializationFailed` if the produced bytes do not fit in `output_cap`.

This makes a too-small (or null) Go buffer a clean error instead of an
out-of-bounds write across the FFI boundary. Always pass `C.size_t(len(buf))`
for the capacity, never a hard-coded constant.

## Error Handling

Rust FFI functions return integer error codes instead of panicking across the FFI
boundary. Go code should map those codes to package errors such as
`ErrInvalidInput`, `ErrVerificationFailed`, or `ErrNullPointer`.

Every `extern "C"` entry point runs its body inside `ffi_guard`, which wraps the
work in `std::panic::catch_unwind`. A panic inside ark-vrf (or any callee) is
caught and surfaced as `FfiError::Internal` rather than unwinding across the C
ABI (which is undefined behaviour). Constructors that return a pointer return a
null pointer on panic.

When modifying Rust FFI code:

- Do not use `.unwrap()` or `.expect()` in production paths.
- Use `ok_or`, `map_err`, and explicit `match` blocks to return `VrfError` or
  `FfiError`.
- Wrap every `extern "C"` body in `ffi_guard` (pointer-returning constructors use
  `catch_unwind` directly and return null on panic).
- Write outputs through `write_output` so null pointers and undersized buffers are
  validated.
- Surface real `VrfError` variants from verify/output helpers (do not collapse
  errors to `()`), so the FFI layer can return precise error codes.

## Secret Material

`Prover` constructors take a copy of the raw secret-key bytes only to derive the
scalar; that owned `Vec<u8>` is wiped with `zeroize` immediately afterwards. The
derived `Secret`/scalar is owned by ark-vrf and cannot be zeroized here without
changing the dependency. Callers should still avoid keeping long-lived copies of
raw secret bytes on the Go side.

## Batch Verification

`RingVerifyBatch` uses the Rust native ark-vrf batch verifier through
`vrf_ring_verify_batch`. It is the preferred path when verifying multiple ring
VRF signatures with the same verifier.

The C prototypes may appear in more than one cgo file because each Go file with
`import "C"` has an independent preamble. Only `verifier.go` calls
`C.vrf_ring_verify_batch`; duplicate prototypes do not add another batch
implementation.

When changing an FFI signature, update the prototype in **all three** cgo files
(`prover.go`, `verifier.go`, `vrf.go`) and every call site, otherwise Go and Rust
disagree on the argument list and the process will segfault at the call.

## Rebuilding the Native Library

The Go package links against
`pkg/Rust-VRF/vrf-func-ffi/target/release/libbandersnatch_vrfs_ffi.so`
(via the cgo `LDFLAGS` `-L${SRCDIR}/../target/release`). After changing Rust
code, rebuild into that exact directory:

```bash
cargo build --release --manifest-path pkg/Rust-VRF/vrf-func-ffi/Cargo.toml
```

If a `CARGO_TARGET_DIR` override points elsewhere, the `.so` Go links against will
be stale and the ABI will mismatch the Go call sites. Force the in-tree target dir
when in doubt:

```bash
CARGO_TARGET_DIR=pkg/Rust-VRF/vrf-func-ffi/target \
  cargo build --release --manifest-path pkg/Rust-VRF/vrf-func-ffi/Cargo.toml
```
