#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

int32_t rs_encode(const uint8_t *data_ptr,
                  uintptr_t data_len,
                  uintptr_t data_shard,
                  uintptr_t parity_shard,
                  uint8_t **out_ptr,
                  uintptr_t *out_len);

int32_t rs_decode(const uint8_t *shards_ptr,
                  const uintptr_t *indices_ptr,
                  uintptr_t shard_count,
                  uintptr_t shard_size,
                  uintptr_t data_shards,
                  uintptr_t parity_shards,
                  uint8_t **out_ptr,
                  uintptr_t *out_len);
