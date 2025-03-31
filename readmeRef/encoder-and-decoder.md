# Encoder and Decoder

This document describes how to use the encoder and decoder in this project.

In the graypaper, you will see the following symbols:

- Encode function: $\mathcal{E}$
- Decode function: $\mathcal{E}^{-1}$

## How to use the encoder and decoder

## Encoder

See more examples in the [encode_test.go](https://github.com/New-JAMneration/JAM-Protocol/blob/main/internal/types/encode_test.go) file.

### Encode a block

```go
block := Block{} // A block which you want to encode

encoder := NewEncoder() // Create a new encoder
encoded, err := encoder.Encode(&block) // Encode the block 
if err != nil {
  t.Errorf("Error encoding Block: %v", err)
}
```

### Custom encoding

If you want to encode some local structures that are not defined in the `types` package, you can concatenate the encoded bytes from encode function output.
(Encode function will execute `e.buf.Reset()` before encoding.)

For example, if you have a structure like this:

graypaper (B.10): $\mathcal{E}(s, \eta^{'}_{0}, \mathbf{H_t})$

```go
serviceId := ServiceId(1)
entropy := Entropy{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
timeSlot := TimeSlot(100)

output := []byte{}
encoder := NewEncoder()

// Encode service id
encoded, err := encoder.Encode(&serviceId)
if err != nil {
  t.Errorf("Error encoding ServiceId: %v", err)
}
output = append(output, encoded...)

// Encode entropy
encoded, err = encoder.Encode(&entropy)
if err != nil {
  t.Errorf("Error encoding Entropy: %v", err)
}
output = append(output, encoded...)

// Encode time slot 
encoded, err = encoder.Encode(&timeSlot)
if err != nil {
  t.Errorf("Error encoding TimeSlot: %v", err)
}
output = append(output, encoded...)

// output: [1 0 0 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 100 0 0 0]
```

## Encode `uint` with a specific length

You can use `encoder.EncodeUintWithLength(value, length)` to encode a `uint` with a specific length.

```go
encoder := NewEncoder()
encoded, err := encoder.EncodeUintWithLength(100, 3)
if err != nil {
  t.Errorf("Error encoding UintWithLength: %v", err)
}
```

## Deocder

See more examples in the [decode_test.go](https://github.com/New-JAMneration/JAM-Protocol/blob/main/internal/types/decode_test.go) file.

### Decode a block

```go
block := &Block{} // A block which you want to decode

decoder := NewDecoder()
// data is your byte array
err = decoder.Decode(data, block)
if err != nil {
  t.Errorf("Error decoding block: %v", err)
}
```

## How to implement encode or decode for a new structure

If you want to encode or decode a new structure, you need to implement the `Encode` and `Decode` functions for that structure.
We define the `Encode` and `Decode` functions in the `encode.go` and `decode.go` files in the `types` package.

For example, if you want to encode or decode a structure like this:

```go
type RefineLoad struct {
  GasUsed        Gas `json:"gas_used,omitempty"`        // u
  Imports        U16 `json:"imports,omitempty"`         // i
  ExtrinsicCount U16 `json:"extrinsic_count,omitempty"` // x
  ExtrinsicSize  U32 `json:"extrinsic_size,omitempty"`  // z
  Exports        U16 `json:"exports,omitempty"`         // e
}
```

### Implement the `Encode` function

Put the following code in the `encode.go` file.

```go
// RefineLoad
func (r *RefineLoad) Encode(e *Encoder) error {
  cLog(Cyan, "Encoding RefineLoad")

  // GasUsed
  if err := r.GasUsed.Encode(e); err != nil {
          return err
  }

  // Imports
  if err := r.Imports.Encode(e); err != nil {
          return err
  }

  // ExtrinsicCount
  if err := r.ExtrinsicCount.Encode(e); err != nil {
          return err
  }

  // ExtrinsicSize
  if err := r.ExtrinsicSize.Encode(e); err != nil {
          return err
  }

  // Exports
  if err := r.Exports.Encode(e); err != nil {
          return err
  }

  return nil
}
```

### Implement the `Decode` function

Put the following code in the `decode.go` file.

```go
// RefineLoad
func (r *RefineLoad) Decode(d *Decoder) error {
  cLog(Cyan, "Decoding RefineLoad")

  // GasUsed
  if err := r.GasUsed.Decode(d); err != nil {
          return err
  }

  // Imports
  if err := r.Imports.Decode(d); err != nil {
          return err
  }

  // ExtrinsicCount
  if err := r.ExtrinsicCount.Decode(d); err != nil {
          return err
  }

  // ExtrinsicSize
  if err := r.ExtrinsicSize.Decode(d); err != nil {
          return err
  }

  // Exports
  if err := r.Exports.Decode(d); err != nil {
          return err
  }

  return nil
}
```

## How to test the encoder and decoder

You can see the test cases in the [encoder_test.go](https://github.com/New-JAMneration/JAM-Protocol/blob/main/internal/types/encoder_test.go) and [decoder_test.go](https://github.com/New-JAMneration/JAM-Protocol/blob/main/internal/types/decoder_test.go)

We test our encoder and decoder with below test vectors:

- [davxy/jam-test-vectors](https://github.com/davxy/jam-test-vectors)
- [jam-duna/jamtestnet](https://github.com/jam-duna/jamtestnet)

> We use git submodule to manage the test vectors. You can clone the test vectors with the following command:
>
> ```bash
> git submodule update --init --recursive
> ```
>
> If you want to update the test vectors, you can use the following command:
>
> ```bash
> git submodule update --recursive --remote
> ```

**Note:** If you want to read json files in your test, you have to implement `UnmarshalJSON` function for your structure. We defined this function in the `unmarshal_json.go`

**Warning:** We haven't implemented the `MarshalJSON` function yet.
