use reed_solomon_simd::{ReedSolomonEncoder, ReedSolomonDecoder};
use std::slice;
use std::alloc::{alloc, Layout};
use std::ptr;
use std::collections::HashMap;

#[no_mangle]
pub extern "C" fn rs_encode(
    data_ptr: *const u8,
    data_len: usize,
    data_shard: usize,
    parity_shard: usize,
    out_ptr: *mut *mut u8,
    out_len: *mut usize,
) -> i32 {
    let w_e = data_shard * 2;

    // Load input data
    let mut data = unsafe { slice::from_raw_parts(data_ptr, data_len).to_vec() };

    // Padding to multiple of w_e
    if data.len() % w_e != 0 {
        let pad_len = w_e - (data.len() % w_e);
        data.extend(std::iter::repeat(0).take(pad_len));
    }

    let shard_size = data.len() / data_shard;
    let total_shards = data_shard + parity_shard;

    let mut chunks: Vec<Vec<u16>> = Vec::with_capacity(shard_size / 2);
    for i in 0..shard_size / 2 {
        let mut encoder = match ReedSolomonEncoder::new(data_shard, parity_shard, 2) {
            Ok(enc) => enc,
            Err(_) => return 1,
        };

        // Add data shards
        for j in 0..data_shard {
            let shard = u16::from_le_bytes([
                data[j * shard_size + i * 2],
                data[j * shard_size + i * 2 + 1],
            ]);
            if encoder.add_original_shard(&shard.to_le_bytes()).is_err() {
                return 2;
            }
        }

        // Encode parity shards
        let encoded = match encoder.encode() {
            Ok(e) => e,
            Err(_) => return 3,
        };

        let mut chunk = Vec::with_capacity(total_shards);
        for j in 0..data_shard {
            let shard = u16::from_le_bytes([
                data[j * shard_size + i * 2],
                data[j * shard_size + i * 2 + 1],
            ]);
            chunk.push(shard);
        }
        for shard in encoded.recovery_iter() {
            chunk.push(u16::from_le_bytes([shard[0], shard[1]]));
        }
        chunks.push(chunk);
    }

    // Flatten in shard-major order
    let total_len = total_shards * shard_size;
    unsafe {
        let layout = Layout::from_size_align(total_len, 1).unwrap();
        let ptr = alloc(layout);
        if ptr.is_null() {
            return 4;
        }
        let mut offset = 0;
        for shard_idx in 0..total_shards {
            for chunk in &chunks {
                let val = chunk[shard_idx];
                let le_bytes = val.to_le_bytes();
                ptr::copy_nonoverlapping(le_bytes.as_ptr(), ptr.add(offset), 2);
                offset += 2;
            }
        }
        *out_ptr = ptr;
        *out_len = total_len;
    }

    0
}

#[no_mangle]
pub extern "C" fn rs_decode(
    shards_ptr: *const u8,
    indices_ptr: *const usize,
    shard_count: usize,
    shard_size: usize,
    data_shards: usize,
    parity_shards: usize,
    out_ptr: *mut *mut u8,
    out_len: *mut usize,
) -> i32 {
    if shards_ptr.is_null() || indices_ptr.is_null() || out_ptr.is_null() || out_len.is_null() {
        return 99;
    }

    if shard_size % 2 != 0 {
        return 97;
    }

    let chunk_count = shard_size / 2;
    let flat = unsafe { slice::from_raw_parts(shards_ptr, shard_count * shard_size) };
    let indices = unsafe { slice::from_raw_parts(indices_ptr, shard_count) };

    // Output buffers per data shard
    let mut output_chunks: Vec<Vec<u16>> = vec![Vec::with_capacity(chunk_count); data_shards];

    for chunk_idx in 0..chunk_count {
        let mut decoder = match ReedSolomonDecoder::new(data_shards, parity_shards, 2) {
            Ok(d) => d,
            Err(_) => return 1,
        };

        let mut provided = vec![None; data_shards];

        for (i, &shard_idx) in indices.iter().enumerate() {
            let offset = i * shard_size + chunk_idx * 2;
            let pair = u16::from_le_bytes([
                flat[offset],
                flat[offset + 1],
            ]);

            let bytes = pair.to_le_bytes();

            if shard_idx < data_shards {
                provided[shard_idx] = Some(pair);
                if decoder.add_original_shard(shard_idx, &bytes).is_err() {
                    return 2;
                }
            } else {
                let recovery_idx = shard_idx - data_shards;
                if decoder.add_recovery_shard(recovery_idx, &bytes).is_err() {
                    return 2;
                }
            }
        }

        let recovered = match decoder.decode() {
            Ok(r) => r,
            Err(_) => return 3,
        };

        let restored: HashMap<usize, &[u8]> =
            recovered.restored_original_iter().collect();

        for i in 0..data_shards {
            if let Some(existing) = provided[i] {
                output_chunks[i].push(existing);
            } else if let Some(recovered) = restored.get(&i) {
                if recovered.len() != 2 {
                    return 6;
                }
                output_chunks[i].push(u16::from_le_bytes([recovered[0], recovered[1]]));
            } else {
                return 5;
            }
        }
    }

    // Concatenate all data shards
    let mut output = Vec::with_capacity(data_shards * shard_size);
    for shard_data in output_chunks {
        for word in shard_data {
            output.extend_from_slice(&word.to_le_bytes());
        }
    }

    unsafe {
        let total_len = output.len();
        let layout = Layout::from_size_align(total_len, 1).unwrap();
        let ptr_mem = alloc(layout);
        if ptr_mem.is_null() {
            return 4;
        }
        ptr::copy_nonoverlapping(output.as_ptr(), ptr_mem, total_len);
        *out_ptr = ptr_mem;
        *out_len = total_len;
    }

    0
}
