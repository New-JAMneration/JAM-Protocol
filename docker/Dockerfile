# BUILD:
#   docker build \
#     --build-arg GP_VERSION="$(cat VERSION_GP)" \
#     --build-arg TARGET_VERSION="$(cat VERSION_TARGET)" \
#     --build-arg OUTPUT=new-jamneration-target \
#     -t new-jamneration-target:latest \
#     -f docker/Dockerfile .
# RUN: or use make fuzz-docker-build、scripts/run_fuzz_target_docker.sh
#   mkdir -p /tmp/jam_fuzz && chmod a+rwx /tmp/jam_fuzz
#   docker run --rm -u "$(id -u):$(id -g)" --cap-add IPC_LOCK \
#     -e JAM_FUZZ=1 \
#     -e JAM_FUZZ_SPEC=tiny \
#     -e JAM_FUZZ_DATA_PATH=/tmp/jam_fuzz \
#     -e JAM_FUZZ_SOCK_PATH=/tmp/jam_fuzz/fuzz.sock \
#     -e JAM_FUZZ_LOG_LEVEL=info \
#     -v /tmp/jam_fuzz:/tmp/jam_fuzz \
#     new-jamneration-target:latest

# First stage: use rust image to build the Rust library
FROM rust:latest AS rust-builder
WORKDIR /
COPY pkg/Rust-VRF /pkg/Rust-VRF
WORKDIR /pkg/Rust-VRF/vrf-func-ffi
RUN rustup target add x86_64-unknown-linux-musl && \
    cargo build --release --target x86_64-unknown-linux-musl

# Second stage: use golang image to build the main project
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache musl-dev build-base tar zip
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy the built Rust library from the previous stage
COPY --from=rust-builder /pkg/Rust-VRF/vrf-func-ffi/target/x86_64-unknown-linux-musl/release/libbandersnatch_vrfs_ffi.a /app/pkg/Rust-VRF/vrf-func-ffi/target/release/

ARG GP_VERSION=
ARG TARGET_VERSION=
ARG GOOS=linux
ARG GOARCH=amd64
ARG OUTPUT=new-jamneration-target

# Build the Go project with CGO enabled for static linking
RUN CGO_ENABLED=1 CC=gcc GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -ldflags "-s -w -linkmode external -extldflags '-static -O3' -X main.GP_VERSION=${GP_VERSION} -X main.TARGET_VERSION=${TARGET_VERSION}" \
    -o ./build/${OUTPUT} ./cmd/fuzz

# Package the binary into tar.gz and zip formats
RUN cd build && \
    tar -czf ${OUTPUT}-${GOOS}-${GOARCH}-${GP_VERSION}.tar.gz ${OUTPUT} && \
    zip ${OUTPUT}-${GOOS}-${GOARCH}-${GP_VERSION}.zip ${OUTPUT}

# Final stage: put the binary in a minimal image
FROM alpine:latest
WORKDIR /
COPY --from=go-builder /app/build /build
COPY --from=go-builder /app/VERSION_GP /VERSION_GP
COPY --from=go-builder /app/VERSION_TARGET /VERSION_TARGET

ENV JAM_FUZZ=1
ENV JAM_FUZZ_SPEC=tiny
ENV JAM_FUZZ_DATA_PATH=/tmp/jam_fuzz/
ENV JAM_FUZZ_SOCK_PATH=/tmp/jam_fuzz/fuzz.sock
ENV JAM_FUZZ_LOG_LEVEL=info

ENTRYPOINT ["/build/new-jamneration-target"]
