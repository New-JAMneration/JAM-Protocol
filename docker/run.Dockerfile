# This Docker image is for running the new-jamneration-target binary directly.
# You can build and run the Docker image with the following commands:
# docker build --build-arg GP_VERSION=0.7.0 --build-arg TARGET_VERSION=0.1.0 -t run-new-jamneration-target -f docker/run.Dockerfile .
# docker run --rm -u $(id -u):$(id -g) -v /tmp:/tmp new-jamneration-target /tmp/jam_target.sock

# First stage: use rust image to build the Rust library
FROM rust:latest AS rust-builder
WORKDIR /app
COPY pkg/Rust-VRF /app/pkg/Rust-VRF

WORKDIR /app/pkg/Rust-VRF/vrf-func-ffi
RUN cargo build --release --quiet

# Second stage: use golang image to build the main project
FROM golang:latest AS go-builder
WORKDIR /app
COPY . /app

ARG GP_VERSION=latest
ARG TARGET_VERSION=latest

# Copy the built Rust library from the previous stage
COPY --from=rust-builder /app/pkg/Rust-VRF/vrf-func-ffi /app/pkg/Rust-VRF/vrf-func-ffi

RUN go mod tidy && \
    go build -ldflags="-s -w -X main.GP_VERSION=${VERSION_GP} -X main.TARGET_VERSION=${VERSION_TARGET}" -o new-jamneration-target ./cmd/fuzz/

# Final stage: use a minimal image to run the application
FROM debian:stable-slim
WORKDIR /

# Copy the binary from the Go build stage
COPY --from=go-builder /app/new-jamneration-target /new-jamneration-target

# Copy the Rust library from the Go build stage
COPY --from=rust-builder /app/pkg/Rust-VRF/vrf-func-ffi/target/release /vrf-func-ffi/target/release

ENV USE_MINI_REDIS=true

ENTRYPOINT ["/new-jamneration-target"]
