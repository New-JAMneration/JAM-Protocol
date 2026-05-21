FROM golang:1.25-bookworm

# For testing the JIT compiler/ASM code

# build: docker build --platform=linux/amd64 -t go-jit-test -f Dockerfile .
# run: docker run --rm -it  -v "$(pwd)":/app \ go-jit-test 

RUN apt-get update && apt-get install -y \
    binutils \
    gdb \
    vim \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

WORKDIR /app
