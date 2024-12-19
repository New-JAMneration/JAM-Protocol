# 開發文件

## 環境

go 1.23.3

## 可用的第三方庫

加密原語

- Hash :
  - [Blake2b256 and keccak256 from gosammer](https://github.com/ChainSafe/gossamer/blob/development/lib/common/hasher.go)
  - [golang crypto blake2b 256-bit](https://pkg.go.dev/golang.org/x/crypto/blake2b#New256)
  - [golang crypto sha3 keccak 256-bit](https://pkg.go.dev/golang.org/x/crypto/sha3#NewLegacyKeccak256) (using NewLegacyKeccak256 function to generate LegacyKeccak256 object)
  - 測試參照 [波卡官方](https://polkadot.js.org/docs/util-crypto/examples/hash-data/)：
    - [Polkadot.js](https://polkadot.js.org/) -> Devolop -> Utility -> hash 可以測試 Blake2b 的 hash

- 糾刪碼（Erasure Coding）：
  - 可以使用 [klauspost/reedsolomon](https://github.com/klauspost/reedsolomon) 庫，該庫提供高效的 Reed-Solomon 編碼實現，適用於資料恢復和錯誤校正。
  
- Bandersnatch：
  - 目前 Go 語言中尚無 Bandersnatch 曲線的主流實現。 (https://github.com/Consensys/gnark-crypto/tree/master/ecc/bls12-381)

- Ed25519：
  - Go 標準庫的 `crypto/ed25519` 包提供了 Ed25519 簽名演算法的實現，可用於生成密鑰對、簽名和驗證。

- RingVRF
  - [VRF and RingVRF Spec](https://github.com/davxy/bandersnatch-vrfs-spec)
  - [Davxy (暫)指定測試用](https://github.com/davxy/ark-ec-vrfs)
  - [RingVRF proof by w3f](https://github.com/w3f/ring-proof)


編解碼器

- **SCALE（Simple Concatenated Aggregate Little-Endian）：**
  - SCALE 是由 Parity Technologies 為 Substrate 區塊鏈框架設計的編解碼格式。 Go 語言中有 [ChainSafe/gossamer](https://github.com/ChainSafe/gossamer) 等專案實現了 SCALE 編解碼器，可用於與 Substrate 節點進行互操作。

網路協議

- **QUIC（Quick UDP Internet Connections）：**
  - [quic-go](https://github.com/quic-go/quic-go) 是用純 Go 語言實現的 QUIC 協議庫，支持 HTTP/3，並實現了多個相關的 RFC 標準。 


## 路徑

### M0

- [參考 davxy 的 M1 conference](https://github.com/w3f/jamtestvectors/issues/21)


### M1 只要證明你讀進來的 block stream 轉換成 state 是正確的就好

- [JAMNetworkProtocol](https://github.com/zdave-parity/jam-np/blob/main/simple.md)

### M2 之後會有網路測試環境 (測試沙盒), 不只要告訴 state 還要有什麼東西他發送出去

