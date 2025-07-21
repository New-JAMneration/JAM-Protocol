package ce

// AssignShardIndex computes the erasure coded shard index assigned to a validator.
//
// Formula: i = (cR + v) mod V
//   c: core index
//   R: recovery threshold
//   v: validator index
//   V: total number of validators
//
// Returns the shard index assigned to the validator.
func AssignShardIndex(coreIndex, recoveryThreshold, validatorIndex, totalValidators int) int {
	return (coreIndex*recoveryThreshold + validatorIndex) % totalValidators
}
