# You can run the built binary in a Docker container with the following commands:
# docker build --build-arg GP_VERSION=0.7.0 --build-arg TARGET_VERSION=0.1.0 --build-arg OUTPUT=new-jamneration-target -t new-jamneration-target -f docker/Dockerfile .
# docker run --rm -u $(id -u):$(id -g) -v /tmp:/tmp new-jamneration-target /tmp/jam_target.sock

# First stage: use rust image to build the Rust library
FROM rust:latest AS rust-builder
WORKDIR /
COPY pkg/Rust-VRF /pkg/Rust-VRF
WORKDIR /pkg/Rust-VRF/vrf-func-ffi
RUN cargo clean && \
    rustup target add x86_64-unknown-linux-musl && \
    cargo build --release --target x86_64-unknown-linux-musl

# Second stage: use golang image to build the main project
FROM golang:1.23-alpine AS go-builder
WORKDIR /app
COPY . .
# Copy the built Rust library from the previous stage
COPY --from=rust-builder /pkg/Rust-VRF/vrf-func-ffi/target/x86_64-unknown-linux-musl/release/libbandersnatch_vrfs_ffi.a /app/pkg/Rust-VRF/vrf-func-ffi/target/release/

RUN apk add --no-cache musl-dev build-base tar zip

ARG GP_VERSION=latest
ARG TARGET_VERSION=latest
ARG GOOS=linux
ARG GOARCH=amd64
ARG OUTPUT=new-jamneration-target

# Build the Go project with CGO enabled for static linking
RUN go mod tidy
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

ENV USE_MINI_REDIS=true

ENTRYPOINT ["/build/new-jamneration-target"]
