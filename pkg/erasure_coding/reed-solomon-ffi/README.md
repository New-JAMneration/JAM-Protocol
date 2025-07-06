# reed-solomon-ffi

This directory contains the Rust FFI library for erasure coding.
- https://docs.rs/reed-solomon-simd/3.0.1/reed_solomon_simd/

## Build instructions

Make sure you have Rust installed

Then build the dynamic library before using:

```bash
cargo build --release
```

## Generating the C header

The `reedsolomon.h` file is generated using [cbindgen](https://github.com/eqrion/cbindgen).
If you need to re-generate it (e.g., after changing FFI signatures), run:

```bash
cbindgen --crate reed-solomon-ffi --output reedsolomon.h

