# JAM Protocol Implementation

<img width="1672" height="383" alt="New-JAMneration JAM Protocol banner" src="https://github.com/user-attachments/assets/181dd824-3310-4703-9220-c9691935b7b5" />

[![Go Format Check](https://github.com/New-JAMneration/JAM-Protocol/actions/workflows/go-format.yml/badge.svg)](https://github.com/New-JAMneration/JAM-Protocol/actions/workflows/go-format.yml)
[![Release](https://github.com/New-JAMneration/JAM-Protocol/actions/workflows/release.yml/badge.svg)](https://github.com/New-JAMneration/JAM-Protocol/actions/workflows/release.yml)
[![M1 Conformance](https://img.shields.io/badge/M1%20Conformance-passed-success)](https://fuzz.jamtoaster.network/attestation/0x4ec2d3b8498bf7938515820be78f187eefe46b7ef3c7e5bd6355c6e70bd4bbfd)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](./LICENSE)

A **Go implementation of the Polkadot [JAM Protocol](https://jam.web3.foundation/)**, built by the **New-JAMneration** team and aligned with **Gray Paper v0.7.2**.

## Project Status

- ✅ **Milestone 1 (M1) — Passed.** Validated for block-import conformance against the cross-team [JAM Conformance](https://github.com/davxy/jam-conformance) fuzzer. See our [on-chain attestation](https://fuzz.jamtoaster.network/attestation/0x4ec2d3b8498bf7938515820be78f187eefe46b7ef3c7e5bd6355c6e70bd4bbfd).

## Table of Contents

- [Documentation](#documentation)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Clone the repository](#clone-the-repository)
  - [Install dependencies](#install-dependencies)
  - [Run](#run-the-jam-protocol)
  - [Build](#build-the-jam-protocol)
- [Testing](#testing)
- [Conformance & Fuzzing](#conformance--fuzzing)
- [Operations](#operations)
- [Contributing](#contributing)
- [License](#license)

## Documentation

Our development documentation is maintained on HackMD for real-time collaboration and easy updates.

This is our mindmap link: [JAM Mindmap](https://new-jamneration.github.io/JAM-mindmap/)

If you want to update the mindmap, you can go to the [JAM-mindmap repository](https://github.com/New-JAMneration/JAM-mindmap).

### Access Documentation

- **Full documentation index**: [READMERef/INDEX.md](./READMERef/INDEX.md)
- Main documentation: [HackMD Development Guide](https://hackmd.io/8ckvpUULSp-HqThsxXE3jg)
- Development documentation: [Github Document](https://github.com/New-JAMneration/JAM-Protocol/blob/main/READMERef/DEVELOPMENT_DOC.md)
- Requires team member access - please contact project maintainers if you need access

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) **1.25.5+**
- [Rust toolchain](https://www.rust-lang.org/tools/install) — required to build the VRF submodule (see the [Rust VRF Compile Guide](./READMERef/RUST_VRF_COMPILE_GUIDE.md))
- `make` and a POSIX shell
- (Optional) [Docker](https://www.docker.com/) — for release builds and fuzz-target runs

### Clone the repository

This project uses git submodules (Rust-VRF and test-data vectors), so clone with submodules:

```bash
git clone --recurse-submodules https://github.com/New-JAMneration/JAM-Protocol.git
```

If you already cloned without `--recurse-submodules`:

```bash
git submodule update --init --recursive
```

### Install dependencies

```bash
go mod tidy
```

### Run the JAM Protocol

```bash
make run
```

### Build the JAM Protocol

```bash
make build
```

## Testing

### Test jam-test-vectors

Run single test:

```bash
make test-jam-test-vectors mode=safrole size=full
```

```bash
make test-jam-test-vectors-trace mode=safrole
```

Run all:

```bash
make test-jam-test-vectors
```

```bash
make test-jam-test-vectors-trace
```

## Conformance & Fuzzing

We continuously validate the node against the JAM Conformance fuzz protocol. See the [Fuzz Validation guide](./READMERef/VALIDATE_FUZZ.md) for vectors, trace, socket, and CI steps.

## Operations

- **Release and Publish**: see the [release and publish guide](./READMERef/RELEASE_AND_PUBLISH.md).
- **Rust Submodule**: for compiling and using the Rust library, see the [Rust VRF Compile Guide](./READMERef/RUST_VRF_COMPILE_GUIDE.md).
- **Encoder/Decoder**: for details of the encoder and decoder, see the [Encoder & Decoder guide](./READMERef/ENCODER_AND_DECODER.md).

## Contributing

- **Coding Style**: our codebase follows the [Google Go Style Guide](https://google.github.io/styleguide/go/) for consistent and maintainable code.
- **Code Formatting**: we use `gofmt` to maintain consistent code formatting. See the [commands here](./READMERef/CODE_FORMATTING.md).
- **Commit Message**: please stick to the [Semantic Commit Messages](./READMERef/SEMANTIC_COMMIT_MESSAGES.md) when submitting a commit.
- **Pull Request**: before creating a pull request, please **rebase** (*instead of merging*) your branch onto the target branch. Also, follow these [instructions](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue) to link your PR to the assigned ticket's issue.

## License

Licensed under the [Apache License 2.0](./LICENSE).

![](https://github.com/user-attachments/assets/6514346b-e691-45da-bd2d-bec332d89d88)
