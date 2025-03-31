# How to Use the Rust VRF Repository

## Prerequisites
- Git
- Golang
- Rust
- C compiler (based on your OS, e.g., GCC)

## How to Add Submodules (skip this cause we've added this repository in our JAM-Protocal)

```bash
git submodule add https://github.com/New-JAMneration/Rust-VRF.git ./pkg/Rust-VRF
```
- Note: you can use either https or ssh for repo, depending on your situaiton.

## Update Submodules

```bash
git submodule update --init --recursive
```

## Compile Rust Library

```bash
cargo build --manifest-path ./pkg/Rust-VRF/vrf-func-ffi/Cargo.toml -r
```

## Test if FFI Works Correctly

### Create the Test File
Create `signature_test.go` under the path `./internal/utilities/signature/signature_test.go`, clone the following code:
```
package ututilities

import (
"encoding/hex"
"log"
"testing"

vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

func createRingData() []byte {
ringSetHex := []string{"7b32d917d5aa771d493c47b0e096886827cd056c82dbdba19e60baa8b2c60313d3b1bdb321123449c6e89d310bc6b7f654315eb471c84778353ce08b951ad471561fdb0dcfb8bd443718b942f82fe717238cbcf8d12b8d22861c8a09a984a3c5a1b1da71cc4682e159b7da23050d8b6261eb11a3247c89b07ef56ccd002fd38b4fd11f89c2a1aaefe856bb1c5d4a1fad73f4de5e41804ca2c17ba26d6e10050c86d06ee2c70da6cf2da2a828d8a9d8ef755ad6e580e838359a10accb086ae437ad6fdeda0dde0a57c51d3226b87e3795e6474393772da46101fd597fbd456c1b3f9dc0c4f67f207974123830c2d66988fb3fb44becbbba5a64143f376edc51d9"}

var ringSetBytes []byte
for _, hexString := range ringSetHex {
bytes, err := hex.DecodeString(hexString)
if err != nil {
log.Fatalf("failed to decode hex string: %v", err)
}
ringSetBytes = append(ringSetBytes, bytes...)
}
return ringSetBytes
}

func getSK() []byte {
ringSetHex := []string{
"3d6406500d4009fdf2604546093665911e753f2213570a29521fd88bc30ede18",
}

var ringSetBytes []byte
for _, hexString := range ringSetHex {
bytes, err := hex.DecodeString(hexString)
if err != nil {
log.Fatalf("failed to decode hex string: %v", err)
}
ringSetBytes = append(ringSetBytes, bytes...)
}
return ringSetBytes
}

func TestVRF(t *testing.T) {
// Mock data
ring := createRingData()
skBytes := getSK()

ringSize := uint(8)
proverIdx := uint(3)

handler, err := vrf.NewHandler(ring, skBytes, ringSize, proverIdx)
if err != nil {
t.Fatalf("Failed to create handler: %v", err)
}
defer handler.Free()

// Test commitment
commitment, err := handler.GetCommitment()
if err != nil {
t.Fatalf("GetCommitment failed: %v", err)
}
if len(commitment) != 144 {
t.Errorf("GetCommitment returned invalid commitment length: %d", len(commitment))
}

// Test ring signature
input := []byte("73616d706c65")
aux := []byte("t")

signature, err := handler.RingSign(input, aux)
if err != nil {
t.Fatalf("RingSign failed: %v", err)
}

if len(signature) != 784 {
t.Fatalf("RingSign returned invalid signature length: %d", len(signature))
}

// Test verification
output, err := handler.RingVerify(input, aux, signature)
if err != nil {
t.Fatalf("RingVerify failed: %v", err)
}
if len(output) == 0 {
t.Error("RingVerify returned empty output")
}

// Test VRF output
vrfOutput, err := handler.VRFOutput(input)
if err != nil {
t.Fatalf("VRFOutput failed: %v", err)
}
if len(vrfOutput) == 0 {
t.Error("VRFOutput returned empty output")
}
}
```

### Run Go Test: 

```bash
go test -v ./internal/utilities/signature
```

1. Ensure you have run `cargo build --manifest-path ./pkg/Rust-VRF/vrf-func-ffi/Cargo.toml -r`. This will generate the files under `./pkg/Rust-VRF/vrf-func-ffi/target/release/bandersnatch_vrfs_ffi*` in your local.
   - Linux(WSL) or Mac users: you can run the `go test` command directly.
   - Windows users: Before running `go test`, copy `bandersnatch_vrfs_ffi.dll` into the same directory as `signature_test.go`:
     ```markdown
       - internal
       - utilities
         - signature
           - signature_test.go
           - bandersnatch_vrfs_ffi.dll
     ```
2. If you see the following output, it indicates success:
   ```bash
      === RUN   TestVRF
      Reading SRS file from: C:\work\jam_test\JAM-Protocol\pkg\Rust-VRF\vrf-func-ffi/data/zcash-srs-2-11-uncompressed.bin
      Ring Size: 8
      Successfully deserialized PcsParams
      Successfully created RingContext
      Successfully set RingContext
      Initializing ring...
      Initializing ring: size = 8
      Ring signature verified
      --- PASS: TestVRF (0.27s)
      PASS
      ok      github.com/New-JAMneration/JAM-Protocol/internal/utilities/signature  0.438s

   ```

### If you encounter any questions or bugs not mentioned above, feel free to ask in [Discord](https://discord.com/channels/1300472335232536710/1324173965593149503)
