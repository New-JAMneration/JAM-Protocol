# BUILD (multi-arch via Docker BuildKit / buildx):
#   docker buildx build --platform=linux/amd64 \
#     --build-arg GP_VERSION="$(cat VERSION_GP)" \
#     --build-arg TARGET_VERSION="$(cat VERSION_TARGET)" \
#     --build-arg OUTPUT=new-jamneration-target \
#     -t new-jamneration-target:latest \
#     -f docker/Dockerfile --load .
#
# RUN: JAM_FUZZ_* env vars are baked into the image (see Stage 3 ENV below).
#   Override any via `-e` at runtime; `-v` mounts the fuzz data directory.
#   mkdir -p /tmp/jam_fuzz && chmod a+rwx /tmp/jam_fuzz
#   docker run --rm -u "$(id -u):$(id -g)" --cap-add IPC_LOCK \
#     -v /tmp/jam_fuzz:/tmp/jam_fuzz \
#     new-jamneration-target:latest
#
# Or use: make fuzz-docker-build / scripts/run_fuzz_target_docker.sh
#
# Notes on cross-compilation:
# - This Dockerfile relies on Docker BuildKit / buildx automatic platform ARGs
#   (BUILDPLATFORM / TARGETPLATFORM / TARGETOS / TARGETARCH).
# - Both rust-builder and go-builder pin themselves to $BUILDPLATFORM (the
#   builder's native arch) and use real cross compilers, avoiding QEMU emulation
#   so builds stay fast on Apple Silicon while still producing linux/amd64
#   binaries (or any other target the user picks).

# tonistiigi/xx provides the xx-* helper scripts (xx-apk / xx-go / xx-cc /
# xx-verify) used to do CGO cross-compilation cleanly on Alpine.
FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.5.0 AS xx

# ---- Stage 1: build the Rust static lib for the *target* platform ----
# rust:1-alpine tracks the latest 1.x stable on Alpine (musl) for security
# patches; pin a more specific tag (e.g. rust:1.86-alpine3.21) if you need
# fully reproducible builds. bandersnatch-vrfs-ffi is pure Rust producing a
# staticlib, so `rustup target add` alone is enough — no C cross toolchain.
FROM --platform=$BUILDPLATFORM rust:latest AS rust-builder
ARG TARGETARCH
WORKDIR /
COPY pkg/Rust-VRF /pkg/Rust-VRF
WORKDIR /pkg/Rust-VRF/vrf-func-ffi
RUN case "$TARGETARCH" in \
        amd64) RUST_TARGET=x86_64-unknown-linux-musl ;; \
        arm64) RUST_TARGET=aarch64-unknown-linux-musl ;; \
        *) echo "unsupported TARGETARCH: $TARGETARCH" >&2 ; exit 1 ;; \
    esac && \
    rustup target add "$RUST_TARGET" && \
    cargo clean && \
    cargo build --release --target "$RUST_TARGET" && \
    mkdir -p /out && \
    cp "target/$RUST_TARGET/release/libbandersnatch_vrfs_ffi.a" /out/

# ---- Stage 2: build the Go binary for the *target* platform ----
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS go-builder
COPY --from=xx / /
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY . .
COPY --from=rust-builder /out/libbandersnatch_vrfs_ffi.a \
    /app/pkg/Rust-VRF/vrf-func-ffi/target/release/

# Build-platform tooling (runs natively on the builder).
RUN apk add --no-cache clang lld tar zip
# Target-platform toolchain (musl headers + libs for $TARGETPLATFORM).
RUN xx-apk add --no-cache musl-dev gcc

ARG GP_VERSION=latest
ARG TARGET_VERSION=latest
ARG OUTPUT=new-jamneration-target
ARG BUILD_TAGS=

RUN go mod tidy

# xx-go wraps `go` with the right CC / GOOS / GOARCH / pkg-config for the
# target platform; CGO_ENABLED=1 is required for the Rust FFI static archive
# and the JIT signal-handler cgo code.
RUN CGO_ENABLED=1 \
    xx-go build \
        -tags "${BUILD_TAGS}" \
        -ldflags "-s -w -linkmode external -extldflags '-static -O3' -X main.GP_VERSION=${GP_VERSION} -X main.TARGET_VERSION=${TARGET_VERSION}" \
        -o ./build/${OUTPUT} ./cmd/fuzz && \
    xx-verify --static ./build/${OUTPUT}

RUN CGO_ENABLED=0 \
    xx-go build \
        -tags "${BUILD_TAGS}" \
        -o ./build/pvmtrace ./cmd/pvmtrace

# Package the binary into tar.gz and zip formats (uses TARGETOS/TARGETARCH so
# the artifact name reflects what was actually built).
RUN cd build && \
    tar -czf ${OUTPUT}-${TARGETOS}-${TARGETARCH}-${GP_VERSION}.tar.gz ${OUTPUT} && \
    zip ${OUTPUT}-${TARGETOS}-${TARGETARCH}-${GP_VERSION}.zip ${OUTPUT}

# ---- Stage 3: minimal runtime image (target platform) ----
# No --platform= here: FROM defaults to $TARGETPLATFORM, which is what we want.
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
ENV USE_MINI_REDIS=true

ENTRYPOINT ["/build/new-jamneration-target"]
