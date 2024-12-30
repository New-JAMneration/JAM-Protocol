# How to use the Rust VRF repository

## How to add submodules

```bash
git submodule add https://github.com/New-JAMneration/Rust-VRF.git ./pkg/Rust-VRF
# Note: you can use http or ssh for repo, and it depends on situaiton.
```

## Update submodules

```bash
git submodule update --init --recursive
```

## Compile rust library

```bash
cargo build --manifest-path ./pkg/Rust-VRF/vrf-func-ffi/Cargo.toml -r
```


