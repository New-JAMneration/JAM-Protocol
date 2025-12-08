#!/bin/bash
# This script builds the new-jamneration-target binary for linux
# Only for local builds without Docker when you want to test changes quickly
# You have to have golang, rust, cargo, and musl-gcc installed locally
# Usage: ./build.sh [GP_VERSION] [TARGET_VERSION]

GP_VERSION=${1:-"0.7.0"}
TARGET_VERSION=${2:-"0.1.0"}
OUTPUT_DIR="./build"
BINARY_NAME="new-jamneration-target"
GOOS=linux
GOARCH=amd64

# Build the Rust-VRF static library for Linux AMD64 using musl
cd pkg/Rust-VRF/vrf-func-ffi
cargo clean
rustup target add x86_64-unknown-linux-musl
cargo build --release --target x86_64-unknown-linux-musl
# Copy the static library to the target directory
mkdir -p target/release
cp target/x86_64-unknown-linux-musl/release/libbandersnatch_vrfs_ffi.a \
  target/release/

# Build the Go project with static linking
cd ../../..
CGO_ENABLED=1 CC=musl-gcc GOOS=${GOOS} GOARCH=${GOARCH} \
  go build -ldflags "-s -w -linkmode external -extldflags '-static -O3' -X main.GP_VERSION=${GP_VERSION} -X main.TARGET_VERSION=${TARGET_VERSION}" \
  -o $OUTPUT_DIR/$BINARY_NAME ./cmd/fuzz

# Package the binary into a tar.gz file
tar -czvf $OUTPUT_DIR/${BINARY_NAME}-${GOOS}-${GOARCH}-${GP_VERSION}.tar.gz -C $OUTPUT_DIR $BINARY_NAME

# Package the binary into a zip file
zip -j $OUTPUT_DIR/${BINARY_NAME}-${GOOS}-${GOARCH}-${GP_VERSION}.zip $OUTPUT_DIR/${BINARY_NAME}
